// sync_service.go
package collector

import (
	collectorv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1/collector"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type SyncService struct {
	collectorv1.UnimplementedCollectorSyncServiceServer
	collectorID string
	metrics     map[string]*collectorv1.DatabaseMetrics
	metricsTTL  map[string]time.Time
	mu          sync.RWMutex

	// Prometheus metrics
	stateSize  *prometheus.GaugeVec
	syncErrors *prometheus.CounterVec
	lastSync   *prometheus.GaugeVec
	dbMetrics  map[string]*dbPrometheusMetrics
	registry   prometheus.Registerer
	metricTTL  time.Duration
	gcInterval time.Duration
}

type dbPrometheusMetrics struct {
	queryTime     *prometheus.GaugeVec
	rowsProcessed *prometheus.GaugeVec
	cpuTime       *prometheus.GaugeVec
	ioTime        *prometheus.GaugeVec
	connections   *prometheus.GaugeVec
}

func newDBMetrics(reg prometheus.Registerer, dbID string) *dbPrometheusMetrics {
	m := &dbPrometheusMetrics{
		queryTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_query_time_seconds",
				Help: "Average query execution time",
			},
			[]string{"database_id", "query_pattern"},
		),
		rowsProcessed: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_rows_processed_total",
				Help: "Number of rows processed",
			},
			[]string{"database_id", "query_pattern"},
		),
		cpuTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_cpu_time_seconds",
				Help: "CPU time used by queries",
			},
			[]string{"database_id", "query_pattern"},
		),
		ioTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_io_time_seconds",
				Help: "IO time used by queries",
			},
			[]string{"database_id", "query_pattern"},
		),
		connections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_active_connections",
				Help: "Number of active connections",
			},
			[]string{"database_id"},
		),
	}

	reg.MustRegister(
		m.queryTime,
		m.rowsProcessed,
		m.cpuTime,
		m.ioTime,
		m.connections,
	)

	return m
}

func NewSyncService(collectorID string, reg prometheus.Registerer,
	metricTTL time.Duration,
	gcInterval time.Duration) *SyncService {
	s := &SyncService{
		collectorID: collectorID,
		metrics:     make(map[string]*collectorv1.DatabaseMetrics),
		metricsTTL:  make(map[string]time.Time),
		dbMetrics:   make(map[string]*dbPrometheusMetrics),
		registry:    reg,
		stateSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "collector_state_size",
				Help: "Number of metrics currently stored in collector state",
			},
			[]string{"collector_id"},
		),
		syncErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "collector_sync_errors_total",
				Help: "Total number of sync errors",
			},
			[]string{"collector_id", "error_type"},
		),
		lastSync: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "collector_last_sync_timestamp",
				Help: "Timestamp of last successful sync",
			},
			[]string{"collector_id"},
		),
		metricTTL:  metricTTL,
		gcInterval: gcInterval,
	}

	reg.MustRegister(s.stateSize, s.syncErrors, s.lastSync)

	go s.runGC()

	return s
}

func (s *SyncService) handleStateUpdate(update *collectorv1.StateUpdate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	//for _, metrics := range update.Metrics {
	//dbID := metrics.DatabaseId

	// Get or create Prometheus metrics for this database
	//dbMetrics, exists := s.dbMetrics[dbID]
	//if !exists {
	//	dbMetrics = newDBMetrics(s.registry, dbID)
	//	s.dbMetrics[dbID] = dbMetrics
	//}
	//
	//// Update metrics and TTL
	//s.metrics[dbID] = metrics
	//s.metricsTTL[dbID] = time.Now().Add(s.metricTTL)

	// Update Prometheus metrics
	//for _, qm := range metrics.QueryMetrics {
	//	labels := prometheus.Labels{
	//		"database_id":   dbID,
	//		"query_pattern": qm.QueryPattern,
	//	}
	//	dbMetrics.queryTime.With(labels).Set(qm.AvgExecutionTime)
	//	dbMetrics.rowsProcessed.With(labels).Set(qm.RowsProcessed)
	//	dbMetrics.cpuTime.With(labels).Set(qm.CpuTime)
	//	dbMetrics.ioTime.With(labels).Set(qm.IoTime)
	//}

	//dbMetrics.connections.With(prometheus.Labels{
	//	"database_id": dbID,
	//}).Set(float64(metrics.SystemMetrics.ActiveConnections))
	//}

	s.stateSize.WithLabelValues(s.collectorID).Set(float64(len(s.metrics)))
	s.lastSync.WithLabelValues(s.collectorID).SetToCurrentTime()

	return nil
}

func (s *SyncService) runGC() {
	ticker := time.NewTicker(s.gcInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.gc()
	}
}

func (s *SyncService) gc() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for dbID, ttl := range s.metricsTTL {
		if now.After(ttl) {
			// Clean up internal state
			delete(s.metrics, dbID)
			delete(s.metricsTTL, dbID)

			// Clean up Prometheus metrics
			if dbMetrics, exists := s.dbMetrics[dbID]; exists {
				// Unregister all metric vectors for this database
				s.registry.Unregister(dbMetrics.queryTime)
				s.registry.Unregister(dbMetrics.rowsProcessed)
				s.registry.Unregister(dbMetrics.cpuTime)
				s.registry.Unregister(dbMetrics.ioTime)
				s.registry.Unregister(dbMetrics.connections)
				delete(s.dbMetrics, dbID)
			}
		}
	}

	s.stateSize.WithLabelValues(s.collectorID).Set(float64(len(s.metrics)))
}

// Helper method to get metrics for testing
func (s *SyncService) GetMetrics() map[string]*collectorv1.DatabaseMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modifications
	metrics := make(map[string]*collectorv1.DatabaseMetrics, len(s.metrics))
	for k, v := range s.metrics {
		metrics[k] = v
	}
	return metrics
}
