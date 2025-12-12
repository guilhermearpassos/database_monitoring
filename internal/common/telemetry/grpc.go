package telemetry

import (
	"context"
	"crypto/tls"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/propagators/b3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"runtime/debug"
)

func OpenInstrumentedClientConn(endpoint string, maxSize int, tlsEnabled bool) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxSize),
			grpc.MaxCallSendMsgSize(maxSize),
		)}
	if tlsEnabled {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})))
	}
	return grpc.NewClient(endpoint, opts...)

}

func NewGrpcServer(maxMessageLength int, tlsEnabled bool, tlsCertPath string, tlsKeyPath string) *grpc.Server {

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
	if tlsEnabled {
		serverCert, err := tls.LoadX509KeyPair(tlsCertPath, tlsKeyPath)
		if err != nil {
			log.Fatalf("failed to load TLS certificate and key: %s", err)
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{serverCert},
			ClientAuth:   tls.NoClientCert,
			MinVersion:   tls.VersionTLS12,
		})))
	}
	grpcServer := grpc.NewServer(opts...)
	return grpcServer

}
