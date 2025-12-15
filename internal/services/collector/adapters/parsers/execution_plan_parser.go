package parsers

import (
	"encoding/xml"
	"fmt"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
)

// ParsedExecutionPlan represents the root of the execution plan XML
type ParsedExecutionPlan struct {
	XMLName    xml.Name    `xml:"ShowPlanXML"`
	Version    string      `xml:"Version,attr"`
	Statements []Statement `xml:"BatchSequence>Batch>Statements>StmtSimple"`
	Raw        string
}

// Statement represents a SQL statement and its plan
type Statement struct {
	StatementText        string    `xml:"StatementText,attr"`
	StatementId          string    `xml:"StatementId,attr"`
	StatementType        string    `xml:"StatementType,attr"`
	StatementSubTreeCost float64   `xml:"StatementSubTreeCost,attr"`
	StatementEstRows     float64   `xml:"StatementEstRows,attr"`
	QueryPlan            QueryPlan `xml:"QueryPlan"`
}

// QueryPlan contains the actual execution plan
type QueryPlan struct {
	RelOp                RelOp                  `xml:"RelOp"`
	PlanAffectingConvert []PlanAffectingConvert `xml:"Warnings>PlanAffectingConvert"`
	OptimizerStatsUsage  struct {
		StatisticsInfo []StatisticsInfo `xml:"StatisticsInfo"`
	} `xml:"OptimizerStatsUsage"`
}

type PlanAffectingConvert struct {
	XMLName      xml.Name `xml:"PlanAffectingConvert"`
	ConvertIssue string   `xml:"ConvertIssue,attr"`
	Expression   string   `xml:"Expression,attr"`
}

type StatisticsInfo struct {
	XMLName           xml.Name `xml:"StatisticsInfo"`
	LastUpdate        string   `xml:"LastUpdate,attr"`
	ModificationCount int      `xml:"ModificationCount,attr"`
	SamplingPercent   float64  `xml:"SamplingPercent,attr"`
	Statistics        string   `xml:"Statistics,attr"`
	Table             string   `xml:"Table,attr"`
}

// OperatorHeader represents common attributes for all operators
type OperatorHeader struct {
	NodeId           string  `xml:"NodeId,attr"`
	PhysicalOp       string  `xml:"PhysicalOp,attr"`
	LogicalOp        string  `xml:"LogicalOp,attr"`
	EstimateCPU      float64 `xml:"EstimateCPU,attr"`
	EstimateIO       float64 `xml:"EstimateIO,attr"`
	EstimateRows     float64 `xml:"EstimateRows,attr"`
	EstimateRowsRead float64 `xml:"EstimateRowsRead,attr"`
	EstimatedCost    float64 `xml:"EstimatedTotalSubtreeCost,attr"`
	TableCardinality string  `xml:"TableCardinality,attr"`
	Parallel         string  `xml:"Parallel,attr"`
}

// RelOp represents a relational operation in the execution plan
// The RelOp tag is ALWAYS the parent, and operation-specific tags are children
type RelOp struct {
	OperatorHeader

	// Common elements that can appear in any RelOp
	OutputList    []OutputColumn `xml:"OutputList>ColumnReference"`
	DefinedValues []DefinedValue `xml:"DefinedValues>DefinedValue"`

	// Operation-specific child elements that may contain nested RelOps
	Action          *ActionOp          `xml:"Action"`
	Aggregate       *AggregateOp       `xml:"Aggregate"`
	Assert          *AssertOp          `xml:"Assert"`
	Compute         *ComputeOp         `xml:"Compute"`
	ComputeScalar   *ComputeScalarOp   `xml:"ComputeScalar"`
	Delete          *DeleteOp          `xml:"Delete"`
	Filter          *FilterOp          `xml:"Filter"`
	Hash            *HashOp            `xml:"Hash"`
	Insert          *InsertOp          `xml:"Insert"`
	Merge           *MergeOp           `xml:"Merge"`
	NestedLoops     *NestedLoopsOp     `xml:"NestedLoops"`
	Parallelism     *ParallelismOp     `xml:"Parallelism"`
	Remote          *RemoteOp          `xml:"Remote"`
	RemoteQuery     *RemoteQueryOp     `xml:"RemoteQuery"`
	Segment         *SegmentOp         `xml:"Segment"`
	Sequence        *SequenceOp        `xml:"Sequence"`
	Sort            *SortOp            `xml:"Sort"`
	Spool           *SpoolOp           `xml:"Spool"`
	StreamAggregate *StreamAggregateOp `xml:"StreamAggregate"`
	Top             *TopOp             `xml:"Top"`
	TopSort         *TopSortOp         `xml:"TopSort"`
	Update          *UpdateOp          `xml:"Update"`

	// Leaf operations (no nested RelOps, just details)
	IndexScan *IndexScanOp `xml:"IndexScan"`
	TableScan *TableScanOp `xml:"TableScan"`
}

