package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	grpcui "github.com/fullstorydev/grpcui/standalone"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
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
	"runtime/debug"
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

func StartCollector(cmd *cobra.Command, args []string) error {
	address := "localhost:7080"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("failed to listen on 8082: %w", err)
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
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses:     []string{"https://localhost:9200"},
		Username:      "elastic",
		Password:      "changeme",
		EnableMetrics: false,
		Transport:     &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	})
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
	ctx := context.Background()
	cc, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	h, err := grpcui.HandlerViaReflection(ctx, cc, "database monitoring collector")
	if err != nil {
		panic(err)
	}
	err2 := http.ListenAndServe(":7081", h)
	if err2 != nil {
		panic(err2)
	}
	return nil

}
