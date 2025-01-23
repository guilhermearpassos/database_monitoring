package domain

import "time"

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

func GenerateSampleData() []TimeSeriesData {
	return []TimeSeriesData{
		{Timestamp: time.Date(2025, time.January, 18, 1, 0, 0, 0, time.UTC), WaitGroups: map[string]int{"wg1": 10, "wg2": 5}},
		{Timestamp: time.Date(2025, time.January, 18, 2, 0, 0, 0, time.UTC), WaitGroups: map[string]int{"wg1": 15, "wg2": 10}},
		{Timestamp: time.Date(2025, time.January, 18, 3, 0, 0, 0, time.UTC), WaitGroups: map[string]int{"wg1": 20, "wg2": 12}},
		{Timestamp: time.Date(2025, time.January, 18, 4, 0, 0, 0, time.UTC), WaitGroups: map[string]int{"wg1": 25, "wg2": 18}},
		{Timestamp: time.Date(2025, time.January, 18, 5, 0, 0, 0, time.UTC), WaitGroups: map[string]int{"wg1": 30, "wg2": 20}},
	}
}

// Snapshot represents a database snapshot
type Snapshot struct {
	ID        string
	Timestamp time.Time
	DBName    string
	Status    string
}

// QuerySample represents a query sample from a snapshot
type QuerySample struct {
	Query         string
	ExecutionTime string
	User          string
}

// Global data storage (in-memory, can replace with database)
var Snapshots = []Snapshot{
	{"1", time.Date(2025, 1, 16, 10, 0, 0, 0, time.UTC), "ProductionDB", "Success"},
	{"a", time.Date(2025, 1, 15, 9, 30, 0, 0, time.UTC), "TestDB", "Failed"},
	{"3", time.Date(2025, 1, 14, 8, 45, 0, 0, time.UTC), "AnalyticsDB", "Success"},
}

var QuerySamples = map[string][]QuerySample{
	"1": {
		{"SELECT * FROM users WHERE id = 1", "5ms", "admin"},
		{"INSERT INTO orders (user_id, total) VALUES (1, 100)", "8ms", "admin"},
	},
	"a": {
		{"SELECT * FROM orders WHERE id = 1", "3ms", "user1"},
		{"UPDATE users SET status = 'active' WHERE id = 1", "4ms", "user2"},
	},
	"3": {
		{"SELECT * FROM products WHERE category = 'electronics'", "6ms", "admin"},
	},
}
