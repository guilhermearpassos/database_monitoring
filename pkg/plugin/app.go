package plugin

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/credentials"
	"net/http"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Make sure App implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. Plugin should not implement all these interfaces - only those which are
// required for a particular task.
var (
	_ backend.CallResourceHandler   = (*App)(nil)
	_ instancemgmt.InstanceDisposer = (*App)(nil)
	_ backend.CheckHealthHandler    = (*App)(nil)
	_ backend.QueryDataHandler      = (*App)(nil)
	_ backend.StreamHandler         = (*App)(nil)
)

// AppConfig holds the configuration for the app
type AppConfig struct {
	APIURL string `json:"apiUrl"`
}

// App is an example app plugin with a backend which can respond to data queries.
type App struct {
	backend.CallResourceHandler
	client        dbmv1.DBMApiClient
	supportClient dbmv1.DBMSupportApiClient
}

// NewApp creates a new example *App instance.
func NewApp(_ context.Context, settings backend.AppInstanceSettings) (instancemgmt.Instance, error) {
	var app App
	cfg := AppConfig{}
	// Parse the configuration from settings
	if err := json.Unmarshal(settings.JSONData, &cfg); err != nil {
		return nil, fmt.Errorf("missing or invalid JSON data in settings: %w", err)
	}
	port := strings.Split(cfg.APIURL, ":")
	if len(port) != 2 {
		return nil, fmt.Errorf("invalid API URL: %s", cfg.APIURL)
	}
	var client *grpc.ClientConn
	var err error
	if port[1] == "443" {
		client, err = grpc.NewClient(cfg.APIURL, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})))
	} else {
		client, err = grpc.NewClient(cfg.APIURL, grpc.WithTransportCredentials(insecure.NewCredentials()))

	}
	if err != nil {
		return nil, err
	}
	app.client = dbmv1.NewDBMApiClient(client)
	app.supportClient = dbmv1.NewDBMSupportApiClient(client)
	// Use a httpadapter (provided by the SDK) for resource calls. This allows us
	// to use a *http.ServeMux for resource calls, so we can map multiple routes
	// to CallResource without having to implement extra logic.
	mux := http.NewServeMux()
	app.registerRoutes(mux)
	app.CallResourceHandler = httpadapter.New(mux)

	return &app, nil
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created.
func (a *App) Dispose() {
	// cleanup
}

// CheckHealth handles health checks sent from Grafana to the plugin.
func (a *App) CheckHealth(_ context.Context, _ *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "ok",
	}, nil
}