// Operation types that contain nested RelOps
type ActionOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type AggregateOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type AssertOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type ComputeOp struct {
	RelOp []RelOp `xml:"RelOp"`
}
type ComputeScalarOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type DeleteOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type FilterOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type HashOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type InsertOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type MergeOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type NestedLoopsOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type ParallelismOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type RemoteOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type RemoteQueryOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type SegmentOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type SequenceOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type SortOp struct {
	OrderBy SortDetails `xml:"OrderBy"`
	RelOp   []RelOp     `xml:"RelOp"`
}

type SpoolOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type StreamAggregateOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type TopOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type TopSortOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

type UpdateOp struct {
	RelOp []RelOp `xml:"RelOp"`
}

// Leaf operations (no nested RelOps)
type IndexScanOp struct {
	Object IndexScanDetails `xml:"Object"`
}

type TableScanOp struct {
	Object TableScanDetails `xml:"Object"`
}

// Details structures
type IndexScanDetails struct {
	Database  string `xml:"Database,attr"`
	Schema    string `xml:"Schema,attr"`
	Table     string `xml:"Table,attr"`
	Index     string `xml:"Index,attr"`
	IndexKind string `xml:"IndexKind,attr"`
	Storage   string `xml:"Storage,attr"`
}

type TableScanDetails struct {
	Database string `xml:"Database,attr"`
	Schema   string `xml:"Schema,attr"`
	Table    string `xml:"Table,attr"`
}

type SortDetails struct {
	Columns []SortColumn `xml:"ColumnReference"`
}

type SortColumn struct {
	Column string `xml:"Column,attr"`
}

type OutputColumn struct {
	Column string `xml:"Column,attr"`
}

type DefinedValue struct {
	Column string `xml:"ColumnReference>Column,attr"`
}

// GetAllChildren returns all child RelOps from all possible operation types
func (r *RelOp) GetAllChildren() []RelOp {
	var children []RelOp

	if r.Action != nil {
		children = append(children, r.Action.RelOp...)
	}
	if r.Aggregate != nil {
		children = append(children, r.Aggregate.RelOp...)
	}
	if r.Assert != nil {
		children = append(children, r.Assert.RelOp...)
	}
	if r.Compute != nil {
		children = append(children, r.Compute.RelOp...)
	}
	if r.ComputeScalar != nil {
		children = append(children, r.ComputeScalar.RelOp...)
	}
	if r.Delete != nil {
		children = append(children, r.Delete.RelOp...)
	}
	if r.Filter != nil {
		children = append(children, r.Filter.RelOp...)
	}
	if r.Hash != nil {
		children = append(children, r.Hash.RelOp...)
	}
	if r.Insert != nil {
		children = append(children, r.Insert.RelOp...)
	}
	if r.Merge != nil {
		children = append(children, r.Merge.RelOp...)
	}
	if r.NestedLoops != nil {
		children = append(children, r.NestedLoops.RelOp...)
	}
	if r.Parallelism != nil {
		children = append(children, r.Parallelism.RelOp...)
	}
	if r.Remote != nil {
		children = append(children, r.Remote.RelOp...)
	}
	if r.RemoteQuery != nil {
		children = append(children, r.RemoteQuery.RelOp...)
	}
	if r.Segment != nil {
		children = append(children, r.Segment.RelOp...)
	}
	if r.Sequence != nil {
		children = append(children, r.Sequence.RelOp...)
	}
	if r.Sort != nil {
		children = append(children, r.Sort.RelOp...)
	}
	if r.Spool != nil {
		children = append(children, r.Spool.RelOp...)
	}
	if r.StreamAggregate != nil {
		children = append(children, r.StreamAggregate.RelOp...)
	}
	if r.Top != nil {
		children = append(children, r.Top.RelOp...)
	}
	if r.TopSort != nil {
		children = append(children, r.TopSort.RelOp...)
	}
	if r.Update != nil {
		children = append(children, r.Update.RelOp...)
	}

	// IndexScan and TableScan are leaf nodes - no children

	return children
}

