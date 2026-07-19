package config

type TLSConfig struct {
	Enabled  bool   `toml:"enabled" yaml:"enabled"`
	CertFile string `toml:"cert_file" yaml:"cert_file"`
	KeyFile  string `toml:"key_file" yaml:"key_file"`
}

type GRPCServerConfig struct {
	Enabled bool       `toml:"enabled" yaml:"enabled"`
	Grpc    GrpcConfig `toml:"grpc" yaml:"grpc"`
}
type GrpcConfig struct {
	Url                string    `toml:"url" yaml:"url"`
	GrpcMessageMaxSize int       `toml:"grpc_message_max_size" yaml:"grpc_message_max_size"`
	TLS                TLSConfig `toml:"tls" yaml:"tls"`
}

type GRPCUIConfig struct {
	Enabled bool   `toml:"enabled" yaml:"enabled"`
	Url     string `toml:"url" yaml:"url"`
}
