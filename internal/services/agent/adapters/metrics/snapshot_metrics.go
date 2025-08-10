package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"sync"
)

var (
	// DatabaseLockDuration tracks the duration of database locks in seconds by server, database, and lock type
	DatabaseLockDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "sqlsights_database_lock_duration_seconds",
			Help: "Duration of database locks in seconds",
		},
		[]string{"server", "database", "wait_type", "table"},
	)

	// DatabaseLocksTotal tracks the total number of database locks by server and database
	DatabaseLocksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sqlsights_database_locks_total",
			Help: "Total number of database locks",
		},
		[]string{"server", "database"},
	)
)

func init() {
	sync.OnceFunc(func() {
		prometheus.MustRegister(DatabaseLockDuration)
		prometheus.MustRegister(DatabaseLocksTotal)
	})
}
