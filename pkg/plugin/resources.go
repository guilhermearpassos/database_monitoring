package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	querierv2 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/querier/v2"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// handlePing is an example HTTP GET resource that returns a {"message": "ok"} JSON response.
func (a *App) handlePing(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write([]byte(`{"message": "ok"}`)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleEcho is an example HTTP POST resource that accepts a JSON with a "message" key and
// returns to the client whatever it is sent.
func (a *App) handleEcho(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// registerRoutes takes a *http.ServeMux and registers some HTTP handlers.
func (a *App) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/ping", a.handlePing)
	mux.HandleFunc("/", a.handlePing)
	mux.HandleFunc("/echo", a.handleEcho)
	mux.HandleFunc("/datasource-options", a.handleDropdownOptions)
	mux.HandleFunc("/query", a.handleQuery)
	mux.HandleFunc("/getQueryDetails", a.handleFetchQueryDetails)
	mux.HandleFunc("/planAnalysis", a.handlePlanAnalysis)
	mux.HandleFunc("/snapshotSummaries", a.handleSnapshotSummaries)
	mux.HandleFunc("/snapshotSamples", a.handleSnapshotSamples)
}

func (a *App) handleFetchQueryDetails(w http.ResponseWriter, r *http.Request) {
	sampleID := r.URL.Query().Get("sampleId")
	snapID := r.URL.Query().Get("snapId")
	if sampleID == "" {
		http.Error(w, "sampleId is required", http.StatusBadRequest)
		return
	}
	if snapID == "" {
		http.Error(w, "snapId is required", http.StatusBadRequest)
		return
	}
	resp, err := a.client.GetSampleDetails(r.Context(), &querierv2.GetSampleDetailsRequest{
		SampleId: sampleID,
		SnapId:   snapID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var plan ParsedExecutionPlan
	if resp.ParsedPlan != nil {
		plan = protoParsedPlanToDomain(resp.ParsedPlan)
	}
	blockChain, err2 := blockChainFromProto(resp.BlockChain)
	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		return
	}
	sample, err := protoSampleToDomain(resp.QuerySample)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	m, err := json.Marshal(struct {
		Plan   ParsedExecutionPlan `json:"plan"`
		Chain  BlockChain          `json:"blocking_chain"`
		Sample QuerySample         `json:"query_sample"`
	}{
		Plan:   plan,
		Chain:  blockChain,
		Sample: sample,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(m)
}

func blockChainFromProto(chain *querierv2.BlockChain) (BlockChain, error) {
	if chain == nil {
		return BlockChain{Roots: []BlockingNode{}}, nil
	}
	roots := make([]BlockingNode, len(chain.Roots))
	for i, root := range chain.Roots {
		node, err2 := nodeFromProto(root, i)
		if err2 != nil {
			return BlockChain{}, err2
		}
		roots[i] = node
	}
	return BlockChain{
		Roots: roots,
	}, nil
}

func protoSampleToDomain(sample *dbmv1.QuerySample) (QuerySample, error) {
	if sample == nil {
		return QuerySample{}, fmt.Errorf("nil query sample")
	}
	var blockTime, blockDetails string
	if sample.Blocker && sample.BlockInfo != nil {
		blockDetails = fmt.Sprintf("%d block waiters", len(sample.BlockInfo.BlockedSessions))
		if sample.Blocked {
			blockDetails += " | "
		}
	}
	if sample.Blocked {
		if sample.WaitInfo != nil {
			blockTime = time.Time{}.Add(time.Duration(sample.WaitInfo.WaitTime * 1_000_000_000)).Format(time.TimeOnly)
		}
		if sample.BlockInfo != nil {
			blockDetails += fmt.Sprintf("blocked by %s", sample.BlockInfo.BlockedBy)
		}
	}
	var sid int
	if sample.GetSession().GetSessionId() != "" {
		var err2 error
		sid, err2 = strconv.Atoi(sample.GetSession().GetSessionId())
		if err2 != nil {
			return QuerySample{}, err2
		}
	}

	dSample := QuerySample{
		SID:                     sid,
		Query:                   sample.Text,
		ExecutionTime:           fmt.Sprintf("%d ms", sample.TimeElapsedMillis),
		User:                    sample.GetSession().GetLoginName(),
		IsBlocker:               sample.Blocker,
		IsWaiter:                sample.Blocked,
		BlockingTime:            blockTime,
		BlockDetails:            blockDetails,
		WaitEvent:               sample.GetWaitInfo().GetWaitType(),
		Database:                sample.GetDb().GetDatabaseName(),
		SampleID:                sample.Id,
		SnapID:                  sample.GetSnapInfo().GetId(),
		SQLHandle:               sample.SqlHandle,
		PlanHandle:              sample.PlanHandle,
		Status:                  sample.Status,
		QueryHash:               sample.QueryHash,
		SessionLoginTime:        sample.GetSession().GetLoginTime().AsTime(),
		SessionHost:             sample.GetSession().GetHost(),
		SessionClientIp:         sample.GetSession().GetClientIp(),
		SessionStatus:           sample.GetSession().GetStatus(),
		SessionProgramName:      sample.GetSession().GetProgramName(),
		SessionLastRequestStart: sample.GetSession().GetLastRequestStart().AsTime(),
		SessionLastRequestEnd:   sample.GetSession().GetLastRequestEnd().AsTime(),
		TraceID:                 sample.GetContext().GetTraceId(),
	}
	return dSample, nil
}

func nodeFromProto(root *querierv2.BlockChain_BlockingNode, i int) (BlockingNode, error) {
	ds, err := protoSampleToDomain(root.QuerySample)
	if err != nil {
		return BlockingNode{}, err
	}
	childNodes := make([]BlockingNode, len(root.ChildNodes))
	for j, child := range root.ChildNodes {
		cn, err2 := nodeFromProto(child, i+1)
		if err2 != nil {
			return BlockingNode{}, err2
		}
		childNodes[j] = cn
	}
	node := BlockingNode{
		QuerySample: ds,
		ChildNodes:  childNodes,
		Level:       i,
	}
	return node, nil
}

// Add this method to handle queries from nested datasource
func (a *App) handleQuery(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body to match Grafana's query format
	var queryReq struct {
		Queries       []json.RawMessage `json:"queries"`
		Range         backend.TimeRange `json:"range"`
		IntervalMs    int64             `json:"intervalMs"`
		MaxDataPoints int64             `json:"maxDataPoints"`
	}

	if err := json.NewDecoder(req.Body).Decode(&queryReq); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Create a backend query request
	backendReq := &backend.QueryDataRequest{
		Queries: make([]backend.DataQuery, len(queryReq.Queries)),
	}

	// Convert each query
	for i, rawQuery := range queryReq.Queries {
		var query map[string]interface{}
		if err := json.Unmarshal(rawQuery, &query); err != nil {
			http.Error(w, fmt.Sprintf("invalid query format: %v", err), http.StatusBadRequest)
			return
		}

		refID, _ := query["refId"].(string)
		if refID == "" {
			refID = fmt.Sprintf("A%d", i)
		}
		queryType, _ := query["queryType"].(string)

		backendReq.Queries[i] = backend.DataQuery{
			RefID:     refID,
			JSON:      rawQuery,
			TimeRange: queryReq.Range,
			QueryType: queryType,
		}
	}

	// Call your existing QueryData method
	resp, err := a.QueryData(req.Context(), backendReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("query failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert response to the format expected by frontend
	result := map[string]interface{}{
		"data": make([]interface{}, 0),
	}

	for _, dataResp := range resp.Responses {
		if dataResp.Error != nil {
			http.Error(w, fmt.Sprintf("query error: %v", dataResp.Error), http.StatusInternalServerError)
			return
		}

		for _, frame := range dataResp.Frames {
			// Convert frame to JSON format
			frameJSON, err := frame.MarshalJSON()
			if err != nil {
				continue
			}

			var frameData interface{}
			if err := json.Unmarshal(frameJSON, &frameData); err != nil {
				continue
			}

			result["data"] = append(result["data"].([]interface{}), frameData)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// QueryData handles data source queries
func (a *App) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	// Process each query in the request
	for _, q := range req.Queries {
		switch q.QueryType {
		case "chart", "":
			res := a.query(ctx, req.PluginContext, q)
			response.Responses[q.RefID] = res
		case "blocking_timeseries":
			res := a.queryBlockingTimeSeries(ctx, req.PluginContext, q)
			response.Responses[q.RefID] = res
		case "wait_breakdown":
			res := a.queryWaitBreakdown(ctx, req.PluginContext, q)
			response.Responses[q.RefID] = res
		case "snapshot-list":
			res := a.querySnapList(ctx, req.PluginContext, q)
			response.Responses[q.RefID] = res
		case "snapshot":
			res := a.querySnap(ctx, req.PluginContext, q)
			response.Responses[q.RefID] = res
		case "metrics":
			res := a.queryMetrics(ctx, req.PluginContext, q)
			response.Responses[q.RefID] = res
		case "metrics_series":
			res := a.queryMetricsTimeSeries(ctx, req.PluginContext, q)
			response.Responses[q.RefID] = res
		case "plan_analysis":
			res := a.queryPlanAnalysis(ctx, req.PluginContext, q)
			response.Responses[q.RefID] = res
		case "plan_analysis_summary":
			res := a.queryPlanAnalysisSummary(ctx, req.PluginContext, q)
			response.Responses[q.RefID] = res
		case "servers":
			res := a.queryServers(ctx, req.PluginContext, q)
			response.Responses[q.RefID] = res
		}
	}

	return response, nil
}

// SubscribeStream handles streaming data
func (a *App) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	return &backend.SubscribeStreamResponse{
		Status: backend.SubscribeStreamStatusPermissionDenied,
	}, nil
}

// RunStream handles running streams
func (a *App) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	return nil
}

// PublishStream handles stream publishing
func (a *App) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

// query processes individual queries
func (a *App) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}
	q := struct {
		Database  string `json:"database"`
		Metric    string `json:"metric"`
		WaitEvent string `json:"waitEvent"`
	}{}
	if err := json.Unmarshal(query.JSON, &q); err != nil {
		response.Error = err
		return response
	}
	timeRange := query.TimeRange
	from := timeRange.From
	to := timeRange.To
	r, err := a.client.ListSnapshotSummaries(ctx, &querierv2.ListSnapshotSummariesRequest{
		Start:  timestamppb.New(from),
		End:    timestamppb.New(to),
		Server: q.Database,
	})
	if err != nil {
		response.Error = err
		return response
	}
	interval := time.Minute * 1 // 1-minute intervals
	points := int(to.Sub(from)/interval) + 1
	if points < 1 {
		points = 1
	}

	times := make([]time.Time, points)
	valuesByWaitType := make(map[string][]float64, points)
	for i := 0; i < points; i++ {
		times[i] = from.Add(time.Duration(i) * interval)
	}
	useWaitTime := q.Metric == "wait_time_ms"
	wantWait := strings.TrimSpace(q.WaitEvent)
	for _, sum := range r.GetSnapSummaries() {
		idx := int(sum.Timestamp.AsTime().Sub(from).Minutes())
		if idx < 0 || idx >= points {
			continue
		}
		source := sum.GetConnectionsByWaitEvent()
		if useWaitTime {
			source = sum.GetTimeMsByWaitEvent()
		}
		for we, c := range source {
			if we == "" {
				we = "cpu"
			}
			if wantWait != "" && !strings.EqualFold(we, wantWait) {
				continue
			}
			if _, ok := valuesByWaitType[we]; !ok {
				valuesByWaitType[we] = make([]float64, points)
			}
			valuesByWaitType[we][idx] = float64(c)
		}
	}
	frameName := "connections by wait event"
	if useWaitTime {
		frameName = "wait time (ms) by wait event"
	}
	frame := data.NewFrame(frameName,
		data.NewField("time", nil, times),
	)
	keys := make([]string, 0, len(valuesByWaitType))
	for we := range valuesByWaitType {
		keys = append(keys, we)
	}
	slices.Sort(keys)
	for _, we := range keys {
		f := data.NewField(we, nil, valuesByWaitType[we])
		if useWaitTime {
			f.Config = &data.FieldConfig{Unit: "ms"}
		}
		frame.Fields = append(frame.Fields, f)
	}

	frame.RefID = query.RefID
	frame.Meta = &data.FrameMeta{
		Type:                       data.FrameTypeTimeSeriesWide,
		PreferredVisualization:     data.VisTypeGraph,
	}

	response.Frames = append(response.Frames, frame)
	return response
}

func (a *App) queryBlockingTimeSeries(ctx context.Context, _ backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}
	q := struct {
		Database       string   `json:"database"`
		BlockingSeries []string `json:"blockingSeries"`
	}{}
	if err := json.Unmarshal(query.JSON, &q); err != nil {
		response.Error = err
		return response
	}
	r, err := a.client.ListSnapshotSummaries(ctx, &querierv2.ListSnapshotSummariesRequest{
		Start:  timestamppb.New(query.TimeRange.From),
		End:    timestamppb.New(query.TimeRange.To),
		Server: q.Database,
	})
	if err != nil {
		response.Error = err
		return response
	}
	summaries := slices.Clone(r.GetSnapSummaries())
	slices.SortFunc(summaries, func(a, b *querierv2.SnapshotSummary) int {
		at, bt := a.GetTimestamp().AsTime(), b.GetTimestamp().AsTime()
		if at.Before(bt) {
			return -1
		}
		if at.After(bt) {
			return 1
		}
		return 0
	})

	size := len(summaries)
	times := make([]time.Time, 0, size)
	connections := make([]float64, 0, size)
	blockers := make([]float64, 0, size)
	waiters := make([]float64, 0, size)
	waitDurationSec := make([]float64, 0, size)
	for _, sum := range summaries {
		times = append(times, sum.GetTimestamp().AsTime())
		connections = append(connections, float64(sum.GetConnections()))
		blockers = append(blockers, float64(sum.GetBlockers()))
		waiters = append(waiters, float64(sum.GetWaiters()))
		waitDurationSec = append(waitDurationSec, sum.GetWaitDuration()/1000.0)
	}

	include := func(name string) bool {
		return len(q.BlockingSeries) == 0 || slices.Contains(q.BlockingSeries, name)
	}

	frame := data.NewFrame("connections_and_blocking",
		data.NewField("time", nil, times),
	)
	if include("connections") {
		frame.Fields = append(frame.Fields, data.NewField("connections", nil, connections))
	}
	if include("blockers") {
		frame.Fields = append(frame.Fields, data.NewField("blockers", nil, blockers))
	}
	if include("waiters") {
		frame.Fields = append(frame.Fields, data.NewField("waiters", nil, waiters))
	}
	if include("waitDurationSec") {
		f := data.NewField("waitDurationSec", nil, waitDurationSec)
		f.Config = &data.FieldConfig{Unit: "s"}
		frame.Fields = append(frame.Fields, f)
	}

	frame.RefID = query.RefID
	frame.Meta = &data.FrameMeta{
		Type:                   data.FrameTypeTimeSeriesWide,
		PreferredVisualization: data.VisTypeGraph,
	}
	response.Frames = append(response.Frames, frame)
	return response
}

func (a *App) queryWaitBreakdown(ctx context.Context, _ backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}
	q := struct {
		Database string `json:"database"`
		Metric   string `json:"metric"`
	}{}
	if err := json.Unmarshal(query.JSON, &q); err != nil {
		response.Error = err
		return response
	}
	r, err := a.client.ListSnapshotSummaries(ctx, &querierv2.ListSnapshotSummariesRequest{
		Start:  timestamppb.New(query.TimeRange.From),
		End:    timestamppb.New(query.TimeRange.To),
		Server: q.Database,
	})
	if err != nil {
		response.Error = err
		return response
	}

	connTotals := make(map[string]int64)
	timeTotals := make(map[string]int64)
	var grandConns int64
	var grandTime int64
	for _, sum := range r.GetSnapSummaries() {
		for rawWait, c := range sum.GetConnectionsByWaitEvent() {
			k := rawWait
			if k == "" {
				k = "cpu"
			}
			connTotals[k] += c
			grandConns += c
		}
		for rawWait, ms := range sum.GetTimeMsByWaitEvent() {
			k := rawWait
			if k == "" {
				k = "cpu"
			}
			timeTotals[k] += ms
			grandTime += ms
		}
	}

	type waitRow struct {
		name  string
		conns int64
		ms    int64
	}
	rows := make([]waitRow, 0, len(connTotals))
	seen := make(map[string]struct{}, len(connTotals))
	for k, c := range connTotals {
		seen[k] = struct{}{}
		rows = append(rows, waitRow{name: k, conns: c, ms: timeTotals[k]})
	}
	for k, ms := range timeTotals {
		if _, ok := seen[k]; !ok {
			rows = append(rows, waitRow{name: k, conns: 0, ms: ms})
		}
	}
	useWaitTime := q.Metric == "wait_time_ms"
	slices.SortFunc(rows, func(a, b waitRow) int {
		if useWaitTime {
			if a.ms > b.ms {
				return -1
			}
			if a.ms < b.ms {
				return 1
			}
		}
		if a.conns > b.conns {
			return -1
		}
		if a.conns < b.conns {
			return 1
		}
		return strings.Compare(a.name, b.name)
	})

	names := make([]string, 0, len(rows))
	conns := make([]float64, 0, len(rows))
	waitMs := make([]float64, 0, len(rows))
	sharePct := make([]float64, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.name)
		conns = append(conns, float64(row.conns))
		waitMs = append(waitMs, float64(row.ms))
		pct := 0.0
		if useWaitTime && grandTime > 0 {
			pct = (float64(row.ms) / float64(grandTime)) * 100
		} else if !useWaitTime && grandConns > 0 {
			pct = (float64(row.conns) / float64(grandConns)) * 100
		}
		sharePct = append(sharePct, pct)
	}

	waitMsField := data.NewField("waitTimeMs", nil, waitMs)
	waitMsField.Config = &data.FieldConfig{Unit: "ms"}
	pctField := data.NewField("sharePct", nil, sharePct)
	pctField.Config = &data.FieldConfig{Unit: "percent"}

	frame := data.NewFrame("wait_breakdown",
		data.NewField("waitEvent", nil, names),
		data.NewField("connections", nil, conns),
		waitMsField,
		pctField,
	)
	frame.RefID = query.RefID
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: data.VisTypeTable,
	}
	response.Frames = append(response.Frames, frame)
	return response
}

func (a *App) queryServers(ctx context.Context, _ backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}
	resp, err := a.client.ListServerSummary(ctx, &querierv2.ListServerSummaryRequest{
		Start: timestamppb.New(query.TimeRange.From),
		End:   timestamppb.New(query.TimeRange.To),
	})
	if err != nil {
		response.Error = err
		return response
	}
	size := len(resp.GetServers())
	names := make([]string, 0, size)
	types := make([]string, 0, size)
	conns := make([]float64, 0, size)
	reqRate := make([]float64, 0, size)
	for _, srv := range resp.GetServers() {
		names = append(names, srv.GetName())
		types = append(types, srv.GetType())
		conns = append(conns, float64(srv.GetConnections()))
		reqRate = append(reqRate, srv.GetRequestRate())
	}
	nameField := data.NewField("server", nil, names)
	nameField.Config = &data.FieldConfig{
		Links: []data.DataLink{
			{
				Title: "Open in Sqlsights Snapshots",
				URL:   "/a/guilhermearpassos-sqlsights-app/snapshots?server=${__value.raw}&from=${__from}&to=${__to}",
			},
			{
				Title: "Open in Sqlsights Plan Analysis",
				URL:   "/a/guilhermearpassos-sqlsights-app/plan_analysis?server=${__value.raw}&from=${__from}&to=${__to}",
			},
		},
	}
	frame := data.NewFrame("servers",
		nameField,
		data.NewField("type", nil, types),
		data.NewField("connections", nil, conns),
		data.NewField("requestRate", nil, reqRate),
	)
	frame.RefID = query.RefID
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: data.VisTypeTable,
	}
	response.Frames = append(response.Frames, frame)
	return response
}

