package plugin

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	querierv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/querier/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
)

var (
	_ backend.CallResourceHandler   = (*App)(nil)
	_ instancemgmt.InstanceDisposer = (*App)(nil)
	_ backend.CheckHealthHandler    = (*App)(nil)
	_ backend.QueryDataHandler      = (*App)(nil)
	_ backend.StreamHandler         = (*App)(nil)
)

// AppConfig holds the configuration for the app.
type AppConfig struct {
	APIURL                string `json:"apiUrl"`
	TLSInsecureSkipVerify bool   `json:"tlsInsecureSkipVerify"`
}

// App is an app plugin with a backend which can respond to data queries.
type App struct {
	backend.CallResourceHandler
	conn   *grpc.ClientConn
	client querierv2.QuerierAPIClient
}

// NewApp creates a new *App instance.
func NewApp(_ context.Context, settings backend.AppInstanceSettings) (instancemgmt.Instance, error) {
	var app App
	cfg := AppConfig{}
	if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
		return nil, fmt.Errorf("missing or invalid JSON data in settings: %w", err)
	}

	target := strings.TrimSpace(cfg.APIURL)
	useTLS := false
	if strings.HasPrefix(target, "https://") {
		useTLS = true
		target = strings.TrimPrefix(target, "https://")
	} else if strings.HasPrefix(target, "http://") {
		target = strings.TrimPrefix(target, "http://")
	}
	target = strings.TrimRight(target, "/")

	_, port, err := net.SplitHostPort(target)
	if err != nil {
		return nil, fmt.Errorf("invalid API URL %q (expected host:port): %w", cfg.APIURL, err)
	}
	if port == "443" {
		useTLS = true
	}

	var creds grpc.DialOption
	if useTLS {
		creds = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: cfg.TLSInsecureSkipVerify, //nolint:gosec // opt-in for self-signed certs
		}))
	} else {
		creds = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	conn, err := grpc.NewClient(
		target,
		creds,
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(100*1024*1024),
			grpc.MaxCallSendMsgSize(100*1024*1024),
		),
	)
	if err != nil {
		return nil, err
	}
	app.conn = conn
	app.client = querierv2.NewQuerierAPIClient(conn)

	mux := http.NewServeMux()
	app.registerRoutes(mux)
	app.CallResourceHandler = httpadapter.New(mux)

	return &app, nil
}

// Dispose cleans up the underlying gRPC client connection when the instance is disposed.
func (a *App) Dispose() {
	if a.conn != nil {
		_ = a.conn.Close()
	}
}

// CheckHealth verifies connectivity to the Sqlsights QuerierAPI gRPC endpoint.
func (a *App) CheckHealth(ctx context.Context, _ *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	_, err := a.client.ListServerSummary(healthCtx, &querierv2.ListServerSummaryRequest{
		Start: timestamppb.New(now.Add(-1 * time.Minute)),
		End:   timestamppb.New(now),
	})
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("failed to connect to Sqlsights API: %v", err),
		}, nil
	}
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "ok",
	}, nil
}
