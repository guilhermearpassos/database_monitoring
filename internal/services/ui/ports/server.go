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
	"sort"
	"strconv"
	"strings"
	"time"
)

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

// SortSnapshots sorts snapshots by the given column
func SortSnapshots(snapshots []domain.Snapshot, column string, sortDirection string) []domain.Snapshot {
	switch column {
	case "timestamp":
		sort.Slice(snapshots, func(i, j int) bool {
			if sortDirection == "asc" {
				return snapshots[i].Timestamp.Before(snapshots[j].Timestamp)
			} else {
				return snapshots[i].Timestamp.After(snapshots[j].Timestamp)

			}
		})
	case "db_name":
		sort.Slice(snapshots, func(i, j int) bool {
			if sortDirection == "asc" {
				return strings.ToLower(snapshots[i].DBName) < strings.ToLower(snapshots[j].DBName)
			} else {
				return strings.ToLower(snapshots[i].DBName) > strings.ToLower(snapshots[j].DBName)

			}
		})
	case "status":
		sort.Slice(snapshots, func(i, j int) bool {
			if sortDirection == "asc" {
				return strings.ToLower(snapshots[i].Status) < strings.ToLower(snapshots[j].Status)
			} else {
				return strings.ToLower(snapshots[i].Status) > strings.ToLower(snapshots[j].Status)

			}
		})
	}
	return snapshots
}

// SortQuerySamples sorts query samples by the given column
func SortQuerySamples(querySamples []domain.QuerySample, column string) []domain.QuerySample {
	switch column {
	case "query":
		sort.Slice(querySamples, func(i, j int) bool {
			return querySamples[i].Query < querySamples[j].Query
		})
	case "execution_time":
		sort.Slice(querySamples, func(i, j int) bool {
			return querySamples[i].ExecutionTime < querySamples[j].ExecutionTime
		})
	case "user":
		sort.Slice(querySamples, func(i, j int) bool {
			return querySamples[i].User < querySamples[j].User
		})
	}
	return querySamples
}

func (s *HtmxServer) HandleBaseLayout(w http.ResponseWriter, r *http.Request) {
	err := s.templates.ExecuteTemplate(w, "layout.html", domain.SampleServers)
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

	timeRange := map[string]string{
		"start": startTime.Format("2006-01-02T15:04:05"),
		"end":   endTime.Format("2006-01-02T15:04:05"),
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
	err = s.templates.ExecuteTemplate(w, "slideover.html", struct {
		State      string
		ServerName string
		ChartData  string
		TimeRange  string
	}{
		State:      "open",
		ServerName: server,
		ChartData:  string(chartData),
		TimeRange:  string(timeRangeJSON),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
	resp, err := s.client.ListSnapshots(r.Context(), &dbmv1.ListSnapshotsRequest{
		Start:      timestamppb.New(startTime),
		End:        timestamppb.New(endTime),
		Host:       server,
		Database:   "",
		PageSize:   5,
		PageNumber: 0,
	})
	var qsdata []domain.Snapshot
	if err == nil {
		qsdata = make([]domain.Snapshot, 0)
		for _, snapshot := range resp.Snapshots {
			qsdata = append(qsdata, domain.Snapshot{
				ID:        snapshot.Id,
				Timestamp: snapshot.Timestamp.AsTime(),
				DBName:    "",
				Status:    "",
			})
		}
	} else {
		qsdata = domain.Snapshots
	}
	sortDirection := r.URL.Query().Get("direction")
	if sortDirection == "" {
		sortDirection = "asc" // default sort direction
	}
	// Parse query parameters for sorting
	column := r.URL.Query().Get("column")
	if column == "" {
		column = "timestamp"   // Default sort column
		sortDirection = "desc" // default sort direction
	}

	// Sort snapshots
	sortedSnapshots := SortSnapshots(qsdata, column, sortDirection)

	// Create the template data
	data := struct {
		Snapshots     []domain.Snapshot
		SortDirection string
		SortColumn    string
	}{
		Snapshots:     sortedSnapshots,
		SortDirection: sortDirection,
		SortColumn:    column,
	}

	err = s.templates.ExecuteTemplate(w, "active_conn_table.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (s *HtmxServer) HandleSamples(w http.ResponseWriter, r *http.Request) {
	// Get snapshot ID from the URL
	var snapshotID string
	_, err := fmt.Sscanf(r.URL.Path, "/samples/%s", &snapshotID)
	if err != nil || snapshotID == "" {
		http.NotFound(w, r)
		return
	}
	//startTime, endTime, err := getTimeRange(r)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	return
	//}
	resp, err := s.client.GetSnapshot(r.Context(), &dbmv1.GetSnapshotRequest{Id: snapshotID})
	var querySamplesForSnapshot []domain.QuerySample
	if err == nil {
		querySamplesForSnapshot = make([]domain.QuerySample, len(resp.GetSnapshot().GetSamples()))
		for i, sample := range resp.Snapshot.Samples {
			querySamplesForSnapshot[i] = domain.QuerySample{
				Query:         sample.Text,
				ExecutionTime: strconv.FormatInt(sample.TimeElapsedMillis, 10),
				User:          sample.Session.LoginName,
			}
		}
		//http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		querySamplesForSnapshot = domain.QuerySamples[snapshotID]

	}
	// Get the query samples for the snapshot

	// Parse query parameters for sorting query samples
	column := r.URL.Query().Get("column")
	if column == "" {
		column = "query" // Default sort column
	}

	// Sort query samples
	sortedQuerySamples := SortQuerySamples(querySamplesForSnapshot, column)

	// Create the template data
	data := struct {
		QuerySamples []domain.QuerySample
		SortColumn   string
		SnapID       string
	}{
		QuerySamples: sortedQuerySamples,
		SortColumn:   column,
		SnapID:       snapshotID,
	}

	err = s.templates.ExecuteTemplate(w, "samples_table.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
