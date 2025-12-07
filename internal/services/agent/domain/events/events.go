package events

import "github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"

type Event interface {
	EventName() string
}

type SampleSnapshotTaken struct {
	Snap *common_domain.DataBaseSnapshot
}

func (e SampleSnapshotTaken) EventName() string {
	return "SampleSnapshotTaken"
}

type MetricsSnapshotTaken struct {
	Metrics []*common_domain.QueryMetric
}

func (e MetricsSnapshotTaken) EventName() string {
	return "MetricsSnapshotTaken"
}

type ExecutionPlanFetched struct {
	Plan *common_domain.ExecutionPlan
}

func (e ExecutionPlanFetched) EventName() string {
	return "ExecutionPlanFetched"
}

type WarningDetected struct {
	Warning *common_domain.Warning
}

func (e WarningDetected) EventName() string {
	return "WarningDetected"
}