// querySnapList processes individual queries
func (a *App) querySnapList(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}
	q := struct {
		Database     string `json:"database"`
		OnlyBlocking bool   `json:"onlyBlocking"`
		WaitEvent    string `json:"waitEvent"`
	}{}
	if err := json.Unmarshal(query.JSON, &q); err != nil {
		response.Error = err
		return response
	}
	timeRange := query.TimeRange
	from := timeRange.From
	to := timeRange.To
	r, err := a.client.ListSnapshotSummaries(ctx, &querierv2.ListSnapshotSummariesRequest{
		Start:  timestamppb.New(from),
		End:    timestamppb.New(to),
		Server: q.Database,
	})
	if err != nil {
		response.Error = err
		return response
	}
	size := len(r.GetSnapSummaries())
	ids := make([]string, 0, size)
	times := make([]time.Time, 0, size)
	connections := make([]float64, 0, size)
	waiters := make([]float64, 0, size)
	blockers := make([]float64, 0, size)
	waitDuration := make([]float64, 0, size)
	avgDuration := make([]float64, 0, size)
	maxDuration := make([]float64, 0, size)
	waitsByType := make([]string, 0, size)
	wantWait := strings.TrimSpace(q.WaitEvent)
	for _, sum := range r.GetSnapSummaries() {
		if q.OnlyBlocking && sum.Waiters <= 0 && sum.Blockers <= 0 {
			continue
		}
		if wantWait != "" {
			key := wantWait
			if strings.EqualFold(wantWait, "cpu") {
				key = ""
			}
			if sum.GetConnectionsByWaitEvent()[key] <= 0 && sum.GetConnectionsByWaitEvent()[wantWait] <= 0 {
				continue
			}
		}
		times = append(times, time.Unix(sum.Timestamp.AsTime().Unix(), 0))
		connections = append(connections, float64(sum.GetConnections()))
		waiters = append(waiters, float64(sum.Waiters))
		blockers = append(blockers, float64(sum.Blockers))
		waitDuration = append(waitDuration, sum.WaitDuration/1000)
		avgDuration = append(avgDuration, sum.AvgDuration)
		maxDuration = append(maxDuration, sum.MaxDuration)
		ids = append(ids, sum.Id)
		cwe, err := json.Marshal(sum.ConnectionsByWaitEvent)
		if err != nil {
			response.Error = err
			return response
		}
		waitsByType = append(waitsByType, string(cwe))
	}
	idField := data.NewField("id", nil, ids)
	idField.Config = &data.FieldConfig{
		Links: []data.DataLink{
			{
				Title: "Inspect Snapshot in Sqlsights",
				URL:   fmt.Sprintf("/a/guilhermearpassos-sqlsights-app/snapshots?server=%s&snapId=${__value.raw}&from=${__from}&to=${__to}", q.Database),
			},
		},
	}
	waitDurField := data.NewField("waitDuration", nil, waitDuration)
	waitDurField.Config = &data.FieldConfig{Unit: "s"}

	frame := data.NewFrame("snapshots",
		data.NewField("time", nil, times),
		idField,
		data.NewField("connections", nil, connections),
		data.NewField("waiters", nil, waiters),
		data.NewField("blockers", nil, blockers),
		waitDurField,
		data.NewField("avgDuration", nil, avgDuration),
		data.NewField("maxDuration", nil, maxDuration),
		data.NewField("waitsByType", nil, waitsByType),
	)

	frame.RefID = query.RefID
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: data.VisTypeTable,
	}

	response.Frames = append(response.Frames, frame)
	return response
}

