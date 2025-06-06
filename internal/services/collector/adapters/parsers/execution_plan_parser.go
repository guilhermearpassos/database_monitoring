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
type RelOp struct {
	OperatorHeader

	// Additional details for specific operations
	OutputList    []OutputColumn `xml:"OutputList>ColumnReference"`
	DefinedValues []DefinedValue `xml:"DefinedValues>DefinedValue"`

	// Handle all possible nested locations
	RelOp  []RelOp `xml:"RelOp"`        // Direct RelOp
	Action []RelOp `xml:"Action>RelOp"` // Action wrapper
	Assert []struct {
		RelOp `xml:"relOp"`
	} `xml:"Assert"` // Action wrapper
	Aggregate       []RelOp            `xml:"Aggregate>RelOp"`       // Aggregate wrapper
	Compute         []RelOp            `xml:"Compute>RelOp"`         // Compute wrapper
	Delete          []RelOp            `xml:"Delete>RelOp"`          // Delete wrapper
	FilterWrapper   []RelOp            `xml:"Filter>RelOp"`          // Filter wrapper
	Hash            []RelOp            `xml:"Hash>RelOp"`            // Hash wrapper
	IndexScan       []IndexScanRelOp   `xml:"IndexScan"`             // Scan operations
	Insert          []RelOp            `xml:"Insert>RelOp"`          // Insert wrapper
	Merge           []RelOp            `xml:"Merge>RelOp"`           // Merge wrapper
	NestedLoops     []NestedLoopsRelOp `xml:"NestedLoops"`           // NestedLoops wrapper
	Parallelism     []RelOp            `xml:"Parallelism>RelOp"`     // Parallelism wrapper
	Remote          []RelOp            `xml:"Remote>RelOp"`          // Remote wrapper
	RemoteQuery     []RelOp            `xml:"RemoteQuery>RelOp"`     // RemoteQuery wrapper
	Segment         []RelOp            `xml:"Segment>RelOp"`         // Segment wrapper
	Sequence        []RelOp            `xml:"Sequence>RelOp"`        // Sequence wrapper
	SortWrapper     []SortRelOp        `xml:"Sort"`                  // Sort wrapper
	Spool           []RelOp            `xml:"Spool>RelOp"`           // Spool wrapper
	StreamAggregate []RelOp            `xml:"StreamAggregate>RelOp"` // StreamAggregate wrapper
	TableScan       []TableScanRelOp   `xml:"TableScan"`             // Scan operations
	Top             []RelOp            `xml:"Top>RelOp"`
	TopSort         []RelOp            `xml:"TopSort>RelOp"` // TopSort wrapper
	Update          []RelOp            `xml:"Update>RelOp"`  // Update wrapper
}

// IndexScanDetails contains details for IndexScan operations
type IndexScanDetails struct {
	Database  string `xml:"Database,attr"`
	Schema    string `xml:"Schema,attr"`
	Table     string `xml:"Table,attr"`
	Index     string `xml:"Index,attr"`
	IndexKind string `xml:"IndexKind,attr"`
	Storage   string `xml:"Storage,attr"`
}

// TableScanDetails contains details for TableScan operations
type TableScanDetails struct {
	Database string `xml:"Database,attr"`
	Schema   string `xml:"Schema,attr"`
	Table    string `xml:"Table,attr"`
}

// SortDetails contains details for Sort operations
type SortDetails struct {
	Columns []SortColumn `xml:"ColumnReference"`
}

// SortColumn represents a column in a Sort operation
type SortColumn struct {
	Column string `xml:"Column,attr"`
}

// OutputColumn represents a column in the output list
type OutputColumn struct {
	Column string `xml:"Column,attr"`
}

// DefinedValue represents a computed or defined value
type DefinedValue struct {
	Column string `xml:"ColumnReference>Column,attr"`
}

type NestedLoopsRelOp struct {
	OuterRelOp RelOp `xml:"OuterRelOp"`
	InnerRelOp RelOp `xml:"InnerRelOp"`
}

type SortRelOp struct {
	RelOp   RelOp       `xml:"RelOp"`
	Details SortDetails `xml:"OrderBy"`
}

type IndexScanRelOp struct {
	RelOp   RelOp            `xml:"RelOp"`
	Details IndexScanDetails `xml:"Object"`
}
type TableScanRelOp struct {
	RelOp   RelOp            `xml:"RelOp"`
	Details TableScanDetails `xml:"Object"`
}

// GetAllChildren returns all child RelOps from all possible locations
func (r *RelOp) GetAllChildren() []RelOp {
	var children []RelOp

	// Add direct RelOps
	children = append(children, r.RelOp...)

	// Add RelOps from wrapper types
	for _, a := range r.Action {
		children = append(children, a)
	}
	for _, t := range r.TopSort {
		children = append(children, t)
	}
	for _, t := range r.Top {
		children = append(children, t)
	}
	for _, n := range r.NestedLoops {
		children = append(children, n.OuterRelOp)
		children = append(children, n.InnerRelOp)
	}
	for _, h := range r.Hash {
		children = append(children, h)
	}
	for _, s := range r.SortWrapper {
		children = append(children, s.RelOp)
	}
	for _, s := range r.IndexScan {
		children = append(children, s.RelOp)
	}
	for _, s := range r.TableScan {
		children = append(children, s.RelOp)
	}
	for _, f := range r.FilterWrapper {
		children = append(children, f)
	}
	for _, a := range r.Aggregate {
		children = append(children, a)
	}
	for _, m := range r.Merge {
		children = append(children, m)
	}
	for _, p := range r.Parallelism {
		children = append(children, p)
	}
	for _, s := range r.StreamAggregate {
		children = append(children, s)
	}
	for _, c := range r.Compute {
		children = append(children, c)
	}
	for _, s := range r.Sequence {
		children = append(children, s)
	}
	for _, s := range r.Segment {
		children = append(children, s)
	}
	for _, s := range r.Spool {
		children = append(children, s)
	}
	for _, r2 := range r.RemoteQuery {
		children = append(children, r2)
	}
	for _, r2 := range r.Remote {
		children = append(children, r2)
	}
	for _, u := range r.Update {
		children = append(children, u)
	}
	for _, d := range r.Delete {
		children = append(children, d)
	}
	for _, i := range r.Insert {
		children = append(children, i)
	}

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

func PlanToProto(handle []byte, server common_domain.ServerMeta, plan *ParsedExecutionPlan) (*dbmv1.ParsedExecutionPlan, error) {
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
			Nodes:         make([]*dbmv1.PlanNode, 0),
		}
		for _, c := range stmt.QueryPlan.RelOp.GetAllChildren() {
			baseNode.Nodes = append(baseNode.Nodes, relOpToProtoNode(c))
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

	baseNode := dbmv1.PlanNode{
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
	return &baseNode
}
