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
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/ports"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

var (
	GrpcCmd = &cobra.Command{
		Use:     "grpc",
		Short:   "run grpc server",
		Long:    "run grpc server",
		Aliases: []string{"server"},
		Example: "dbm grpc",
		RunE:    StartGrpc,
	}
)

func init() {
	GrpcCmd.Flags().StringVar(&configFileName, "config", "local/grpc.toml", "--config=local/grpc.toml")
}

func StartGrpc(cmd *cobra.Command, args []string) error {
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
	lis, err := net.Listen("tcp", config.GRPCServerConfig.GrpcConfig.Url)
	if err != nil {
		log.Fatalf("failed to listen on %s: %s", config.GRPCServerConfig.GrpcConfig.Url, err)
	}
	grpcServer := telemetry.NewGrpcServer(int(config.GRPCServerConfig.GrpcConfig.GrpcMessageMaxSize), config.GRPCServerConfig.GrpcConfig.TLS.Enabled, config.GRPCServerConfig.GrpcConfig.TLS.CertFile, config.GRPCServerConfig.GrpcConfig.TLS.KeyFile)
	db, err := config.PostgresConfig.Get(context.Background())
	if err != nil {
		panic(err)
	}
	elk := adapters.NewPostgresRepo(db)
	application := app.NewApplication(elk, elk, elk)
	server := ports.NewGRPCServer(application)
	dbmv1.RegisterDBMApiServer(grpcServer, server)
	dbmv1.RegisterDBMSupportApiServer(grpcServer, server)
	reflection.Register(grpcServer)
	go func() {
		err = grpcServer.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()
	ctx := context.Background()
	if config.GRPCServerConfig.GrpcUiConfig.Enabled {

		cc, err := telemetry.OpenInstrumentedClientConn(config.GRPCServerConfig.GrpcConfig.Url, int(config.GRPCServerConfig.GrpcConfig.GrpcMessageMaxSize), config.GRPCServerConfig.GrpcConfig.TLS.Enabled)
		if err != nil {
			panic(err)
		}
		h, err := grpcui.HandlerViaReflection(ctx, cc, "database monitoring")
		if err != nil {
			panic(err)
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
