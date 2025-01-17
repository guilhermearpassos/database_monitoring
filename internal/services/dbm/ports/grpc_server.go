package ports

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/dbm/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/dbm/app/query"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type GRPCServer struct {
	dbmv1.UnimplementedDBMApiServer
	dbmv1.UnimplementedDBMSupportApiServer
	app app.Application
}

func NewGRPCServer(app app.Application) GRPCServer {
	return GRPCServer{app: app}
}

func (s GRPCServer) ListDatabases(ctx context.Context, request *dbmv1.ListDatabasesRequest) (*dbmv1.ListDatabasesResponse, error) {
	servers, err := s.app.Queries.ListDatabases.Handle(ctx, request.GetStart().AsTime(), request.GetEnd().AsTime())
	if err != nil {
		return nil, err
	}
	protoServers := make([]*dbmv1.InstrumentedServer, len(servers))
	for i, server := range servers {
		dbs := make([]*dbmv1.DBMetadata, len(server.DataBaseMetadata))
		for j, dataBaseMetadata := range server.DataBaseMetadata {
			dbs[j] = &dbmv1.DBMetadata{
				DatabaseId:   dataBaseMetadata.DatabaseID,
				DatabaseName: dataBaseMetadata.DatabaseName,
			}
		}
		protoServers[i] = &dbmv1.InstrumentedServer{
			Server: &dbmv1.ServerMetadata{
				Host: server.ServerMeta.Host,
				Type: server.ServerMeta.Type,
			},
			Db: dbs,
		}
	}
	return &dbmv1.ListDatabasesResponse{Servers: protoServers}, nil
}

func (s GRPCServer) ListSnapshots(ctx context.Context, request *dbmv1.ListSnapshotsRequest) (*dbmv1.ListSnapshotsResponse, error) {
	pageNumber := request.PageNumber
	if pageNumber == 0 {
		pageNumber = 1
	}

	snaps, total, err := s.app.Queries.ListSnapshots.Handle(ctx, query.SnapshotsQuery{
		Start:      request.Start.AsTime(),
		End:        request.End.AsTime(),
		PageSize:   int(request.PageSize),
		PageNumber: int(pageNumber),
	})
	if err != nil {
		return nil, err
	}
	protoSnaps := make([]*dbmv1.DBSnapshot, len(snaps))
	for i, snap := range snaps {
		protoSamples := make([]*dbmv1.QuerySample, len(snap.Samples))
		for j, sample := range snap.Samples {
			var waitType string
			if sample.Wait.WaitType != nil {
				waitType = *sample.Wait.WaitType
			}
			protoSamples[j] = &dbmv1.QuerySample{
				Status:    sample.Status,
				SqlHandle: sample.SqlHandle,
				Text:      sample.Text,
				Blocked:   sample.IsBlocked,
				Blocker:   sample.IsBlocker,
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
			}
		}
		protoSnaps[i] = &dbmv1.DBSnapshot{
			Id:        "",
			Timestamp: timestamppb.New(snap.SnapInfo.Timestamp.In(time.UTC)),
			Server: &dbmv1.ServerMetadata{
				Host: snap.SnapInfo.Server.Host,
				Type: snap.SnapInfo.Server.Type,
			},
			Database: &dbmv1.DBMetadata{
				DatabaseId:   snap.SnapInfo.Database.DatabaseID,
				DatabaseName: snap.SnapInfo.Database.DatabaseName,
			},
			Samples: protoSamples,
		}
	}
	return &dbmv1.ListSnapshotsResponse{
		Snapshots:  protoSnaps,
		PageNumber: pageNumber,
		TotalCount: int64(total),
	}, nil
}
