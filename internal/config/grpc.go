package config

type TLSConfig struct {
	Enabled  bool   `toml:"enabled"`
	CertFile string `toml:"cert_file"`
	KeyFile  string `toml:"key_file"`
}

type GRPCServerConfig struct {
	Enabled bool       `toml:"enabled"`
	Grpc    GrpcConfig `toml:"grpc"`
}
type GrpcConfig struct {
	Url                string    `toml:"url"`
	GrpcMessageMaxSize int       `toml:"grpc_message_max_size"`
	TLS                TLSConfig `toml:"tls"`
}

type GRPCUIConfig struct {
	Enabled bool   `toml:"enabled"`
	Url     string `toml:"url"`
}
