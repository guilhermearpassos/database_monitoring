package common_domain

import "time"

type QuerySample struct {
	Id              string
	Status          string
	Cmd             string
	SqlHandle       []byte
	PlanHandle      []byte `json:"PlanHandle"`
	Text            string
	IsBlocked       bool
	IsBlocker       bool
	Session         SessionMetadata
	Database        DataBaseMetadata
	Block           BlockMetadata
	Wait            WaitMetadata
	Snapshot        SnapshotMetadata
	TimeElapsedMs   int64
	CommandMetadata CommandMetadata
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
	ConnectionId         string
}

type DataBaseMetadata struct {
	DatabaseID   string
	DatabaseName string
}

type SnapshotMetadata struct {
	ID        string
	Timestamp time.Time
}
type CommandMetadata struct {
	TransactionId           string
	RequestId               string
	EstimatedCompletionTime int64
	PercentComplete         float64
}
