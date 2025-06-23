package common_domain

import "time"

type DataBaseSnapshot struct {
	SnapInfo SnapInfo
	Samples  []*QuerySample
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
