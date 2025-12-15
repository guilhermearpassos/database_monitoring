package parsers_test

import (
	"embed"
	"testing"

	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/adapters/parsers"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"github.com/stretchr/testify/require"
)

//go:embed testdata
var testData embed.FS

func fetchChildNodes(node *dbmv1.PlanNode) []*dbmv1.PlanNode {
	n := make([]*dbmv1.PlanNode, 0)
	n = append(n, node.Nodes...)
	for _, childNode := range node.Nodes {
		n = append(n, fetchChildNodes(childNode)...)
	}
	return n
}
func TestParseExecutionPlan(t *testing.T) {
	tcs := []struct {
		Name      string
		file      string
		nodeCount int
	}{
		{
			Name:      "NestedLoops",
			file:      "testdata/nested_loops_plan.xml",
			nodeCount: 10,
		},
		{
			Name:      "BulkInsert",
			file:      "testdata/bulk_insert.xml",
			nodeCount: 7,
		},
		{
			Name:      "Select with Groupby",
			file:      "testdata/select_groupby.xml",
			nodeCount: 4,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			data, err := testData.ReadFile(tc.file)
			require.NoError(t, err)
			parsed, err := parsers.ParseExecutionPlan(string(data))
			require.NoError(t, err)
			require.NotNil(t, parsed)
			protoParsed, err := parsers.PlanToProto("aaa", common_domain.ServerMeta{Host: "server"}, parsed)
			require.NoError(t, err)
			require.NotNil(t, protoParsed)
			require.NotNil(t, protoParsed.Plan)
			require.Equal(t, "aaa", protoParsed.Plan.PlanHandle)
			require.Equal(t, string(data), protoParsed.Plan.XmlPlan)
			allNodes := make([]*dbmv1.PlanNode, 0)
			nodes := protoParsed.Nodes
			for _, node := range nodes {
				allNodes = append(allNodes, fetchChildNodes(node)...)
				allNodes = append(allNodes, node)
			}
			for _, node := range allNodes {
				require.NotNil(t, node)
				require.NotZero(t, node.NodeCost)
				require.NotZero(t, node.SubtreeCost)
				require.NotZero(t, node.Name)
			}
			require.Equal(t, tc.nodeCount, len(allNodes))
		})
	}
}
