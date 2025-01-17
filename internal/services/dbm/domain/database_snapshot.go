package domain

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
	ID        string           `json:"id"`
	Timestamp time.Time        `json:"timestamp"`
	Server    ServerMeta       `json:"server"`
	Database  DataBaseMetadata `json:"database"`
}
