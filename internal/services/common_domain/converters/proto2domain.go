package converters

import (
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
)

func DatabaseSnapshotToDomain(p *dbmv1.DBSnapshot) common_domain.DataBaseSnapshot {
	if p == nil {
		return common_domain.DataBaseSnapshot{}
	}
	samples := make([]*common_domain.QuerySample, 0, len(p.Samples))
	for _, sample := range p.Samples {
		if ds := SampleToDomain(sample); ds != nil {
			samples = append(samples, ds)
		}
	}
	return common_domain.DataBaseSnapshot{
		SnapInfo: common_domain.SnapInfo{
			ID:        p.GetId(),
			Timestamp: p.GetTimestamp().AsTime(),
			Server: common_domain.ServerMeta{
				Host: p.GetServer().GetHost(),
				Type: p.GetServer().GetType(),
			},
		},
		Samples: samples,
	}
}

func SampleToDomain(sample *dbmv1.QuerySample) *common_domain.QuerySample {
	if sample == nil {
		return nil
	}
	ses := sample.GetSession()
	sesCtx := common_domain.ContextInfo{}
	if sc := ses.GetContext(); sc != nil {
		sesCtx = common_domain.ContextInfo{
			IsTraceParent: sc.GetIsTraceParent(),
			TraceId:       sc.GetTraceId(),
			SpanId:        sc.GetSpanId(),
			Raw:           sc.GetRaw(),
		}
	}
	sampCtx := common_domain.ContextInfo{}
	if sc := sample.GetContext(); sc != nil {
		sampCtx = common_domain.ContextInfo{
			IsTraceParent: sc.GetIsTraceParent(),
			TraceId:       sc.GetTraceId(),
			SpanId:        sc.GetSpanId(),
			Raw:           sc.GetRaw(),
		}
	}
	waitInfo := sample.GetWaitInfo()
	var waitTypePtr *string
	if waitInfo != nil {
		wt := waitInfo.GetWaitType()
		waitTypePtr = &wt
	}
	return &common_domain.QuerySample{
		Status:        sample.GetStatus(),
		SqlHandle:     sample.GetSqlHandle(),
		QueryHash:     sample.GetQueryHash(),
		Text:          sample.GetText(),
		IsBlocked:     sample.GetBlocked(),
		IsBlocker:     sample.GetBlocker(),
		TimeElapsedMs: sample.GetTimeElapsedMillis(),
		Session: common_domain.SessionMetadata{
			SessionID:            ses.GetSessionId(),
			LoginTime:            ses.GetLoginTime().AsTime(),
			HostName:             ses.GetHost(),
			ProgramName:          ses.GetProgramName(),
			LoginName:            ses.GetLoginName(),
			Status:               ses.GetStatus(),
			LastRequestStartTime: ses.GetLastRequestStart().AsTime(),
			LastRequestEndTime:   ses.GetLastRequestEnd().AsTime(),
			ConnectionId:         ses.GetConnectionId(),
			ClientIP:             ses.GetClientIp(),

			Context: sesCtx,
		},
		Database: common_domain.DataBaseMetadata{
			DatabaseID:   sample.GetDb().GetDatabaseId(),
			DatabaseName: sample.GetDb().GetDatabaseName(),
		},
		Block: common_domain.BlockMetadata{
			BlockedBy:       sample.GetBlockInfo().GetBlockedBy(),
			BlockedSessions: sample.GetBlockInfo().GetBlockedSessions(),
		},
		Wait: common_domain.WaitMetadata{
			WaitType:     waitTypePtr,
			WaitTime:     int(waitInfo.GetWaitTime()),
			LastWaitType: waitInfo.GetLastWaitType(),
			WaitResource: waitInfo.GetWaitResource(),
		},
		Snapshot: common_domain.SnapshotMetadata{
			ID:        sample.GetSnapInfo().GetId(),
			Timestamp: sample.GetSnapInfo().GetTimestamp().AsTime(),
		},
		Cmd:             "",
		PlanHandle:      sample.GetPlanHandle(),
		Id:              sample.GetId(),
		CommandMetadata: CommandMetaToDomain(sample.GetCommand()),
		Context:         sampCtx,
	}
}

func QueryMetricToDomain(metric *dbmv1.QueryMetric) (*common_domain.QueryMetric, error) {
	if metric == nil {
		return nil, nil
	}
	return &common_domain.QueryMetric{
		QueryHash:         metric.GetQueryHash(),
		Text:              metric.GetText(),
		Database:          common_domain.DataBaseMetadata{DatabaseID: metric.GetDb().GetDatabaseId(), DatabaseName: metric.GetDb().GetDatabaseName()},
		LastExecutionTime: metric.GetLastExecutionTime().AsTime(),
		LastElapsedTime:   time.Duration(metric.GetLastElapsedTimeMicros()) * time.Microsecond,
		Counters:          metric.GetCounters(),
		Rates:             metric.GetRates(),
		CollectionTime:    metric.GetCollectedAt().AsTime(),
	}, nil
}

func ExecutionPlanToDomain(plan *dbmv1.ExecutionPlan) (*common_domain.ExecutionPlan, error) {
	if plan == nil {
		return nil, nil
	}
	return &common_domain.ExecutionPlan{
		PlanHandle: plan.GetPlanHandle(),
		Server: common_domain.ServerMeta{
			Host: plan.GetServer().GetHost(),
			Type: plan.GetServer().GetType(),
		},
		XmlData: plan.GetXmlPlan(),
	}, nil
}
func CommandMetaToDomain(cm *dbmv1.CommandMetadata) common_domain.CommandMetadata {
	if cm == nil {
		return common_domain.CommandMetadata{}
	}
	return common_domain.CommandMetadata{
		TransactionId:           cm.GetTransactionId(),
		RequestId:               cm.GetRequestId(),
		EstimatedCompletionTime: cm.GetEstimatedCompletionTime(),
		PercentComplete:         cm.GetPercentComplete(),
	}
}
