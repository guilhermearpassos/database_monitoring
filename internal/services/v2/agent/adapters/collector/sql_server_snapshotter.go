package collector

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/domain/collector"
	"github.com/jmoiron/sqlx"
	mssql "github.com/microsoft/go-mssqldb"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Since CONTEXT_INFO is a raw binary slab, adding a 4-byte signature allows your monitoring tool to distinguish trace data from other garbage or different uses of the context.
// Proposed Structure (30 bytes total):
// 1.Magic Header (4 bytes): 0x54505243 (ASCII for TPRC - Trace PaRent Context).
// 2.Version (1 byte): 0x00 (Standard W3C version).
// 3.Trace ID (16 bytes).
// 4.Span ID (8 bytes).
// 5.Flags (1 byte).
var MagicTraceparent = []byte("TPRC")

type queryCounterState struct {
	counters map[string]int64
	lastSeen time.Time
}

type SqlServerSnapshotter struct {
	db                *sqlx.DB
	Server            common_domain.ServerMeta
	tracer            trace.Tracer
	metricsMu         *sync.Mutex
	lastQueryCounters map[string]queryCounterState
}

func NewSqlServerSnapshotter(db *sqlx.DB, server common_domain.ServerMeta) *SqlServerSnapshotter {
	return &SqlServerSnapshotter{
		db:                db,
		tracer:            otel.Tracer("SQLServerSnapshotter"),
		Server:            server,
		metricsMu:         &sync.Mutex{},
		lastQueryCounters: make(map[string]queryCounterState),
	}
}

var _ collector.Snapshotter = (*SqlServerSnapshotter)(nil)

func parseContext(raw []byte) common_domain.ContextInfo {
	// Check that raw starts with MagicTraceparent and contains full TraceID + SpanID
	if len(raw) >= 29 && bytes.HasPrefix(raw, MagicTraceparent) {
		return common_domain.ContextInfo{
			// Trace ID: 16 bytes starting at index 5
			TraceId: hex.EncodeToString(raw[5:21]),
			// Span ID: 8 bytes starting at index 21
			SpanId:        hex.EncodeToString(raw[21:29]),
			IsTraceParent: true,
			Raw:           raw,
		}
	}
	return common_domain.ContextInfo{Raw: raw}
}

// isValidPlanHandle reports whether raw bytes can be a valid SQL Server plan_handle
// for sys.dm_exec_query_plan (non-empty, <= 64 bytes, and not all zeroes like 0x0 or 0x00...00).
func isValidPlanHandle(raw []byte) bool {
	if len(raw) == 0 || len(raw) > 64 {
		return false
	}
	for _, b := range raw {
		if b != 0 {
			return true
		}
	}
	return false
}

func encodeValidPlanHandle(raw []byte) string {
	if !isValidPlanHandle(raw) {
		return ""
	}
	return base64.StdEncoding.EncodeToString(raw)
}

const (
	monitoringSessionPrefix = "SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED; SET DEADLOCK_PRIORITY LOW; SET LOCK_TIMEOUT 2000; "
	maxPlansPerFetch        = 50
)

func isInvalidPlanHandleErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "handle that was passed to dm_exec_query_plan was invalid") ||
		strings.Contains(msg, "handle that was passed to dm_exec_text_query_plan was invalid")
}

