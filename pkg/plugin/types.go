package plugin

import (
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/deadlock"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
)

type ExecutionPlan struct {
	PlanHandle string `json:"plan_handle"`
	XmlPlan    string `json:"xml_plan"`
}

type ParsedExecutionPlan struct {
	Plan       ExecutionPlan    `json:"plan"`
	StatsUsage []StatisticsInfo `json:"stats_usage"`
	Warnings   []PlanWarning    `json:"warnings"`
	Nodes      []PlanNode       `json:"nodes"`
}

type PlanNode struct {
	Name          string         `json:"name"`
	EstimatedRows float64        `json:"estimated_rows"`
	SubtreeCost   float64        `json:"subtree_cost"`
	NodeCost      float64        `json:"node_cost"`
	Header        PlanNodeHeader `json:"header"`
	Nodes         []PlanNode     `json:"nodes"`
	Level         int            `json:"level"`
}

type PlanNodeHeader struct {
	PhysicalOp    string  `json:"physical_op"`
	LogicalOp     string  `json:"logical_op"`
	EstimateCpu   float64 `json:"estimate_cpu"`
	EstimateIO    float64 `json:"estimate_io"`
	EstimateRows  float64 `json:"estimate_rows"`
	EstimatedCost float64 `json:"estimated_cost"`
	Parallel      string  `json:"parallel"`
}

type StatisticsInfo struct {
	LastUpdate        string  `json:"last_update"`
	ModificationCount int64   `json:"modification_count"`
	SamplingPercent   float64 `json:"sampling_percent"`
	Statistics        string  `json:"statistics"`
	Table             string  `json:"table"`
}

type PlanWarning struct {
	Convert      *PlanAffectingConvert `json:"convert"`
	MissingIndex *MissingIndex         `json:"missing_index"`
}

type MissingIndex struct {
	Database          string   `json:"database"`
	Schema            string   `json:"schema"`
	Table             string   `json:"table"`
	Impact            float64  `json:"impact"`
	EqualityColumns   []string `json:"equality_columns"`
	InequalityColumns []string `json:"inequality_columns"`
	IncludeColumns    []string `json:"include_columns"`
}

type PlanAffectingConvert struct {
	ConvertIssue string `json:"convert_issue"`
	Expression   string `json:"expression"`
}

func protoParsedPlanToDomain(p *dbmv1.ParsedExecutionPlan) ParsedExecutionPlan {
	if p == nil {
		return ParsedExecutionPlan{}
	}
	stats := make([]StatisticsInfo, 0, len(p.StatsUsage))
	warns := make([]PlanWarning, 0, len(p.Warnings))
	nodes := make([]PlanNode, 0, len(p.Nodes))
	for _, info := range p.StatsUsage {
		if info == nil {
			continue
		}
		stats = append(stats, StatisticsInfo{
			LastUpdate:        info.LastUpdate,
			ModificationCount: info.ModificationCount,
			SamplingPercent:   info.SamplingPercent,
			Statistics:        info.Statistics,
			Table:             info.Table,
		})
	}
	for _, warn := range p.Warnings {
		if warn == nil {
			continue
		}
		switch warn.Warning.(type) {
		case *dbmv1.PlanWarning_Convert:
			warns = append(warns, PlanWarning{
				Convert: &PlanAffectingConvert{
					ConvertIssue: warn.GetConvert().GetConvertIssue(),
					Expression:   warn.GetConvert().GetExpression(),
				},
			})
		case *dbmv1.PlanWarning_MissingIndex:
			warns = append(warns, PlanWarning{
				MissingIndex: &MissingIndex{
					Database:          warn.GetMissingIndex().GetDatabase(),
					Schema:            warn.GetMissingIndex().GetSchema(),
					Table:             warn.GetMissingIndex().GetTable(),
					Impact:            warn.GetMissingIndex().GetImpact(),
					EqualityColumns:   warn.GetMissingIndex().GetEqualityColumns(),
					InequalityColumns: warn.GetMissingIndex().GetInequalityColumns(),
					IncludeColumns:    warn.GetMissingIndex().GetIncludeColumns(),
				},
			})
		}
	}
	for _, node := range p.Nodes {
		if node == nil {
			continue
		}
		nodes = append(nodes, planNodeFromProto(node, 0))
	}
	var execPlan ExecutionPlan
	if p.Plan != nil {
		execPlan = ExecutionPlan{
			PlanHandle: p.Plan.PlanHandle,
			XmlPlan:    p.Plan.XmlPlan,
		}
	}
	return ParsedExecutionPlan{
		Plan:       execPlan,
		StatsUsage: stats,
		Warnings:   warns,
		Nodes:      nodes,
	}
}

