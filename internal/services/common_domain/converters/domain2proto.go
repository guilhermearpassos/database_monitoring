package converters

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func DatabaseSnapshotToProto(d *common_domain.DataBaseSnapshot) *dbmv1.DBSnapshot {
	samples := make([]*dbmv1.QuerySample, len(d.Samples))
	for i, sample := range d.Samples {
		samples[i] = SampleToProto(sample)
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
	}
}