func filterPlanItemsExtended(items []PlanAnalysisItem, onlyLocks bool, minImpact float64, search string) []PlanAnalysisItem {
	q := strings.ToLower(strings.TrimSpace(search))
	if !onlyLocks && minImpact <= 0 && q == "" {
		return items
	}
	out := make([]PlanAnalysisItem, 0, len(items))
	for _, item := range items {
		if onlyLocks && item.RelatedLockCount <= 0 {
			continue
		}
		if minImpact > 0 && item.MaxImpact < minImpact {
			continue
		}
		if q != "" {
			matched := strings.Contains(strings.ToLower(item.Text), q) ||
				strings.Contains(strings.ToLower(item.Database), q) ||
				strings.Contains(strings.ToLower(item.QueryHash), q)
			if !matched {
				for _, c := range item.Converts {
					if strings.Contains(strings.ToLower(c.ConvertIssue+" "+c.Expression), q) {
						matched = true
						break
					}
				}
			}
			if !matched {
				for _, m := range item.MissingIndexes {
					if strings.Contains(strings.ToLower(m.Schema+"."+m.Table), q) {
						matched = true
						break
					}
				}
			}
			if !matched {
				continue
			}
		}
		out = append(out, item)
	}
	return out
}

func formatPlanWarningSummary(item PlanAnalysisItem) string {
	parts := make([]string, 0, len(item.Converts)+len(item.MissingIndexes))
	for _, c := range item.Converts {
		parts = append(parts, fmt.Sprintf("%s: %s", c.ConvertIssue, c.Expression))
	}
	for _, m := range item.MissingIndexes {
		cols := append(slices.Clone(m.EqualityColumns), m.InequalityColumns...)
		parts = append(parts, fmt.Sprintf("MissingIndex(%.1f%%) %s.%s(%s)", m.Impact, m.Schema, m.Table, strings.Join(cols, ",")))
	}
	return strings.Join(parts, " | ")
}

