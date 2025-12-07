package common_domain

import "time"

type DataBaseSnapshot struct {
	SnapInfo SnapInfo
	Samples  []*QuerySample
}

func (s *DataBaseSnapshot) GetPlanHandles() []string {
	handles := make([]string, len(s.Samples))
	for i, sample := range s.Samples {
		if sample.PlanHandle == "" {
			continue
		}
		handles[i] = sample.PlanHandle
	}
	return handles
}

type ServerMeta struct {
	Host string
	Type string
}

type SnapInfo struct {
	ID        string     `json:"id"`
	Timestamp time.Time  `json:"timestamp"`
	Server    ServerMeta `json:"server"`
}

type SnapshotSummary struct {
	ID               string
	Timestamp        time.Time
	Server           ServerMeta
	ConnsByWaitType  map[string]int64
	TimeMsByWaitType map[string]int64
}
