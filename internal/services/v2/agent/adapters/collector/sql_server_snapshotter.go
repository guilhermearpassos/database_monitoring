package collector

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/agent/domain/collector"
	"github.com/jmoiron/sqlx"
	mssql "github.com/microsoft/go-mssqldb"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type SqlServerSnapshotter struct {
	db     *sqlx.DB
	Server common_domain.ServerMeta
	tracer trace.Tracer
}

func NewSqlServerSnapshotter(db *sqlx.DB, server common_domain.ServerMeta) *SqlServerSnapshotter {
	return &SqlServerSnapshotter{
		db:     db,
		tracer: otel.Tracer("SQLServerSnapshotter"),
		Server: server,
	}
}

var _ collector.Snapshotter = (*SqlServerSnapshotter)(nil)

func (s SqlServerSnapshotter) TakeSnapshot(ctx context.Context, databases []string) (*common_domain.DataBaseSnapshot, error) {
	qDBName := `select database_id, name from sys.databases`
	db := s.db
	rowsDB, err := db.QueryContext(ctx, qDBName)
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
	query := `
SELECT s.session_id,
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
       query_hash, isnull(c.client_net_address, '') as client_net_address
FROM sys.dm_exec_sessions s
         inner join sys.dm_exec_requests  p on p.session_id = s.session_id
left JOIN sys.dm_exec_connections AS c on s.session_id = c.session_id
         CROSS APPLY sys.dm_exec_sql_text(sql_handle)
	 where text is not null
`
	rows, err := db.QueryxContext(ctx, query)
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
			PlanHandle: base64.StdEncoding.EncodeToString(planHandle),
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

	query := fmt.Sprintf(`
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
LEFT JOIN sys.dm_exec_query_stats qs
    ON c.most_recent_sql_handle = qs.sql_handle
WHERE s.status = 'sleeping'
  AND c.most_recent_sql_handle IS NOT NULL
  AND t.text IS NOT NULL
    AND s.session_id IN (%s)`, strings.Join(placeholders, ","))

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
			PlanHandle: base64.StdEncoding.EncodeToString(planHandle),
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

func (s SqlServerSnapshotter) FetchExecutionPlans(ctx context.Context, handles []string) (*collector.ExecPlanChunk, error) {
	ctx, span := s.tracer.Start(ctx, "GetPlanHandles")
	defer span.End()
	handles2 := make([]interface{}, 0, len(handles))
	for _, handle := range handles {
		decoded, err := base64.StdEncoding.DecodeString(handle)
		if err == nil {
			handles2 = append(handles2, decoded)
		}
	}
	if len(handles2) == 0 {
		return &collector.ExecPlanChunk{Server: s.Server}, nil
	}

	ret, err3 := s.batchFetchPlanHandles(ctx, s.db, handles2, s.Server)
	if err3 != nil {
		ret = make(map[string]*common_domain.ExecutionPlan)
		//fallback to 1 by 1 strategy, as there might be a problem with tempdb
		query := "select query_plan from sys.dm_exec_query_plan(?)"
		for _, handle := range handles {
			decodeString, err3 := base64.StdEncoding.DecodeString(handle)
			if err3 != nil {
				span.RecordError(err3)
				continue
			}
			row := s.db.QueryRowContext(ctx, query, decodeString)
			err := row.Err()
			if err != nil {
				return &collector.ExecPlanChunk{Server: s.Server, Plans: ret}, fmt.Errorf("fetch plan handle - %w", err)
			}
			var queryPlan *string
			err = row.Scan(&queryPlan)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					continue
				}
				return &collector.ExecPlanChunk{Server: s.Server, Plans: ret}, fmt.Errorf("fetch plan handle scan - %w", err)
			}
			if queryPlan == nil {
				continue
			}
			ret[handle] = &common_domain.ExecutionPlan{
				PlanHandle: handle,
				XmlData:    *queryPlan,
				Server:     s.Server,
			}
		}
	}
	return &collector.ExecPlanChunk{Server: s.Server, Plans: ret}, nil
}

func (s SqlServerSnapshotter) batchFetchPlanHandles(ctx context.Context, db *sqlx.DB, handles2 []interface{}, server common_domain.ServerMeta) (map[string]*common_domain.ExecutionPlan, error) {

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadUncommitted})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)
	createTempTable := `
create table #temp_plans (handle varbinary(64) not null)`
	insertIds := fmt.Sprintf(`
insert into #temp_plans (handle) values %s
`, strings.Join(slices.Repeat([]string{"(?)"}, len(handles2)), ","))
	query := `
select handle, query_plan from #temp_plans
    cross apply sys.dm_exec_query_plan(handle)
`
	_, err = tx.ExecContext(ctx, createTempTable)
	if err != nil {
		return nil, fmt.Errorf("create id table: %w", err)
	}
	_, err = tx.ExecContext(ctx, insertIds, handles2...)
	if err != nil {
		return nil, fmt.Errorf("insert ids: %w", err)
	}
	rows, err2 := tx.QueryContext(ctx, query)
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
