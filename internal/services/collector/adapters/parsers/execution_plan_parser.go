package parsers

import (
	"encoding/xml"
	"fmt"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
)

// ParsedExecutionPlan represents the root of the execution plan XML
type ParsedExecutionPlan struct {
	XMLName xml.Name `xml:"ShowPlanXML"`
	Version string   `xml:"Version,attr"`
	Batches []Batch  `xml:"BatchSequence>Batch"`
	Raw     string
}

// Batch represents a batch of statements
type Batch struct {
	Statements Statements `xml:"Statements"`
}

// Statements can contain both simple statements and conditional statements (like while loops)
type Statements struct {
	StmtSimple []StmtSimple `xml:"StmtSimple"`
	StmtCond   []StmtCond   `xml:"StmtCond"`
}

// StmtSimple represents a simple SQL statement
type StmtSimple struct {
	StatementText        string     `xml:"StatementText,attr"`
	StatementId          string     `xml:"StatementId,attr"`
	StatementType        string     `xml:"StatementType,attr"`
	StatementSubTreeCost float64    `xml:"StatementSubTreeCost,attr"`
	StatementEstRows     float64    `xml:"StatementEstRows,attr"`
	QueryPlan            *QueryPlan `xml:"QueryPlan"` // Pointer because not all statements have query plans
}

// StmtCond represents a conditional statement (like while loops, if statements)
type StmtCond struct {
	StatementText string     `xml:"StatementText,attr"`
	StatementId   string     `xml:"StatementId,attr"`
	StatementType string     `xml:"StatementType,attr"`
	Condition     *Condition `xml:"Condition"`
	Then          *Then      `xml:"Then"`
	Else          *Else      `xml:"Else"` // For if/else statements
}

// Condition represents the condition part of a conditional statement
type Condition struct {
	// The condition might have a query plan if it's complex
	QueryPlan *QueryPlan `xml:"QueryPlan"`
}

// Then represents the "then" branch of a conditional statement
type Then struct {
	Statements Statements `xml:"Statements"`
}

// Else represents the "else" branch of a conditional statement
type Else struct {
	Statements Statements `xml:"Statements"`
}

// QueryPlan contains the actual execution plan
type QueryPlan struct {
	RelOp                RelOp                  `xml:"RelOp"`
	PlanAffectingConvert []PlanAffectingConvert `xml:"Warnings>PlanAffectingConvert"`
	MissingIndexes       *MissingIndexes        `xml:"MissingIndexes"`
	OptimizerStatsUsage  struct {
		StatisticsInfo []StatisticsInfo `xml:"StatisticsInfo"`
	} `xml:"OptimizerStatsUsage"`
}

type PlanAffectingConvert struct {
	XMLName      xml.Name `xml:"PlanAffectingConvert"`
	ConvertIssue string   `xml:"ConvertIssue,attr"`
	Expression   string   `xml:"Expression,attr"`
}

// MissingIndexes contains information about indexes that could improve performance
type MissingIndexes struct {
	MissingIndexGroups []MissingIndexGroup `xml:"MissingIndexGroup"`
}

// MissingIndexGroup represents a group of missing indexes with their impact
type MissingIndexGroup struct {
	Impact       float64      `xml:"Impact,attr"`
	MissingIndex MissingIndex `xml:"MissingIndex"`
}

// MissingIndex contains details about a specific missing index
type MissingIndex struct {
	Database     string        `xml:"Database,attr"`
	Schema       string        `xml:"Schema,attr"`
	Table        string        `xml:"Table,attr"`
	ColumnGroups []ColumnGroup `xml:"ColumnGroup"`
}

// ColumnGroup represents a group of columns (EQUALITY, INEQUALITY, or INCLUDE)
type ColumnGroup struct {
	Usage   string   `xml:"Usage,attr"`
	Columns []Column `xml:"Column"`
}

// Column represents a column in a missing index
type Column struct {
	Name     string `xml:"Name,attr"`
	ColumnId string `xml:"ColumnId,attr"`
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

// Helper function to recursively collect all statements (including those in loops)
func collectAllStatements(statements Statements) []StmtSimple {
	var result []StmtSimple

	// Add simple statements
	result = append(result, statements.StmtSimple...)

	// Recursively process conditional statements
	for _, cond := range statements.StmtCond {
		if cond.Then != nil {
			result = append(result, collectAllStatements(cond.Then.Statements)...)
		}
		if cond.Else != nil {
			result = append(result, collectAllStatements(cond.Else.Statements)...)
		}
	}

	return result
}

func PlanToProto(handle string, server common_domain.ServerMeta, plan *ParsedExecutionPlan) (*dbmv1.ParsedExecutionPlan, error) {
	stats := make([]*dbmv1.StatisticsInfo, 0)
	warnings := make([]*dbmv1.PlanWarning, 0)
	nodes := make([]*dbmv1.PlanNode, 0)

	// Process all batches
	for _, batch := range plan.Batches {
		// Collect all statements, including those nested in loops
		allStatements := collectAllStatements(batch.Statements)

		for _, stmt := range allStatements {
			// Skip statements without query plans (like ASSIGN, WAITFOR, etc.)
			if stmt.QueryPlan == nil {
				continue
			}

			// Collect statistics
			for _, stat := range stmt.QueryPlan.OptimizerStatsUsage.StatisticsInfo {
				stats = append(stats, &dbmv1.StatisticsInfo{
					LastUpdate:        stat.LastUpdate,
					ModificationCount: int64(stat.ModificationCount),
					SamplingPercent:   stat.SamplingPercent,
					Statistics:        stat.Statistics,
					Table:             stat.Table,
				})
			}

			// Collect warnings
			for _, warning := range stmt.QueryPlan.PlanAffectingConvert {
				warnings = append(warnings, &dbmv1.PlanWarning{
					Warning: &dbmv1.PlanWarning_Convert{Convert: &dbmv1.PlanWarning_PlanAffectingConvert{
						ConvertIssue: warning.ConvertIssue,
						Expression:   warning.Expression,
					}},
				})
			}

			// Collect Missing Index warnings
			if stmt.QueryPlan.MissingIndexes != nil {
				for _, missingIndexGroup := range stmt.QueryPlan.MissingIndexes.MissingIndexGroups {
					// Build column group information
					var equalityColumns, inequalityColumns, includeColumns []string

					for _, colGroup := range missingIndexGroup.MissingIndex.ColumnGroups {
						for _, col := range colGroup.Columns {
							switch colGroup.Usage {
							case "EQUALITY":
								equalityColumns = append(equalityColumns, col.Name)
							case "INEQUALITY":
								inequalityColumns = append(inequalityColumns, col.Name)
							case "INCLUDE":
								includeColumns = append(includeColumns, col.Name)
							}
						}
					}

					warnings = append(warnings, &dbmv1.PlanWarning{
						Warning: &dbmv1.PlanWarning_MissingIndex{MissingIndex: &dbmv1.PlanWarning_MissingIndexWarning{
							Database:          missingIndexGroup.MissingIndex.Database,
							Schema:            missingIndexGroup.MissingIndex.Schema,
							Table:             missingIndexGroup.MissingIndex.Table,
							Impact:            missingIndexGroup.Impact,
							EqualityColumns:   equalityColumns,
							InequalityColumns: inequalityColumns,
							IncludeColumns:    includeColumns,
						}},
					})
				}
			}
			// Build node tree
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
