package converters

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
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
	}
}
