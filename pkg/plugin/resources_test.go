package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	querierv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/querier/v2"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockCallResourceResponseSender struct {
	response *backend.CallResourceResponse
}

func (s *mockCallResourceResponseSender) Send(response *backend.CallResourceResponse) error {
	s.response = response
	return nil
}

type mockQuerierService struct {
	querierv2.UnimplementedQuerierAPIServer
}

func (m *mockQuerierService) ListServerSummary(_ context.Context, _ *querierv2.ListServerSummaryRequest) (*querierv2.ListServerSummaryResponse, error) {
	return &querierv2.ListServerSummaryResponse{
		Servers: []*querierv2.ServerSummary{
			{Name: "localhost", Type: "mssql", Connections: 12, RequestRate: 42.5},
		},
	}, nil
}

func (m *mockQuerierService) ListPlansWithIssues(_ context.Context, _ *querierv2.ListPlansWithIssuesRequest) (*querierv2.ListPlansWithIssuesResponse, error) {
	now := timestamppb.Now()
	return &querierv2.ListPlansWithIssuesResponse{
		Plans: []*querierv2.PlansWithIssues{
			{
				IssueCount:       1,
				RelatedLockCount: 3,
				Occurrences:      12,
				LatestSample: &dbmv1.QuerySample{
					Id:         "sample-1",
					Text:       "SELECT * FROM orders WHERE code = @p1",
					PlanHandle: "plan-1",
					Db:         &dbmv1.DBMetadata{DatabaseName: "orders_db"},
					SnapInfo:   &dbmv1.SnapMetadata{Id: "snap-1", Timestamp: now},
					Context:    &dbmv1.ContextInfo{TraceId: "0123456789abcdef0123456789abcdef"},
				},
				ParsedPlan: &dbmv1.ParsedExecutionPlan{
					Warnings: []*dbmv1.PlanWarning{
						{
							Warning: &dbmv1.PlanWarning_Convert{
								Convert: &dbmv1.PlanWarning_PlanAffectingConvert{
									ConvertIssue: "Seek Plan",
									Expression:   "CONVERT_IMPLICIT(nvarchar(50),[code],0)",
								},
							},
						},
					},
				},
			},
			{
				IssueCount:  1,
				Occurrences: 5,
				LatestSample: &dbmv1.QuerySample{
					Id:         "sample-2",
					Text:       "SELECT * FROM users WHERE email = @p1",
					PlanHandle: "plan-2",
					Db:         &dbmv1.DBMetadata{DatabaseName: "users_db"},
					SnapInfo:   &dbmv1.SnapMetadata{Id: "snap-2", Timestamp: now},
				},
				ParsedPlan: &dbmv1.ParsedExecutionPlan{
					Warnings: []*dbmv1.PlanWarning{
						{
							Warning: &dbmv1.PlanWarning_MissingIndex{
								MissingIndex: &dbmv1.PlanWarning_MissingIndexWarning{
									Database:        "[users_db]",
									Schema:          "[dbo]",
									Table:           "[users]",
									Impact:          84.5,
									EqualityColumns: []string{"[email]"},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func (m *mockQuerierService) ListSnapshotSummaries(_ context.Context, _ *querierv2.ListSnapshotSummariesRequest) (*querierv2.ListSnapshotSummariesResponse, error) {
	return &querierv2.ListSnapshotSummariesResponse{
		SnapSummaries: []*querierv2.SnapshotSummary{
			{
				Id:                     "snap-1",
				Timestamp:              timestamppb.Now(),
				Connections:            4,
				Waiters:                1,
				Blockers:               1,
				WaitDuration:           1500,
				ConnectionsByWaitEvent: map[string]int64{"LCK_M_X": 1, "": 3},
				TimeMsByWaitEvent:      map[string]int64{"LCK_M_X": 1500},
			},
		},
	}, nil
}

func (m *mockQuerierService) GetSnapshot(_ context.Context, req *querierv2.GetSnapshotRequest) (*querierv2.GetSnapshotResponse, error) {
	return &querierv2.GetSnapshotResponse{
		Snapshot: &dbmv1.DBSnapshot{
			Id: req.GetId(),
			Samples: []*dbmv1.QuerySample{
				{
					Id:                "sample-1",
					Text:              "UPDATE orders SET status = 1",
					Status:            "running",
					Blocker:           true,
					TimeElapsedMillis: 2500,
					Db:                &dbmv1.DBMetadata{DatabaseName: "orders_db"},
					Session:           &dbmv1.SessionMetadata{SessionId: "51", LoginName: "app_user"},
					BlockInfo:         &dbmv1.BlockMetadata{BlockedSessions: []string{"52"}},
					SnapInfo:          &dbmv1.SnapMetadata{Id: req.GetId()},
					Context:           &dbmv1.ContextInfo{TraceId: "abcdef0123456789abcdef0123456789"},
				},
			},
		},
	}, nil
}

func (m *mockQuerierService) ListQueryMetrics(_ context.Context, _ *querierv2.ListQueryMetricsRequest) (*querierv2.ListQueryMetricsResponse, error) {
	now := timestamppb.Now()
	return &querierv2.ListQueryMetricsResponse{
		Metrics: []*dbmv1.QueryMetric{
			{
				QueryHash:             "hash-spill",
				Text:                  "SELECT * FROM orders ORDER BY created_at",
				Db:                    &dbmv1.DBMetadata{DatabaseName: "orders_db"},
				LastExecutionTime:     now,
				LastElapsedTimeMicros: 250000,
				Counters: map[string]int64{
					"executionCount":     10,
					"totalElapsedTime":   2500000, // 2500ms total -> 250ms avg (slow_query)
					"totalWorkerTime":    1200000, // 1200ms total -> 120ms avg (high_cpu)
					"totalLogicalReads":  25000,   // 2500 avg (high_reads)
					"totalPhysicalReads": 600,     // 60 avg (physical_io)
					"totalSpills":        4,       // spills
					"totalGrantKb":       204800,  // 20480KB (20MB) avg
					"totalUsedGrantKb":   10240,   // 1024KB (1MB) avg -> 95% waste (memory_overgrant)
					"totalRows":          1000,
				},
			},
			{
				QueryHash:         "hash-zero",
				Text:              "SELECT 1",
				Db:                &dbmv1.DBMetadata{DatabaseName: "users_db"},
				LastExecutionTime: now,
				Counters: map[string]int64{
					"executionCount": 0,
				},
			},
		},
	}, nil
}

func (m *mockQuerierService) GetQueryMetricsTimeSeries(_ context.Context, req *querierv2.GetQueryMetricsTimeSeriesRequest) (*querierv2.GetQueryMetricsTimeSeriesResponse, error) {
	t0 := time.Now().Add(-10 * time.Minute)
	return &querierv2.GetQueryMetricsTimeSeriesResponse{
		Metrics: []*dbmv1.QueryMetric{
			{
				QueryHash:   req.GetSqlHandle(),
				CollectedAt: timestamppb.New(t0),
				Counters: map[string]int64{
					"executionCount":    4,
					"totalElapsedTime":  800000,
					"totalWorkerTime":   400000,
					"totalLogicalReads": 4000,
					"totalSpills":       2,
				},
			},
			{
				QueryHash:   req.GetSqlHandle(),
				CollectedAt: timestamppb.New(t0.Add(5 * time.Minute)),
				Counters: map[string]int64{
					"executionCount": 0,
				},
			},
		},
	}, nil
}

var testServerAddress string

func TestMain(m *testing.M) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	testServerAddress = fmt.Sprintf("127.0.0.1:%d", lis.Addr().(*net.TCPAddr).Port)

	grpcServer := grpc.NewServer()
	querierv2.RegisterQuerierAPIServer(grpcServer, &mockQuerierService{})

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			panic(err)
		}
	}()

	code := m.Run()
	grpcServer.Stop()
	os.Exit(code)
}

func TestCallResourceAndHealth(t *testing.T) {
	data := map[string]interface{}{
		"apiUrl": testServerAddress,
	}
	marshal, err := json.Marshal(data)
	require.NoError(t, err)
	inst, err := NewApp(context.Background(), backend.AppInstanceSettings{
		JSONData: marshal,
		Updated:  time.Time{},
	})
	require.NoError(t, err)
	require.NotNil(t, inst)

	app, ok := inst.(*App)
	require.True(t, ok)
	defer app.Dispose()

	health, err := app.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	require.Equal(t, backend.HealthStatusOk, health.Status)

	for _, tc := range []struct {
		name      string
		method    string
		path      string
		body      []byte
		expStatus int
		expBody   []byte
	}{
		{
			name:      "get ping 200",
			method:    http.MethodGet,
			path:      "ping",
			expStatus: http.StatusOK,
		},
		{
			name:      "get echo 405",
			method:    http.MethodGet,
			path:      "echo",
			expStatus: http.StatusMethodNotAllowed,
		},
		{
			name:      "post echo 200",
			method:    http.MethodPost,
			path:      "echo",
			body:      []byte(`{"message":"ok"}`),
			expStatus: http.StatusOK,
			expBody:   []byte(`{"message":"ok"}`),
		},
		{
			name:      "getQueryDetails missing snapId 400",
			method:    http.MethodGet,
			path:      "getQueryDetails?sampleId=abc",
			expStatus: http.StatusBadRequest,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var r mockCallResourceResponseSender
			err = app.CallResource(context.Background(), &backend.CallResourceRequest{
				Method: tc.method,
				Path:   tc.path,
				Body:   tc.body,
			}, &r)
			require.NoError(t, err)
			require.NotNil(t, r.response)
			if tc.expStatus > 0 {
				require.Equal(t, tc.expStatus, r.response.Status)
			}
			if len(tc.expBody) > 0 {
				require.Equal(t, tc.expBody, bytes.TrimSpace(r.response.Body))
			}
		})
	}

	t.Run("planAnalysis filters single database and seek_plan_convert in memory", func(t *testing.T) {
		var r mockCallResourceResponseSender
		err := app.CallResource(context.Background(), &backend.CallResourceRequest{
			Method: http.MethodGet,
			URL:    "planAnalysis?server=localhost&databases=orders_db&issueTags=seek_plan_convert",
			Path:   "planAnalysis",
		}, &r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, r.response.Status)

		var parsed PlanAnalysisResponse
		require.NoError(t, json.Unmarshal(r.response.Body, &parsed))
		require.Equal(t, []string{"orders_db", "users_db"}, parsed.AvailableDatabases)
		require.Len(t, parsed.Plans, 1)
		require.Equal(t, "orders_db", parsed.Plans[0].Database)
		require.Contains(t, parsed.Plans[0].IssueTags, "seek_plan_convert")
		require.Len(t, parsed.Plans[0].Converts, 1)
		require.Equal(t, "Seek Plan", parsed.Plans[0].Converts[0].ConvertIssue)
	})

	t.Run("snapshotSummaries and snapshotSamples return typed JSON", func(t *testing.T) {
		var r1 mockCallResourceResponseSender
		err := app.CallResource(context.Background(), &backend.CallResourceRequest{
			Method: http.MethodGet,
			URL:    "snapshotSummaries?server=localhost",
			Path:   "snapshotSummaries",
		}, &r1)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, r1.response.Status)
		var sums []SnapshotSummaryDTO
		require.NoError(t, json.Unmarshal(r1.response.Body, &sums))
		require.Len(t, sums, 1)
		require.Equal(t, int32(4), sums[0].Connections)

		var r2 mockCallResourceResponseSender
		err = app.CallResource(context.Background(), &backend.CallResourceRequest{
			Method: http.MethodGet,
			URL:    "snapshotSamples?snapId=snap-1",
			Path:   "snapshotSamples",
		}, &r2)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, r2.response.Status)
		var samples []SnapshotSampleDTO
		require.NoError(t, json.Unmarshal(r2.response.Body, &samples))
		require.Len(t, samples, 1)
		require.Equal(t, "abcdef0123456789abcdef0123456789", samples[0].TraceID)
		require.True(t, samples[0].IsBlocker)
	})

	t.Run("QueryData supports blocking_timeseries, wait_breakdown, plan_analysis_summary, and servers", func(t *testing.T) {
		now := time.Now()
		tr := backend.TimeRange{From: now.Add(-1 * time.Hour), To: now}

		resp, err := app.QueryData(context.Background(), &backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{RefID: "A", QueryType: "blocking_timeseries", TimeRange: tr, JSON: []byte(`{"database":"localhost","blockingSeries":["blockers","waiters"]}`)},
				{RefID: "B", QueryType: "wait_breakdown", TimeRange: tr, JSON: []byte(`{"database":"localhost","metric":"wait_time_ms"}`)},
				{RefID: "C", QueryType: "plan_analysis_summary", TimeRange: tr, JSON: []byte(`{"database":"localhost","onlyLocks":true}`)},
				{RefID: "D", QueryType: "servers", TimeRange: tr, JSON: []byte(`{}`)},
			},
		})
		require.NoError(t, err)

		require.NoError(t, resp.Responses["A"].Error)
		require.Len(t, resp.Responses["A"].Frames, 1)
		require.Len(t, resp.Responses["A"].Frames[0].Fields, 3) // time + blockers + waiters

		require.NoError(t, resp.Responses["B"].Error)
		require.Len(t, resp.Responses["B"].Frames, 1)
		require.Equal(t, 2, resp.Responses["B"].Frames[0].Rows()) // LCK_M_X + cpu

		require.NoError(t, resp.Responses["C"].Error)
		require.Len(t, resp.Responses["C"].Frames, 1)
		require.Equal(t, float64(1), resp.Responses["C"].Frames[0].Fields[0].At(0)) // totalPlans with locks == 1
		require.Equal(t, float64(1), resp.Responses["C"].Frames[0].Fields[1].At(0)) // seekPlanConversions == 1

		require.NoError(t, resp.Responses["D"].Error)
		require.Len(t, resp.Responses["D"].Frames, 1)
		require.Equal(t, 1, resp.Responses["D"].Frames[0].Rows())
		require.Equal(t, "localhost", resp.Responses["D"].Frames[0].Fields[0].At(0))
	})

	t.Run("queryMetrics and queryMetricsSeries return computed flags and time series", func(t *testing.T) {
		var r1 mockCallResourceResponseSender
		err := app.CallResource(context.Background(), &backend.CallResourceRequest{
			Method: http.MethodGet,
			URL:    "queryMetrics?server=localhost&databases=orders_db",
			Path:   "queryMetrics",
		}, &r1)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, r1.response.Status)

		var parsed QueryMetricsResponse
		require.NoError(t, json.Unmarshal(r1.response.Body, &parsed))
		require.Equal(t, []string{"orders_db", "users_db"}, parsed.AvailableDatabases)
		require.Len(t, parsed.Metrics, 1)
		m := parsed.Metrics[0]
		require.Equal(t, "hash-spill", m.QueryHash)
		require.InDelta(t, 250.0, m.AvgElapsedMs, 0.01)
		require.InDelta(t, 120.0, m.AvgWorkerMs, 0.01)
		require.Equal(t, []string{"spills", "memory_overgrant"}, m.IssueFlags)

		var r2 mockCallResourceResponseSender
		err = app.CallResource(context.Background(), &backend.CallResourceRequest{
			Method: http.MethodGet,
			URL:    "queryMetricsSeries?server=localhost&queryHash=hash-spill",
			Path:   "queryMetricsSeries",
		}, &r2)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, r2.response.Status)

		var pts []QueryMetricsSeriesPointDTO
		require.NoError(t, json.Unmarshal(r2.response.Body, &pts))
		require.Len(t, pts, 2)
		require.Equal(t, int64(4), pts[0].ExecutionCount)
		require.InDelta(t, 200.0, pts[0].AvgElapsedMs, 0.01)

		// Also verify QueryData("metrics") and QueryData("metrics_series") have uniform field lengths even with 0-exec rows
		now := time.Now()
		tr := backend.TimeRange{From: now.Add(-1 * time.Hour), To: now}
		qdResp, err := app.QueryData(context.Background(), &backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{RefID: "M1", QueryType: "metrics", TimeRange: tr, JSON: []byte(`{"database":"localhost"}`)},
				{RefID: "M2", QueryType: "metrics_series", TimeRange: tr, JSON: []byte(`{"database":"localhost","queryHash":"hash-spill"}`)},
			},
		})
		require.NoError(t, err)
		for _, ref := range []string{"M1", "M2"} {
			require.NoError(t, qdResp.Responses[ref].Error)
			frame := qdResp.Responses[ref].Frames[0]
			expectedRows := frame.Fields[0].Len()
			require.Equal(t, 2, expectedRows)
			for _, f := range frame.Fields {
				require.Equal(t, expectedRows, f.Len(), "field %s length mismatch in %s", f.Name, ref)
			}
		}
	})

	t.Run("datasource-options supports server-databases, snapshots, and issue-tags", func(t *testing.T) {
		for _, tc := range []struct {
			url      string
			minCount int
		}{
			{"datasource-options?type=server-databases&server=localhost", 2},
			{"datasource-options?type=snapshots&server=localhost", 1},
			{"datasource-options?type=issue-tags", 5},
		} {
			var r mockCallResourceResponseSender
			err := app.CallResource(context.Background(), &backend.CallResourceRequest{
				Method: http.MethodGet,
				URL:    tc.url,
				Path:   "datasource-options",
			}, &r)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, r.response.Status)
			var opts []DropdownOption
			require.NoError(t, json.Unmarshal(r.response.Body, &opts))
			require.GreaterOrEqual(t, len(opts), tc.minCount)
		}
	})
}
