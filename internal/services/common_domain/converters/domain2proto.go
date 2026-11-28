package converters

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func DatabaseSnapshotToProto(d *common_domain.DataBaseSnapshot) *dbmv1.DBSnapshot {
	if d == nil {
		return nil
	}
	samples := make([]*dbmv1.QuerySample, 0, len(d.Samples))
	for _, sample := range d.Samples {
		if ps := SampleToProto(sample); ps != nil {
			samples = append(samples, ps)
		}
	}
	return &dbmv1.DBSnapshot{
		Id:        d.SnapInfo.ID,
		Timestamp: timestamppb.New(d.SnapInfo.Timestamp),
		Server: &dbmv1.ServerMetadata{Host: d.SnapInfo.Server.Host,
			Type: d.SnapInfo.Server.Type,
		},
		Samples: samples,
	}
}

func SampleToProto(sample *common_domain.QuerySample) *dbmv1.QuerySample {
	if sample == nil {
		return nil
	}
	var waitType string
	if sample.Wait.WaitType != nil {
		waitType = *sample.Wait.WaitType
	}
	return &dbmv1.QuerySample{
		Status:            sample.Status,
		SqlHandle:         sample.SqlHandle,
		Text:              sample.Text,
		Blocked:           sample.IsBlocked,
		Blocker:           sample.IsBlocker,
		QueryHash:         sample.QueryHash,
		TimeElapsedMillis: sample.TimeElapsedMs,
		Session: &dbmv1.SessionMetadata{
			SessionId:        sample.Session.SessionID,
			LoginTime:        timestamppb.New(sample.Session.LoginTime),
			Host:             sample.Session.HostName,
			ProgramName:      sample.Session.ProgramName,
			LoginName:        sample.Session.LoginName,
			Status:           sample.Session.Status,
			LastRequestStart: timestamppb.New(sample.Session.LastRequestStartTime),
			LastRequestEnd:   timestamppb.New(sample.Session.LastRequestEndTime),
			ConnectionId:     sample.Session.ConnectionId,
			ClientIp:         sample.Session.ClientIP,
			Context: &dbmv1.ContextInfo{
				IsTraceParent: sample.Session.Context.IsTraceParent,
				TraceId:       sample.Session.Context.TraceId,
				SpanId:        sample.Session.Context.SpanId,
				Raw:           sample.Session.Context.Raw,
			},
		},
		Db: &dbmv1.DBMetadata{
			DatabaseId:   sample.Database.DatabaseID,
			DatabaseName: sample.Database.DatabaseName,
		},
		BlockInfo: &dbmv1.BlockMetadata{
			BlockedBy:       sample.Block.BlockedBy,
			BlockedSessions: sample.Block.BlockedSessions,
		},
		WaitInfo: &dbmv1.WaitMetadata{
			WaitType:     waitType,
			WaitTime:     int64(sample.Wait.WaitTime),
			LastWaitType: sample.Wait.LastWaitType,
			WaitResource: sample.Wait.WaitResource,
		},
		SnapInfo: &dbmv1.SnapMetadata{
			Id:        sample.Snapshot.ID,
			Timestamp: timestamppb.New(sample.Snapshot.Timestamp),
		},
		PlanHandle: sample.PlanHandle,
		Id:         sample.Id,
		Command:    CommandMetaToProto(&sample.CommandMetadata),

		Context: &dbmv1.ContextInfo{
			IsTraceParent: sample.Context.IsTraceParent,
			TraceId:       sample.Context.TraceId,
			SpanId:        sample.Context.SpanId,
			Raw:           sample.Context.Raw,
		},
	}
}

func QueryMetricToProto(metric *common_domain.QueryMetric) (*dbmv1.QueryMetric, error) {
	if metric == nil {
		return nil, nil
	}
	return &dbmv1.QueryMetric{
		QueryHash:             metric.QueryHash,
		Text:                  metric.Text,
		Db:                    &dbmv1.DBMetadata{DatabaseId: metric.Database.DatabaseID, DatabaseName: metric.Database.DatabaseName},
		LastExecutionTime:     timestamppb.New(metric.LastExecutionTime),
		LastElapsedTimeMicros: metric.LastElapsedTime.Microseconds(),
		Counters:              metric.Counters,
		Rates:                 metric.Rates,
		CollectedAt:           timestamppb.New(metric.CollectionTime),
	}, nil
}

func ExecutionPlanToProto(plan *common_domain.ExecutionPlan) (*dbmv1.ExecutionPlan, error) {
	if plan == nil {
		return nil, nil
	}
	return &dbmv1.ExecutionPlan{
		PlanHandle: plan.PlanHandle,
		Server: &dbmv1.ServerMetadata{
			Host: plan.Server.Host,
			Type: plan.Server.Type,
		},
		XmlPlan: plan.XmlData,
	}, nil
}

func CommandMetaToProto(cm *common_domain.CommandMetadata) *dbmv1.CommandMetadata {
	if cm == nil {
		return &dbmv1.CommandMetadata{}
	}
	return &dbmv1.CommandMetadata{
		TransactionId:           cm.TransactionId,
		RequestId:               cm.RequestId,
		EstimatedCompletionTime: cm.EstimatedCompletionTime,
		PercentComplete:         cm.PercentComplete,
	}
}
