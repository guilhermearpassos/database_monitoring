package main

import (
	"context"
	"fmt"
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
	"net/http"
	"os"
	"os/signal"
	"time"
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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	var config config2.AgentConfig
	// Check if file exists
	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		panic(fmt.Errorf("config file does not exist: %s", configFileName))
	}
	if _, err := toml.DecodeFile(configFileName, &config); err != nil {
		panic(fmt.Errorf("failed to parse config file: %s", err))
	}
	err := telemetry.InitTelemetryFromConfig(config.Telemetry)
	if err != nil {
		panic(fmt.Errorf("failed to init telemetry: %v", err))
	}
	cc, err := telemetry.OpenInstrumentedClientConn(config.CollectorConfig.Url, int(config.CollectorConfig.GrpcMessageMaxSize), config.CollectorConfig.TLS.Enabled)

	if err != nil {
		panic(err)
	}
	if config.Telemetry.Metrics.Enabled {
		go func() {
			promHost := config.Telemetry.Metrics.Host
			mux := http.NewServeMux()
			mux.Handle("/metrics", promhttp.Handler())
			fmt.Sprintf("serving metrics on %s", promHost)
			err2 := http.ListenAndServe(promHost, mux)
			if err2 != nil {
				panic(err2)
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
		db, err := telemetry.OpenInstrumentedDB(tgt.Driver, tgt.ConnString)
		if err != nil {
			panic(fmt.Errorf("error connecting to %s: %w", tgt.Alias, err))
		}
		dbByHost[tgt.Alias] = db
	}
	reader := adapters.NewSQLServerDataReader(dbByHost)
	router := events.NewEventRouter()
	go router.StartMetrics(ctx)
	a := app.NewApplication(reader, reader, adapters.NewGRPCIngestionClient(client), router)
	pf := event_processors.NewPlanFetcher(*a)
	mc := event_processors.NewPrometheusMetricsCollector()
	sp := event_processors.NewDefaultSQLParser()
	ld := event_processors.NewMetricsDetector(a, mc, sp)
	pf.Register(router)
	ld.Register(router)
	go pf.Run(ctx)
	go ld.Run(ctx)
	for _, tgt := range config.TargetHosts {
		startTarget(ctx, a, tgt, config.CollectMetrics)
	}
	<-ctx.Done()
	return nil
}
func startTarget(ctx context.Context, a *app.Application, config config2.DBDataCollectionConfig, collectMetrics bool) {

	serverMeta := common_domain.ServerMeta{
		Host: config.Alias,
		Type: config.Driver,
	}
	sc := background_agent.NewSnapshotCollector(*a)
	mc := background_agent.NewMetricsCollector(*a)
	go sc.Run(ctx, serverMeta, 10*time.Second)
	if collectMetrics {
		go mc.Run(ctx, serverMeta, 1*time.Minute)
	}
}
