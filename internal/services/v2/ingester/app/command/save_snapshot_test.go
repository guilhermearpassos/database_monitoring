package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/adapters/state"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
)

type sliceIter struct {
	msgs []*domain.SnapUploadMessage
	idx  int
}

func (s *sliceIter) Next(ctx context.Context) (*domain.SnapUploadMessage, error) {
	if s.idx >= len(s.msgs) {
		return nil, nil
	}
	m := s.msgs[s.idx]
	s.idx++
	return m, nil
}

type flakyRepo struct {
	failFirstSave bool
	savedSeqs     []uint32
}

func (f *flakyRepo) EnsureSnapshot(ctx context.Context, header domain.SnapshotHeader) error {
	return nil
}
func (f *flakyRepo) SaveSamples(ctx context.Context, snapshotID string, seq uint32, samples []*dbmv1.QuerySample) error {
	if f.failFirstSave {
		f.failFirstSave = false
		return errors.New("transient db timeout")
	}
	f.savedSeqs = append(f.savedSeqs, seq)
	return nil
}
func (f *flakyRepo) FinalizeSnapshot(ctx context.Context, snapshotID string) error { return nil }
func (f *flakyRepo) SaveExecutionPlans(ctx context.Context, executionPlan []*common_domain.ExecutionPlan) error {
	return nil
}
func (f *flakyRepo) GetMissingPlans(ctx context.Context, server common_domain.ServerMeta, start time.Time, end time.Time) ([]string, error) {
	return nil, nil
}
func (f *flakyRepo) PurgeSnapshots(ctx context.Context, start time.Time, end time.Time, size int) error {
	return nil
}
func (f *flakyRepo) PurgeAllSnapshots(ctx context.Context) error                       { return nil }
func (f *flakyRepo) PurgeUnboundedPlans(ctx context.Context, batchSize int) error      { return nil }
func (f *flakyRepo) GetUnanalizedPlans(ctx context.Context, size int) ([]*common_domain.ExecutionPlan, error) {
	return nil, nil
}
func (f *flakyRepo) SetPlanAnalisysBatch(ctx context.Context, results []domain.PlanAnalisysBatchResult) error {
	return nil
}

func TestSaveSnapshot_RepoFailureDoesNotMarkChunkReceived(t *testing.T) {
	store := state.NewMemoryStore()
	repo := &flakyRepo{failFirstSave: true}
	h := NewSaveSnapshotStreamHandler(store, repo)

	msgs := []*domain.SnapUploadMessage{
		{SnapshotID: "snap-1", Header: &domain.SnapshotHeader{SnapshotID: "snap-1", ExpectedChunks: 1, ExpectedTotalSamples: 1}},
		{SnapshotID: "snap-1", Chunk: &domain.SampleChunk{ChunkSeq: 1, Samples: []*dbmv1.QuerySample{{Id: "q1"}}}},
		{SnapshotID: "snap-1", Finalize: &domain.Finalize{TotalSamples: 1}},
	}

	if _, err := h.Handle(context.Background(), &sliceIter{msgs: msgs}); err == nil {
		t.Fatalf("expected first attempt to fail on transient repo error")
	}

	// Chunk 1 must still be reported as missing!
	st, missing, err := store.GetMissing("snap-1")
	if err != nil {
		t.Fatalf("GetMissing: %v", err)
	}
	if st != domain.UploadStatusReceiving || len(missing) != 1 || missing[0] != 1 {
		t.Fatalf("expected chunk 1 to remain missing after repo failure, got st=%v missing=%v", st, missing)
	}

	// Retry succeeds and persists chunk 1.
	if _, err := h.Handle(context.Background(), &sliceIter{msgs: msgs}); err != nil {
		t.Fatalf("expected retry to succeed, got: %v", err)
	}
	if len(repo.savedSeqs) != 1 || repo.savedSeqs[0] != 1 {
		t.Fatalf("expected chunk 1 to be saved on retry, got %v", repo.savedSeqs)
	}
}
