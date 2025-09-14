package converters

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"time"
)

func DatabaseSnapshotToDomain(p *dbmv1.DBSnapshot) common_domain.DataBaseSnapshot {
	samples := make([]*common_domain.QuerySample, len(p.Samples))
	for i, sample := range p.Samples {
		samples[i] = SampleToDomain(sample)
	}
	return common_domain.DataBaseSnapshot{
		SnapInfo: common_domain.SnapInfo{
			ID:        p.Id,
			Timestamp: p.Timestamp.AsTime(),
			Server: common_domain.ServerMeta{
				Host: p.Server.Host,
				Type: p.Server.Type,
			},
		},
		Samples: samples,
	}
}

func SampleToDomain(sample *dbmv1.QuerySample) *common_domain.QuerySample {
	return &common_domain.QuerySample{
		Status:        sample.Status,
		SqlHandle:     sample.SqlHandle,
		QueryHash:     sample.QueryHash,
		Text:          sample.Text,
		IsBlocked:     sample.Blocked,
		IsBlocker:     sample.Blocker,
		TimeElapsedMs: sample.TimeElapsedMillis,
		Session: common_domain.SessionMetadata{
			SessionID:            sample.Session.SessionId,
			LoginTime:            sample.Session.LoginTime.AsTime(),
			HostName:             sample.Session.Host,
			ProgramName:          sample.Session.ProgramName,
			LoginName:            sample.Session.LoginName,
			Status:               sample.Session.Status,
			LastRequestStartTime: sample.Session.LastRequestStart.AsTime(),
			LastRequestEndTime:   sample.Session.LastRequestEnd.AsTime(),
			ConnectionId:         sample.Session.ConnectionId,
			ClientIP:             sample.Session.ClientIp,
		},
		Database: common_domain.DataBaseMetadata{
			DatabaseID:   sample.Db.DatabaseId,
			DatabaseName: sample.Db.DatabaseName,
		},
		Block: common_domain.BlockMetadata{
			BlockedBy:       sample.BlockInfo.BlockedBy,
			BlockedSessions: sample.BlockInfo.BlockedSessions,
		},
		Wait: common_domain.WaitMetadata{
			WaitType:     &sample.WaitInfo.WaitType,
			WaitTime:     int(sample.WaitInfo.WaitTime),
			LastWaitType: sample.WaitInfo.LastWaitType,
			WaitResource: sample.WaitInfo.WaitResource,
		},
		Snapshot: common_domain.SnapshotMetadata{
			ID:        sample.SnapInfo.Id,
			Timestamp: sample.SnapInfo.Timestamp.AsTime(),
		},
		Cmd:             "",
		PlanHandle:      sample.PlanHandle,
		Id:              sample.Id,
		CommandMetadata: CommandMetaToDomain(sample.Command),
	}
}

func QueryMetricToDomain(metric *dbmv1.QueryMetric) (*common_domain.QueryMetric, error) {
	return &common_domain.QueryMetric{
		QueryHash:         metric.QueryHash,
		Text:              metric.Text,
		Database:          common_domain.DataBaseMetadata{DatabaseID: metric.Db.DatabaseId, DatabaseName: metric.Db.DatabaseName},
		LastExecutionTime: metric.LastExecutionTime.AsTime(),
		LastElapsedTime:   time.Duration(metric.LastElapsedTimeMicros) * time.Microsecond,
		Counters:          metric.Counters,
		Rates:             metric.Rates,
	}, nil
}

func ExecutionPlanToDomain(plan *dbmv1.ExecutionPlan) (*common_domain.ExecutionPlan, error) {
	return &common_domain.ExecutionPlan{
		PlanHandle: plan.PlanHandle,
		Server: common_domain.ServerMeta{
			Host: plan.Server.Host,
			Type: plan.Server.Type,
		},
		XmlData: plan.XmlPlan,
	}, nil
}
func CommandMetaToDomain(cm *dbmv1.CommandMetadata) common_domain.CommandMetadata {
	return common_domain.CommandMetadata{
		TransactionId:           cm.TransactionId,
		RequestId:               cm.RequestId,
		EstimatedCompletionTime: cm.EstimatedCompletionTime,
		PercentComplete:         cm.PercentComplete,
	}
}
