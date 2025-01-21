// sync_service_test.go
package collector

import (
	"context"
	collectorv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1/collector"
	io_prometheus_client "github.com/prometheus/client_model/go"
	timestamppb1 "google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type mockSyncStateServer struct {
	grpc.ServerStream
	ctx      context.Context
	received []*collectorv1.StateUpdate
	toSend   []*collectorv1.StateUpdate
	sendIdx  int
}

func (m *mockSyncStateServer) Context() context.Context {
	return m.ctx
}

func (m *mockSyncStateServer) Send(update *collectorv1.StateUpdate) error {
	m.received = append(m.received, update)
	return nil
}

func (m *mockSyncStateServer) Recv() (*collectorv1.StateUpdate, error) {
	if m.sendIdx >= len(m.toSend) {
		return nil, nil
	}
	update := m.toSend[m.sendIdx]
	m.sendIdx++
	return update, nil
}

func TestSyncService_HandleStateUpdate(t *testing.T) {
	reg := prometheus.NewRegistry()
	service := NewSyncService("test-collector", reg, 5*time.Second, 5*time.Second)

	tests := []struct {
		name    string
		update  *collectorv1.StateUpdate
		wantErr bool
	}{
		{
			name: "valid update",
			update: &collectorv1.StateUpdate{
				CollectorId: "collector-1",
				Metrics: []*collectorv1.DatabaseMetrics{
					{
						DatabaseId:    "db-1",
						Timestamp:     timestamppb1.New(time.Now()),
						SystemMetrics: &collectorv1.SystemMetrics{},
					},
				},
				SequenceNumber: 1,
			},
			wantErr: false,
		},
		{
			name: "multiple databases",
			update: &collectorv1.StateUpdate{
				CollectorId: "collector-1",
				Metrics: []*collectorv1.DatabaseMetrics{
					{DatabaseId: "db-1",

						SystemMetrics: &collectorv1.SystemMetrics{}},
					{DatabaseId: "db-2",

						SystemMetrics: &collectorv1.SystemMetrics{}},
				},
				SequenceNumber: 2,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.handleStateUpdate(tt.update)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			metrics := service.GetMetrics()
			for _, m := range tt.update.Metrics {
				assert.Contains(t, metrics, m.DatabaseId)
			}
		})
	}
}

func TestSyncService_GC(t *testing.T) {
	reg := prometheus.NewRegistry()
	service := NewSyncService("test-collector", reg, 5*time.Second, 5*time.Second)

	// Add metrics with expired TTL
	service.mu.Lock()
	service.metrics["expired-db"] = &collectorv1.DatabaseMetrics{DatabaseId: "expired-db",
		SystemMetrics: &collectorv1.SystemMetrics{}}
	service.metricsTTL["expired-db"] = time.Now().Add(-5 * time.Second * 2)
	service.mu.Unlock()

	// Run GC
	service.gc()

	// Verify expired metrics are removed
	metrics := service.GetMetrics()
	assert.NotContains(t, metrics, "expired-db")
}

func TestSyncService_GetPeerState(t *testing.T) {
	reg := prometheus.NewRegistry()
	service := NewSyncService("test-collector", reg, 5*time.Second, 5*time.Second)

	// Add test data
	now := time.Now()
	oldTimestamp := now.Add(-time.Hour)
	newTimestamp := now

	update := &collectorv1.StateUpdate{
		CollectorId: "collector-1",
		Metrics: []*collectorv1.DatabaseMetrics{
			{DatabaseId: "db-1", Timestamp: timestamppb1.New(oldTimestamp),
				SystemMetrics: &collectorv1.SystemMetrics{}},
			{DatabaseId: "db-2", Timestamp: timestamppb1.New(newTimestamp),
				SystemMetrics: &collectorv1.SystemMetrics{}},
		},
	}

	require.NoError(t, service.handleStateUpdate(update))

	tests := []struct {
		name           string
		sinceTimestamp *timestamppb1.Timestamp
		wantDBs        []string
	}{
		{
			name:           "get all metrics",
			sinceTimestamp: timestamppb1.New(oldTimestamp.Add(-3600 * time.Second)),
			wantDBs:        []string{"db-1", "db-2"},
		},
		{
			name:           "get only new metrics",
			sinceTimestamp: timestamppb1.New(oldTimestamp.Add(1 * time.Second)),
			wantDBs:        []string{"db-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &collectorv1.PeerStateRequest{
				CollectorId:    "collector-2",
				SinceTimestamp: tt.sinceTimestamp,
			}

			resp, err := service.GetPeerState(context.Background(), req)
			require.NoError(t, err)

			assert.Equal(t, len(tt.wantDBs), len(resp.CurrentState))
			for _, dbID := range tt.wantDBs {
				assert.Contains(t, resp.CurrentState, dbID)
			}
		})
	}
}

