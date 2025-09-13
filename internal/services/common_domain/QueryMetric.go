package common_domain

import "time"

type QueryMetric struct {
	QueryHash         string
	Text              string
	Database          DataBaseMetadata
	LastExecutionTime time.Time
	LastElapsedTime   time.Duration
	Counters          map[string]int64
	Rates             map[string]float64
	CollectionTime    time.Time
}
