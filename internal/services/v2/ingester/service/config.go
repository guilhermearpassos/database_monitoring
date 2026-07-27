package service

import (
	"context"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/guilhermearpassos/database-monitoring/internal/appcommon"
	"github.com/guilhermearpassos/database-monitoring/internal/runtimes"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/ports"
	ingestorv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/ingestor/v2"
)

// IngesterConfig mirrors the style of the v2 agent service configuration.
// It provides minimal knobs for the ingestion server application layer.
// Adapter wiring (e.g., gRPC) is intentionally left out for now.
type IngesterConfig struct {
	Enabled bool `yaml:"enabled" toml:"enabled"`
}

// GetService constructs the transport-agnostic ports.GrpcIngester using
// in-memory/default adapters so the service can run without external deps.
func (c *IngesterConfig) GetService(ctx context.Context, inproc *inprocgrpc.Channel) (appcommon.Service, error) { //nolint:revive,unused
	application := app.NewApplication()
	p := ports.NewService(application)
	return IngesterService{Port: p}, nil
}

// IngesterService conforms to appcommon.Service and is the entry point the
// bootstrap will register. For now, it does not bind any transport.
type IngesterService struct {
	Port *ports.GrpcIngester
}

func (s IngesterService) Register(grpcRuntime *runtimes.GRPCServerRuntime, TaskRuntime *runtimes.BackGroundTaskRuntime) error {
	// No transport adapter yet; nothing to register.
	// The Port is ready to be bound by a future gRPC adapter and/or inproc hub.
	grpcRuntime.RegisterService(&ingestorv2.IngestionService_ServiceDesc, s.Port)
	return nil
}

var _ appcommon.Service = (*IngesterService)(nil)