func TestSyncService_MetricCleanup(t *testing.T) {
	reg := prometheus.NewRegistry()
	service := NewSyncService("test-collector", reg, 5*time.Second, 5*time.Second)

	// Add metrics and verify they're present
	update := &collectorv1.StateUpdate{
		CollectorId: "collector-1",
		Metrics: []*collectorv1.DatabaseMetrics{
			{
				DatabaseId: "db-1",
				Timestamp:  timestamppb1.New(time.Now()),
				SystemMetrics: &collectorv1.SystemMetrics{
					CpuUsage:          100,
					MemoryUsage:       80,
					ActiveConnections: 32,
					DiskIoRate:        82,
					NetworkIoRate:     12,
					Counters: []*collectorv1.PerformanceCounters{
						{CounterName: "a",
							CounterValue: 5,
							CounterRate:  1},
					},
				},
			},
		},
	}
	require.NoError(t, service.handleStateUpdate(update))

	// Modify TTL to simulate expiration
	service.mu.Lock()
	service.metricsTTL["db-1"] = time.Now().Add(-5 * time.Second * 2)
	service.mu.Unlock()

	// Wait for GC cycle
	time.Sleep(5*time.Second + time.Second)

	// Verify metrics were cleaned up
	metrics := service.GetMetrics()
	assert.Empty(t, metrics)

	// Verify Prometheus metrics were updated
	metricFamilies, err := reg.Gather()
	require.NoError(t, err)

	var foundStateSize bool
	for _, mf := range metricFamilies {
		if mf.GetName() == "collector_state_size" {
			foundStateSize = true
			assert.Equal(t, float64(0), mf.Metric[0].Gauge.GetValue())
		}
	}
	assert.True(t, foundStateSize, "collector_state_size metric not found")
}

func TestPrometheusMetricCleanup(t *testing.T) {
	reg := prometheus.NewRegistry()
	service := NewSyncService("test-collector", reg, 5*time.Second, 5*time.Second)

	// Create test data with two databases
	now := time.Now()
	updates := []*collectorv1.StateUpdate{
		{
			CollectorId: "collector-1",
			Metrics: []*collectorv1.DatabaseMetrics{
				{
					DatabaseId: "db-fresh",
					//QueryMetrics: []*collectorv1.QueryMetric{
					//	{
					//		QueryPattern:      "SELECT *",
					//		AvgExecutionTime: 1.5,
					//		RowsProcessed:    1000,
					//		CpuTime:          0.5,
					//		IoTime:           1.0,
					//	},
					//},
					SystemMetrics: &collectorv1.SystemMetrics{
						ActiveConnections: 100,
					},
				},
				{
					DatabaseId: "db-stale",
					//QueryMetrics: []*collectorv1.QueryMetric{
					//	{
					//		QueryPattern:      "INSERT",
					//		AvgExecutionTime: 0.5,
					//		RowsProcessed:    50,
					//		CpuTime:          0.1,
					//		IoTime:           0.4,
					//	},
					//},
					SystemMetrics: &collectorv1.SystemMetrics{
						ActiveConnections: 50,
					},
				},
			},
		},
	}

	// Add metrics
	for _, update := range updates {
		require.NoError(t, service.handleStateUpdate(update))
	}

	// Verify both databases have metrics in registry
	assertMetricExists(t, reg, "database_query_time_seconds", "db-fresh")
	assertMetricExists(t, reg, "database_query_time_seconds", "db-stale")

	// Make db-stale expire
	service.mu.Lock()
	service.metricsTTL["db-stale"] = now.Add(-5 * time.Second * 2)
	service.mu.Unlock()

	// Run GC
	service.gc()

	// Verify db-fresh metrics still exist
	assertMetricExists(t, reg, "database_query_time_seconds", "db-fresh")
	assertMetricExists(t, reg, "database_rows_processed_total", "db-fresh")
	assertMetricExists(t, reg, "database_cpu_time_seconds", "db-fresh")
	assertMetricExists(t, reg, "database_io_time_seconds", "db-fresh")
	assertMetricExists(t, reg, "database_active_connections", "db-fresh")

	// Verify specific metric values for db-fresh
	assertMetricValue(t, reg, "database_query_time_seconds", 1.5, map[string]string{
		"database_id":   "db-fresh",
		"query_pattern": "SELECT *",
	})
	assertMetricValue(t, reg, "database_active_connections", 100, map[string]string{
		"database_id": "db-fresh",
	})

	// Verify db-stale metrics are gone
	assertMetricNotExists(t, reg, "database_query_time_seconds", "db-stale")
	assertMetricNotExists(t, reg, "database_rows_processed_total", "db-stale")
	assertMetricNotExists(t, reg, "database_cpu_time_seconds", "db-stale")
	assertMetricNotExists(t, reg, "database_io_time_seconds", "db-stale")
	assertMetricNotExists(t, reg, "database_active_connections", "db-stale")
}

