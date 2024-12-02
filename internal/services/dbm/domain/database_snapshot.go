package domain

import "time"

type DataBaseSnapshot struct {
	Timestamp time.Time
	Samples   []*QuerySample
	Server    ServerMeta
	Database  DataBaseMetadata
}

type ServerMeta struct {
	Host string
	Type string
}
