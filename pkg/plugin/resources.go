package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/ui/domain"

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
	w.WriteHeader(http.StatusOK)
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
	w.WriteHeader(http.StatusOK)
}

// registerRoutes takes a *http.ServeMux and registers some HTTP handlers.
func (a *App) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/ping", a.handlePing)
	mux.HandleFunc("/", a.handlePing)
	mux.HandleFunc("/echo", a.handleEcho)
	mux.HandleFunc("/datasource-options", a.handleDropdownOptions)
	mux.HandleFunc("/query", a.handleQuery)
	mux.HandleFunc("/getQueryDetails", a.handleFetchQueryDetails)

}

func (a *App) handleFetchQueryDetails(w http.ResponseWriter, r *http.Request) {
	// handle the request
	// e.g. call a third-party API
	sampleID := r.URL.Query().Get("sampleId")
	snapID := r.URL.Query().Get("snapId")
	if sampleID == "" {
		http.Error(w, "sampleId is required", http.StatusBadRequest)
		return
	}
	if snapID == "" {
		http.Error(w, "snapId is required", http.StatusBadRequest)
	}
	resp, err := a.client.GetSampleDetails(r.Context(), &dbmv1.GetSampleDetailsRequest{
		SampleId: sampleID,
		SnapId:   snapID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var plan domain.ParsedExecutionPlan
	if resp.ParsedPlan != nil {
		plan = domain.ProtoParsedPlanToDomain(resp.ParsedPlan)

	}
	blockChain, err2 := blockChainFromProto(resp.BlockChain)
	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		return
	}
	m, err := json.Marshal(struct {
		Plan  domain.ParsedExecutionPlan `json:"plan"`
		Chain domain.BlockChain          `json:"blocking_chain"`
	}{
		Plan:  plan,
		Chain: blockChain,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(m)
	w.WriteHeader(http.StatusOK)
}

func blockChainFromProto(chain *dbmv1.BlockChain) (domain.BlockChain, error) {
	roots := make([]domain.BlockingNode, len(chain.Roots))
	for i, root := range chain.Roots {
		node, err2 := nodeFromProto(root, i)
		if err2 != nil {
			return domain.BlockChain{}, err2
		}
		roots[i] = node
	}
	return domain.BlockChain{
		Roots: roots,
	}, nil
}

func protoSampleToDomain(sample *dbmv1.QuerySample) (domain.QuerySample, error) {
	var blockTime, blockDetails string
	if sample.Blocker {
		blockDetails = fmt.Sprintf("%d block waiters", len(sample.BlockInfo.BlockedSessions))
		if sample.Blocked {
			blockDetails += " | "
		}
	}
	if sample.Blocked {
		blockTime = time.Time{}.Add(time.Duration(sample.WaitInfo.WaitTime * 1_000_000_000)).Format(time.TimeOnly)
		blockDetails += fmt.Sprintf("blocked by %s", sample.BlockInfo.BlockedBy)
	}
	sid, err2 := strconv.Atoi(sample.Session.SessionId)
	if err2 != nil {
		return domain.QuerySample{}, err2
	}

	dSample := domain.QuerySample{
		SID:                     sid,
		Query:                   sample.Text,
		ExecutionTime:           fmt.Sprintf("%d ms", sample.TimeElapsedMillis),
		User:                    sample.Session.LoginName,
		IsBlocker:               sample.Blocker,
		IsWaiter:                sample.Blocked,
		BlockingTime:            blockTime,
		BlockDetails:            blockDetails,
		WaitEvent:               sample.WaitInfo.WaitType,
		Database:                sample.Db.DatabaseName,
		SampleID:                sample.Id,
		SnapID:                  sample.SnapInfo.Id,
		SQLHandle:               sample.SqlHandle,
		PlanHandle:              sample.PlanHandle,
		Status:                  sample.Status,
		QueryHash:               sample.QueryHash,
		SessionLoginTime:        sample.Session.LoginTime.AsTime(),
		SessionHost:             sample.Session.Host,
		SessionClientIp:         sample.Session.ClientIp,
		SessionStatus:           sample.Session.Status,
		SessionProgramName:      sample.Session.ProgramName,
		SessionLastRequestStart: sample.Session.LastRequestStart.AsTime(),
		SessionLastRequestEnd:   sample.Session.LastRequestEnd.AsTime(),
	}
	return dSample, nil
}

func nodeFromProto(root *dbmv1.BlockChain_BlockingNode, i int) (domain.BlockingNode, error) {
	ds, err := protoSampleToDomain(root.QuerySample)
	if err != nil {
		return domain.BlockingNode{}, err
	}
	childNodes := make([]domain.BlockingNode, len(root.ChildNodes))
	for j, child := range root.ChildNodes {
		cn, err2 := nodeFromProto(child, i+1)
		if err2 != nil {
			return domain.BlockingNode{}, err2
		}
		childNodes[j] = cn
	}
	node := domain.BlockingNode{
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
		case "chart":
			res := a.query(ctx, req.PluginContext, q)
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
	// Implement your SQL query logic here
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
	r, err := a.client.ListSnapshotSummaries(ctx, &dbmv1.ListSnapshotSummariesRequest{
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

	times := make([]time.Time, points)
	valuesByWaitType := make(map[string][]float64, points)
	for i := 0; i < points; i++ {
		times[i] = from.Add(time.Duration(i) * interval)
	}
	for _, sum := range r.GetSnapSummaries() {
		idx := int(sum.Timestamp.AsTime().Sub(from).Minutes())
		for we, c := range sum.GetConnectionsByWaitEvent() {

			if we == "" {
				we = "cpu"
			}
			if _, ok := valuesByWaitType[we]; !ok {
				valuesByWaitType[we] = make([]float64, points)
			}
			valuesByWaitType[we][idx] = float64(c)
		}
	}
	frame := data.NewFrame("locks by type",
		data.NewField("time", nil, times),
	)
	for we, v := range valuesByWaitType {
		frame.Fields = append(frame.Fields, data.NewField(we, nil, v))
	}

	// Set the RefID to match the query
	frame.RefID = query.RefID

	// Add metadata for proper visualization
	frame.Meta = &data.FrameMeta{
		Type: data.FrameTypeTimeSeriesWide,
	}

	response.Frames = append(response.Frames, frame)
	return response
}

// querySnapList processes individual queries
func (a *App) querySnapList(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	// Implement your SQL query logic here
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
	r, err := a.client.ListSnapshotSummaries(ctx, &dbmv1.ListSnapshotSummariesRequest{
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
	for _, sum := range r.GetSnapSummaries() {
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
		waitsByType = append(waitsByType, string(cwe)) //"["+strings.Join(wbt, ",")+"]")
	}
	frame := data.NewFrame("snapshots",
		data.NewField("time", nil, times),
		data.NewField("id", nil, ids),
		data.NewField("connections", nil, connections),
		data.NewField("waiters", nil, waiters),
		data.NewField("blockers", nil, blockers),
		data.NewField("waitDuration", nil, waitDuration),
		data.NewField("avgDuration", nil, avgDuration),
		data.NewField("maxDuration", nil, maxDuration),
		data.NewField("waitsByType", nil, waitsByType),
	)

	// Set the RefID to match the query
	frame.RefID = query.RefID

	// Add metadata for proper visualization
	frame.Meta = &data.FrameMeta{
		Type: data.FrameTypeTimeSeriesWide,
	}

	response.Frames = append(response.Frames, frame)
	return response
}

// querySnap processes individual queries
func (a *App) querySnap(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	// Implement your SQL query logic here
	response := backend.DataResponse{}
	q := struct {
		Database string `json:"database"`
		SnapID   string `json:"snapshotID"`
	}{}
	if err := json.Unmarshal(query.JSON, &q); err != nil {
		response.Error = err
		return response
	}
	//timeRange := query.TimeRange
	//from := timeRange.From
	//to := timeRange.To
	r, err := a.client.GetSnapshot(ctx, &dbmv1.GetSnapshotRequest{
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
	blockingOrSelf := make([]string, 0, size)
	for _, sample := range r.GetSnapshot().GetSamples() {
		ids = append(ids, sample.Id)
		sessionIDs = append(sessionIDs, sample.Session.SessionId)
		statuses = append(statuses, sample.Status)
		text = append(text, sample.Text)
		users = append(users, sample.Session.LoginName)
		waitEvents = append(waitEvents, sample.GetWaitInfo().GetWaitType())
		databases = append(databases, sample.Db.DatabaseName)
		durations = append(durations, float64(sample.TimeElapsedMillis/1000))
		bos := sample.Session.SessionId
		if sample.Blocked {
			bos = sample.BlockInfo.BlockedBy
		}
		blockingOrSelf = append(blockingOrSelf, bos)
		impactText := ""
		if sample.Blocked {
			impactText += fmt.Sprintf("blocked by %s ", sample.BlockInfo.BlockedBy)
		}
		if sample.Blocker {
			if impactText != "" {
				impactText += "| "
			}
			impactText += fmt.Sprintf("%d waiting", len(sample.BlockInfo.BlockedSessions))

		}
		blockingImpact = append(blockingImpact, impactText)
	}
	frame := data.NewFrame("snapshots",
		data.NewField("bsid", nil, blockingOrSelf),
		data.NewField("sampleID", nil, ids),
		data.NewField("sessionID", nil, sessionIDs),
		data.NewField("text", nil, text),
		data.NewField("Elapsed", nil, durations),
		data.NewField("Blocking Impact", nil, blockingImpact),
		data.NewField("wait event", nil, waitEvents),
		data.NewField("database", nil, databases),
		data.NewField("Status", nil, statuses),
		data.NewField("user", nil, users),
	)

	// Set the RefID to match the query
	frame.RefID = query.RefID

	// Add metadata for proper visualization
	frame.Meta = &data.FrameMeta{
		Type: data.FrameTypeTimeSeriesWide,
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
	resp, err := a.client.ListQueryMetrics(ctx, &dbmv1.ListQueryMetricsRequest{
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
		Type: data.FrameTypeTimeSeriesWide,
	}

	response.Frames = append(response.Frames, frame)
	return response
}

// DropdownOption represents a single dropdown option
type DropdownOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// handleDropdownOptions returns available options for query parameters
func (a *App) handleDropdownOptions(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the option type from query parameters
	optionType := req.URL.Query().Get("type")

	var options []DropdownOption
	// Customize this based on your needs
	switch optionType {
	case "databases":
		start := req.URL.Query().Get("start")
		end := req.URL.Query().Get("end")
		var startTimestamp, endTimestamp time.Time
		var err error
		if start == "" {
			startTimestamp = time.Now().Add(-1 * time.Hour)
			//http.Error(w, "start and parameter is required", http.StatusBadRequest)
			//return
		} else {

			startTimestamp, err = time.Parse(time.RFC3339, start)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if end == "" {
			endTimestamp = time.Now()
			//http.Error(w, "end and parameter is required", http.StatusBadRequest)
			//return
		} else {
			endTimestamp, err = time.Parse(time.RFC3339, end)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		options, err = a.getDatabaseOptions(req.Context(), startTimestamp, endTimestamp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
	resp, err := a.client.ListServerSummary(ctx, &dbmv1.ListServerSummaryRequest{
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
