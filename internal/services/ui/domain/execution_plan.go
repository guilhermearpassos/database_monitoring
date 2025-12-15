package domain

import (
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"time"
)

type ExecutionPlan struct {
	PlanHandle string `json:"plan_handle"`
	//Server     ServerMetadata
	XmlPlan string `json:"xml_plan"`
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
	Level         int            `json:"level"` // Added for template rendering
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
	Convert *PlanAffectingConvert `json:"convert"`
}

type PlanAffectingConvert struct {
	ConvertIssue string `json:"convert_issue"`
	Expression   string `json:"expression"`
}

func ProtoParsedPlanToDomain(p *dbmv1.ParsedExecutionPlan) ParsedExecutionPlan {
	stats := make([]StatisticsInfo, len(p.StatsUsage))
	warns := make([]PlanWarning, len(p.Warnings))
	nodes := make([]PlanNode, len(p.Nodes))
	for i, info := range p.StatsUsage {
		stat := StatisticsInfo{
			LastUpdate:        info.LastUpdate,
			ModificationCount: info.ModificationCount,
			SamplingPercent:   info.SamplingPercent,
			Statistics:        info.Statistics,
			Table:             info.Table,
		}
		stats[i] = stat
	}
	for i, warn := range p.Warnings {
		w := PlanWarning{
			Convert: &PlanAffectingConvert{
				ConvertIssue: warn.GetConvert().ConvertIssue,
				Expression:   warn.GetConvert().Expression,
			},
		}
		warns[i] = w
	}
	for i, node := range p.Nodes {
		n := planNodeFromProto(node, 0)
		nodes[i] = n
	}
	return ParsedExecutionPlan{
		Plan: ExecutionPlan{
			PlanHandle: p.Plan.PlanHandle,
			XmlPlan:    p.Plan.XmlPlan,
		},
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
	return PlanNode{
		Name:          node.Name,
		EstimatedRows: node.EstimatedRows,
		SubtreeCost:   node.SubtreeCost,
		NodeCost:      node.NodeCost,
		Header: PlanNodeHeader{
			PhysicalOp:    node.Header.PhysicalOp,
			LogicalOp:     node.Header.LogicalOp,
			EstimateCpu:   node.Header.EstimateCpu,
			EstimateIO:    node.Header.EstimateIo,
			EstimateRows:  node.Header.EstimateRows,
			EstimatedCost: node.Header.EstimatedCost,
			Parallel:      node.Header.Parallel,
		},
		Nodes: childNodes,
		Level: level,
	}
}

type QueryMetric struct {
	// Basic query information
	LastExecutionTime time.Time
	LastElapsedTime   time.Duration

	// Metrics directly accessible in templates
	CpuTime            float64 // For template usage, derived from avgWorkerTime
	ExecutionCount     int64   // For template usage, derived from counters.executionCount
	AvgElapsedTime     float64 // For template usage, derived from rates.avgElapsedTime
	TotalWorkerTime    int64   // For template usage, derived from counters.totalWorkerTime
	AvgWorkerTime      float64 // For template usage, derived from rates.avgWorkerTime
	TotalLogicalReads  int64   // For template usage, derived from counters.totalLogicalReads
	AvgLogicalReads    float64 // For template usage, derived from rates.avgLogicalReads
	TotalLogicalWrites int64   // For template usage, derived from counters.totalLogicalWrites
	AvgLogicalWrites   float64 // For template usage, derived from rates.avgLogicalWrites
	TotalPhysicalReads int64   // For template usage, derived from counters.totalPhysicalReads
	AvgPhysicalReads   float64 // For template usage, derived from rates.avgPhysicalReads
	TotalRows          int64   // For template usage, derived from counters.totalRows
	AvgRows            float64 // For template usage, derived from rates.avgRows

	// Original counter metrics from the response
	Counters map[string]int64 `json:"counters"` // Using string to support both numeric formats

	// Original rate metrics from the response
	Rates map[string]float64 `json:"rates"`
}

// NewQueryMetric creates a new QueryMetric with values mapped from counters and rates
func NewQueryMetric(
	LastExecutionTime time.Time,
	LastElapsedTime time.Duration,
	Counters map[string]int64,
	Rates map[string]float64) *QueryMetric {
	return &QueryMetric{
		LastExecutionTime: LastExecutionTime,
		LastElapsedTime:   LastElapsedTime,
		Counters:          Counters,
		Rates:             Rates,
	}
}

// PopulateMetrics processes the counters and rates to fill in the direct access fields
func (qm *QueryMetric) PopulateMetrics() {
	// Extract execution count
	if execCount, ok := qm.Counters["executionCount"]; ok {
		// Convert string to int64 (handling error appropriately in real code)
		// This is simplified; you should handle conversion errors
		qm.ExecutionCount = execCount
	}

	// Extract worker time (CPU time)
	if totalWorkerTime, ok := qm.Counters["totalWorkerTime"]; ok {
		// Convert string to int64
		qm.TotalWorkerTime = totalWorkerTime
	}

	// Extract total logical reads
	if totalLogicalReads, ok := qm.Counters["totalLogicalReads"]; ok {
		qm.TotalLogicalReads = totalLogicalReads
	}

	// Extract total logical writes
	if totalLogicalWrites, ok := qm.Counters["totalLogicalWrites"]; ok {
		qm.TotalLogicalWrites = totalLogicalWrites
	}

	// Extract total physical reads
	if totalPhysicalReads, ok := qm.Counters["totalPhysicalReads"]; ok {
		qm.TotalPhysicalReads = totalPhysicalReads
	}

	// Extract total rows
	if totalRows, ok := qm.Counters["totalRows"]; ok {
		qm.TotalRows = totalRows
	}

	// Extract rates
	if avgWorkerTime, ok := qm.Rates["avgWorkerTime"]; ok {
		qm.AvgWorkerTime = avgWorkerTime
		qm.CpuTime = avgWorkerTime / 1000.0 // Set CpuTime to match the template's expectation
	}

	if avgElapsedTime, ok := qm.Rates["avgElapsedTime"]; ok {
		qm.AvgElapsedTime = avgElapsedTime / 1000.0
	}

	if avgLogicalReads, ok := qm.Rates["avgLogicalReads"]; ok {
		qm.AvgLogicalReads = avgLogicalReads
	}

	if avgLogicalWrites, ok := qm.Rates["avgLogicalWrites"]; ok {
		qm.AvgLogicalWrites = avgLogicalWrites
	}

	if avgPhysicalReads, ok := qm.Rates["avgPhysicalReads"]; ok {
		qm.AvgPhysicalReads = avgPhysicalReads
	}

	if avgRows, ok := qm.Rates["avgRows"]; ok {
		qm.AvgRows = avgRows
	}
}