func ParseExecutionPlan(data string) (*ParsedExecutionPlan, error) {
	var plan ParsedExecutionPlan
	if err := xml.Unmarshal([]byte(data), &plan); err != nil {
		return nil, fmt.Errorf("Error parsing execution plan XML: %w", err)
	}
	plan.Raw = data
	return &plan, nil
}

func PlanToProto(handle string, server common_domain.ServerMeta, plan *ParsedExecutionPlan) (*dbmv1.ParsedExecutionPlan, error) {
	stats := make([]*dbmv1.StatisticsInfo, 0)
	warnings := make([]*dbmv1.PlanWarning, 0)
	nodes := make([]*dbmv1.PlanNode, 0)
	for _, stmt := range plan.Statements {
		for _, stat := range stmt.QueryPlan.OptimizerStatsUsage.StatisticsInfo {
			stats = append(stats, &dbmv1.StatisticsInfo{
				LastUpdate:        stat.LastUpdate,
				ModificationCount: int64(stat.ModificationCount),
				SamplingPercent:   stat.SamplingPercent,
				Statistics:        stat.Statistics,
				Table:             stat.Table,
			})
		}
		for _, warning := range stmt.QueryPlan.PlanAffectingConvert {
			warnings = append(warnings, &dbmv1.PlanWarning{
				Warning: &dbmv1.PlanWarning_Convert{Convert: &dbmv1.PlanWarning_PlanAffectingConvert{
					ConvertIssue: warning.ConvertIssue,
					Expression:   warning.Expression,
				}},
			})
		}
		baseNode := dbmv1.PlanNode{
			Name:          stmt.StatementType,
			EstimatedRows: stmt.StatementEstRows,
			SubtreeCost:   stmt.StatementSubTreeCost,
			NodeCost:      stmt.StatementSubTreeCost,
			Header:        &dbmv1.PlanNode_Header{},
			Nodes:         []*dbmv1.PlanNode{relOpToProtoNode(stmt.QueryPlan.RelOp)},
		}
		nodes = append(nodes, &baseNode)

	}
	return &dbmv1.ParsedExecutionPlan{
		Plan: &dbmv1.ExecutionPlan{
			PlanHandle: handle,
			Server: &dbmv1.ServerMetadata{
				Host: server.Host,
				Type: server.Type,
			},
			XmlPlan: plan.Raw,
		},
		StatsUsage: stats,
		Warnings:   warnings,
		Nodes:      nodes,
	}, nil
}

func relOpToProtoNode(n RelOp) *dbmv1.PlanNode {
	baseNode := &dbmv1.PlanNode{
		Name:          n.PhysicalOp,
		EstimatedRows: n.EstimateRows,
		SubtreeCost:   n.EstimatedCost,
		NodeCost:      n.EstimatedCost,
		Header: &dbmv1.PlanNode_Header{
			PhysicalOp:    n.PhysicalOp,
			LogicalOp:     n.LogicalOp,
			EstimateCpu:   n.EstimateCPU,
			EstimateIo:    n.EstimateIO,
			EstimateRows:  n.EstimateRows,
			EstimatedCost: n.EstimatedCost,
			Parallel:      n.Parallel,
		},
		Nodes: make([]*dbmv1.PlanNode, 0),
	}

	for _, c := range n.GetAllChildren() {
		baseNode.Nodes = append(baseNode.Nodes, relOpToProtoNode(c))
	}

	return baseNode
}
