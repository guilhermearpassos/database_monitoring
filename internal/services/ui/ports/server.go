package ports

import (
	"encoding/json"
	"fmt"
	"github.com/guilhermearpassos/database-monitoring/internal/services/ui/domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"html/template"
	"log"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

var groupToBaseColor = map[string]string{
	"Locks":         "239, 68, 68, 1",   // Red
	"I/O":           "249, 115, 22, 1",  // Orange
	"CPU":           "34, 197, 94, 1",   // Green
	"Memory":        "234, 179, 8, 1",   // Yellow
	"Network":       "168, 85, 247, 1",  // Purple
	"Background":    "59, 130, 246, 1",  // Blue
	"Idle":          "107, 114, 128, 1", // Gray
	"Miscellaneous": "132, 204, 22, 1",  // Lime
}

var waitEventToIntensity = map[string]float64{
	// Locks and Blocking
	"LCK_M_BU":  0.2,
	"LCK_M_IS":  0.25,
	"LCK_M_IU":  0.3,
	"LCK_M_S":   0.35,
	"LCK_M_IX":  0.4,
	"LCK_M_X":   0.45,
	"LCK_M_U":   0.5,
	"LCK_M_SIU": 0.55,
	"LCK_M_SIX": 0.6,
	"LCK_M_UIX": 0.65,

	// I/O-Related
	"ASYNC_IO_COMPLETION": 0.2,
	"IO_COMPLETION":       0.25,
	"PAGEIOLATCH_SH":      0.3,
	"PAGEIOLATCH_EX":      0.35,
	"BACKUPIO":            0.4,
	"WRITELOG":            0.45,
	"LOGBUFFER":           0.5,

	// CPU and Parallelism
	"CXPACKET":                         0.2,
	"SOS_SCHEDULER_YIELD":              0.25,
	"THREADPOOL":                       0.3,
	"RESOURCE_SEMAPHORE":               0.35,
	"RESOURCE_SEMAPHORE_QUERY_COMPILE": 0.4,
	"none":                             0.45,

	// Memory-Related
	"CMEMTHREAD":            0.2,
	"MEMORY_ALLOCATION_EXT": 0.25,
	"PAGELATCH_EX":          0.3,
	"PAGELATCH_SH":          0.35,

	// Network and Latency
	"ASYNC_NETWORK_IO": 0.2,
	"NETWORK_IO":       0.25,
	"OLEDB":            0.3,

	// Background and Maintenance
	"CHECKPOINT_QUEUE":            0.2,
	"LAZYWRITER_SLEEP":            0.25,
	"XE_TIMER_EVENT":              0.3,
	"TRACEWRITE":                  0.35,
	"FT_IFTS_SCHEDULER_IDLE_WAIT": 0.4,

	// Idle and Sleep
	"SLEEP_TASK":             0.2,
	"WAITFOR":                0.25,
	"BROKER_RECEIVE_WAITFOR": 0.3,
	"BROKER_TO_FLUSH":        0.35,
	"BROKER_TRANSMITTER":     0.4,

	// Miscellaneous
	"PREEMPTIVE_OS_AUTHENTICATIONOPS": 0.2,
	"PREEMPTIVE_OS_GETPROCADDRESS":    0.25,
	"CLR_AUTO_EVENT":                  0.3,
	"CLR_CRST":                        0.35,
	"CLR_JOIN":                        0.4,
	"CLR_MANUAL_EVENT":                0.45,
}

var waitEventToGroup = map[string]string{
	// Locks and Blocking
	"LCK_M_S":   "Locks",
	"LCK_M_X":   "Locks",
	"LCK_M_U":   "Locks",
	"LCK_M_IS":  "Locks",
	"LCK_M_IU":  "Locks",
	"LCK_M_IX":  "Locks",
	"LCK_M_SIU": "Locks",
	"LCK_M_SIX": "Locks",
	"LCK_M_UIX": "Locks",
	"LCK_M_BU":  "Locks",

	// I/O-Related
	"PAGEIOLATCH_SH":      "I/O",
	"PAGEIOLATCH_EX":      "I/O",
	"WRITELOG":            "I/O",
	"ASYNC_IO_COMPLETION": "I/O",
	"IO_COMPLETION":       "I/O",
	"BACKUPIO":            "I/O",
	"LOGBUFFER":           "I/O",

	// CPU and Parallelism
	"none":                             "CPU",
	"CXPACKET":                         "CPU",
	"SOS_SCHEDULER_YIELD":              "CPU",
	"THREADPOOL":                       "CPU",
	"RESOURCE_SEMAPHORE":               "CPU",
	"RESOURCE_SEMAPHORE_QUERY_COMPILE": "CPU",

	// Memory-Related
	"CMEMTHREAD":            "Memory",
	"MEMORY_ALLOCATION_EXT": "Memory",
	"PAGELATCH_EX":          "Memory",
	"PAGELATCH_SH":          "Memory",

	// Network and Latency
	"ASYNC_NETWORK_IO": "Network",
	"NETWORK_IO":       "Network",
	"OLEDB":            "Network",

	// Background and Maintenance
	"CHECKPOINT_QUEUE":            "Background",
	"LAZYWRITER_SLEEP":            "Background",
	"XE_TIMER_EVENT":              "Background",
	"TRACEWRITE":                  "Background",
	"FT_IFTS_SCHEDULER_IDLE_WAIT": "Background",

	// Idle and Sleep
	"SLEEP_TASK":             "Idle",
	"WAITFOR":                "Idle",
	"BROKER_RECEIVE_WAITFOR": "Idle",
	"BROKER_TO_FLUSH":        "Idle",
	"BROKER_TRANSMITTER":     "Idle",

	// Miscellaneous
	"PREEMPTIVE_OS_AUTHENTICATIONOPS": "Miscellaneous",
	"PREEMPTIVE_OS_GETPROCADDRESS":    "Miscellaneous",
	"CLR_AUTO_EVENT":                  "Miscellaneous",
	"CLR_CRST":                        "Miscellaneous",
	"CLR_JOIN":                        "Miscellaneous",
	"CLR_MANUAL_EVENT":                "Miscellaneous",
}

func getWaitEventColor(waitEvent string) string {
	// Get the group for the wait event
	group, ok := waitEventToGroup[waitEvent]
	if !ok {
		return "107, 114, 128, 1" // Default gray color
	}

	// Get the base color for the group
	rgba, ok := groupToBaseColor[group]
	if !ok {
		return "107, 114, 128, 1" // Default gray color
	}

	// Get the intensity level for the wait event
	alpha, ok := waitEventToIntensity[waitEvent]
	if !ok {
		alpha = 0.5 // Default intensity
	}

	// Parse the base color
	parts := strings.Split(rgba, ",")
	if len(parts) != 4 {
		return "107, 114, 128, 1" // Default gray color
	}

	// Extract RGB values
	r := strings.TrimSpace(parts[0])
	g := strings.TrimSpace(parts[1])
	b := strings.TrimSpace(parts[2])

	// Generate the RGBA color
	return fmt.Sprintf("%s, %s, %s, %.2f", r, g, b, alpha)
}

type HtmxServer struct {
	client        dbmv1.DBMApiClient
	supportClient dbmv1.DBMSupportApiClient
	templates     *template.Template
}

func NewServer(cc grpc.ClientConnInterface) (*HtmxServer, error) {
	// Parse all templates
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		return nil, err
	}

	// Parse partials
	partials, err := template.ParseGlob("templates/partials/*.html")
	if err != nil {
		return nil, err
	}

	// Add partials to main template
	for _, t := range partials.Templates() {
		_, err = tmpl.AddParseTree(t.Name(), t.Tree)
		if err != nil {
			return nil, err
		}
	}
	// Parse pages
	pages, err := template.ParseGlob("templates/pages/*.html")
	if err != nil {
		return nil, err
	}

	// Add pages to main template
	for _, t := range pages.Templates() {
		_, err = tmpl.AddParseTree(t.Name(), t.Tree)
		if err != nil {
			return nil, err
		}
	}
	return &HtmxServer{
		client:        dbmv1.NewDBMApiClient(cc),
		supportClient: dbmv1.NewDBMSupportApiClient(cc),
		templates:     tmpl,
	}, nil
}

