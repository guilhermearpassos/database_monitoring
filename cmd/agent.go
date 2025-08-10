package main

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	config2 "github.com/guilhermearpassos/database-monitoring/internal/common/config"
	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/adapters"
	_ "github.com/guilhermearpassos/database-monitoring/internal/services/agent/adapters/metrics"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	collectorv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1/collector"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	"os"
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
	ctx := context.Background()
	tracer := otel.Tracer("agent")
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
	cc, err := telemetry.OpenInstrumentedClientConn(config.CollectorConfig.Url, int(config.CollectorConfig.GrpcMessageMaxSize))

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
	for _, tgt := range config.TargetHosts {
		startTarget(ctx, tgt, client, tracer)
	}
	for {
		time.Sleep(5 * time.Second)
	}
	return nil
}
func startTarget(ctx context.Context, config config2.DBDataCollectionConfig, collectorClient collectorv1.IngestionServiceClient, tracer trace.Tracer) {
	db, err := telemetry.OpenInstrumentedDB(config.Driver, config.ConnString)
	if err != nil {
		panic(fmt.Errorf("error connecting to %s: %w", config.Alias, err))
	}
	knownHandles, err := collectorClient.GetKnownPlanHandles(context.Background(), &collectorv1.GetKnownPlanHandlesRequest{Server: &dbmv1.ServerMetadata{
		Host: config.Alias,
		Type: config.Driver,
	}})
	var knownHandlesSlice []string
	if err != nil {
		if grpcErr, ok := status.FromError(err); ok {
			if grpcErr.Code() == codes.NotFound {
				knownHandlesSlice = []string{}
			} else {
				panic(fmt.Errorf("error getting known plan handles for %s: %w", config.Alias, err))
			}

		} else {

			panic(fmt.Errorf("error getting known plan handles for %s: %w", config.Alias, err))
		}
	} else {
		knownHandlesSlice = make([]string, len(knownHandles.Handles))
		for idx, data := range knownHandles.Handles {
			knownHandlesSlice[idx] = string(data)
		}
	}

	serverMeta := common_domain.ServerMeta{
		Host: config.Alias,
		Type: config.Driver,
	}
	dataReader := adapters.NewSQLServerDataReader(db, serverMeta, knownHandlesSlice)
	metricsSnapshotProcessor := adapters.NewSnapshotMetricsProcessor(600)
	go collectSnapshots(dataReader, collectorClient, metricsSnapshotProcessor, tracer)
	go collectQueryMetrics(dataReader, collectorClient, serverMeta, tracer)
	go metricsSnapshotProcessor.Run(ctx)
}

func collectQueryMetrics(reader domain.QueryMetricsReader, client collectorv1.IngestionServiceClient, serverName common_domain.ServerMeta, tracer trace.Tracer) {
	for {
		ctx, span := tracer.Start(context.Background(), "QueryMetrics")

		sampleTime := time.Now()
		metrics, err := reader.CollectMetrics(ctx)
		if err != nil {
			span.End()
			panic(err)
		}
		protoMetrics := make([]*dbmv1.QueryMetric, len(metrics))

		for i, m := range metrics {
			protoMetrics[i], err = converters.QueryMetricToProto(m)
			if err != nil {
				span.End()
				panic(err)
			}
		}
		_, err = client.IngestMetrics(ctx, &collectorv1.DatabaseMetrics{
			Server:    &dbmv1.ServerMetadata{Host: serverName.Host, Type: serverName.Type},
			Timestamp: timestamppb.New(sampleTime),
			Metrics:   &collectorv1.DatabaseMetrics_QueryMetrics{QueryMetrics: &collectorv1.DatabaseMetrics_QueryMetricSample{QueryMetrics: protoMetrics}},
		})
		if err != nil {
			span.End()
			panic(err)
		}
		span.End()

		select {
		case <-time.After(time.Until(sampleTime.Add(1 * time.Minute))):
			break

		}
	}
}

func collectSnapshots(dataReader domain.SamplesReader, client collectorv1.IngestionServiceClient, metricsSnapshotProcessor domain.MetricsProcessor, tracer trace.Tracer) {

	for {
		ctx, span := tracer.Start(context.Background(), "CollectSnapshots")
		var snapshots []*common_domain.DataBaseSnapshot
		snapshots, err := dataReader.TakeSnapshot(ctx)
		if err != nil {
			panic(err)
			span.End()
		}
		planHandleStrings := make(map[string]struct{})
		for _, snapshot := range snapshots {
			metricsSnapshotProcessor.QueueSnapshot(snapshot)
			if snapshot == nil {
				continue
			}
			_, err = client.IngestSnapshot(ctx, &collectorv1.IngestSnapshotRequest{
				Snapshot: converters.DatabaseSnapshotToProto(snapshot),
			})
			if err != nil {
				panic(err)
				span.End()
			}
			for _, qs := range snapshot.Samples {
				if qs.PlanHandle == nil {
					continue
				}
				planHandleStrings[string(qs.PlanHandle)] = struct{}{}
			}

		}
		planHandles := make([][]byte, 0)
		for k := range planHandleStrings {
			planHandles = append(planHandles, []byte(k))
		}
		executionPlans, err := dataReader.GetPlanHandles(ctx, planHandles, true)
		if err != nil {
			panic(err)
			span.End()
		}
		protoPlans := make([]*dbmv1.ExecutionPlan, 0, len(executionPlans))
		for _, p := range executionPlans {
			protoPlan, err2 := converters.ExecutionPlanToProto(p)
			if err2 != nil {
				panic(err2)
				span.End()
			}
			protoPlans = append(protoPlans, protoPlan)
		}
		if len(protoPlans) != 0 {

			_, err = client.IngestExecutionPlans(ctx, &collectorv1.IngestExecutionPlansRequest{Plans: protoPlans})
			if err != nil {
				panic(err)
				span.End()
			}
		}
		span.End()
		time.Sleep(10 * time.Second)
	}
}
