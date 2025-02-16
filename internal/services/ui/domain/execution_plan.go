package domain

import dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"

type ExecutionPlan struct {
	PlanHandle []byte
	//Server     ServerMetadata
	XmlPlan string
}

type ParsedExecutionPlan struct {
	Plan       ExecutionPlan
	StatsUsage []StatisticsInfo
	Warnings   []PlanWarning
	Nodes      []PlanNode
}

type PlanNode struct {
	Name          string
	EstimatedRows float64
	SubtreeCost   float64
	NodeCost      float64
	Header        PlanNodeHeader
	Nodes         []PlanNode
	Level         int // Added for template rendering
}

type PlanNodeHeader struct {
	PhysicalOp    string
	LogicalOp     string
	EstimateCpu   float64
	EstimateIO    float64
	EstimateRows  float64
	EstimatedCost float64
	Parallel      string
}

type StatisticsInfo struct {
	LastUpdate        string
	ModificationCount int64
	SamplingPercent   float64
	Statistics        string
	Table             string
}

type PlanWarning struct {
	Convert *PlanAffectingConvert
}

type PlanAffectingConvert struct {
	ConvertIssue string
	Expression   string
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