func planNodeFromProto(node *dbmv1.PlanNode, level int) PlanNode {
	if node == nil {
		return PlanNode{Level: level}
	}
	childNodes := make([]PlanNode, 0, len(node.Nodes))
	for _, child := range node.Nodes {
		if child == nil {
			continue
		}
		childNodes = append(childNodes, planNodeFromProto(child, level+1))
	}
	var header PlanNodeHeader
	if node.Header != nil {
		header = PlanNodeHeader{
			PhysicalOp:    node.Header.PhysicalOp,
			LogicalOp:     node.Header.LogicalOp,
			EstimateCpu:   node.Header.EstimateCpu,
			EstimateIO:    node.Header.EstimateIo,
			EstimateRows:  node.Header.EstimateRows,
			EstimatedCost: node.Header.EstimatedCost,
			Parallel:      node.Header.Parallel,
		}
	}
	return PlanNode{
		Name:          node.Name,
		EstimatedRows: node.EstimatedRows,
		SubtreeCost:   node.SubtreeCost,
		NodeCost:      node.NodeCost,
		Header:        header,
		Nodes:         childNodes,
		Level:         level,
	}
}

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
	TraceID                 string    `json:"trace_id"`
}

type BlockingNode struct {
	QuerySample QuerySample    `json:"query_sample"`
	ChildNodes  []BlockingNode `json:"child_nodes"`
	Level       int            `json:"level"`
}

type BlockChain struct {
	Roots []BlockingNode `json:"roots"`
}

type PlanAnalysisItem struct {
	PlanHandle       string                 `json:"plan_handle"`
	QueryHash        string                 `json:"query_hash"`
	SQLHandle        string                 `json:"sql_handle"`
	Text             string                 `json:"text"`
	Database         string                 `json:"database"`
	SnapID           string                 `json:"snap_id"`
	SampleID         string                 `json:"sample_id"`
	LastSnapTime     time.Time              `json:"last_snap_time"`
	TraceID          string                 `json:"trace_id"`
	IssueCount       int32                  `json:"issue_count"`
	Occurrences      int32                  `json:"occurrences"`
	RelatedLockCount int32                  `json:"related_lock_count"`
	IssueTags        []string               `json:"issue_tags"`
	Converts         []PlanAffectingConvert `json:"converts"`
	MissingIndexes   []MissingIndex         `json:"missing_indexes"`
	MaxImpact        float64                `json:"max_impact"`
}

type PlanAnalysisResponse struct {
	Plans              []PlanAnalysisItem `json:"plans"`
	AvailableDatabases []string           `json:"available_databases"`
}

type SnapshotSummaryDTO struct {
	ID                     string           `json:"id"`
	Timestamp              time.Time        `json:"timestamp"`
	Connections            int32            `json:"connections"`
	Waiters                int32            `json:"waiters"`
	Blockers               int32            `json:"blockers"`
	WaitDurationMs         float64          `json:"wait_duration_ms"`
	AvgDuration            float64          `json:"avg_duration"`
	MaxDuration            float64          `json:"max_duration"`
	ConnectionsByWaitEvent map[string]int64 `json:"connections_by_wait_event"`
	TimeMsByWaitEvent      map[string]int64 `json:"time_ms_by_wait_event"`
}

type SnapshotSampleDTO struct {
	SampleID        string   `json:"sample_id"`
	SnapID          string   `json:"snap_id"`
	SessionID       string   `json:"session_id"`
	Status          string   `json:"status"`
	Text            string   `json:"text"`
	User            string   `json:"user"`
	ElapsedMs       int64    `json:"elapsed_ms"`
	WaitEvent       string   `json:"wait_event"`
	WaitTimeMs      int64    `json:"wait_time_ms"`
	WaitResource    string   `json:"wait_resource"`
	Database        string   `json:"database"`
	IsBlocked       bool     `json:"is_blocked"`
	IsBlocker       bool     `json:"is_blocker"`
	BlockedBy       string   `json:"blocked_by"`
	BlockedSessions []string `json:"blocked_sessions"`
	TraceID         string   `json:"trace_id"`
	QueryHash       string   `json:"query_hash"`
	PlanHandle      string   `json:"plan_handle"`
	ProgramName     string   `json:"program_name"`
	HostName        string   `json:"host_name"`
	ClientIP        string   `json:"client_ip"`
	SessionStatus   string   `json:"session_status"`
}

