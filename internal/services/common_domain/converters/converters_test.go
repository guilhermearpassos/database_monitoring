package converters

import (
	"testing"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
)

func TestQueryMetric_LastElapsedTimeRoundTrip(t *testing.T) {
	orig := &common_domain.QueryMetric{
		QueryHash:       "abc",
		Text:            "SELECT 1",
		Database:        common_domain.DataBaseMetadata{DatabaseID: "5", DatabaseName: "OrdersDB"},
		LastElapsedTime: 12500 * time.Microsecond, // 12.5 ms
	}

	// Simulate 3 round trips: Agent -> Ingester -> Postgres -> Querier -> Plugin
	cur := orig
	for i := 0; i < 3; i++ {
		pb, err := QueryMetricToProto(cur)
		if err != nil {
			t.Fatalf("round %d QueryMetricToProto error: %v", i, err)
		}
		if pb.LastElapsedTimeMicros != 12500 {
			t.Fatalf("round %d LastElapsedTimeMicros = %d, want 12500", i, pb.LastElapsedTimeMicros)
		}
		cur, err = QueryMetricToDomain(pb)
		if err != nil {
			t.Fatalf("round %d QueryMetricToDomain error: %v", i, err)
		}
		if cur.LastElapsedTime != 12500*time.Microsecond {
			t.Fatalf("round %d LastElapsedTime = %v, want %v", i, cur.LastElapsedTime, 12500*time.Microsecond)
		}
	}
}

func TestConverters_NilSafety(t *testing.T) {
	if got := DatabaseSnapshotToProto(nil); got != nil {
		t.Fatalf("expected nil from DatabaseSnapshotToProto(nil)")
	}
	if got := SampleToProto(nil); got != nil {
		t.Fatalf("expected nil from SampleToProto(nil)")
	}
	if got, err := ExecutionPlanToProto(nil); err != nil || got != nil {
		t.Fatalf("expected nil, nil from ExecutionPlanToProto(nil), got %v, %v", got, err)
	}
	if got := SampleToDomain(nil); got != nil {
		t.Fatalf("expected nil from SampleToDomain(nil)")
	}
	if got := SampleToDomain(&dbmv1.QuerySample{Id: "s1"}); got == nil || got.Id != "s1" {
		t.Fatalf("expected partial QuerySample to convert safely without panic, got %+v", got)
	}
	if got, err := QueryMetricToDomain(&dbmv1.QueryMetric{QueryHash: "h1"}); err != nil || got == nil {
		t.Fatalf("expected partial QueryMetric to convert safely, got %+v, %v", got, err)
	}
	if got, err := ExecutionPlanToDomain(&dbmv1.ExecutionPlan{PlanHandle: "p1"}); err != nil || got == nil {
		t.Fatalf("expected partial ExecutionPlan to convert safely, got %+v, %v", got, err)
	}
	_ = DatabaseSnapshotToDomain(&dbmv1.DBSnapshot{Id: "snap1"})
}
