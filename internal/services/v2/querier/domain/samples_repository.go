package domain

import (
	"context"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/ports/tasks/parsers"
)

type SampleRepository interface {
	ListServers(ctx context.Context, start time.Time, end time.Time) ([]ServerSummary, error)
	ListSnapshots(ctx context.Context, databaseID string, start time.Time, end time.Time, pageNumber int, pageSize int, serverID string) ([]common_domain.DataBaseSnapshot, int, error)
	GetSnapshot(ctx context.Context, id string) (common_domain.DataBaseSnapshot, error)
	ListSnapshotSummaries(ctx context.Context, serverID string, start time.Time, end time.Time) ([]common_domain.SnapshotSummary, error)
	GetExecutionPlan(ctx context.Context, planHandle string, server common_domain.ServerMeta) (*common_domain.ExecutionPlan, error)
	ListPlansWithIssues(ctx context.Context, ServerID string, Start, End time.Time, pageSize, pageNumber int) ([]*PlanWithIssue, error)
}
type PlanWithIssue struct {
	ParsedPlan  *parsers.ParsedExecutionPlan
	Sample      *common_domain.QuerySample
	IssueCount  int
	LockCount   int
	Occurrences int
}