// queryPlanAnalysis processes individual queries
func (a *App) queryPlanAnalysis(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}
	q := struct {
		Database  string   `json:"database"`
		Databases []string `json:"databases"`
		IssueTags []string `json:"issueTags"`
		OnlyLocks bool     `json:"onlyLocks"`
		MinImpact float64  `json:"minImpact"`
		Search    string   `json:"search"`
	}{}
	if err := json.Unmarshal(query.JSON, &q); err != nil {
		response.Error = err
		return response
	}
	timeRange := query.TimeRange
	from := timeRange.From
	to := timeRange.To
	r, err := a.client.ListPlansWithIssues(ctx, &querierv2.ListPlansWithIssuesRequest{
		Start:     timestamppb.New(from),
		End:       timestamppb.New(to),
		Host:      q.Database,
		Databases: q.Databases,
	})
	if err != nil {
		response.Error = err
		return response
	}
	items := filterPlanItemsExtended(buildPlanAnalysisItems(r.GetPlans(), q.IssueTags), q.OnlyLocks, q.MinImpact, q.Search)
	size := len(items)
	queries := make([]string, 0, size)
	databases := make([]string, 0, size)
	issues := make([]float64, 0, size)
	issueTags := make([]string, 0, size)
	warningSummaries := make([]string, 0, size)
	maxImpacts := make([]float64, 0, size)
	occurrences := make([]float64, 0, size)
	locks := make([]float64, 0, size)
	times := make([]time.Time, 0, size)
	snapIds := make([]string, 0, size)
	sampleIds := make([]string, 0, size)
	traceIds := make([]string, 0, size)
	for _, item := range items {
		queries = append(queries, item.Text)
		issues = append(issues, float64(item.IssueCount))
		issueTags = append(issueTags, strings.Join(item.IssueTags, ", "))
		warningSummaries = append(warningSummaries, formatPlanWarningSummary(item))
		maxImpacts = append(maxImpacts, item.MaxImpact)
		occurrences = append(occurrences, float64(item.Occurrences))
		locks = append(locks, float64(item.RelatedLockCount))
		times = append(times, item.LastSnapTime)
		snapIds = append(snapIds, item.SnapID)
		sampleIds = append(sampleIds, item.SampleID)
		databases = append(databases, item.Database)
		traceIds = append(traceIds, item.TraceID)
	}
	textField := data.NewField("text", nil, queries)
	textField.Config = &data.FieldConfig{
		Links: []data.DataLink{
			{
				Title: "Inspect Execution Plan in Sqlsights",
				URL:   fmt.Sprintf("/a/guilhermearpassos-sqlsights-app/plan_analysis?server=%s&snapId=${__data.fields.snapID}&sampleId=${__data.fields.sampleId}&from=${__from}&to=${__to}", q.Database),
			},
		},
	}
	impactField := data.NewField("maxImpact", nil, maxImpacts)
	impactField.Config = &data.FieldConfig{Unit: "percent"}

	frame := data.NewFrame("plans_with_issues",
		data.NewField("lastSnapTime", nil, times),
		data.NewField("database", nil, databases),
		data.NewField("issueTags", nil, issueTags),
		textField,
		data.NewField("warningSummary", nil, warningSummaries),
		impactField,
		data.NewField("issues", nil, issues),
		data.NewField("occurrences", nil, occurrences),
		data.NewField("locks", nil, locks),
		data.NewField("traceId", nil, traceIds),
		data.NewField("snapID", nil, snapIds),
		data.NewField("id", nil, snapIds),
		data.NewField("sampleId", nil, sampleIds),
	)

	frame.RefID = query.RefID
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: data.VisTypeTable,
	}

	response.Frames = append(response.Frames, frame)
	return response
}

