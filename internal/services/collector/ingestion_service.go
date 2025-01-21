package collector

import (
	"context"
	"github.com/google/uuid"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	collectorv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1/collector"
)

type IngestionSvc struct {
	*collectorv1.UnimplementedIngestionServiceServer
	agents map[string]*domain.AgentConfig
	app    app.Application
}

func NewIngestionSvc(app app.Application) *IngestionSvc {
	return &IngestionSvc{agents: make(map[string]*domain.AgentConfig), app: app}
}

func (i IngestionSvc) RegisterAgent(ctx context.Context, request *collectorv1.RegisterAgentRequest) (*collectorv1.RegisterAgentResponse, error) {
	i.agents[request.TargetHost] = &domain.AgentConfig{
		ID:           uuid.NewString(),
		TargetHost:   request.TargetHost,
		TargetType:   request.TargetType,
		AgentVersion: request.AgentVersion,
		Tags:         nil,
	}
	return &collectorv1.RegisterAgentResponse{}, nil
}

func (i IngestionSvc) IngestMetrics(ctx context.Context, metrics *collectorv1.DatabaseMetrics) (*collectorv1.IngestMetricsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (i IngestionSvc) IngestSnapshot(ctx context.Context, request *collectorv1.IngestSnapshotRequest) (*collectorv1.IngestSnapshotResponse, error) {

	domain_snap := converters.DatabaseSnapshotToDomain(request.GetSnapshot())
	err := i.app.Commands.StoreSnapshot.Handle(ctx, domain_snap)
	if err != nil {
		return nil, err
	}
	return &collectorv1.IngestSnapshotResponse{
		Success: true,
		Message: "",
	}, nil
}
