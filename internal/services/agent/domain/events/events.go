package events

import (
	"context"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
)

type Event interface {
	EventName() string
	Context() context.Context
}

type SampleSnapshotTaken struct {
	Snap *common_domain.DataBaseSnapshot
	Ctx  context.Context
}

func (e SampleSnapshotTaken) EventName() string {
	return "SampleSnapshotTaken"
}

func (e SampleSnapshotTaken) Context() context.Context {
	return e.Ctx
}

type MetricsSnapshotTaken struct {
	Metrics []*common_domain.QueryMetric
	Ctx     context.Context
}

func (e MetricsSnapshotTaken) EventName() string {
	return "MetricsSnapshotTaken"
}

func (e MetricsSnapshotTaken) Context() context.Context {
	return e.Ctx
}

type ExecutionPlanFetched struct {
	Plan *common_domain.ExecutionPlan
}

func (e ExecutionPlanFetched) EventName() string {
	return "ExecutionPlanFetched"
}

func (e ExecutionPlanFetched) Context() context.Context {
	return context.Background()
}

type WarningDetected struct {
	Warning *common_domain.Warning
}

func (e WarningDetected) EventName() string {
	return "WarningDetected"
}

func (e WarningDetected) Context() context.Context {
	return context.Background()
}
