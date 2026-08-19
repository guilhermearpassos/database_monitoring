package state

import (
	"sort"
	"sync"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
)

// MemoryStore is an in-memory implementation of domain.SessionStore for MVP/testing.
// It deduplicates chunks by (snapshotID, seq) and tracks session completion.
// Not suitable for multi-process deployments.
type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]*sessionState
	ttl      time.Duration
}

type sessionState struct {
	header    domain.SnapshotHeader
	received  map[uint32]struct{}
	finalized bool
	completed bool
}

func NewMemoryStore() *MemoryStore { // default, ttl is driven per-header UpsertHeader arg
	return &MemoryStore{sessions: make(map[string]*sessionState)}
}

func (m *MemoryStore) UpsertHeader(h domain.SnapshotHeader) (*domain.SnapshotSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st, ok := m.sessions[h.SnapshotID]
	if !ok {
		st = &sessionState{
			header:   h,
			received: make(map[uint32]struct{}),
		}
		m.sessions[h.SnapshotID] = st
	}
	st.header = h // refresh header fields idempotently
	return m.asDomainSession(st), nil
}

func (m *MemoryStore) SaveChunk(snapshotID string, seq uint32, samples []*dbmv1.QuerySample) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st, ok := m.sessions[snapshotID]
	if !ok {
		return false, nil // unknown; caller will decide status later
	}
	_, had := st.received[seq]
	if had {
		return true, nil
	}
	st.received[seq] = struct{}{}
	m.maybeCompleteLocked(st)
	return false, nil
}

func (m *MemoryStore) Finalize(snapshotID string, _ uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	st, ok := m.sessions[snapshotID]
	if !ok {
		return nil
	}
	st.finalized = true
	m.maybeCompleteLocked(st)
	return nil
}

func (m *MemoryStore) GetMissing(snapshotID string) (domain.UploadStatus, []uint32, error) {
	m.mu.RLock()
	st, ok := m.sessions[snapshotID]
	if !ok {
		m.mu.RUnlock()
		return domain.UploadStatusUnknown, nil, nil
	}
	// copy needed fields while holding RLock, then release for sort/work
	header := st.header
	received := make([]uint32, 0, len(st.received))
	for k := range st.received {
		received = append(received, k)
	}
	finalized := st.finalized
	completed := st.completed
	m.mu.RUnlock()

	if completed {
		return domain.UploadStatusComplete, nil, nil
	}
	// Receiving state
	missing := make([]uint32, 0)
	if header.ExpectedChunks > 0 {
		// compute missing in 1..ExpectedChunks
		present := make(map[uint32]struct{}, len(received))
		for _, s := range received {
			present[s] = struct{}{}
		}
		for i := uint32(1); i <= header.ExpectedChunks; i++ {
			if _, ok := present[i]; !ok {
				missing = append(missing, i)
			}
		}
	} else if finalized {
		// No expected chunks; if finalized and we don't track expected, assume complete when
		// there are no gaps between 1..maxSeq; otherwise report gaps as missing.
		if len(received) > 0 {
			sort.Slice(received, func(i, j int) bool { return received[i] < received[j] })
			max := received[len(received)-1]
			present := make(map[uint32]struct{}, len(received))
			for _, s := range received {
				present[s] = struct{}{}
			}
			for i := uint32(1); i <= max; i++ {
				if _, ok := present[i]; !ok {
					missing = append(missing, i)
				}
			}
		}
	}
	return domain.UploadStatusReceiving, missing, nil
}

func (m *MemoryStore) MarkComplete(snapshotID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, snapshotID)
	return nil
}

func (m *MemoryStore) maybeCompleteLocked(st *sessionState) {
	if st.finalized {
		if st.header.ExpectedChunks > 0 {
			st.completed = uint32(len(st.received)) == st.header.ExpectedChunks
		} else {
			// Without an expected count, consider completed if there are no gaps
			if len(st.received) == 0 {
				return
			}
			max := uint32(0)
			for s := range st.received {
				if s > max {
					max = s
				}
			}
			for i := uint32(1); i <= max; i++ {
				if _, ok := st.received[i]; !ok {
					return
				}
			}
			st.completed = true
		}
	}
}

func (m *MemoryStore) asDomainSession(st *sessionState) *domain.SnapshotSession {
	return &domain.SnapshotSession{
		Header:         st.header,
		Received:       cloneSet(st.received),
		ExpectedChunks: st.header.ExpectedChunks,
		Finalized:      st.finalized,
		Completed:      st.completed,
	}
}

func cloneSet(mset map[uint32]struct{}) map[uint32]struct{} {
	cp := make(map[uint32]struct{}, len(mset))
	for k := range mset {
		cp[k] = struct{}{}
	}
	return cp
}
