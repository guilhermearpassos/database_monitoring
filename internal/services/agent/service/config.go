package service

type AgentConfig struct {
	CollectorConfig CollectorConfig    `toml:"collector"`
	TargetHosts     []TargetHostConfig `toml:"target_hosts"`
}

type CollectorConfig struct {
	Url string `toml:"url"`
}

type TargetHostConfig struct {
	Alias      string `toml:"alias"`
	Driver     string `toml:"driver"`
	ConnString string `toml:"conn_string"`
}
