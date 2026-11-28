package collector

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
)

var lockMetricLabels = []string{"server", "database", "table", "index", "wait_type", "resource_type"}

type lockMetricKey struct {
	database     string
	table        string
	index        string
	waitType     string
	resourceType string
}

type lockMetricAgg struct {
	count      int
	maxWaitSec float64
}

// LockMetricsRecorder maintains Prometheus metrics for blocked sessions resolved to
// specific SQL Server tables and indexes.
type LockMetricsRecorder struct {
	blockedSessions   *prometheus.GaugeVec
	blockedMaxWaitSec *prometheus.GaugeVec
	locksTotal        *prometheus.CounterVec
	lockDurationSec   *prometheus.HistogramVec

	mu             sync.Mutex
	activeByServer map[string]map[lockMetricKey]struct{}
}

func newLockMetricsRecorder(reg prometheus.Registerer) *LockMetricsRecorder {
	r := &LockMetricsRecorder{
		blockedSessions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "sqlsights",
				Name:      "database_blocked_sessions",
				Help:      "Current number of blocked sessions waiting on a locked database table/index in the latest snapshot.",
			},
			lockMetricLabels,
		),
		blockedMaxWaitSec: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "sqlsights",
				Name:      "database_blocked_max_wait_seconds",
				Help:      "Longest current wait duration (seconds) among blocked sessions on a table/index in the latest snapshot.",
			},
			lockMetricLabels,
		),
		locksTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "sqlsights",
				Name:      "database_locks_total",
				Help:      "Total blocked query samples observed per table/index and wait type.",
			},
			lockMetricLabels,
		),
		lockDurationSec: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "sqlsights",
				Name:      "database_lock_duration_seconds",
				Help:      "Observed wait duration (seconds) of blocked query samples per table/index and wait type.",
				Buckets:   []float64{5, 10, 30, 60, 120, 300},
			},
			lockMetricLabels,
		),
		activeByServer: make(map[string]map[lockMetricKey]struct{}),
	}
	if reg != nil {
		reg.MustRegister(r.blockedSessions, r.blockedMaxWaitSec, r.locksTotal, r.lockDurationSec)
	}
	return r
}

var defaultLockMetrics = newLockMetricsRecorder(prometheus.DefaultRegisterer)

// RecordSnapshot processes all blocked query samples in a snapshot, resolves their wait_resource
// to database/table/index via the cached resolver, updates gauges/counters/histograms, and resets
// gauges to 0 for any previously locked table/index that is no longer blocked.
func (m *LockMetricsRecorder) RecordSnapshot(
	ctx context.Context,
	server string,
	samples []*common_domain.QuerySample,
	resolver *WaitResourceResolver,
) {
	if m == nil {
		return
	}
	span := trace.SpanFromContext(ctx)

	current := make(map[lockMetricKey]*lockMetricAgg)
	for _, sample := range samples {
		if sample == nil || !sample.IsBlocked {
			continue
		}

		var target ResolvedLockTarget
		if resolver != nil {
			var err error
			target, err = resolver.Resolve(ctx, sample.Wait.WaitResource, sample.Database.DatabaseName)
			if err != nil {
				span.RecordError(err)
				slog.WarnContext(ctx, "failed to resolve lock wait resource",
					"server", server,
					"database", sample.Database.DatabaseName,
					"wait_resource", sample.Wait.WaitResource,
					"error", err,
				)
			}
		} else {
			target = ResolvedLockTarget{
				Database:     sample.Database.DatabaseName,
				ResourceType: "UNKNOWN",
			}
		}

		waitType := "UNKNOWN"
		if sample.Wait.WaitType != nil && strings.TrimSpace(*sample.Wait.WaitType) != "" {
			waitType = strings.TrimSpace(*sample.Wait.WaitType)
		}
		dbName := target.Database
		if dbName == "" {
			dbName = sample.Database.DatabaseName
		}
		if dbName == "" {
			dbName = "unknown"
		}
		tableName := target.Table
		if tableName == "" {
			tableName = "unknown"
		}
		resType := target.ResourceType
		if resType == "" {
			resType = "UNKNOWN"
		}

		key := lockMetricKey{
			database:     dbName,
			table:        tableName,
			index:        target.Index,
			waitType:     waitType,
			resourceType: resType,
		}

		waitSec := float64(sample.Wait.WaitTime) / 1000.0
		if waitSec < 0 {
			waitSec = 0
		}

		agg, exists := current[key]
		if !exists {
			agg = &lockMetricAgg{}
			current[key] = agg
		}
		agg.count++
		if waitSec > agg.maxWaitSec {
			agg.maxWaitSec = waitSec
		}

		m.locksTotal.WithLabelValues(server, key.database, key.table, key.index, key.waitType, key.resourceType).Inc()
		m.lockDurationSec.WithLabelValues(server, key.database, key.table, key.index, key.waitType, key.resourceType).Observe(waitSec)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	prev := m.activeByServer[server]
	next := make(map[lockMetricKey]struct{}, len(current))

	for key, agg := range current {
		m.blockedSessions.WithLabelValues(server, key.database, key.table, key.index, key.waitType, key.resourceType).Set(float64(agg.count))
		m.blockedMaxWaitSec.WithLabelValues(server, key.database, key.table, key.index, key.waitType, key.resourceType).Set(agg.maxWaitSec)
		next[key] = struct{}{}
	}

	// Reset gauges to 0 for locks that cleared since the previous snapshot so alerting rules resolve cleanly.
	for oldKey := range prev {
		if _, stillActive := current[oldKey]; !stillActive {
			m.blockedSessions.WithLabelValues(server, oldKey.database, oldKey.table, oldKey.index, oldKey.waitType, oldKey.resourceType).Set(0)
			m.blockedMaxWaitSec.WithLabelValues(server, oldKey.database, oldKey.table, oldKey.index, oldKey.waitType, oldKey.resourceType).Set(0)
		}
	}

	m.activeByServer[server] = next
}
