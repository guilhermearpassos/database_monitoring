package domain

type AgentConfig struct {
	ID           string
	TargetHost   string
	TargetType   string
	AgentVersion string
	Tags         map[string]struct{}
}
