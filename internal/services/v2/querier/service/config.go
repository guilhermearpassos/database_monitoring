package service

import (
	"context"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/guilhermearpassos/database-monitoring/internal/appcommon"
	"github.com/guilhermearpassos/database-monitoring/internal/common/config"
	"github.com/guilhermearpassos/database-monitoring/internal/runtimes"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/querier/adapters/repository"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/querier/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/querier/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/querier/ports"
	querierv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/querier/v2"
)

// IngesterConfig mirrors the style of the v2 agent service configuration.
// It provides minimal knobs for the ingestion server application layer.
// Adapter wiring (e.g., gRPC) is intentionally left out for now.
type QuerierConfig struct {
	Enabled  bool                  `yaml:"enabled" toml:"enabled"`
	Postgres config.PostgresConfig `yaml:"postgres" toml:"postgres"`
}

// GetService constructs the transport-agnostic ports.GrpcIngester using
// in-memory/default adapters so the service can run without external deps.
func (c *QuerierConfig) GetService(ctx context.Context, inproc *inprocgrpc.Channel) (appcommon.Service, error) { //nolint:revive,unused
	var repo domain.SampleRepository = repository.NewNoopSampleRepo()
	if c.Postgres.Connstring != "" {
		db, err := c.Postgres.Get(ctx)
		if err != nil {
			return nil, err
		}
		repo = repository.NewPostgresRepo(db)
	}
	application := app.NewApplication(repo)
	p := ports.NewGRPCServer(application)
	return QuerierService{Port: &p}, nil
}

// QuerierService conforms to appcommon.Service and is the entry point the
// bootstrap will register. For now, it does not bind any transport.
type QuerierService struct {
	Port *ports.GRPCServer
}

func (s QuerierService) Register(grpcRuntime *runtimes.GRPCServerRuntime, TaskRuntime *runtimes.BackGroundTaskRuntime) error {
	// No transport adapter yet; nothing to register.
	// The Port is ready to be bound by a future gRPC adapter and/or inproc hub.
	grpcRuntime.RegisterService(&querierv2.QuerierAPI_ServiceDesc, s.Port)
	return nil
}

var _ appcommon.Service = (*QuerierService)(nil)
