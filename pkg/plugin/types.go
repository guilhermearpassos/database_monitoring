package plugin

import (
	"time"

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
	stats := make([]StatisticsInfo, len(p.StatsUsage))
	warns := make([]PlanWarning, len(p.Warnings))
	nodes := make([]PlanNode, len(p.Nodes))
	for i, info := range p.StatsUsage {
		stats[i] = StatisticsInfo{
			LastUpdate:        info.LastUpdate,
			ModificationCount: info.ModificationCount,
			SamplingPercent:   info.SamplingPercent,
			Statistics:        info.Statistics,
			Table:             info.Table,
		}
	}
	for i, warn := range p.Warnings {
		switch warn.Warning.(type) {
		case *dbmv1.PlanWarning_Convert:
			warns[i] = PlanWarning{
				Convert: &PlanAffectingConvert{
					ConvertIssue: warn.GetConvert().GetConvertIssue(),
					Expression:   warn.GetConvert().GetExpression(),
				},
			}
		case *dbmv1.PlanWarning_MissingIndex:
			warns[i] = PlanWarning{
				MissingIndex: &MissingIndex{
					Database:          warn.GetMissingIndex().GetDatabase(),
					Schema:            warn.GetMissingIndex().GetSchema(),
					Table:             warn.GetMissingIndex().GetTable(),
					Impact:            warn.GetMissingIndex().GetImpact(),
					EqualityColumns:   warn.GetMissingIndex().GetEqualityColumns(),
					InequalityColumns: warn.GetMissingIndex().GetInequalityColumns(),
					IncludeColumns:    warn.GetMissingIndex().GetIncludeColumns(),
				},
			}
		}
	}
	for i, node := range p.Nodes {
		nodes[i] = planNodeFromProto(node, 0)
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
	childNodes := make([]PlanNode, len(node.Nodes))
	for i, child := range node.Nodes {
		childNodes[i] = planNodeFromProto(child, level+1)
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
}

type BlockingNode struct {
	QuerySample QuerySample    `json:"query_sample"`
	ChildNodes  []BlockingNode `json:"child_nodes"`
	Level       int            `json:"level"`
}

type BlockChain struct {
	Roots []BlockingNode `json:"roots"`
}
