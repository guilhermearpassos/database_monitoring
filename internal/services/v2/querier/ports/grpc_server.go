package ports

import (
	"context"
	"fmt"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/querier/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/querier/app/query"
	querierv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/querier/v2"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCServer struct {
	querierv2.UnimplementedQuerierAPIServer
	app    *app.Application
	tracer trace.Tracer
}

func NewGRPCServer(app *app.Application) GRPCServer {
	return GRPCServer{app: app,
		tracer: otel.Tracer("grpc-server"),
	}
}

func (s GRPCServer) ListSnapshotSummaries(ctx context.Context, in *querierv2.ListSnapshotSummariesRequest) (*querierv2.ListSnapshotSummariesResponse, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("request.start", in.Start.AsTime().Format(time.RFC3339)),
		attribute.String("request.end", in.End.AsTime().Format(time.RFC3339)),
		attribute.String("request.server", in.Server),
	)

	resp, err := s.app.Queries.ListSnapshotSummaries.Handle(ctx, query.SnapshotSummariesQuery{
		Start:      in.Start.AsTime(),
		End:        in.End.AsTime(),
		DatabaseID: "",
		ServerID:   in.Server,
	})
	if err != nil {
		return nil, fmt.Errorf("listing snapshot summaries: %w", err)
	}
	protoSnaps := make([]*querierv2.SnapshotSummary, len(resp))
	for i, snap := range resp {
		protoSnaps[i] = SnapSummaryToProto(&snap)
	}

	span.SetAttributes(attribute.Int("response.summaries_count", len(protoSnaps)))
	return &querierv2.ListSnapshotSummariesResponse{SnapSummaries: protoSnaps}, nil
}

func (s GRPCServer) ListServerSummary(ctx context.Context, request *querierv2.ListServerSummaryRequest) (*querierv2.ListServerSummaryResponse, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("request.start", request.Start.AsTime().Format(time.RFC3339)),
		attribute.String("request.end", request.End.AsTime().Format(time.RFC3339)),
	)

	resp, err := s.app.Queries.ListServerSummary.Handle(ctx, request.Start.AsTime(), request.End.AsTime())
	if err != nil {
		return nil, err
	}
	protoServers := make([]*querierv2.ServerSummary, len(resp))
	for i, srv := range resp {
		protoServers[i] = &querierv2.ServerSummary{
			Name:                   srv.Name,
			Type:                   srv.Type,
			Connections:            int32(srv.Connections),
			RequestRate:            srv.RequestRate,
			ConnectionsByWaitGroup: srv.ConnsByWaitGroup,
		}
	}

	span.SetAttributes(attribute.Int("response.servers_count", len(protoServers)))
	return &querierv2.ListServerSummaryResponse{Servers: protoServers}, nil
}

func (s GRPCServer) GetSnapshot(ctx context.Context, request *querierv2.GetSnapshotRequest) (*querierv2.GetSnapshotResponse, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("request.id", request.Id),
	)

	snap, err := s.app.Queries.GetSnapshot.Handle(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &querierv2.GetSnapshotResponse{Snapshot: converters.DatabaseSnapshotToProto(&snap)}, nil
}

func SnapSummaryToProto(summary *common_domain.SnapshotSummary) *querierv2.SnapshotSummary {
	return &querierv2.SnapshotSummary{
		Id:                     summary.ID,
		Timestamp:              timestamppb.New(summary.Timestamp),
		Server:                 &dbmv1.ServerMetadata{Host: summary.Server.Host, Type: summary.Server.Type},
		ConnectionsByWaitEvent: summary.ConnsByWaitType,
		TimeMsByWaitEvent:      summary.TimeMsByWaitType,
		Connections:            int32(summary.Connections),
		Waiters:                int32(summary.Waiters),
		Blockers:               int32(summary.Blockers),
		WaitDuration:           summary.WaitDuration,
		AvgDuration:            summary.AvgDuration,
		MaxDuration:            summary.MaxDuration,
	}
}

func (s GRPCServer) GetSampleDetails(ctx context.Context, in *querierv2.GetSampleDetailsRequest) (*querierv2.GetSampleDetailsResponse, error) {

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("request.snap_id", in.GetSnapId()),
		attribute.String("request.sample_id", in.SampleId),
	)

	resp, err := s.app.Queries.GetQuerySampleDetails.Handle(ctx, in.GetSnapId(), in.SampleId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s GRPCServer) ListQueryMetrics(ctx context.Context, in *querierv2.ListQueryMetricsRequest) (*querierv2.ListQueryMetricsResponse, error) {
	resp, err := s.app.Queries.ListQueryMetrics.Handle(ctx, in.Start.AsTime(), in.End.AsTime(), in.Host)
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

	return &querierv2.ListQueryMetricsResponse{Metrics: ret}, nil
}

func (s GRPCServer) GetQueryMetricsTimeSeries(ctx context.Context, in *querierv2.GetQueryMetricsTimeSeriesRequest) (*querierv2.GetQueryMetricsTimeSeriesResponse, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("request.start", in.Start.AsTime().Format(time.RFC3339)),
		attribute.String("request.end", in.End.AsTime().Format(time.RFC3339)),
		attribute.String("request.host", in.Host),
		attribute.String("request.sql_handle", in.SqlHandle),
		attribute.String("request.interval", in.Interval),
	)
	interval, err := time.ParseDuration(in.Interval)
	if err != nil {
		return nil, err
	}
	ret, err := s.app.Queries.GetQueryMetricsSlice.Handle(ctx, in.Start.AsTime(), in.End.AsTime(), in.Host, in.SqlHandle, interval)
	if err != nil {
		return nil, err
	}

	protoM := make([]*dbmv1.QueryMetric, 0, len(ret))
	for _, metric := range ret {
		pm, err := converters.QueryMetricToProto(metric)
		if err != nil {
			return nil, err
		}
		if pm == nil {
			continue
		}
		protoM = append(protoM, pm)
	}
	return &querierv2.GetQueryMetricsTimeSeriesResponse{Metrics: protoM}, nil
}
