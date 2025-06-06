package ports

import (
	"context"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/app/query"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type GRPCServer struct {
	dbmv1.UnimplementedDBMApiServer
	dbmv1.UnimplementedDBMSupportApiServer
	app    *app.Application
	tracer trace.Tracer
}

func NewGRPCServer(app *app.Application) GRPCServer {
	return GRPCServer{app: app,
		tracer: otel.Tracer("grpc-server"),
	}
}

func (s GRPCServer) ListDatabases(ctx context.Context, request *dbmv1.ListDatabasesRequest) (*dbmv1.ListDatabasesResponse, error) {
	panic("implement me")
	//servers, err := s.app.Queries.ListServerSummary.Handle(ctx, request.GetStart().AsTime(), request.GetEnd().AsTime())
	//if err != nil {
	//	return nil, err
	//}
	//protoServers := make([]*dbmv1.InstrumentedServer, len(servers))
	//for i, server := range servers {
	//	//dbs := make([]*dbmv1.DBMetadata, len(server.DataBaseMetadata))
	//	//for j, dataBaseMetadata := range server.DataBaseMetadata {
	//	//	dbs[j] = &dbmv1.DBMetadata{
	//	//		DatabaseId:   dataBaseMetadata.DatabaseID,
	//	//		DatabaseName: dataBaseMetadata.DatabaseName,
	//	//	}
	//	//}
	//	protoServers[i] = &dbmv1.InstrumentedServer{
	//		Server: &dbmv1.ServerMetadata{
	//			Host: server.Host,
	//			Type: server.Type,
	//		},
	//		Db: nil,
	//	}
	//}
	//return &dbmv1.ListDatabasesResponse{Servers: protoServers}, nil
}

func (s GRPCServer) ListSnapshots(ctx context.Context, request *dbmv1.ListSnapshotsRequest) (*dbmv1.ListSnapshotsResponse, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("request.start", request.Start.AsTime().Format(time.RFC3339)),
		attribute.String("request.end", request.End.AsTime().Format(time.RFC3339)),
		attribute.String("request.host", request.Host),
		attribute.Int64("request.page_number", request.PageNumber),
		attribute.Int64("request.page_size", int64(request.PageSize)),
	)
	pageNumber := request.PageNumber
	if pageNumber == 0 {
		pageNumber = 1
	}

	snaps, total, err := s.app.Queries.ListSnapshots.Handle(ctx, query.SnapshotsQuery{
		Start:      request.Start.AsTime(),
		End:        request.End.AsTime(),
		PageSize:   int(request.PageSize),
		PageNumber: int(pageNumber),
		ServerID:   request.Host,
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
		protoSnaps[i] = converters.DatabaseSnapshotToProto(&snap)
	}
	span.SetAttributes(
		attribute.Int64("response.size", int64(len(snaps))),
		attribute.Int64("response.total_items", int64(total)),
	)
	return &dbmv1.ListSnapshotsResponse{
		Snapshots:  protoSnaps,
		PageNumber: pageNumber,
		TotalCount: int64(total),
	}, nil
}

func (s GRPCServer) ListServerSummary(ctx context.Context, request *dbmv1.ListServerSummaryRequest) (*dbmv1.ListServerSummaryResponse, error) {
	resp, err := s.app.Queries.ListServerSummary.Handle(ctx, request.Start.AsTime(), request.End.AsTime())
	if err != nil {
		return nil, err
	}
	protoServers := make([]*dbmv1.ServerSummary, len(resp))
	for i, srv := range resp {
		protoServers[i] = &dbmv1.ServerSummary{
			Name:                   srv.Name,
			Type:                   srv.Type,
			Connections:            int32(srv.Connections),
			RequestRate:            srv.RequestRate,
			ConnectionsByWaitGroup: srv.ConnsByWaitGroup,
		}
	}
	return &dbmv1.ListServerSummaryResponse{Servers: protoServers}, nil
	//servers, err := s.app.Queries.ListServerSummary.Handle(ctx, request.GetStart().AsTime(), request.GetEnd().AsTime())
	//if err != nil {
	//	return nil, err
	//}
	//protoServers := make([]*dbmv1.ServerSummary, len(servers))
	//for i, server := range servers {
	//	protoServers[i] = &dbmv1.ServerSummary{
	//		Name:                   server.Name,
	//		Type:                   server.Type,
	//		Connections:            int32(server.Connections),
	//		RequestRate:            server.RequestRate,
	//		ConnectionsByWaitGroup: server.ConnsByWaitGroup,
	//	}
	//}
	//return &dbmv1.ListServerSummaryResponse{Servers: protoServers}, nil
}

func (s GRPCServer) GetSnapshot(ctx context.Context, request *dbmv1.GetSnapshotRequest) (*dbmv1.GetSnapshotResponse, error) {
	snap, err := s.app.Queries.GetSnapshot.Handle(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &dbmv1.GetSnapshotResponse{Snapshot: converters.DatabaseSnapshotToProto(&snap)}, nil
}

func (s GRPCServer) ListQueryMetrics(ctx context.Context, in *dbmv1.ListQueryMetricsRequest) (*dbmv1.ListQueryMetricsResponse, error) {
	resp, err := s.app.Queries.ListQueryMetrics.Handle(ctx, in.Start.AsTime(), in.End.AsTime())
	if err != nil {
		return nil, err
	}
	ret := make([]*dbmv1.QueryMetric, len(resp))
	for i, metric := range resp {
		protoMetric, err2 := converters.QueryMetricToProto(metric)
		if err2 != nil {
			return nil, err2
		}
		ret[i] = protoMetric
	}

	return &dbmv1.ListQueryMetricsResponse{Metrics: ret}, nil
}

func (s GRPCServer) GetSampleDetails(ctx context.Context, in *dbmv1.GetSampleDetailsRequest) (*dbmv1.GetSampleDetailsResponse, error) {
	resp, err := s.app.Queries.GetQuerySampleDetails.Handle(ctx, in.GetSnapId(), in.SampleId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