func PackTraceContext(ctx context.Context) []byte {
	sc := trace.SpanFromContext(ctx).SpanContext()
	buf := make([]byte, 128)
	if !sc.IsValid() {
		return buf
	}

	// 1. Add Magic Header 'TPRC'
	copy(buf[0:4], MagicTraceparent)
	// 2. Version
	buf[4] = 0
	// 3. TraceID (16 bytes)
	tid := sc.TraceID()
	copy(buf[5:21], tid[:])
	// 4. SpanID (8 bytes)
	sid := sc.SpanID()
	copy(buf[21:29], sid[:])
	// 5. Flags (1 byte)
	buf[29] = byte(sc.TraceFlags())

	return buf
}
func (s SqlServerSnapshotter) TakeSnapshot(ctx context.Context, databases []string) (*common_domain.DataBaseSnapshot, error) {
	ctx, span := s.tracer.Start(ctx, "TakeSnapshot")
	defer span.End()
	traceContext := PackTraceContext(ctx)
	qDBName := monitoringSessionPrefix + `SET CONTEXT_INFO ?;select database_id, name from sys.databases OPTION (MAXDOP 1)`
	db := s.db
	rowsDB, err := db.QueryContext(ctx, qDBName, traceContext)
	if err != nil {
		return nil, fmt.Errorf("queryDatabases: %w", err)
	}
	dbInfo := make(map[string]common_domain.DataBaseMetadata)
	defer func(rowsDB *sql.Rows) {
		_ = rowsDB.Close()
	}(rowsDB)
	for rowsDB.Next() {
		var dbID string
		var name string
		err = rowsDB.Scan(&dbID, &name)
		if err != nil {
			return nil, fmt.Errorf("queryDatabases scan: %w", err)
		}
		dbInfo[dbID] = common_domain.DataBaseMetadata{
			DatabaseID:   dbID,
			DatabaseName: name,
		}
	}
	err = rowsDB.Err()
	if err != nil {
		return nil, fmt.Errorf("queryDatabases orwsErr: %w", err)
	}
	_ = rowsDB.Close()
	snapID := uuid.NewString()
	snapTime := time.Now().In(time.UTC)
	query := monitoringSessionPrefix + `
SET CONTEXT_INFO ?;SELECT s.session_id,
       s.login_time,
       s.host_name,
       s.program_name,
       s.login_name,
       s.status,
       s.cpu_time,
       s.memory_usage,
       p.total_elapsed_time,
       s.last_request_start_time,
       s.last_request_end_time,
       s.reads,
       s.writes,
       s.logical_reads,
       s.row_count,
       s.database_id,
       p.blocking_session_id,
       p.wait_type,
       p.wait_time,
       p.last_wait_type,
       p.wait_resource,
       p.status,
       sql_handle,
  	   plan_handle,
       text, p.request_id, p.transaction_id, p.connection_id, p.percent_complete, p.estimated_completion_time, s.transaction_isolation_level,
       query_hash, isnull(c.client_net_address, '') as client_net_address,
       p.context_info,
       s.context_info
FROM sys.dm_exec_sessions s
         inner join sys.dm_exec_requests  p on p.session_id = s.session_id
left JOIN sys.dm_exec_connections AS c on p.connection_id = c.connection_id
         CROSS APPLY sys.dm_exec_sql_text(sql_handle)
	 where text is not null
OPTION (MAXDOP 1)
`
	rows, err := db.QueryxContext(ctx, query, traceContext)
	if err != nil {
		return nil, err
	}
	defer func(rows *sqlx.Rows) {
		_ = rows.Close()
	}(rows)
	querySamplesByDB := make(map[string][]*common_domain.QuerySample)
	blockingMap := make(map[int][]string)
	for rows.Next() {
		var sessionID int
		var loginTime time.Time
		var hostName string
		var programName string
		var loginName string
		var status string
		var cpuTime int
		var memoryUsage int
		var totalElapsedTime int
		var lastRequestStartTime time.Time
		var lastRequestEndTime time.Time
		var reads string
		var writes string
		var logicalReads string
		var rowCount int
		var databaseId int
		var blockingSessionId int
		var waitType *string
		var waitTime int
		var lastWaitType string
		var waitResource string
		var pStatus string
		var sqlHandle []byte
		var planHandle []byte
		var text string
		var requestId int
		var transactionId int
		var connectionId mssql.UniqueIdentifier
		var percentComplete float64
		var estimatedCompletionTime int
		var transactionIsolationLevel int
		var queryHash []byte
		var clientNetAddress string
		var reqContext []byte
		var sesContext []byte
		err = rows.Scan(&sessionID,
			&loginTime,
			&hostName,
			&programName,
			&loginName,
			&status,
			&cpuTime,
			&memoryUsage,
			&totalElapsedTime,
			&lastRequestStartTime,
			&lastRequestEndTime,
			&reads,
			&writes,
			&logicalReads,
			&rowCount,
			&databaseId,
			&blockingSessionId,
			&waitType,
			&waitTime,
			&lastWaitType,
			&waitResource,
			&pStatus,
			&sqlHandle,
			&planHandle,
			&text,
			&requestId,
			&transactionId,
			&connectionId,
			&percentComplete,
			&estimatedCompletionTime,
			&transactionIsolationLevel,
			&queryHash,
			&clientNetAddress,
			&reqContext,
			&sesContext,
		)
		if err != nil {
			return nil, err
		}
		if len(databases) > 0 && !slices.Contains(databases, dbInfo[strconv.Itoa(databaseId)].DatabaseName) {
			continue
		}
		var blockedBy string
		if blockingSessionId != 0 {
			blockedBy = strconv.Itoa(blockingSessionId)
			bl, ok := blockingMap[blockingSessionId]
			if !ok {
				bl = make([]string, 0)
			}
			bl = append(bl, strconv.Itoa(sessionID))
			blockingMap[blockingSessionId] = bl
		}
		sampleId := []byte(fmt.Sprintf("%s_%d_%d_%d", connectionId, sessionID, transactionId, requestId))
		qs := common_domain.QuerySample{
			Id:         base64.StdEncoding.EncodeToString(sampleId),
			Status:     pStatus,
			Cmd:        "",
			SqlHandle:  base64.StdEncoding.EncodeToString(sqlHandle),
			PlanHandle: encodeValidPlanHandle(planHandle),
			QueryHash:  base64.StdEncoding.EncodeToString(queryHash),
			Text:       text,
			IsBlocked:  blockingSessionId != 0,
			IsBlocker:  false,
			Session: common_domain.SessionMetadata{
				SessionID:            strconv.Itoa(sessionID),
				LoginTime:            loginTime,
				HostName:             hostName,
				ProgramName:          programName,
				LoginName:            loginName,
				Status:               status,
				LastRequestStartTime: lastRequestStartTime,
				LastRequestEndTime:   lastRequestEndTime,
				ConnectionId:         connectionId.String(),
				ClientIP:             clientNetAddress,
				Context:              parseContext(sesContext),
			},
			Database: common_domain.DataBaseMetadata{
				DatabaseID:   strconv.Itoa(databaseId),
				DatabaseName: dbInfo[strconv.Itoa(databaseId)].DatabaseName,
			},
			Block: common_domain.BlockMetadata{
				BlockedBy:       blockedBy,
				BlockedSessions: make([]string, 0),
			},
			Wait: common_domain.WaitMetadata{
				WaitType:     waitType,
				WaitTime:     waitTime,
				LastWaitType: lastWaitType,
				WaitResource: waitResource,
			},
			Snapshot: common_domain.SnapshotMetadata{
				ID:        snapID,
				Timestamp: snapTime,
			},
			TimeElapsedMs: int64(totalElapsedTime),
			CommandMetadata: common_domain.CommandMetadata{
				TransactionId:           strconv.Itoa(transactionId),
				RequestId:               strconv.Itoa(requestId),
				EstimatedCompletionTime: int64(estimatedCompletionTime),
				PercentComplete:         percentComplete,
			},
			Context: parseContext(reqContext),
		}
		if _, ok := querySamplesByDB[strconv.Itoa(databaseId)]; !ok {
			querySamplesByDB[strconv.Itoa(databaseId)] = make([]*common_domain.QuerySample, 0)
		}
		querySamplesByDB[strconv.Itoa(databaseId)] = append(querySamplesByDB[strconv.Itoa(databaseId)], &qs)

	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	querySamples := make([]*common_domain.QuerySample, 0)
	for _, qs2 := range querySamplesByDB {
		for _, qs := range qs2 {
			var sessionID int
			sessionID, err = strconv.Atoi(qs.Session.SessionID)
			if bl, ok := blockingMap[sessionID]; ok {
				qs.SetBlockedIds(bl)
				delete(blockingMap, sessionID)
			}
		}
		querySamples = append(querySamples, qs2...)
	}
	missingBlockingSessionIds := make([]int, 0, len(blockingMap))
	for i := range blockingMap {
		missingBlockingSessionIds = append(missingBlockingSessionIds, i)
	}

	sleepingSamples, err := s.getSleepingBlockingSessions(ctx, missingBlockingSessionIds, dbInfo,
		snapID,
		snapTime)
	if err != nil {
		return nil, fmt.Errorf("getSleepingBlockingSessions: %w", err)
	}
	for _, qs := range sleepingSamples {
		var sessionID int
		sessionID, err = strconv.Atoi(qs.Session.SessionID)
		if bl, ok := blockingMap[sessionID]; ok {
			qs.SetBlockedIds(bl)
			delete(blockingMap, sessionID)
		}
	}
	querySamples = append(querySamples, sleepingSamples...)
	return &common_domain.DataBaseSnapshot{
		Samples: querySamples,
		SnapInfo: common_domain.SnapInfo{
			ID:        snapID,
			Timestamp: snapTime,
			Server:    s.Server,
		},
	}, nil
}
func (s SqlServerSnapshotter) getSleepingBlockingSessions(ctx context.Context, ids []int, dbInfo map[string]common_domain.DataBaseMetadata, snapID string, snapTime time.Time) ([]*common_domain.QuerySample, error) {
	if len(ids) == 0 {
		return []*common_domain.QuerySample{}, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(monitoringSessionPrefix+`
SELECT s.session_id,
       s.login_time,
       s.host_name,
       s.program_name,
       s.login_name,
       s.status,
       s.cpu_time,
       s.memory_usage,
       0 AS total_elapsed_time,
       s.last_request_start_time,
       s.last_request_end_time,
       s.reads,
       s.writes,
       s.logical_reads,
       s.row_count,
       s.database_id,
       0 AS blocking_session_id,
       NULL AS wait_type,
       0 AS wait_time,
       '' AS last_wait_type,
       '' AS wait_resource,
       s.status,
       c.most_recent_sql_handle AS sql_handle,
       0x0 AS plan_handle,
       t.text,
       0 AS request_id,
       0 AS transaction_id,
       c.connection_id,
       0 AS percent_complete,
       0 AS estimated_completion_time,
       s.transaction_isolation_level,
       0x0 as query_hash,
       isnull(c.client_net_address, '') as client_net_address 

FROM sys.dm_exec_sessions s
INNER JOIN sys.dm_exec_connections c
    ON s.session_id = c.session_id
CROSS APPLY sys.dm_exec_sql_text(c.most_recent_sql_handle) t
WHERE s.status = 'sleeping'
  AND c.most_recent_sql_handle IS NOT NULL
  AND t.text IS NOT NULL
    AND s.session_id IN (%s)
OPTION (MAXDOP 1)`, strings.Join(placeholders, ","))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying sleeping sessions: %w", err)
	}
	defer rows.Close()

	var result []*common_domain.QuerySample
	for rows.Next() {

		var sessionID int
		var loginTime time.Time
		var hostName string
		var programName string
		var loginName string
		var status string
		var cpuTime int
		var memoryUsage int
		var totalElapsedTime int
		var lastRequestStartTime time.Time
		var lastRequestEndTime time.Time
		var reads string
		var writes string
		var logicalReads string
		var rowCount int
		var databaseId int
		var blockingSessionId int
		var waitType *string
		var waitTime int
		var lastWaitType string
		var waitResource string
		var pStatus string
		var sqlHandle []byte
		var planHandle []byte
		var text string
		var requestId int
		var transactionId int
		var connectionId mssql.UniqueIdentifier
		var percentComplete float64
		var estimatedCompletionTime int
		var transactionIsolationLevel int
		var queryHash []byte
		var clientNetAddress string
		err = rows.Scan(&sessionID,
			&loginTime,
			&hostName,
			&programName,
			&loginName,
			&status,
			&cpuTime,
			&memoryUsage,
			&totalElapsedTime,
			&lastRequestStartTime,
			&lastRequestEndTime,
			&reads,
			&writes,
			&logicalReads,
			&rowCount,
			&databaseId,
			&blockingSessionId,
			&waitType,
			&waitTime,
			&lastWaitType,
			&waitResource,
			&pStatus,
			&sqlHandle,
			&planHandle,
			&text,
			&requestId,
			&transactionId,
			&connectionId,
			&percentComplete,
			&estimatedCompletionTime,
			&transactionIsolationLevel,
			&queryHash,
			&clientNetAddress,
		)
		if err != nil {
			return nil, err
		}
		var blockedBy string
		sampleId := []byte(fmt.Sprintf("%s_%d_%d_%d", connectionId, sessionID, transactionId, requestId))
		qs := common_domain.QuerySample{
			Id:         base64.StdEncoding.EncodeToString(sampleId),
			Status:     pStatus,
			Cmd:        "",
			SqlHandle:  base64.StdEncoding.EncodeToString(sqlHandle),
			PlanHandle: encodeValidPlanHandle(planHandle),
			QueryHash:  base64.StdEncoding.EncodeToString(queryHash),
			Text:       text,
			IsBlocked:  blockingSessionId != 0,
			IsBlocker:  false,
			Session: common_domain.SessionMetadata{
				SessionID:            strconv.Itoa(sessionID),
				LoginTime:            loginTime,
				HostName:             hostName,
				ProgramName:          programName,
				LoginName:            loginName,
				Status:               status,
				LastRequestStartTime: lastRequestStartTime,
				LastRequestEndTime:   lastRequestEndTime,
				ConnectionId:         connectionId.String(),
				ClientIP:             clientNetAddress,
			},
			Database: common_domain.DataBaseMetadata{
				DatabaseID:   strconv.Itoa(databaseId),
				DatabaseName: dbInfo[strconv.Itoa(databaseId)].DatabaseName,
			},
			Block: common_domain.BlockMetadata{
				BlockedBy:       blockedBy,
				BlockedSessions: make([]string, 0),
			},
			Wait: common_domain.WaitMetadata{
				WaitType:     waitType,
				WaitTime:     waitTime,
				LastWaitType: lastWaitType,
				WaitResource: waitResource,
			},
			Snapshot: common_domain.SnapshotMetadata{
				ID:        snapID,
				Timestamp: snapTime,
			},
			TimeElapsedMs: int64(totalElapsedTime),
			CommandMetadata: common_domain.CommandMetadata{
				TransactionId:           strconv.Itoa(transactionId),
				RequestId:               strconv.Itoa(requestId),
				EstimatedCompletionTime: int64(estimatedCompletionTime),
				PercentComplete:         percentComplete,
			},
		}
		result = append(result, &qs)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("row scan failed: %v", err)
	}
	return result, nil
}

func encodePlanHandlesJSON(rawHandles [][]byte) (string, error) {
	hexHandles := make([]string, len(rawHandles))
	for i, raw := range rawHandles {
		hexHandles[i] = "0x" + hex.EncodeToString(raw)
	}
	b, err := json.Marshal(hexHandles)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s SqlServerSnapshotter) FetchExecutionPlans(ctx context.Context, handles []string) (*collector.ExecPlanChunk, error) {
	ctx, span := s.tracer.Start(ctx, "GetPlanHandles")
	defer span.End()

	type validHandle struct {
		encoded string
		raw     []byte
	}
	seen := make(map[string]struct{}, len(handles))
	validHandles := make([]validHandle, 0, min(len(handles), maxPlansPerFetch))
	rawHandles := make([][]byte, 0, min(len(handles), maxPlansPerFetch))
	for _, handle := range handles {
		if len(validHandles) >= maxPlansPerFetch {
			break
		}
		if handle == "" {
			continue
		}
		if _, dup := seen[handle]; dup {
			continue
		}
		decoded, err := base64.StdEncoding.DecodeString(handle)
		if err != nil || !isValidPlanHandle(decoded) {
			continue
		}
		seen[handle] = struct{}{}
		validHandles = append(validHandles, validHandle{encoded: handle, raw: decoded})
		rawHandles = append(rawHandles, decoded)
	}
	if len(rawHandles) == 0 {
		return &collector.ExecPlanChunk{Server: s.Server}, nil
	}

	ret, err3 := s.batchFetchPlanHandles(ctx, s.db, rawHandles, s.Server)
	if err3 != nil {
		span.RecordError(err3)
		ret = make(map[string]*common_domain.ExecutionPlan)
		// Fallback to 1-by-1 strategy if OPENJSON is unavailable (e.g. compat level < 130) or a single invalid handle broke the batch.
		query := monitoringSessionPrefix + "select query_plan from sys.dm_exec_text_query_plan(?, 0, -1) OPTION (MAXDOP 1)"
		for _, h := range validHandles {
			row := s.db.QueryRowContext(ctx, query, h.raw)
			if err := row.Err(); err != nil {
				if isInvalidPlanHandleErr(err) {
					span.RecordError(err)
					continue
				}
				return &collector.ExecPlanChunk{Server: s.Server, Plans: ret}, fmt.Errorf("fetch plan handle - %w", err)
			}
			var queryPlan *string
			if err := row.Scan(&queryPlan); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					continue
				}
				if isInvalidPlanHandleErr(err) {
					span.RecordError(err)
					continue
				}
				return &collector.ExecPlanChunk{Server: s.Server, Plans: ret}, fmt.Errorf("fetch plan handle scan - %w", err)
			}
			if queryPlan == nil {
				continue
			}
			ret[h.encoded] = &common_domain.ExecutionPlan{
				PlanHandle: h.encoded,
				XmlData:    *queryPlan,
				Server:     s.Server,
			}
		}
	}
	return &collector.ExecPlanChunk{Server: s.Server, Plans: ret}, nil
}

