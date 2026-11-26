package collector

import (
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

func TestPlanHandleValidation(t *testing.T) {
	validBytes := make([]byte, 44)
	validBytes[0] = 0x06
	validBytes[10] = 0xAB
	validB64 := base64.StdEncoding.EncodeToString(validBytes)

	tests := []struct {
		name string
		raw  []byte
		want string
	}{
		{name: "nil", raw: nil, want: ""},
		{name: "empty", raw: []byte{}, want: ""},
		{name: "single_zero_byte_0x0", raw: []byte{0x00}, want: ""},
		{name: "all_zero_44_bytes", raw: make([]byte, 44), want: ""},
		{name: "too_long", raw: append(make([]byte, 64), 0x01), want: ""},
		{name: "valid_handle", raw: validBytes, want: validB64},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := encodeValidPlanHandle(tc.raw); got != tc.want {
				t.Fatalf("encodeValidPlanHandle() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestIsInvalidPlanHandleErr(t *testing.T) {
	if !isInvalidPlanHandleErr(errors.New("mssql: The handle that was passed to dm_exec_query_plan was invalid.")) {
		t.Fatalf("expected invalid dm_exec_query_plan handle error to be recognized")
	}
	if !isInvalidPlanHandleErr(errors.New("mssql: The handle that was passed to dm_exec_text_query_plan was invalid.")) {
		t.Fatalf("expected invalid dm_exec_text_query_plan handle error to be recognized")
	}
	if isInvalidPlanHandleErr(errors.New("connection reset by peer")) {
		t.Fatalf("did not expect unrelated error to match")
	}
}

func TestEncodePlanHandlesJSON(t *testing.T) {
	got, err := encodePlanHandlesJSON([][]byte{{0x06, 0x00, 0xAB}, {0x05, 0xFF}})
	if err != nil {
		t.Fatalf("encodePlanHandlesJSON error: %v", err)
	}
	want := `["0x0600ab","0x05ff"]`
	if got != want {
		t.Fatalf("encodePlanHandlesJSON = %s, want %s", got, want)
	}
}

func TestComputeQueryMetricDelta(t *testing.T) {
	// First observation with large count (>30) is skipped to avoid startup spike
	if _, emit := computeQueryMetricDelta(nil, map[string]int64{"executionCount": 500, "totalWorkerTime": 10000}); emit {
		t.Fatalf("expected first observation >30 executions to be skipped as baseline only")
	}

	// First observation with small count (<=30) emits immediately
	delta, emit := computeQueryMetricDelta(nil, map[string]int64{"executionCount": 5, "totalWorkerTime": 1000})
	if !emit || delta["executionCount"] != 5 || delta["totalWorkerTime"] != 1000 {
		t.Fatalf("expected small initial count to emit, got emit=%v delta=%v", emit, delta)
	}

	// Normal delta
	prev := map[string]int64{"executionCount": 5, "totalWorkerTime": 1000, "totalSpills": 1}
	curr := map[string]int64{"executionCount": 8, "totalWorkerTime": 2500, "totalSpills": 3}
	delta, emit = computeQueryMetricDelta(prev, curr)
	if !emit || delta["executionCount"] != 3 || delta["totalWorkerTime"] != 1500 || delta["totalSpills"] != 2 {
		t.Fatalf("unexpected delta: emit=%v delta=%v", emit, delta)
	}

	// Zero delta (idle query) is skipped
	if _, emit := computeQueryMetricDelta(curr, curr); emit {
		t.Fatalf("expected zero-delta query to be skipped")
	}

	// Plan cache reset (curr < prev) with small count emits fresh counts without negatives
	resetCurr := map[string]int64{"executionCount": 2, "totalWorkerTime": 400}
	delta, emit = computeQueryMetricDelta(curr, resetCurr)
	if !emit || delta["executionCount"] != 2 || delta["totalWorkerTime"] != 400 {
		t.Fatalf("expected plan cache reset to emit positive fresh counters, got emit=%v delta=%v", emit, delta)
	}
}

func TestPruneStaleQueryCounters(t *testing.T) {
	ss := NewSqlServerSnapshotter(nil, common_domain.ServerMeta{Host: "localhost", Type: "mssql"})
	now := time.Now().UTC()
	ss.lastQueryCounters["old"] = queryCounterState{counters: map[string]int64{"executionCount": 1}, lastSeen: now.Add(-20 * time.Minute)}
	ss.lastQueryCounters["fresh"] = queryCounterState{counters: map[string]int64{"executionCount": 2}, lastSeen: now.Add(-5 * time.Minute)}

	ss.pruneStaleQueryCounters(now, 15*time.Minute)

	if _, exists := ss.lastQueryCounters["old"]; exists {
		t.Fatalf("expected stale query counter entry to be pruned")
	}
	if _, exists := ss.lastQueryCounters["fresh"]; !exists {
		t.Fatalf("expected fresh query counter entry to remain")
	}
}
