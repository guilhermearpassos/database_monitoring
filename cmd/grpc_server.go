package main

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	grpcui "github.com/fullstorydev/grpcui/standalone"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	config2 "github.com/guilhermearpassos/database-monitoring/common/config"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/adapters"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/ports"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"os"
	"runtime/debug"
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
	lis, err := net.Listen("tcp", config.GRPCServerConfig.GrpcConfig.Url)
	if err != nil {
		log.Fatalf("failed to listen on %s: %s", config.GRPCServerConfig.GrpcConfig.Url, err)
	}
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		grpc_prometheus.UnaryServerInterceptor,
		grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandlerContext(func(ctx context.Context, r interface{}) error {
			panicMessage := fmt.Sprintf("%v", r)
			//span := trace.
			fmt.Printf("%s\n", debug.Stack())
			return fmt.Errorf(panicMessage)
		})),
	}
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryInterceptors...)),
	}
	grpcServer := grpc.NewServer(opts...)
	client, err := config.ELKConfig.Get(context.Background())
	if err != nil {
		panic(err)
	}
	elk := adapters.NewELKRepository(client)
	application := app.NewApplication(elk, elk)
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
		cc, err := grpc.NewClient(config.GRPCServerConfig.GrpcConfig.Url, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
