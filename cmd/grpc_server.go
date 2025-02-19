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
	"runtime/debug"
)

var (
	grpcServerAddr   string
	grpcServeruiAddr string
	GrpcCmd          = &cobra.Command{
		Use:     "grpc",
		Short:   "run grpc server",
		Long:    "run grpc server",
		Aliases: []string{"server"},
		Example: "dbm grpc",
		RunE:    StartGrpc,
	}
)

func init() {
	GrpcCmd.Flags().StringVar(&grpcServerAddr, "grpc-addr", "", "")
	GrpcCmd.Flags().StringVar(&grpcServeruiAddr, "grpcui-addr", "", "")
	GrpcCmd.Flags().StringVar(&elkAddr, "elk-addr", "", "")
	GrpcCmd.Flags().StringVar(&elkUsername, "elk-user", "", "")
	GrpcCmd.Flags().StringVar(&elkPassword, "elk-pwd", "", "")
}

func StartGrpc(cmd *cobra.Command, args []string) error {
	lis, err := net.Listen("tcp", grpcServerAddr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %s", grpcuiAddr, err)
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
		Addresses:     []string{elkAddr},
		Username:      elkUsername,
		Password:      elkPassword,
		EnableMetrics: false,
		Transport:     &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	})
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
	cc, err := grpc.NewClient(grpcServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	h, err := grpcui.HandlerViaReflection(ctx, cc, "database monitoring")
	if err != nil {
		panic(err)
	}
	err2 := http.ListenAndServe(grpcServeruiAddr, h)
	if err2 != nil {
		panic(err2)
	}
	return nil
}