type QueryMetricItemDTO struct {
	QueryHash           string    `json:"query_hash"`
	Text                string    `json:"text"`
	Database            string    `json:"database"`
	LastExecutionTime   time.Time `json:"last_execution_time"`
	LastElapsedMs       float64   `json:"last_elapsed_ms"`
	ExecutionCount      int64     `json:"execution_count"`
	TotalElapsedMs      float64   `json:"total_elapsed_ms"`
	AvgElapsedMs        float64   `json:"avg_elapsed_ms"`
	TotalWorkerMs       float64   `json:"total_worker_ms"`
	AvgWorkerMs         float64   `json:"avg_worker_ms"`
	TotalClrMs          float64   `json:"total_clr_ms"`
	TotalLogicalReads   int64     `json:"total_logical_reads"`
	AvgLogicalReads     float64   `json:"avg_logical_reads"`
	TotalPhysicalReads  int64     `json:"total_physical_reads"`
	AvgPhysicalReads    float64   `json:"avg_physical_reads"`
	TotalLogicalWrites  int64     `json:"total_logical_writes"`
	AvgLogicalWrites    float64   `json:"avg_logical_writes"`
	TotalRows           int64     `json:"total_rows"`
	AvgRows             float64   `json:"avg_rows"`
	TotalSpills         int64     `json:"total_spills"`
	AvgSpills           float64   `json:"avg_spills"`
	TotalGrantKb        int64     `json:"total_grant_kb"`
	AvgGrantKb          float64   `json:"avg_grant_kb"`
	AvgUsedGrantKb      float64   `json:"avg_used_grant_kb"`
	AvgIdealGrantKb     float64   `json:"avg_ideal_grant_kb"`
	MemoryGrantWastePct float64   `json:"memory_grant_waste_pct"`
	AvgDop              float64   `json:"avg_dop"`
	CpuSharePct         float64   `json:"cpu_share_pct"`
	ElapsedSharePct     float64   `json:"elapsed_share_pct"`
	ReadsSharePct       float64   `json:"reads_share_pct"`
	IssueFlags          []string  `json:"issue_flags"`
}

type QueryMetricsResponse struct {
	Metrics            []QueryMetricItemDTO `json:"metrics"`
	AvailableDatabases []string             `json:"available_databases"`
}

type QueryMetricsSeriesPointDTO struct {
	Timestamp         time.Time `json:"timestamp"`
	ExecutionCount    int64     `json:"execution_count"`
	TotalElapsedMs    float64   `json:"total_elapsed_ms"`
	AvgElapsedMs      float64   `json:"avg_elapsed_ms"`
	TotalWorkerMs     float64   `json:"total_worker_ms"`
	AvgWorkerMs       float64   `json:"avg_worker_ms"`
	TotalLogicalReads int64     `json:"total_logical_reads"`
	AvgLogicalReads   float64   `json:"avg_logical_reads"`
	AvgPhysicalReads  float64   `json:"avg_physical_reads"`
	AvgLogicalWrites  float64   `json:"avg_logical_writes"`
	AvgRows           float64   `json:"avg_rows"`
	TotalSpills       int64     `json:"total_spills"`
	AvgSpills         float64   `json:"avg_spills"`
	AvgGrantKb        float64   `json:"avg_grant_kb"`
	AvgUsedGrantKb    float64   `json:"avg_used_grant_kb"`
}

type DeadlockItemDTO struct {
	ID                 string               `json:"id"`
	Server             string               `json:"server"`
	Timestamp          time.Time            `json:"timestamp"`
	XMLHash            string               `json:"xml_hash"`
	Fingerprint        string               `json:"fingerprint"`
	PatternOccurrences int                  `json:"pattern_occurrences"`
	VictimSPID         string               `json:"victim_spid"`
	Databases          []string             `json:"databases"`
	MaxWaitMs          int64                `json:"max_wait_ms"`
	Processes          []deadlock.Process   `json:"processes"`
	Resources          []deadlock.Resource  `json:"resources"`
	Diagnosis          deadlock.Diagnosis   `json:"diagnosis"`
	XMLReport          string               `json:"xml_report"`
}

type DeadlocksResponse struct {
	Deadlocks          []DeadlockItemDTO `json:"deadlocks"`
	AvailableDatabases []string          `json:"available_databases"`
}
