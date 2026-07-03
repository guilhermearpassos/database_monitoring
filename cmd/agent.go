package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/BurntSushi/toml"
	config2 "github.com/guilhermearpassos/database-monitoring/internal/common/config"
	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/adapters"
	_ "github.com/guilhermearpassos/database-monitoring/internal/services/agent/adapters/metrics"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain/events"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/ports/background_agent"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/ports/event_processors"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	collectorv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1/collector"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var (
	AgentCmd = &cobra.Command{
		Use:     "agent",
		Short:   "run dbm agent",
		Long:    "run dbm agent",
		Aliases: []string{},
		Example: "dbm agent --config=local/agent.toml",
		RunE:    StartAgent,
	}
)

func init() {
	AgentCmd.Flags().StringVar(&configFileName, "config", "local/agent.toml", "--config=local/agent.toml")
}

func StartAgent(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	var config config2.AgentConfig
	// Check if file exists
	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", configFileName)
	}
	if _, err := toml.DecodeFile(configFileName, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	err := telemetry.InitTelemetryFromConfig(config.Telemetry)
	if err != nil {
		return fmt.Errorf("failed to init telemetry: %w", err)
	}
	telemetry.Info(ctx, "telemetry initialized", "otlp_endpoint", config.Telemetry.OTLP.Endpoint)
	cc, err := telemetry.OpenInstrumentedClientConn(config.CollectorConfig.Url, int(config.CollectorConfig.GrpcMessageMaxSize), config.CollectorConfig.TLS.Enabled)
	if err != nil {
		return fmt.Errorf("open collector client: %w", err)
	}
	// Prometheus server with graceful shutdown
	var promServer *http.Server
	if config.Telemetry.Metrics.Enabled {
		promHost := config.Telemetry.Metrics.Host
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		promServer = &http.Server{Addr: promHost, Handler: mux}
		go func() {
			telemetry.Info(ctx, "serving prometheus metrics", "host", promHost)
			if err2 := promServer.ListenAndServe(); err2 != nil && err2 != http.ErrServerClosed {
				telemetry.Error(ctx, err2, "prometheus metrics server failed", "host", promHost)
			}
		}()
	}
	client := collectorv1.NewIngestionServiceClient(cc)
	GetPlanPageSize := int32(config.GetKnownPlanPageSize)
	if GetPlanPageSize == 0 {
		GetPlanPageSize = 100
	}
	dbByHost := make(map[string]*sqlx.DB, len(config.TargetHosts))
	for _, tgt := range config.TargetHosts {
		telemetry.Info(ctx, "opening instrumented DB", "alias", tgt.Alias, "driver", tgt.Driver)
		db, err := telemetry.OpenInstrumentedDB(tgt.Driver, tgt.ConnString)
		if err != nil {
			telemetry.Error(ctx, err, "db connect failed", "alias", tgt.Alias, "driver", tgt.Driver)
			return fmt.Errorf("error connecting to %s: %w", tgt.Alias, err)
		}
		telemetry.Info(ctx, "db connected", "alias", tgt.Alias, "driver", tgt.Driver)
		dbByHost[tgt.Alias] = db
	}
	reader := adapters.NewSQLServerDataReader(dbByHost)
	for _, tgt := range config.TargetHosts {
		router := events.NewEventRouter(tgt.Alias)
		go router.StartMetrics(ctx)
		a := app.NewApplication(reader, reader, adapters.NewGRPCIngestionClient(client), router)
		pf := event_processors.NewPlanFetcher(*a)
		mc := event_processors.NewPrometheusMetricsCollector()
		sp := event_processors.NewDefaultSQLParser()
		ld := event_processors.NewMetricsDetector(a, mc, sp)
		pf.Register(router)
		ld.Register(router)
		go pf.Run()
		go ld.Run()
		startTarget(ctx, a, tgt, config.CollectMetrics, config.Databases)
	}
	<-ctx.Done()
	// Begin graceful shutdown
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if promServer != nil {
		_ = promServer.Shutdown(shutdownCtx)
	}
	for alias, db := range dbByHost {
		if db != nil {
			if err := db.Close(); err != nil {
				telemetry.Error(shutdownCtx, err, "db close error", "alias", alias)
			}
		}
	}
	if cc != nil {
		_ = cc.Close()
	}
	_ = telemetry.Shutdown(shutdownCtx)
	return nil
}
func startTarget(ctx context.Context, a *app.Application, config config2.DBDataCollectionConfig, collectMetrics bool, databases []string) {

	serverMeta := common_domain.ServerMeta{
		Host: config.Alias,
		Type: config.Driver,
	}
	sc := background_agent.NewSnapshotCollector(*a)
	mc := background_agent.NewMetricsCollector(*a)
	go sc.Run(ctx, serverMeta, databases, 10*time.Second)
	if collectMetrics {
		go mc.Run(ctx, serverMeta, databases, 1*time.Minute)
	}
}
