package common_domain

import "time"

type QuerySample struct {
	Status        string
	Cmd           string
	SqlHandle     []byte
	PlanHandle    []byte
	Text          string
	IsBlocked     bool
	IsBlocker     bool
	Session       SessionMetadata
	Database      DataBaseMetadata
	Block         BlockMetadata
	Wait          WaitMetadata
	Snapshot      SnapshotMetadata
	TimeElapsedMs int64
}

func (q *QuerySample) SetBlockedIds(sessionIds []string) {
	q.IsBlocker = true
	q.Block.AddBlockedIds(sessionIds)
}

type WaitMetadata struct {
	WaitType     *string
	WaitTime     int
	LastWaitType string
	WaitResource string
}

type BlockMetadata struct {
	BlockedBy       string
	BlockedSessions []string
}

func (m *BlockMetadata) AddBlockedIds(sessionIds []string) {
	m.BlockedSessions = append(m.BlockedSessions, sessionIds...)
}

type SessionMetadata struct {
	SessionID            string
	LoginTime            time.Time
	HostName             string
	ProgramName          string
	LoginName            string
	Status               string
	LastRequestStartTime time.Time
	LastRequestEndTime   time.Time
}

type DataBaseMetadata struct {
	DatabaseID   string
	DatabaseName string
}

type SnapshotMetadata struct {
	ID        string
	Timestamp time.Time
}