func (s *HtmxServer) StartServer(addr string) error {

	// Route handlers
	http.HandleFunc("/", s.HandleBaseLayout)
	http.HandleFunc("/servers", s.HandleServerRefresh)
	http.HandleFunc("/snapshots/", s.HandleSnapshots)
	http.HandleFunc("/server-drilldown", s.HandleServerDrillDown)
	http.HandleFunc("/query-details", s.HandleQuerySampleDetails)
	http.HandleFunc("/samples/", s.HandleSamples)
	// Serve static files
	staticFS := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", staticFS))

	// Start server
	log.Printf("Server starting on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (s *HtmxServer) HandleBaseLayout(w http.ResponseWriter, r *http.Request) {
	err := s.templates.ExecuteTemplate(w, "base.html", map[string]interface{}{
		"ServerList": domain.SampleServers,
		"Slideover": map[string]interface{}{
			"State":      "",
			"ServerName": "",
			"ModalState": "",
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *HtmxServer) HandleServerRefresh(w http.ResponseWriter, r *http.Request) {
	startTime, endTime, err := getTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := s.supportClient.ListDatabases(r.Context(),
		&dbmv1.ListDatabasesRequest{
			Start: timestamppb.New(startTime),
			End:   timestamppb.New(endTime),
		})
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		//return
		fmt.Println(err)
	}
	_ = resp

	err = s.templates.ExecuteTemplate(w, "server_list.html", domain.SampleServers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *HtmxServer) HandleServerDrillDown(w http.ResponseWriter, r *http.Request) {
	startTime, endTime, err := getTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	server := r.URL.Query().Get("server")

	pageNumber := int64(1)
	resp, err := s.client.ListSnapshots(r.Context(), &dbmv1.ListSnapshotsRequest{
		Start:      timestamppb.New(startTime),
		End:        timestamppb.New(endTime),
		Host:       server,
		Database:   "",
		PageSize:   30,
		PageNumber: pageNumber,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	snaps := resp.GetSnapshots()
	for int64(len(snaps)) < resp.TotalCount {
		pageNumber++

		resp2, err2 := s.client.ListSnapshots(r.Context(), &dbmv1.ListSnapshotsRequest{
			Start:      timestamppb.New(startTime),
			End:        timestamppb.New(endTime),
			Host:       server,
			Database:   "",
			PageSize:   30,
			PageNumber: pageNumber,
		})
		if err2 != nil {
			http.Error(w, err2.Error(), http.StatusInternalServerError)
			return
		}
		snaps = append(snaps, resp2.GetSnapshots()...)
	}
	colorMap := make(map[string]string)
	filteredData := make([]domain.TimeSeriesData, 0)
	for _, snap := range snaps {
		waitGroups := make(map[string]int)
		for _, sample := range snap.Samples {
			waitType := sample.WaitInfo.WaitType
			if waitType == "" {
				waitType = "none"
			}
			if _, ok := waitGroups[waitType]; ok {
				waitGroups[waitType]++
			} else {
				waitGroups[waitType] = 1
			}
			if _, ok := colorMap[waitType]; !ok {
				colorMap[waitType] = fmt.Sprintf("rgba(%s)", getWaitEventColor(waitType))
			}
		}
		filteredData = append(filteredData, domain.TimeSeriesData{
			Timestamp:  snap.Timestamp.AsTime(),
			WaitGroups: waitGroups,
		})
	}

	chartData, err := json.Marshal(filteredData)
	if err != nil {
		http.Error(w, "Unable to marshal data", http.StatusInternalServerError)
		return
	}
	colorMapJson, err := json.Marshal(colorMap)
	if err != nil {
		http.Error(w, "Unable to marshal color data", http.StatusInternalServerError)
		return
	}
	timeRange := map[string]string{
		"start": startTime.Add(-1 * time.Minute).Format("2006-01-02T15:04:05"),
		"end":   endTime.Add(1 * time.Minute).Format("2006-01-02T15:04:05"),
	}
	timeRangeJSON, err := json.Marshal(timeRange)
	if err != nil {
		http.Error(w, "Unable to marshal time range", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	//timeRange := r.URL.Query().Get("time-range")
	//_ = timeRange // TODO: Implement filtering
	//server-drilldown
	if r.Header.Get("Hx-request") == "true" {
		//partial render
		err = s.templates.ExecuteTemplate(w, "slideover.html", struct {
			State        string
			ServerName   string
			DatabaseType string
			ChartData    string
			TimeRange    string
			ColorMap     string
		}{
			State:        "open",
			ServerName:   server,
			ChartData:    string(chartData),
			TimeRange:    string(timeRangeJSON),
			DatabaseType: "mssql",
			ColorMap:     string(colorMapJson),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	err = s.templates.ExecuteTemplate(w, "base.html", map[string]interface{}{
		"ServerList": domain.SampleServers,
		"Slideover": struct {
			State        string
			ServerName   string
			DatabaseType string
			ChartData    string
			TimeRange    string
			ColorMap     string
		}{
			State:        "open",
			ServerName:   server,
			ChartData:    string(chartData),
			TimeRange:    string(timeRangeJSON),
			DatabaseType: "mssql",
			ColorMap:     string(colorMapJson),
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (s *HtmxServer) HandleQuerySampleDetails(w http.ResponseWriter, r *http.Request) {
	snapID := r.URL.Query().Get("snapID")
	sampleID := r.URL.Query().Get("sampleID")
	resp, err := s.client.GetSampleDetails(r.Context(), &dbmv1.GetSampleDetailsRequest{
		SampleId: []byte(sampleID),
		SnapId:   snapID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = resp.QuerySample
	w.Header().Set("Content-Type", "text/html")
	dSample, err := protoSampleToDomain(resp.QuerySample)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	blockChain, err2 := blockChainFromProto(resp.BlockChain)
	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		return
	}
	var plan domain.ParsedExecutionPlan
	if resp.ParsedPlan != nil {
		plan = domain.ProtoParsedPlanToDomain(resp.ParsedPlan)

	}
	if r.Header.Get("Hx-request") == "true" {
		//partial render
		err = s.templates.ExecuteTemplate(w, "samples_modal.html", struct {
			State         string
			QuerySample   domain.QuerySample
			BlockChain    domain.BlockChain
			ExecutionPlan domain.ParsedExecutionPlan
		}{
			State:         "open",
			QuerySample:   dSample,
			BlockChain:    blockChain,
			ExecutionPlan: plan,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	err = s.templates.ExecuteTemplate(w, "base.html", map[string]interface{}{
		"ServerList": domain.SampleServers,
		"Slideover": struct {
			State string
		}{
			State: "closed",
		},
		"SampleModal": struct {
			State string
		}{State: "open"},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

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

func getTimeRange(r *http.Request) (time.Time, time.Time, error) {
	timeRangeInput := r.URL.Query().Get("selected-timerange")
	var startTime time.Time
	endTime := time.Now()
	switch timeRangeInput {
	case "Last 15 minutes":
		startTime = endTime.Add(-15 * time.Minute)
	case "Last 30 minutes":
		startTime = endTime.Add(-30 * time.Minute)
	case "Last 1 hour":
		startTime = endTime.Add(-1 * time.Hour)
	case "Last 3 hours":
		startTime = endTime.Add(-3 * time.Hour)
	case "Last 12 hours":
		startTime = endTime.Add(-12 * time.Hour)
	case "Last 1 day":
		startTime = endTime.Add(-24 * time.Hour)
	case "Last 2 days":
		startTime = endTime.Add(-48 * time.Hour)
	default:
		var startStr, endStr string
		_, err := fmt.Sscanf(timeRangeInput, "%s - %s", &startStr, &endStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("scanning timeRange: %w", err)
		}
		startTime, err = time.Parse("2006-01-02T15:04:05.999Z", startStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("scanning startTime: %w", err)
		}
		endTime, err = time.Parse("2006-01-02T15:04:05.999Z", endStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("scanning endTime: %w", err)
		}
	}
	return startTime, endTime, nil
}

func (s *HtmxServer) HandleSnapshots(w http.ResponseWriter, r *http.Request) {
	startTime, endTime, err := getTimeRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	server := r.URL.Query().Get("selected-server")
	currPageStr := r.URL.Query().Get("page")
	currPage := 1
	if currPageStr != "" {
		currPage, err = strconv.Atoi(currPageStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	pageSize := 5
	resp, err := s.client.ListSnapshots(r.Context(), &dbmv1.ListSnapshotsRequest{
		Start:      timestamppb.New(startTime),
		End:        timestamppb.New(endTime),
		Host:       server,
		Database:   "",
		PageSize:   int32(pageSize),
		PageNumber: int64(currPage - 1),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	qsdata := make([]domain.Snapshot, 0)
	for _, snapshot := range resp.Snapshots {
		users := make(map[string]struct{})
		WaitersNo := 0
		BlockersNo := 0
		WaitDuration := 0
		SumDuration := 0
		MaxDuration := 0
		waitersPerGroup := make(map[string]int)
		for _, sample := range snapshot.Samples {
			users[sample.Session.LoginName] = struct{}{}
			if int(sample.TimeElapsedMillis) > MaxDuration {
				MaxDuration = int(sample.TimeElapsedMillis)
			}
			SumDuration += int(sample.TimeElapsedMillis)
			if sample.Blocked {
				WaitersNo++
				WaitDuration += int(sample.WaitInfo.WaitTime)

			}
			if _, ok := waitersPerGroup[sample.WaitInfo.WaitType]; !ok {
				waitersPerGroup[sample.WaitInfo.WaitType] = 1
			} else {
				waitersPerGroup[sample.WaitInfo.WaitType]++
			}
			if sample.Blocker {
				BlockersNo++
			}

		}
		waitGroups := make([]domain.WaitType, 0)
		for k, v := range waitersPerGroup {
			wait := "none"
			if k != "" {
				wait = k
			}
			color := getWaitEventColor(wait)
			waitGroups = append(waitGroups, domain.WaitType{
				Type:    wait,
				Percent: v * 100 / len(snapshot.Samples),
				Color:   color,
			})
		}
		usersSlice := make([]string, len(users))
		for user := range users {
			usersSlice = append(usersSlice, user)
		}
		AvgDuration := float64(SumDuration) / float64(len(snapshot.Samples))
		slices.SortFunc(waitGroups, func(a, b domain.WaitType) int {
			return b.Percent - a.Percent
		})
		qsdata = append(qsdata, domain.Snapshot{
			ID:           snapshot.Id,
			Timestamp:    snapshot.Timestamp.AsTime(),
			Connections:  len(snapshot.Samples),
			WaitEvGroups: waitGroups,
			Users:        usersSlice,
			WaitersNo:    WaitersNo,
			BlockersNo:   BlockersNo,
			WaitDuration: fmt.Sprintf("%.2d ms", WaitDuration),
			AvgDuration:  fmt.Sprintf("%.2f ms", AvgDuration),
			MaxDuration:  fmt.Sprintf("%.2d ms", MaxDuration),
		})
	}

	// Create the template data
	totalPages := int(resp.TotalCount) / pageSize
	data := struct {
		Snapshots     []domain.Snapshot
		SortDirection string
		SortColumn    string
		CurrentPage   int
		TotalPages    int
		PageRange     []string
		NextPage      int
		PreviousPage  int
	}{
		Snapshots:     qsdata,
		SortDirection: "desc",
		SortColumn:    "",
		CurrentPage:   currPage,
		TotalPages:    totalPages,
		PageRange:     calculatePageRange(currPage, totalPages),
		NextPage:      currPage + 1,
		PreviousPage:  currPage - 1,
	}

	err = s.templates.ExecuteTemplate(w, "active_conn_table.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func calculatePageRange(currentPage, totalPages int) []string {
	pageRange := []string{}

	for i := 1; i < 4; i++ {
		if i <= totalPages {
			pageRange = append(pageRange, strconv.Itoa(i))
		}
	}
	if currentPage > 4 {
		pageRange = append(pageRange, "...")
	}
	for i := currentPage - 2; i < currentPage+3; i++ {
		if i > 3 && i <= totalPages {
			pageRange = append(pageRange, strconv.Itoa(i))
		}
	}
	if currentPage < totalPages-3 {
		pageRange = append(pageRange, "...")
		for i := totalPages - 2; i <= totalPages; i++ {
			pageRange = append(pageRange, strconv.Itoa(i))
		}
	}
	fmt.Println(pageRange)
	return pageRange
}

func (s *HtmxServer) HandleSamples(w http.ResponseWriter, r *http.Request) {
	// Get snapshot ID from the URL
	var snapshotID string
	_, err := fmt.Sscanf(r.URL.Path, "/samples/%s", &snapshotID)
	if err != nil || snapshotID == "" {
		http.Error(w, "Invalid/missing snapshot ID", http.StatusBadRequest)
		return
	}

	resp, err := s.client.GetSnapshot(r.Context(), &dbmv1.GetSnapshotRequest{Id: snapshotID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	querySamplesForSnapshot := make([]domain.QuerySample, len(resp.GetSnapshot().GetSamples()))
	for i, sample := range resp.Snapshot.Samples {
		dSample, err3 := protoSampleToDomain(sample)
		if err3 != nil {
			http.Error(w, err3.Error(), http.StatusInternalServerError)
			return
		}
		querySamplesForSnapshot[i] = dSample
	}

	sort.Slice(querySamplesForSnapshot, func(i, j int) bool {
		a := querySamplesForSnapshot[i]
		b := querySamplesForSnapshot[j]
		if a.IsBlocker && b.IsBlocker {
			return a.SID < b.SID
		}
		if a.IsWaiter && b.IsWaiter && !a.IsBlocker && !b.IsBlocker {
			return a.SID < b.SID
		}
		if a.IsBlocker {
			return true
		}
		if b.IsBlocker {
			return false
		}
		if a.IsWaiter {
			return true
		}
		if b.IsWaiter {
			return false
		}
		return a.SID < b.SID
	})
	// Get the query samples for the snapshot

	// Parse query parameters for sorting query samples
	column := r.URL.Query().Get("column")
	if column == "" {
		column = "query" // Default sort column
	}

	// Sort query samples
	//sortedQuerySamples := SortQuerySamples(querySamplesForSnapshot, column)

	// Create the template data
	data := struct {
		QuerySamples []domain.QuerySample
		SortColumn   string
		SnapID       string
	}{
		QuerySamples: querySamplesForSnapshot,
		SortColumn:   column,
		SnapID:       snapshotID,
	}

	err = s.templates.ExecuteTemplate(w, "samples_table.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

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
		SID:           sid,
		Query:         sample.Text,
		ExecutionTime: fmt.Sprintf("%d ms", sample.TimeElapsedMillis),
		User:          sample.Session.LoginName,
		IsBlocker:     sample.Blocker,
		IsWaiter:      sample.Blocked,
		BlockingTime:  blockTime,
		BlockDetails:  blockDetails,
		WaitEvent:     sample.WaitInfo.WaitType,
		Database:      sample.Db.DatabaseName,
		SampleID:      string(sample.Id),
		SnapID:        sample.SnapInfo.Id,
		SQLHandle:     string(sample.SqlHandle),
		PlanHandle:    string(sample.PlanHandle),
		Status:        sample.Status,
	}
	return dSample, nil
}
