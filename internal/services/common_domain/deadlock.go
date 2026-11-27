package common_domain

import "time"

type DeadlockReport struct {
	ID            string
	Server        ServerMeta
	OccurredAt    time.Time
	XMLReport     string
	XMLHash       string
	Fingerprint   string
	Pattern       string
	Databases     []string
	VictimSPID    string
	MaxWaitMs     int64
	ProcessCount  int32
	ResourceCount int32
}
