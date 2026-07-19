package domain

import (
	"time"
)

type WaitType struct {
	Type    string
	Percent int
	Color   string
}

type Server struct {
	Name           string
	Connections    int
	RequestRate    string
	DatabaseType   string
	BlockedPercent int
	WaitTypes      []WaitType
}

// Sample data - in a real app, this would come from a database
var SampleServers = []Server{
	{
		Name:           "localhost",
		Connections:    120,
		RequestRate:    "350 req/s",
		DatabaseType:   "mssql",
		BlockedPercent: 15,
		WaitTypes: []WaitType{
			{"CPU", 40, "bg-blue-500"},
			{"IO", 35, "bg-green-500"},
			{"Network", 25, "bg-yellow-500"},
		},
	},
	{
		Name:           "Server 2",
		Connections:    80,
		RequestRate:    "210 req/s",
		DatabaseType:   "postgres",
		BlockedPercent: 20,
		WaitTypes: []WaitType{
			{"CPU", 50, "bg-blue-500"},
			{"IO", 30, "bg-green-500"},
			{"Network", 20, "bg-yellow-500"},
		},
	},
	{
		Name:           "Server 2",
		Connections:    22,
		RequestRate:    "10 req/s",
		DatabaseType:   "mysql",
		BlockedPercent: 10,
		WaitTypes: []WaitType{
			{"CPU", 20, "bg-blue-500"},
			{"IO", 15, "bg-green-500"},
			{"Network", 30, "bg-yellow-500"},
			{"Locks", 25, "bg-yellow-500"},
			{"Memory", 10, "bg-yellow-500"},
		},
	},
}

type TimeSeriesData struct {
	Timestamp  time.Time      `json:"timestamp"`
	WaitGroups map[string]int `json:"wait_groups"`
}

// Snapshot represents a database snapshot
type Snapshot struct {
	ID           string
	Timestamp    time.Time
	Connections  int
	WaitEvGroups []WaitType
	Users        []string
	WaitersNo    int
	BlockersNo   int
	WaitDuration string
	AvgDuration  string
	MaxDuration  string
}

// QuerySample represents a query sample from a snapshot
type QuerySample struct {
	SID                     int       `json:"sid"`
	Query                   string    `json:"query"`
	ExecutionTime           string    `json:"execution_time"`
	User                    string    `json:"user"`
	IsBlocker               bool      `json:"is_blocker"`
	IsWaiter                bool      `json:"is_waiter"`
	BlockingTime            string    `json:"blocking_time"`
	BlockDetails            string    `json:"block_details"`
	WaitEvent               string    `json:"wait_event"`
	Database                string    `json:"database"`
	SampleID                string    `json:"sample_id"`
	SnapID                  string    `json:"snap_id"`
	SQLHandle               string    `json:"sql_handle"`
	PlanHandle              string    `json:"plan_handle"`
	Status                  string    `json:"status"`
	QueryHash               string    `json:"query_hash"`
	SessionLoginTime        time.Time `json:"session_login_time"`
	SessionHost             string    `json:"session_host"`
	SessionStatus           string    `json:"session_status"`
	SessionProgramName      string    `json:"session_program_name"`
	SessionLastRequestStart time.Time `json:"session_last_request_start"`
	SessionLastRequestEnd   time.Time `json:"session_last_request_end"`
	SessionClientIp         string    `json:"session_client_ip"`
}

type BlockingNode struct {
	QuerySample QuerySample    `json:"query_sample"`
	ChildNodes  []BlockingNode `json:"child_nodes"`
	Level       int            `json:"level"` // Added for template rendering
}
type BlockChain struct {
	Roots []BlockingNode `json:"roots"`
}

type QueryDetailsData struct {
	QuerySample QuerySample
	BlockChain  BlockChain
}
