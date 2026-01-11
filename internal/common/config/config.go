package config

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/jmoiron/sqlx"
)

type AgentConfig struct {
	CollectorConfig      GrpcConfig                `toml:"collector"`
	TargetHosts          []DBDataCollectionConfig  `toml:"target_hosts"`
	MaxSamplesBatchSize  int                       `toml:"max_samples_batch_size"`
	GetKnownPlanPageSize int                       `toml:"get_known_plan_page_size"`
	Databases            []string                  `toml:"databases"`
	Telemetry            telemetry.TelemetryConfig `toml:"telemetry"`
	CollectMetrics       bool                      `toml:"collect_metrics"`
}

type GrpcConfig struct {
	Url                string    `toml:"url"`
	GrpcMessageMaxSize int64     `toml:"grpc_message_max_size"`
	TLS                TLSConfig `toml:"tls"`
}

type TLSConfig struct {
	Enabled  bool   `toml:"enabled"`
	CertFile string `toml:"cert_file"`
	KeyFile  string `toml:"key_file"`
}

type DBDataCollectionConfig struct {
	Alias      string `toml:"alias"`
	Driver     string `toml:"driver"`
	ConnString string `toml:"conn_string"`
}

type GRPCServerConfig struct {
	GrpcConfig   GrpcConfig `toml:"grpc"`
	GrpcUiConfig struct {
		Enabled bool   `toml:"enabled"`
		Url     string `toml:"url"`
	} `toml:"grpc_ui"`
}

type ELKConfig struct {
	Addr     string `toml:"address"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

type PostgresConfig struct {
	Connstring string `toml:"connstring"`
}

func (c PostgresConfig) Get(ctx context.Context) (*sqlx.DB, error) {

	return telemetry.OpenInstrumentedDB("postgres", c.Connstring)
}

func (c ELKConfig) Get(ctx context.Context) (*elasticsearch.Client, error) {
	return elasticsearch.NewClient(elasticsearch.Config{
		Addresses:     []string{c.Addr},
		Username:      c.User,
		Password:      c.Password,
		EnableMetrics: false,
		Transport:     &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	})
}

type CollectorConfig struct {
	GRPCServerConfig GRPCServerConfig          `toml:"grpc_server"`
	PostgresConfig   PostgresConfig            `toml:"postgres"`
	Telemetry        telemetry.TelemetryConfig `toml:"telemetry"`
	PurgeConfig      PurgeConfig               `toml:"purge"`
}

type PurgeConfig struct {
	Enabled  bool          `toml:"enabled"`
	MaxAge   time.Duration `toml:"max_age"`
	Interval time.Duration `toml:"interval"`
}

type GrpcAPIConfig struct {
	GRPCServerConfig GRPCServerConfig `toml:"grpc_server"`
	ELK              ELKConfig        `toml:"elk"`
}
