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
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/adapters"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/ports"
	collectorv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1/collector"
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
	ctx := context.Background()
	collectorAddr := config.GRPCServerConfig.GrpcConfig.Url
	lis, err := net.Listen("tcp", collectorAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %s", collectorAddr, err)
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
		grpc.MaxRecvMsgSize(int(config.GRPCServerConfig.GrpcConfig.GrpcMessageMaxSize)),
		grpc.MaxSendMsgSize(int(config.GRPCServerConfig.GrpcConfig.GrpcMessageMaxSize)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryInterceptors...)),
	}
	grpcServer := grpc.NewServer(opts...)
	client, err := config.ELKConfig.Get(ctx)
	if err != nil {
		panic(err)
	}
	elk := adapters.NewELKRepository(client)
	application := app.NewApplication(elk, elk)
	svc := ports.NewIngestionSvc(*application)
	collectorv1.RegisterIngestionServiceServer(grpcServer, svc)
	reflection.Register(grpcServer)
	go func() {
		err = grpcServer.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()
	if config.GRPCServerConfig.GrpcUiConfig.Enabled {
		cc, err3 := grpc.NewClient(collectorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
