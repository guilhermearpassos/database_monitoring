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
	SID           int
	Query         string
	ExecutionTime string
	User          string
	IsBlocker     bool
	IsWaiter      bool
	BlockingTime  string
	BlockDetails  string
	WaitEvent     string
	Database      string
	SampleID      string
}
