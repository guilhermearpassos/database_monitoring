package config

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

type Target struct {
	Alias      string
	Connection *sqlx.DB
}
type TargetMonitoring struct {
	SnapshotMonitoring SnapshotMonitoring
	MetricsMonitoring  MetricsMonitoring
	DeadlockMonitoring DeadlockMonitoring
}
type DeadlockMonitoring struct {
	Enabled  bool
	DBFilter DBFilter
}
type MetricsMonitoring struct {
	Enabled              bool
	CollectInfraMetrics  bool
	CollectSampleMetrics bool
	DBFilter             DBFilter
}

type DBFilterMode string

const (
	DBFilterModeAll       DBFilterMode = "all"
	DBFilterModeWhitelist DBFilterMode = "whitelist"
	DBFilterModeBlacklist DBFilterMode = "blacklist"
)

func ToDBFilterMode(mode string) (DBFilterMode, error) {
	switch mode {
	case "all":
		return DBFilterModeAll, nil
	case "":
		return DBFilterModeAll, nil
	case "whitelist":
		return DBFilterModeWhitelist, nil
	case "blacklist":
		return DBFilterModeBlacklist, nil
	default:
		return DBFilterModeAll, fmt.Errorf("invalid filter mode: %s", mode)

	}
}

type DBFilter struct {
	DBFilterMode      DBFilterMode
	IncludedDatabases []string
	ExcludedDatabases []string
}

type SnapshotFilter struct {
	IncludeSqlsightsQueries bool
	DBFilter                DBFilter
}
type SnapshotMonitoring struct {
	Enabled  bool
	Interval time.Duration
	Filter   SnapshotFilter
}

type PlanMonitoring struct {
	EnableCollection         bool
	BackwardsSearchThreshold time.Duration // 0 to disable
	EnableAnalyser           bool
}
