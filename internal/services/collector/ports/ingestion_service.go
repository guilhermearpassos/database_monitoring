package ports

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/guilhermearpassos/database-monitoring/internal/common/custom_errors"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/app/command"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	collectorv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1/collector"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IngestionSvc struct {
	*collectorv1.UnimplementedIngestionServiceServer
	*collectorv1.UnimplementedCollectorSyncServiceServer
	agents map[string]*domain.AgentConfig
	app    app.Application
}

func NewIngestionSvc(app app.Application) *IngestionSvc {
	return &IngestionSvc{agents: make(map[string]*domain.AgentConfig), app: app}
}

func (s IngestionSvc) RegisterAgent(ctx context.Context, request *collectorv1.RegisterAgentRequest) (*collectorv1.RegisterAgentResponse, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("request.target_host", request.TargetHost),
		attribute.String("request.target_type", request.TargetType),
		attribute.String("request.agent_version", request.AgentVersion),
	)

	s.agents[request.TargetHost] = &domain.AgentConfig{
		ID:           uuid.NewString(),
		TargetHost:   request.TargetHost,
		TargetType:   request.TargetType,
		AgentVersion: request.AgentVersion,
		Tags:         nil,
	}
	return &collectorv1.RegisterAgentResponse{}, nil
}

func (s IngestionSvc) IngestMetrics(ctx context.Context, metrics *collectorv1.DatabaseMetrics) (*collectorv1.IngestMetricsResponse, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("request.timestamp", metrics.Timestamp.AsTime().Format(time.RFC3339)),
		attribute.String("request.server.host", metrics.Server.Host),
		attribute.String("request.server.type", metrics.Server.Type),
		attribute.Int("request.metrics_count", len(metrics.GetQueryMetrics().QueryMetrics)),
	)

	timestamp := metrics.Timestamp.AsTime()
	domainMetrics := make([]*common_domain.QueryMetric, len(metrics.GetQueryMetrics().QueryMetrics))
	for i, m := range metrics.GetQueryMetrics().QueryMetrics {
		domainMetric, err := converters.QueryMetricToDomain(m)
		if err != nil {
			return nil, err
		}
		domainMetric.CollectionTime = timestamp
		domainMetrics[i] = domainMetric
	}
	err := s.app.Commands.StoreQueryMetrics.Handle(ctx, domainMetrics, common_domain.ServerMeta{
		Host: metrics.Server.Host,
		Type: metrics.Server.Type,
	}, timestamp)
	if err != nil {
		return nil, err
	}
	return &collectorv1.IngestMetricsResponse{
		Success: true,
		Message: "",
	}, nil
}

func (s IngestionSvc) IngestSnapshot(ctx context.Context, request *collectorv1.IngestSnapshotRequest) (*collectorv1.IngestSnapshotResponse, error) {
	span := trace.SpanFromContext(ctx)
	snapshot := request.GetSnapshot()
	span.SetAttributes(
		attribute.String("request.snapshot.id", snapshot.Id),
		attribute.String("request.snapshot.server_host", snapshot.Server.Host),
		attribute.String("request.snapshot.server_type", snapshot.Server.Type),
		attribute.Int("request.snapshot.samples_count", len(snapshot.Samples)),
	)

	domain_snap := converters.DatabaseSnapshotToDomain(snapshot)
	err := s.app.Commands.StoreSnapshot.Handle(ctx, domain_snap)
	if err != nil {
		return nil, err
	}
	return &collectorv1.IngestSnapshotResponse{}, nil
}

func (s IngestionSvc) IngestSnapshotSamples(ctx context.Context, request *collectorv1.IngestSnapshotSamplesRequest) (*collectorv1.IngestSnapshotSamplesResponse, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("request.id", request.GetId()),
		attribute.Int("request.samples_count", len(request.GetSamples())),
	)

	samples := make([]*common_domain.QuerySample, len(request.GetSamples()))
	for i, sample := range request.GetSamples() {
		samples[i] = converters.SampleToDomain(sample)
	}
	err := s.app.Commands.StoreSnapshotSamples.Handle(ctx, request.GetId(), samples)
	if err != nil {
		return nil, err
	}
	return &collectorv1.IngestSnapshotSamplesResponse{}, nil
}

func (s IngestionSvc) IngestExecutionPlans(ctx context.Context, in *collectorv1.IngestExecutionPlansRequest) (*collectorv1.IngestExecutionPlansResponse, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int("request.plans_count", len(in.GetPlans())),
	)

	domainPlans := make([]*common_domain.ExecutionPlan, len(in.GetPlans()))
	for i, plan := range in.GetPlans() {
		protoPlan, err := converters.ExecutionPlanToDomain(plan)
		if err != nil {
			return nil, err
		}
		domainPlans[i] = protoPlan
	}
	err := s.app.Commands.StoreExecutionPlans.Handle(ctx, domainPlans)
	if err != nil {
		return nil, err
	}
	return &collectorv1.IngestExecutionPlansResponse{}, nil
}

func (s IngestionSvc) GetKnownPlanHandles(ctx context.Context, in *collectorv1.GetKnownPlanHandlesRequest) (*collectorv1.GetKnownPlanHandlesResponse, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("host", in.Server.Host),
		attribute.Int("page_size", int(in.PageSize)),
		attribute.Int("page_number", int(in.PageNumber)))
	ret, totalPages, err := s.app.Queries.GetKnownPlanHandlesHandler.Handle(ctx, &common_domain.ServerMeta{
		Host: in.GetServer().Host,
		Type: in.GetServer().Type,
	}, int(in.PageNumber), int(in.PageSize))
	if err != nil {
		if errors.As(err, &custom_errors.NotFoundErr{}) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, err
	}
	return &collectorv1.GetKnownPlanHandlesResponse{
		Handles:    ret,
		PageNumber: in.PageNumber,
		PageSize:   in.PageSize,
		TotalPages: int32(totalPages),
	}, nil
}

func (s IngestionSvc) IngestWarnings(ctx context.Context, in *collectorv1.IngestWarningsRequest) (*collectorv1.IngestWarningsResponse, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int("request.warnings_count", len(in.GetWarnings())),
	)
	domainWarnings := make([]*common_domain.Warning, len(in.GetWarnings()))
	for i, warning := range in.GetWarnings() {
		domainWarnings[i] = common_domain.NewWarning(warning)
	}
	err := s.app.Commands.StoreWarnings.Handle(ctx, command.StoreWarnings{
		Warnings: domainWarnings,
		ServerMeta: common_domain.ServerMeta{
			Host: in.GetServer().GetHost(),
			Type: in.GetServer().GetType(),
		},
	})
	if err != nil {
		return nil, err
	}
	return &collectorv1.IngestWarningsResponse{}, nil
}

func (s IngestionSvc) GetKnownWarnings(ctx context.Context, in *collectorv1.GetKnownWarningsRequest) (*collectorv1.GetKnownWarningsResponse, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("request.host", in.GetServer().GetHost()),
		attribute.Int("request.page_size", int(in.GetPageSize())),
		attribute.Int("request.page_number", int(in.GetPageNumber())),
	)
	warns, err := s.app.Queries.GetKnownWarnings.Handle(ctx, in.GetServer().GetHost(), int(in.GetPageSize()), int(in.GetPageNumber()))
	if err != nil {
		return nil, err
	}
	protoW := make([]*dbmv1.Warning, len(warns))
	for i, w := range warns {
		protoW[i] = w.WarningData
	}
	return &collectorv1.GetKnownWarningsResponse{Warnings: protoW}, nil
}
