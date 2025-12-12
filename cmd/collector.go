package main

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	grpcui "github.com/fullstorydev/grpcui/standalone"
	config2 "github.com/guilhermearpassos/database-monitoring/internal/common/config"
	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/adapters"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/app/command"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/ports"
	collectorv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1/collector"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	CollectorCmd = &cobra.Command{
		Use:     "collector",
		Short:   "run dbm collector",
		Long:    "run dbm collector",
		Aliases: []string{},
		Example: "dbm collector",
		RunE:    StartCollector,
	}
)

func init() {
	CollectorCmd.Flags().StringVar(&configFileName, "config", "local/collector.toml", "--config=local/collector.toml")
}

func StartCollector(cmd *cobra.Command, args []string) error {
	var config config2.CollectorConfig
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
	ctx := context.Background()
	collectorAddr := config.GRPCServerConfig.GrpcConfig.Url
	lis, err := net.Listen("tcp", collectorAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %s", collectorAddr, err)
	}
	grpcServer := telemetry.NewGrpcServer(int(config.GRPCServerConfig.GrpcConfig.GrpcMessageMaxSize), config.GRPCServerConfig.GrpcConfig.TLS.Enabled, config.GRPCServerConfig.GrpcConfig.TLS.CertFile, config.GRPCServerConfig.GrpcConfig.TLS.KeyFile)
	db, err := config.PostgresConfig.Get(ctx)
	if err != nil {
		panic(err)
	}
	repo := adapters.NewPostgresRepo(db)
	application := app.NewApplication(repo, repo, repo)
	svc := ports.NewIngestionSvc(*application)
	collectorv1.RegisterIngestionServiceServer(grpcServer, svc)
	reflection.Register(grpcServer)
	go func() {
		err = grpcServer.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()
	if config.PurgeConfig.Enabled {
		maxAge := config.PurgeConfig.MaxAge
		interval := config.PurgeConfig.Interval
		go func() {
			for {
				keepUntil := time.Now().Add(-maxAge)
				wg := sync.WaitGroup{}
				wg.Add(2)
				go func() {
					defer wg.Done()
					errM := application.Commands.PurgeQueryMetrics.Handle(context.Background(), command.PurgeQueryMetrics{
						Start:     time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
						End:       keepUntil,
						BatchSize: 1000,
					})
					if errM != nil {
						log.Println(errM)
						panic(errM)
					}
				}()
				go func() {
					defer wg.Done()
					errM := application.Commands.PurgeSnapshots.Handle(context.Background(), command.PurgeSnapshots{
						Start:     time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
						End:       keepUntil,
						BatchSize: 100,
					})
					if errM != nil {
						log.Println(errM)
						panic(errM)
					}
				}()
				//go func() {
				//	defer wg.Done()
				//	errM := application.Commands.PurgeQueryPlans.Handle(context.Background(), 1000)
				//	if errM != nil {
				//		log.Println(errM)
				//		panic(errM)
				//	}
				//}()
				wg.Wait()
				time.Sleep(interval)
			}
		}()
	}
	if config.GRPCServerConfig.GrpcUiConfig.Enabled {

		cc, err3 := telemetry.OpenInstrumentedClientConn(collectorAddr, int(config.GRPCServerConfig.GrpcConfig.GrpcMessageMaxSize), config.GRPCServerConfig.GrpcConfig.TLS.Enabled)
		if err3 != nil {
			panic(err3)
		}
		h, err3 := grpcui.HandlerViaReflection(ctx, cc, "database monitoring collector")
		if err3 != nil {
			panic(err3)
		}
		err2 := http.ListenAndServe(config.GRPCServerConfig.GrpcUiConfig.Url, h)
		if err2 != nil {
			panic(err2)
		}
	}
	for {
		time.Sleep(10 * time.Second)
	}
	return nil

}