func (a *App) queryPlanAnalysisSummary(ctx context.Context, _ backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}
	q := struct {
		Database  string   `json:"database"`
		Databases []string `json:"databases"`
		IssueTags []string `json:"issueTags"`
		OnlyLocks bool     `json:"onlyLocks"`
		MinImpact float64  `json:"minImpact"`
		Search    string   `json:"search"`
	}{}
	if err := json.Unmarshal(query.JSON, &q); err != nil {
		response.Error = err
		return response
	}
	r, err := a.client.ListPlansWithIssues(ctx, &querierv2.ListPlansWithIssuesRequest{
		Start:     timestamppb.New(query.TimeRange.From),
		End:       timestamppb.New(query.TimeRange.To),
		Host:      q.Database,
		Databases: q.Databases,
	})
	if err != nil {
		response.Error = err
		return response
	}
	items := filterPlanItemsExtended(buildPlanAnalysisItems(r.GetPlans(), q.IssueTags), q.OnlyLocks, q.MinImpact, q.Search)

	var seekPlanCount, cardinalityCount, missingIdxCount, relatedLocks, totalOccurrences float64
	maxImpact := 0.0
	for _, item := range items {
		if slices.Contains(item.IssueTags, "seek_plan_convert") {
			seekPlanCount++
		}
		if slices.Contains(item.IssueTags, "cardinality_convert") {
			cardinalityCount++
		}
		if slices.Contains(item.IssueTags, "missing_index") {
			missingIdxCount++
		}
		relatedLocks += float64(item.RelatedLockCount)
		totalOccurrences += float64(item.Occurrences)
		if item.MaxImpact > maxImpact {
			maxImpact = item.MaxImpact
		}
	}

	maxImpactField := data.NewField("maxImpact", nil, []float64{maxImpact})
	maxImpactField.Config = &data.FieldConfig{Unit: "percent"}

	frame := data.NewFrame("plan_analysis_summary",
		data.NewField("totalPlans", nil, []float64{float64(len(items))}),
		data.NewField("seekPlanConversions", nil, []float64{seekPlanCount}),
		data.NewField("cardinalityConversions", nil, []float64{cardinalityCount}),
		data.NewField("missingIndexes", nil, []float64{missingIdxCount}),
		data.NewField("relatedLocks", nil, []float64{relatedLocks}),
		data.NewField("totalOccurrences", nil, []float64{totalOccurrences}),
		maxImpactField,
	)
	frame.RefID = query.RefID
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: data.VisTypeTable,
	}
	response.Frames = append(response.Frames, frame)
	return response
}

