package telemetry

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"runtime/debug"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/propagators/b3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type GRPCClientConfig struct {
	Endpoint              string `toml:"endpoint" yaml:"endpoint"`
	MaxMsgSize            int    `toml:"max_msg_size" yaml:"max_msg_size"`
	TlsEnabled            bool   `toml:"tls_enabled" yaml:"tls_enabled"`
	TlsInsecureSkipVerify bool   `toml:"tls_insecure_skip_verify" yaml:"tls_insecure_skip_verify"`
}

func OpenInstrumentedClientConn(endpoint string, maxSize int, tlsEnabled bool) (*grpc.ClientConn, error) {
	return OpenInstrumentedClientConnWithTLS(endpoint, maxSize, tlsEnabled, false)
}

func OpenInstrumentedClientConnWithTLS(endpoint string, maxSize int, tlsEnabled bool, insecureSkipVerify bool) (*grpc.ClientConn, error) {
	// Shared dial options for both network and inproc
	baseOpts := []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
	if maxSize > 0 {
		baseOpts = append(baseOpts, grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxSize),
			grpc.MaxCallSendMsgSize(maxSize),
		))
	}
	var creds grpc.DialOption
	if tlsEnabled {
		creds = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: insecureSkipVerify, //nolint:gosec // opt-in for local self-signed certs
		}))
	} else {
		creds = grpc.WithTransportCredentials(insecure.NewCredentials())
	}
	opts := append([]grpc.DialOption{creds}, baseOpts...)
	return grpc.DialContext(context.Background(), endpoint, opts...)

}

func NewGrpcServer(maxMessageLength int, tlsEnabled bool, tlsCertPath string, tlsKeyPath string) *grpc.Server {
	recoveryOpt := grpc_recovery.WithRecoveryHandlerContext(func(ctx context.Context, r interface{}) error {
		fmt.Printf("gRPC panic recovered: %v\n%s\n", r, debug.Stack())
		return status.Error(codes.Internal, "internal server error")
	})

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		grpc_prometheus.UnaryServerInterceptor,
		grpc_recovery.UnaryServerInterceptor(recoveryOpt),
	}
	streamInterceptors := []grpc.StreamServerInterceptor{
		grpc_prometheus.StreamServerInterceptor,
		grpc_recovery.StreamServerInterceptor(recoveryOpt),
	}
	opts := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithPropagators(b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader | b3.B3SingleHeader))))),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryInterceptors...)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(streamInterceptors...)),
	}
	if maxMessageLength > 0 {
		opts = append(opts,
			grpc.MaxRecvMsgSize(maxMessageLength),
			grpc.MaxSendMsgSize(maxMessageLength),
		)
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
