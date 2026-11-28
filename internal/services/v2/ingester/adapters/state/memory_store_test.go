package state

import (
	"testing"

	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
)

func TestMemoryStore_EmptySnapshotCompletes(t *testing.T) {
	store := NewMemoryStore()
	_, err := store.UpsertHeader(domain.SnapshotHeader{
		SnapshotID:           "empty-snap",
		ExpectedChunks:       0,
		ExpectedTotalSamples: 0,
	})
	if err != nil {
		t.Fatalf("UpsertHeader: %v", err)
	}
	if err := store.Finalize("empty-snap", 0); err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	st, missing, err := store.GetMissing("empty-snap")
	if err != nil {
		t.Fatalf("GetMissing: %v", err)
	}
	if st != domain.UploadStatusComplete || len(missing) != 0 {
		t.Fatalf("got status=%v missing=%v, want COMPLETE with no missing chunks", st, missing)
	}
}