// querySnap processes individual queries
func (a *App) querySnap(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}
	q := struct {
		Database     string   `json:"database"`
		SnapID       string   `json:"snapshotID"`
		Databases    []string `json:"databases"`
		OnlyBlocking bool     `json:"onlyBlocking"`
	}{}
	if err := json.Unmarshal(query.JSON, &q); err != nil {
		response.Error = err
		return response
	}
	r, err := a.client.GetSnapshot(ctx, &querierv2.GetSnapshotRequest{
		Id: q.SnapID,
	})
	if err != nil {
		response.Error = err
		return response
	}
	size := len(r.GetSnapshot().GetSamples())
	ids := make([]string, 0, size)
	sessionIDs := make([]string, 0, size)
	statuses := make([]string, 0, size)
	text := make([]string, 0, size)
	users := make([]string, 0, size)
	durations := make([]float64, 0, size)
	blockingImpact := make([]string, 0, size)
	waitEvents := make([]string, 0, size)
	databases := make([]string, 0, size)
	traceIds := make([]string, 0, size)
	for _, sample := range r.GetSnapshot().GetSamples() {
		dbName := sample.GetDb().GetDatabaseName()
		if len(q.Databases) > 0 && !slices.Contains(q.Databases, dbName) {
			continue
		}
		if q.OnlyBlocking && !sample.Blocked && !sample.Blocker {
			continue
		}
		ids = append(ids, sample.Id)
		sessionIDs = append(sessionIDs, sample.GetSession().GetSessionId())
		statuses = append(statuses, sample.Status)
		traceIds = append(traceIds, sample.GetContext().GetTraceId())
		text = append(text, sample.Text)
		users = append(users, sample.GetSession().GetLoginName())
		waitEvents = append(waitEvents, sample.GetWaitInfo().GetWaitType())
		databases = append(databases, dbName)
		durations = append(durations, float64(sample.TimeElapsedMillis)/1000.0)
		impactText := ""
		if sample.Blocked {
			impactText += fmt.Sprintf("blocked by %s ", sample.GetBlockInfo().GetBlockedBy())
		}
		if sample.Blocker {
			if impactText != "" {
				impactText += "| "
			}
			impactText += fmt.Sprintf("%d waiting", len(sample.GetBlockInfo().GetBlockedSessions()))
		}
		blockingImpact = append(blockingImpact, impactText)
	}
	sampleIDField := data.NewField("sampleID", nil, ids)
	sampleIDField.Config = &data.FieldConfig{
		Links: []data.DataLink{
			{
				Title: "Inspect Sample in Sqlsights",
				URL:   fmt.Sprintf("/a/guilhermearpassos-sqlsights-app/snapshots?server=%s&snapId=%s&sampleId=${__value.raw}&from=${__from}&to=${__to}", q.Database, q.SnapID),
			},
		},
	}
	elapsedField := data.NewField("Elapsed", nil, durations)
	elapsedField.Config = &data.FieldConfig{Unit: "s"}

	frame := data.NewFrame("snapshot_samples",
		data.NewField("sessionID", nil, sessionIDs),
		sampleIDField,
		data.NewField("database", nil, databases),
		elapsedField,
		data.NewField("Blocking Impact", nil, blockingImpact),
		data.NewField("wait event", nil, waitEvents),
		data.NewField("Status", nil, statuses),
		data.NewField("user", nil, users),
		data.NewField("traceId", nil, traceIds),
		data.NewField("text", nil, text),
	)

	frame.RefID = query.RefID
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: data.VisTypeTable,
	}

	response.Frames = append(response.Frames, frame)
	return response
}