func (s SqlServerSnapshotter) batchFetchPlanHandles(ctx context.Context, db *sqlx.DB, rawHandles [][]byte, server common_domain.ServerMeta) (map[string]*common_domain.ExecutionPlan, error) {
	handlesJSON, err := encodePlanHandlesJSON(rawHandles)
	if err != nil {
		return nil, fmt.Errorf("encode handles json: %w", err)
	}
	query := monitoringSessionPrefix + `
select h.handle, p.query_plan
from OPENJSON(?) j
cross apply (select CONVERT(varbinary(64), j.[value], 1) as handle) h
cross apply sys.dm_exec_text_query_plan(h.handle, 0, -1) p
OPTION (MAXDOP 1)
`
	rows, err2 := db.QueryContext(ctx, query, handlesJSON)
	if err2 != nil {
		return nil, fmt.Errorf("fetch plans: %w", err2)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)
	ret := make(map[string]*common_domain.ExecutionPlan)
	for rows.Next() {
		var handle []byte
		var queryPlan *string
		err = rows.Scan(&handle, &queryPlan)
		if err != nil {
			return nil, fmt.Errorf("fetch plans - scan: %w", err)
		}
		if queryPlan == nil {
			continue
		}
		b64Handle := base64.StdEncoding.EncodeToString(handle)
		ret[b64Handle] = &common_domain.ExecutionPlan{
			PlanHandle: b64Handle,
			XmlData:    *queryPlan,
			Server:     server,
		}

	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("fetch plans - err: %w", err)
	}
	return ret, nil
}

