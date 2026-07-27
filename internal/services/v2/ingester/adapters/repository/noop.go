package repository

import (
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
)

// NoopRepo is a minimal SnapshotRepository that discards data.
// Useful for wiring and tests before a real persistence adapter is provided.
type NoopRepo struct{}

func NewNoopRepo() *NoopRepo { return &NoopRepo{} }

func (n *NoopRepo) SaveSamples(_ string, _ uint32, _ []*dbmv1.QuerySample) error { return nil }
func (n *NoopRepo) FinalizeSnapshot(_ string) error { return nil }