func TestPrometheusMetricUpdateAfterCleanup(t *testing.T) {
	reg := prometheus.NewRegistry()
	service := NewSyncService("test-collector", reg, 5*time.Second, 5*time.Second)

	// Add initial metrics
	update1 := &collectorv1.StateUpdate{
		CollectorId: "collector-1",
		Metrics: []*collectorv1.DatabaseMetrics{
			{
				DatabaseId: "db-1",
				//QueryMetrics: []*collectorv1.QueryMetric{
				//	{
				//		QueryPattern:      "SELECT *",
				//		AvgExecutionTime: 1.0,
				//	},
				//},
				SystemMetrics: &collectorv1.SystemMetrics{
					ActiveConnections: 100,
				},
			},
		},
	}
	require.NoError(t, service.handleStateUpdate(update1))

	// Force GC
	service.mu.Lock()
	service.metricsTTL["db-1"] = time.Now().Add(-5 * time.Second * 2)
	service.mu.Unlock()
	service.gc()

	// Add new metrics for same database
	update2 := &collectorv1.StateUpdate{
		CollectorId: "collector-1",
		Metrics: []*collectorv1.DatabaseMetrics{
			{
				DatabaseId: "db-1",
				//QueryMetrics: []*collectorv1.QueryMetric{
				//	{
				//		QueryPattern:      "SELECT *",
				//		AvgExecutionTime: 2.0,
				//	},
				//},
				SystemMetrics: &collectorv1.SystemMetrics{
					ActiveConnections: 200,
				},
			},
		},
	}
	require.NoError(t, service.handleStateUpdate(update2))

	// Verify new metrics are present with updated values
	assertMetricValue(t, reg, "database_query_time_seconds", 2.0, map[string]string{
		"database_id":   "db-1",
		"query_pattern": "SELECT *",
	})
	assertMetricValue(t, reg, "database_active_connections", 200, map[string]string{
		"database_id": "db-1",
	})
}

// Helper functions for metric assertions
func assertMetricExists(t *testing.T, reg prometheus.Gatherer, name, dbID string) {
	metrics, err := reg.Gather()
	require.NoError(t, err)

	found := false
	for _, mf := range metrics {
		if mf.GetName() == name {
			for _, m := range mf.GetMetric() {
				for _, l := range m.GetLabel() {
					if l.GetName() == "database_id" && l.GetValue() == dbID {
						found = true
						break
					}
				}
			}
		}
	}
	assert.True(t, found, "metric %s for database %s should exist", name, dbID)
}

func assertMetricNotExists(t *testing.T, reg prometheus.Gatherer, name, dbID string) {
	metrics, err := reg.Gather()
	require.NoError(t, err)

	found := false
	for _, mf := range metrics {
		if mf.GetName() == name {
			for _, m := range mf.GetMetric() {
				for _, l := range m.GetLabel() {
					if l.GetName() == "database_id" && l.GetValue() == dbID {
						found = true
						break
					}
				}
			}
		}
	}
	assert.False(t, found, "metric %s for database %s should not exist", name, dbID)
}

func assertMetricValue(t *testing.T, reg prometheus.Gatherer, name string, expected float64, labels map[string]string) {
	metrics, err := reg.Gather()
	require.NoError(t, err)

	for _, mf := range metrics {
		if mf.GetName() == name {
			for _, m := range mf.GetMetric() {
				if matchLabels(m, labels) {
					assert.Equal(t, expected, *m.Gauge.Value)
					return
				}
			}
		}
	}
	t.Errorf("metric %s with labels %v not found", name, labels)
}

func matchLabels(metric *io_prometheus_client.Metric, labels map[string]string) bool {
	for _, l := range metric.GetLabel() {
		if expected, ok := labels[l.GetName()]; ok {
			if expected != l.GetValue() {
				return false
			}
		}
	}
	return true
}
