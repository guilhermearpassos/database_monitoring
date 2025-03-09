package config

import (
	"context"
	"crypto/tls"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"net/http"
)

type AgentConfig struct {
	CollectorConfig     GrpcConfig                `toml:"collector"`
	TargetHosts         []DBDataCollectionConfig  `toml:"target_hosts"`
	MaxSamplesBatchSize int                       `toml:"max_samples_batch_size"`
	Telemetry           telemetry.TelemetryConfig `toml:"telemetry"`
}

type GrpcConfig struct {
	Url                string `toml:"url"`
	GrpcMessageMaxSize int64  `toml:"grpc_message_max_size"`
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
	ELKConfig        ELKConfig                 `toml:"elk"`
	Telemetry        telemetry.TelemetryConfig `toml:"telemetry"`
}

type GrpcAPIConfig struct {
	GRPCServerConfig GRPCServerConfig `toml:"grpc_server"`
	ELK              ELKConfig        `toml:"elk"`
}
