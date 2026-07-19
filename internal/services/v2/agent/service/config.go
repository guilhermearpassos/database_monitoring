package service

type TargetConfig struct {
	Alias      string `json:"alias"`
	Driver     string `json:"driver"`
	ConnString string `json:"conn_string"`
}

type AgentConfig struct {
	Targets []TargetConfig `yaml:"targets" toml:"targets"`
}