func (a *App) queryMetrics(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {

	response := backend.DataResponse{}
	q := struct {
		Database string `json:"database"`
	}{}
	if err := json.Unmarshal(query.JSON, &q); err != nil {
		response.Error = err
		return response
	}
	timeRange := query.TimeRange
	from := timeRange.From
	to := timeRange.To
	resp, err := a.client.ListQueryMetrics(ctx, &querierv2.ListQueryMetricsRequest{
		Start:      timestamppb.New(from),
		End:        timestamppb.New(to),
		Host:       q.Database,
		Database:   "",
		PageSize:   0,
		PageNumber: 0,
	})
	if err != nil {
		response.Error = err
		return response
	}
	text := make([]string, 0, len(resp.GetMetrics()))
	lastExecutionTime := make([]time.Time, 0, len(resp.GetMetrics()))
	queryHash := make([]string, 0, len(resp.GetMetrics()))
	databaseName := make([]string, 0, len(resp.GetMetrics()))
	executionCount := make([]float64, 0, len(resp.GetMetrics()))
	rates := make(map[string][]float64)
	for _, m := range resp.GetMetrics() {
		text = append(text, m.Text)
		lastExecutionTime = append(lastExecutionTime, m.LastExecutionTime.AsTime())
		queryHash = append(queryHash, m.QueryHash)
		databaseName = append(databaseName, m.Db.DatabaseName)

		execCount, ok := m.Counters["executionCount"]
		executionCount = append(executionCount, float64(execCount))
		if ok && execCount != 0 {
			for k, v := range m.Counters {
				if !strings.HasPrefix(k, "total") {
					continue
				}
				avgName := strings.Replace(k, "total", "avg", 1)
				rates[avgName] = append(rates[avgName], float64(v)/float64(execCount))
			}
		}
	}

	frame := data.NewFrame("metrics",
		data.NewField("text", nil, text),
		data.NewField("lastExecutionTime", nil, lastExecutionTime),
		data.NewField("queryHash", nil, queryHash),
		data.NewField("databaseName", nil, databaseName),
		data.NewField("executionCount", nil, executionCount),
	)
	for k, rate := range rates {
		frame.Fields = append(frame.Fields, data.NewField(k, nil, rate))

	}

	// Set the RefID to match the query
	frame.RefID = query.RefID

	// Add metadata for proper visualization
	frame.Meta = &data.FrameMeta{
		PreferredVisualization: data.VisTypeTable,
	}

	response.Frames = append(response.Frames, frame)
	return response
}

func (a *App) queryMetricsTimeSeries(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {

	response := backend.DataResponse{}
	q := struct {
		Database string   `json:"database"`
		Handle   string   `json:"queryHash"`
		Metrics  []string `json:"metrics"`
	}{}
	if err := json.Unmarshal(query.JSON, &q); err != nil {
		response.Error = err
		return response
	}
	timeRange := query.TimeRange
	from := timeRange.From
	to := timeRange.To
	interval := "5m"
	if to.Sub(from) < 25*time.Minute {
		interval = "1m"
	}
	if to.Sub(from) > 15*time.Hour {
		interval = "30m"
	}
	resp, err := a.client.GetQueryMetricsTimeSeries(ctx, &querierv2.GetQueryMetricsTimeSeriesRequest{
		Start:     timestamppb.New(from),
		End:       timestamppb.New(to),
		Host:      q.Database,
		SqlHandle: q.Handle,
		Interval:  interval,
	})
	if err != nil {
		response.Error = err
		return response
	}
	collectedAt := make([]time.Time, 0, len(resp.GetMetrics()))
	executionCount := make([]float64, 0, len(resp.GetMetrics()))
	rates := make(map[string][]float64)
	for _, m := range resp.GetMetrics() {
		collectedAt = append(collectedAt, m.GetCollectedAt().AsTime())

		execCount, ok := m.Counters["executionCount"]
		executionCount = append(executionCount, float64(execCount))
		if ok && execCount != 0 {
			for k, v := range m.Counters {
				if !strings.HasPrefix(k, "total") {
					continue
				}
				avgName := strings.Replace(k, "total", "avg", 1)
				rates[avgName] = append(rates[avgName], float64(v)/float64(execCount))
			}
		}
	}

	frame := data.NewFrame("metrics",
		data.NewField("time", nil, collectedAt),
	)
	if len(q.Metrics) == 0 || slices.Contains(q.Metrics, "executionCount") {
		frame.Fields = append(frame.Fields, data.NewField("executionCount", nil, executionCount))
	}
	for k, rate := range rates {
		if len(q.Metrics) == 0 || slices.Contains(q.Metrics, k) {
			frame.Fields = append(frame.Fields, data.NewField(k, nil, rate))
		}
	}

	// Set the RefID to match the query
	frame.RefID = query.RefID

	// Add metadata for proper visualization
	frame.Meta = &data.FrameMeta{
		Type:                   data.FrameTypeTimeSeriesWide,
		PreferredVisualization: data.VisTypeGraph,
	}

	response.Frames = append(response.Frames, frame)
	return response
}

// DropdownOption represents a single dropdown option
type DropdownOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// handleDropdownOptions returns available options for query parameters and dashboard variables
func (a *App) handleDropdownOptions(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	optionType := req.URL.Query().Get("type")
	startTimestamp, endTimestamp, err := parseTimeBounds(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var options []DropdownOption
	switch optionType {
	case "databases", "servers":
		options, err = a.getDatabaseOptions(req.Context(), startTimestamp, endTimestamp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "server-databases":
		server := strings.TrimSpace(req.URL.Query().Get("server"))
		if server == "" {
			options = []DropdownOption{}
		} else {
			resp, qErr := a.client.ListPlansWithIssues(req.Context(), &querierv2.ListPlansWithIssuesRequest{
				Start: timestamppb.New(startTimestamp),
				End:   timestamppb.New(endTimestamp),
				Host:  server,
			})
			if qErr != nil {
				http.Error(w, qErr.Error(), http.StatusInternalServerError)
				return
			}
			dbSet := make(map[string]struct{})
			for _, p := range resp.GetPlans() {
				if name := p.GetLatestSample().GetDb().GetDatabaseName(); name != "" {
					dbSet[name] = struct{}{}
				}
			}
			names := make([]string, 0, len(dbSet))
			for n := range dbSet {
				names = append(names, n)
			}
			slices.Sort(names)
			options = make([]DropdownOption, len(names))
			for i, n := range names {
				options[i] = DropdownOption{Label: n, Value: n}
			}
		}
	case "snapshots":
		server := strings.TrimSpace(req.URL.Query().Get("server"))
		if server == "" {
			options = []DropdownOption{}
		} else {
			resp, qErr := a.client.ListSnapshotSummaries(req.Context(), &querierv2.ListSnapshotSummariesRequest{
				Start:  timestamppb.New(startTimestamp),
				End:    timestamppb.New(endTimestamp),
				Server: server,
			})
			if qErr != nil {
				http.Error(w, qErr.Error(), http.StatusInternalServerError)
				return
			}
			snaps := resp.GetSnapSummaries()
			options = make([]DropdownOption, 0, len(snaps))
			for _, s := range snaps {
				label := fmt.Sprintf("%s (%d conns, %d blockers)", s.GetTimestamp().AsTime().Format(time.RFC3339), s.GetConnections(), s.GetBlockers())
				options = append(options, DropdownOption{Label: label, Value: s.GetId()})
			}
		}
	case "issue-tags":
		options = []DropdownOption{
			{Label: "Implicit Conversion (Seek Plan)", Value: "seek_plan_convert"},
			{Label: "Missing Index", Value: "missing_index"},
			{Label: "Implicit Conversion (Cardinality)", Value: "cardinality_convert"},
			{Label: "Implicit Conversion (Any)", Value: "implicit_conversion"},
			{Label: "Large Table Scan", Value: "large_table_scan"},
		}
	default:
		http.Error(w, "invalid option type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(options); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getDatabaseOptions fetches available databases
func (a *App) getDatabaseOptions(ctx context.Context, startTimestamp time.Time, endTimestamp time.Time) ([]DropdownOption, error) {
	resp, err := a.client.ListServerSummary(ctx, &querierv2.ListServerSummaryRequest{
		Start: timestamppb.New(startTimestamp),
		End:   timestamppb.New(endTimestamp),
	})
	if err != nil {
		return nil, fmt.Errorf("list databases: %w", err)
	}
	ret := make([]DropdownOption, len(resp.GetServers()))
	for i, server := range resp.GetServers() {
		ret[i] = DropdownOption{
			Label: server.GetName(),
			Value: server.GetName(),
		}
	}
	return ret, nil
}

func parseTimeBounds(r *http.Request) (time.Time, time.Time, error) {
	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now

	if s := r.URL.Query().Get("start"); s != "" {
		parsed, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid start timestamp: %w", err)
		}
		start = parsed
	}
	if e := r.URL.Query().Get("end"); e != "" {
		parsed, err := time.Parse(time.RFC3339, e)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid end timestamp: %w", err)
		}
		end = parsed
	}
	return start, end, nil
}

func splitCSVParam(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func deriveIssueTagsFromProto(p *querierv2.PlansWithIssues, converts []PlanAffectingConvert, missingIndexes []MissingIndex) []string {
	tagSet := make(map[string]struct{}, 4)
	for _, t := range p.GetIssueTag() {
		if t != "" {
			tagSet[t] = struct{}{}
		}
	}
	if len(missingIndexes) > 0 {
		tagSet["missing_index"] = struct{}{}
	}
	for _, c := range converts {
		issue := strings.ToLower(c.ConvertIssue)
		switch {
		case strings.Contains(issue, "seek"):
			tagSet["seek_plan_convert"] = struct{}{}
		case strings.Contains(issue, "cardinality"):
			tagSet["cardinality_convert"] = struct{}{}
		default:
			tagSet["implicit_conversion"] = struct{}{}
		}
	}
	tags := make([]string, 0, len(tagSet))
	for t := range tagSet {
		tags = append(tags, t)
	}
	slices.Sort(tags)
	return tags
}

func matchesIssueTagFilter(itemTags []string, filterTags []string) bool {
	if len(filterTags) == 0 {
		return true
	}
	for _, want := range filterTags {
		if want == "implicit_conversion" {
			for _, t := range itemTags {
				if t == "implicit_conversion" || t == "seek_plan_convert" || t == "cardinality_convert" {
					return true
				}
			}
			continue
		}
		if slices.Contains(itemTags, want) {
			return true
		}
	}
	return false
}

func buildPlanAnalysisItems(plans []*querierv2.PlansWithIssues, filterIssueTags []string) []PlanAnalysisItem {
	items := make([]PlanAnalysisItem, 0, len(plans))
	for _, p := range plans {
		converts := make([]PlanAffectingConvert, 0)
		missingIndexes := make([]MissingIndex, 0)
		maxImpact := 0.0

		for _, w := range p.GetParsedPlan().GetWarnings() {
			switch w.Warning.(type) {
			case *dbmv1.PlanWarning_Convert:
				c := w.GetConvert()
				converts = append(converts, PlanAffectingConvert{
					ConvertIssue: c.GetConvertIssue(),
					Expression:   c.GetExpression(),
				})
			case *dbmv1.PlanWarning_MissingIndex:
				mi := w.GetMissingIndex()
				if mi.GetImpact() > maxImpact {
					maxImpact = mi.GetImpact()
				}
				missingIndexes = append(missingIndexes, MissingIndex{
					Database:          mi.GetDatabase(),
					Schema:            mi.GetSchema(),
					Table:             mi.GetTable(),
					Impact:            mi.GetImpact(),
					EqualityColumns:   mi.GetEqualityColumns(),
					InequalityColumns: mi.GetInequalityColumns(),
					IncludeColumns:    mi.GetIncludeColumns(),
				})
			}
		}

		tags := deriveIssueTagsFromProto(p, converts, missingIndexes)
		if !matchesIssueTagFilter(tags, filterIssueTags) {
			continue
		}

		sample := p.GetLatestSample()
		items = append(items, PlanAnalysisItem{
			PlanHandle:       sample.GetPlanHandle(),
			QueryHash:        sample.GetQueryHash(),
			SQLHandle:        sample.GetSqlHandle(),
			Text:             sample.GetText(),
			Database:         sample.GetDb().GetDatabaseName(),
			SnapID:           sample.GetSnapInfo().GetId(),
			SampleID:         sample.GetId(),
			LastSnapTime:     sample.GetSnapInfo().GetTimestamp().AsTime(),
			TraceID:          sample.GetContext().GetTraceId(),
			IssueCount:       p.GetIssueCount(),
			Occurrences:      p.GetOccurrences(),
			RelatedLockCount: p.GetRelatedLockCount(),
			IssueTags:        tags,
			Converts:         converts,
			MissingIndexes:   missingIndexes,
			MaxImpact:        maxImpact,
		})
	}
	return items
}

func (a *App) handlePlanAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	server := strings.TrimSpace(r.URL.Query().Get("server"))
	if server == "" {
		http.Error(w, "server is required", http.StatusBadRequest)
		return
	}
	start, end, err := parseTimeBounds(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	databasesFilter := splitCSVParam(r.URL.Query().Get("databases"))
	issueTagsFilter := splitCSVParam(r.URL.Query().Get("issueTags"))

	// Query server plans without DB filter first so we can return the full list of
	// available databases for the dropdown alongside the filtered plans.
	resp, err := a.client.ListPlansWithIssues(r.Context(), &querierv2.ListPlansWithIssuesRequest{
		Start: timestamppb.New(start),
		End:   timestamppb.New(end),
		Host:  server,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dbSet := make(map[string]struct{})
	filteredProtoPlans := make([]*querierv2.PlansWithIssues, 0, len(resp.GetPlans()))
	for _, p := range resp.GetPlans() {
		dbName := p.GetLatestSample().GetDb().GetDatabaseName()
		if dbName != "" {
			dbSet[dbName] = struct{}{}
		}
		if len(databasesFilter) == 0 || slices.Contains(databasesFilter, dbName) {
			filteredProtoPlans = append(filteredProtoPlans, p)
		}
	}
	availDBs := make([]string, 0, len(dbSet))
	for db := range dbSet {
		availDBs = append(availDBs, db)
	}
	slices.Sort(availDBs)

	items := buildPlanAnalysisItems(filteredProtoPlans, issueTagsFilter)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(PlanAnalysisResponse{
		Plans:              items,
		AvailableDatabases: availDBs,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *App) handleSnapshotSummaries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	server := strings.TrimSpace(r.URL.Query().Get("server"))
	if server == "" {
		http.Error(w, "server is required", http.StatusBadRequest)
		return
	}
	start, end, err := parseTimeBounds(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := a.client.ListSnapshotSummaries(r.Context(), &querierv2.ListSnapshotSummariesRequest{
		Start:  timestamppb.New(start),
		End:    timestamppb.New(end),
		Server: server,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	summaries := make([]SnapshotSummaryDTO, 0, len(resp.GetSnapSummaries()))
	for _, s := range resp.GetSnapSummaries() {
		connsByWait := s.GetConnectionsByWaitEvent()
		if connsByWait == nil {
			connsByWait = map[string]int64{}
		}
		timeByWait := s.GetTimeMsByWaitEvent()
		if timeByWait == nil {
			timeByWait = map[string]int64{}
		}
		summaries = append(summaries, SnapshotSummaryDTO{
			ID:                     s.GetId(),
			Timestamp:              s.GetTimestamp().AsTime(),
			Connections:            s.GetConnections(),
			Waiters:                s.GetWaiters(),
			Blockers:               s.GetBlockers(),
			WaitDurationMs:         s.GetWaitDuration(),
			AvgDuration:            s.GetAvgDuration(),
			MaxDuration:            s.GetMaxDuration(),
			ConnectionsByWaitEvent: connsByWait,
			TimeMsByWaitEvent:      timeByWait,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(summaries); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *App) handleSnapshotSamples(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	snapID := strings.TrimSpace(r.URL.Query().Get("snapId"))
	if snapID == "" {
		http.Error(w, "snapId is required", http.StatusBadRequest)
		return
	}

	resp, err := a.client.GetSnapshot(r.Context(), &querierv2.GetSnapshotRequest{Id: snapID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	samples := make([]SnapshotSampleDTO, 0, len(resp.GetSnapshot().GetSamples()))
	for _, s := range resp.GetSnapshot().GetSamples() {
		blockedSessions := s.GetBlockInfo().GetBlockedSessions()
		if blockedSessions == nil {
			blockedSessions = []string{}
		}
		samples = append(samples, SnapshotSampleDTO{
			SampleID:        s.GetId(),
			SnapID:          s.GetSnapInfo().GetId(),
			SessionID:       s.GetSession().GetSessionId(),
			Status:          s.GetStatus(),
			Text:            s.GetText(),
			User:            s.GetSession().GetLoginName(),
			ElapsedMs:       s.GetTimeElapsedMillis(),
			WaitEvent:       s.GetWaitInfo().GetWaitType(),
			WaitTimeMs:      s.GetWaitInfo().GetWaitTime(),
			WaitResource:    s.GetWaitInfo().GetWaitResource(),
			Database:        s.GetDb().GetDatabaseName(),
			IsBlocked:       s.GetBlocked(),
			IsBlocker:       s.GetBlocker(),
			BlockedBy:       s.GetBlockInfo().GetBlockedBy(),
			BlockedSessions: blockedSessions,
			TraceID:         s.GetContext().GetTraceId(),
			QueryHash:       s.GetQueryHash(),
			PlanHandle:      s.GetPlanHandle(),
			ProgramName:     s.GetSession().GetProgramName(),
			HostName:        s.GetSession().GetHost(),
			ClientIP:        s.GetSession().GetClientIp(),
			SessionStatus:   s.GetSession().GetStatus(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(samples); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