// computeQueryMetricDelta computes per-interval counter deltas from cumulative sys.dm_exec_query_stats counters.
// Returns (delta, true) when a sample should be emitted, or (nil, false) when only the baseline should be updated.
func computeQueryMetricDelta(prev, curr map[string]int64) (map[string]int64, bool) {
	currExec := curr["executionCount"]
	if currExec <= 0 {
		return nil, false
	}
	if prev == nil {
		// On first observation, skip very old cached queries with large cumulative counts to avoid startup spikes.
		if currExec > 30 {
			return nil, false
		}
		out := make(map[string]int64, len(curr))
		for k, v := range curr {
			out[k] = v
		}
		return out, true
	}

	prevExec := prev["executionCount"]
	if currExec == prevExec {
		return nil, false
	}
	if currExec < prevExec {
		// Plan cache entry was evicted/reset and recreated; emit only if small, otherwise reset baseline.
		if currExec > 30 {
			return nil, false
		}
		out := make(map[string]int64, len(curr))
		for k, v := range curr {
			out[k] = v
		}
		return out, true
	}

	out := make(map[string]int64, len(curr))
	for k, v := range curr {
		diff := v - prev[k]
		if diff < 0 {
			diff = 0
		}
		out[k] = diff
	}
	return out, true
}

func (s SqlServerSnapshotter) CollectQueryMetrics(ctx context.Context, databases []string) ([]*common_domain.QueryMetric, error) {
	ctx, span := s.tracer.Start(ctx, "CollectQueryMetrics")
	defer span.End()

	collectionTime := time.Now().UTC()
	query := monitoringSessionPrefix + `
with qstats as (
    select query_hash,
           plan_handle,
           last_execution_time,
           last_elapsed_time,
           CONCAT(
               CONVERT(binary(64), plan_handle),
               CONVERT(binary(4), statement_start_offset),
               CONVERT(binary(4), statement_end_offset)
           ) as plan_handle_and_offsets,
           (select value from sys.dm_exec_plan_attributes(plan_handle) where attribute = 'dbid') as dbid,
           execution_count,
           total_worker_time,
           total_physical_reads,
           total_logical_writes,
           total_logical_reads,
           total_clr_time,
           total_elapsed_time,
           total_rows,
           total_dop,
           total_grant_kb,
           total_used_grant_kb,
           total_ideal_grant_kb,
           total_reserved_threads,
           total_used_threads,
           total_columnstore_segment_reads,
           total_columnstore_segment_skips,
           total_spills
    from sys.dm_exec_query_stats
    where last_execution_time > dateadd(second, -120, getdate())
),
qstats_aggr as (
    select query_hash,
           isnull(cast(s.dbid as int), 0)       as dbid,
           isnull(d.name, '')                   as db_name,
           max(plan_handle_and_offsets)         as plan_handle_and_offsets,
           max(last_execution_time)             as last_execution_time,
           max(last_elapsed_time)               as last_elapsed_time,
           sum(execution_count)                 as execution_count,
           sum(total_worker_time)               as total_worker_time,
           sum(total_physical_reads)            as total_physical_reads,
           sum(total_logical_writes)            as total_logical_writes,
           sum(total_logical_reads)             as total_logical_reads,
           sum(total_clr_time)                  as total_clr_time,
           sum(total_elapsed_time)              as total_elapsed_time,
           sum(total_rows)                      as total_rows,
           sum(total_dop)                       as total_dop,
           sum(total_grant_kb)                  as total_grant_kb,
           sum(total_used_grant_kb)             as total_used_grant_kb,
           sum(total_ideal_grant_kb)            as total_ideal_grant_kb,
           sum(total_reserved_threads)          as total_reserved_threads,
           sum(total_used_threads)              as total_used_threads,
           sum(total_columnstore_segment_reads) as total_columnstore_segment_reads,
           sum(total_columnstore_segment_skips) as total_columnstore_segment_skips,
           sum(total_spills)                    as total_spills
    from qstats s
             left join sys.databases d on s.dbid = d.database_id
    group by query_hash, isnull(cast(s.dbid as int), 0), isnull(d.name, '')
),
qstats_aggr_split as (
    select convert(varbinary(64), substring(plan_handle_and_offsets, 1, 64))  as plan_handle,
           convert(int, convert(varbinary(4), substring(plan_handle_and_offsets, 65, 4))) as statement_start_offset,
           convert(int, convert(varbinary(4), substring(plan_handle_and_offsets, 69, 4))) as statement_end_offset,
           query_hash,
           dbid,
           db_name,
           last_execution_time,
           last_elapsed_time,
           execution_count,
           total_worker_time,
           total_physical_reads,
           total_logical_writes,
           total_logical_reads,
           total_clr_time,
           total_elapsed_time,
           total_rows,
           total_dop,
           total_grant_kb,
           total_used_grant_kb,
           total_ideal_grant_kb,
           total_reserved_threads,
           total_used_threads,
           total_columnstore_segment_reads,
           total_columnstore_segment_skips,
           total_spills
    from qstats_aggr
)
select qas.query_hash,
       qas.dbid,
       qas.db_name,
       qas.last_execution_time,
       qas.last_elapsed_time,
       qas.execution_count,
       qas.total_worker_time,
       qas.total_physical_reads,
       qas.total_logical_writes,
       qas.total_logical_reads,
       qas.total_clr_time,
       qas.total_elapsed_time,
       qas.total_rows,
       qas.total_dop,
       qas.total_grant_kb,
       qas.total_used_grant_kb,
       qas.total_ideal_grant_kb,
       qas.total_reserved_threads,
       qas.total_used_threads,
       qas.total_columnstore_segment_reads,
       qas.total_columnstore_segment_skips,
       qas.total_spills,
       CASE
           WHEN qas.statement_start_offset >= 0 AND DATALENGTH(st.text) > 0 THEN
               SUBSTRING(
                   st.text,
                   (qas.statement_start_offset / 2) + 1,
                   ((CASE qas.statement_end_offset
                       WHEN -1 THEN DATALENGTH(st.text)
                       ELSE qas.statement_end_offset
                     END - qas.statement_start_offset) / 2) + 1
               )
           ELSE st.text
       END as statement_text
from qstats_aggr_split qas
         cross apply sys.dm_exec_sql_text(qas.plan_handle) st
where st.text is not null
OPTION (MAXDOP 1)
`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("collecting query metrics: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	ret := make([]*common_domain.QueryMetric, 0)
	for rows.Next() {
		var queryHash []byte
		var dbId int
		var dbName string
		var lastExecutionTime time.Time
		var lastElapsedTime int64
		var executionCount int64
		var totalWorkerTime int64
		var totalPhysicalReads int64
		var totalLogicalWrites int64
		var totalLogicalReads int64
		var totalClrTime int64
		var totalElapsedTime int64
		var totalRows int64
		var totalDop int64
		var totalGrantKb int64
		var totalUsedGrantKb int64
		var totalIdealGrantKb int64
		var totalReservedThreads int64
		var totalUsedThreads int64
		var totalColumnstoreSegmentReads int64
		var totalColumnstoreSegmentSkips int64
		var totalSpills int64
		var text sql.NullString

		err = rows.Scan(
			&queryHash,
			&dbId,
			&dbName,
			&lastExecutionTime,
			&lastElapsedTime,
			&executionCount,
			&totalWorkerTime,
			&totalPhysicalReads,
			&totalLogicalWrites,
			&totalLogicalReads,
			&totalClrTime,
			&totalElapsedTime,
			&totalRows,
			&totalDop,
			&totalGrantKb,
			&totalUsedGrantKb,
			&totalIdealGrantKb,
			&totalReservedThreads,
			&totalUsedThreads,
			&totalColumnstoreSegmentReads,
			&totalColumnstoreSegmentSkips,
			&totalSpills,
			&text,
		)
		if err != nil {
			return nil, fmt.Errorf("collecting query metrics - scan: %w", err)
		}
		if len(queryHash) == 0 || !text.Valid || strings.TrimSpace(text.String) == "" {
			continue
		}
		if len(databases) > 0 && !slices.Contains(databases, dbName) {
			continue
		}

		counters := map[string]int64{
			"executionCount":               executionCount,
			"totalWorkerTime":              totalWorkerTime,
			"totalPhysicalReads":           totalPhysicalReads,
			"totalLogicalWrites":           totalLogicalWrites,
			"totalLogicalReads":            totalLogicalReads,
			"totalClrTime":                 totalClrTime,
			"totalElapsedTime":             totalElapsedTime,
			"totalRows":                    totalRows,
			"totalDop":                     totalDop,
			"totalGrantKb":                 totalGrantKb,
			"totalUsedGrantKb":             totalUsedGrantKb,
			"totalIdealGrantKb":            totalIdealGrantKb,
			"totalReservedThreads":         totalReservedThreads,
			"totalUsedThreads":             totalUsedThreads,
			"totalColumnstoreSegmentReads": totalColumnstoreSegmentReads,
			"totalColumnstoreSegmentSkips": totalColumnstoreSegmentSkips,
			"totalSpills":                  totalSpills,
		}

		encodedHash := base64.StdEncoding.EncodeToString(queryHash)
		counterKey := fmt.Sprintf("%s:%d", encodedHash, dbId)

		var prev map[string]int64
		if s.metricsMu != nil {
			s.metricsMu.Lock()
			if state, ok := s.lastQueryCounters[counterKey]; ok {
				prev = state.counters
			}
			s.lastQueryCounters[counterKey] = queryCounterState{
				counters: counters,
				lastSeen: collectionTime,
			}
			s.metricsMu.Unlock()
		}

		digestedCounters, emit := computeQueryMetricDelta(prev, counters)
		if !emit {
			continue
		}

		ret = append(ret, &common_domain.QueryMetric{
			QueryHash: encodedHash,
			Text:      strings.TrimSpace(text.String),
			Database: common_domain.DataBaseMetadata{
				DatabaseID:   strconv.Itoa(dbId),
				DatabaseName: dbName,
			},
			LastExecutionTime: lastExecutionTime,
			LastElapsedTime:   time.Duration(lastElapsedTime) * time.Microsecond,
			Counters:          digestedCounters,
			CollectionTime:    collectionTime,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("collecting query metrics - rows.Err: %w", err)
	}
	s.pruneStaleQueryCounters(collectionTime, 15*time.Minute)
	return ret, nil
}

func (s SqlServerSnapshotter) pruneStaleQueryCounters(now time.Time, maxAge time.Duration) {
	if s.metricsMu == nil {
		return
	}
	cutoff := now.Add(-maxAge)
	s.metricsMu.Lock()
	for k, state := range s.lastQueryCounters {
		if state.lastSeen.Before(cutoff) {
			delete(s.lastQueryCounters, k)
		}
	}
	s.metricsMu.Unlock()
}
