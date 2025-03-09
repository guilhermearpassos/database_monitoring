package telemetry

import (
	"context"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/propagators/b3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"runtime/debug"
)

func OpenInstrumentedClientConn(endpoint string, maxSize int) (*grpc.ClientConn, error) {

	return grpc.NewClient(endpoint,
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxSize),
			grpc.MaxCallSendMsgSize(maxSize),
		))

}

func NewGrpcServer(maxMessageLength int) *grpc.Server {

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
		grpc.MaxRecvMsgSize(maxMessageLength),
		grpc.MaxSendMsgSize(maxMessageLength),
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithPropagators(b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader | b3.B3SingleHeader))))),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryInterceptors...)),
	}
	grpcServer := grpc.NewServer(opts...)
	return grpcServer

}
