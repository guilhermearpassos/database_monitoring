package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/jmoiron/sqlx"
	_ "github.com/microsoft/go-mssqldb"
	"strconv"
	"time"
)

type SQLServerDataReader struct {
	db                *sqlx.DB
	lastQueryCounters map[string]map[string]int64
}

var _ domain.SamplesReader = (*SQLServerDataReader)(nil)
var _ domain.QueryMetricsReader = (*SQLServerDataReader)(nil)

func NewSQLServerDataReader(db *sqlx.DB) SQLServerDataReader {
	return SQLServerDataReader{db: db, lastQueryCounters: make(map[string]map[string]int64)}
}

func (S SQLServerDataReader) TakeSnapshot(ctx context.Context) ([]*common_domain.DataBaseSnapshot, error) {
	qDBName := `select database_id, name from sys.databases`
	rowsDB, err := S.db.QueryContext(ctx, qDBName)
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
       text
FROM sys.dm_exec_sessions s
         inner join sys.dm_exec_requests  p on p.session_id = s.session_id
         CROSS APPLY sys.dm_exec_sql_text(sql_handle)
`
	rows, err := S.db.QueryxContext(ctx, query)
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
		var text string
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
			&text,
		)
		if err != nil {
			return nil, err
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
		qs := common_domain.QuerySample{
			Status:    pStatus,
			Cmd:       "",
			SqlHandle: sqlHandle,
			Text:      text,
			IsBlocked: blockingSessionId != 0,
			IsBlocker: false,
			Session: common_domain.SessionMetadata{
				SessionID:            strconv.Itoa(sessionID),
				LoginTime:            loginTime,
				HostName:             hostName,
				ProgramName:          programName,
				LoginName:            loginName,
				Status:               status,
				LastRequestStartTime: lastRequestStartTime,
				LastRequestEndTime:   lastRequestEndTime,
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
	snapshots := make([]*common_domain.DataBaseSnapshot, 0)
	querySamples := make([]*common_domain.QuerySample, 0)
	for _, qs2 := range querySamplesByDB {
		for _, qs := range qs2 {
			var sessionID int
			sessionID, err = strconv.Atoi(qs.Session.SessionID)
			if bl, ok := blockingMap[sessionID]; ok {
				qs.SetBlockedIds(bl)
			}
		}
		querySamples = append(querySamples, qs2...)
	}
	snapshots = append(snapshots, &common_domain.DataBaseSnapshot{
		Samples: querySamples,
		SnapInfo: common_domain.SnapInfo{
			ID:        snapID,
			Timestamp: snapTime,
			Server: common_domain.ServerMeta{
				Host: "localhost",
				Type: "sqlserver",
			},
		},
	})

	return snapshots, nil
}

func (S SQLServerDataReader) CollectMetrics(ctx context.Context) ([]*common_domain.QueryMetric, error) {
	ret := make([]*common_domain.QueryMetric, 0)
	query := `
with qstats as (select query_hash,
                       query_plan_hash,
                       plan_handle,
                       last_execution_time,
                       last_elapsed_time,

                       CONCAT(
                               CONVERT(binary(64), plan_handle),
                               CONVERT(binary(4), statement_start_offset),
                               CONVERT(binary(4), statement_end_offset))                                     as plan_handle_and_offsets,
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


                from sys.dm_exec_query_stats),
     qstats_aggr as (select query_hash,
                            query_plan_hash,
                            cast(dbid as int)                    as dbid,
                            d.name                               as db_name,
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
                     group by query_hash, query_plan_hash, s.dbid, d.name),
     qstats_aggr_split
         as (select convert(varbinary(64), substring(plan_handle_and_offsets, 1, 64))                  as plan_handle,
                    convert(int, convert(varbinary(4), substring(plan_handle_and_offsets, 64 + 1, 4))) as statement_start_offset,
                    convert(int, convert(varbinary(4), substring(plan_handle_and_offsets, 64 + 6, 4))) as statement_end_offset,
                    *
             from qstats_aggr
             where last_execution_time > dateadd(second, -60, getdate()))

select plan_handle,
       statement_start_offset,
       statement_end_offset,
       query_hash,
       query_plan_hash,
       qas.dbid,
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
       total_spills, 
       text
from qstats_aggr_split qas
         cross apply sys.dm_exec_sql_text(plan_handle)
`
	rows, err := S.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("collecting metrics: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)
	for rows.Next() {
		var planHandle []byte
		var statementStartOffset int
		var statementEndOffset int
		var queryHash []byte
		var queryPlanHash []byte
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
		var text string
		err = rows.Scan(&planHandle, &statementStartOffset, &statementEndOffset,
			&queryHash,
			&queryPlanHash,
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
			&text)
		if err != nil {
			return nil, fmt.Errorf("collecting metrics - scan: %w", err)
		}
		var counters map[string]int64
		lastCounters, ok := S.lastQueryCounters[string(queryHash)]
		if ok {
			counters = map[string]int64{
				"executionCount":               executionCount - lastCounters["executionCount"],
				"totalWorkerTime":              totalWorkerTime - lastCounters["totalWorkerTime"],
				"totalPhysicalReads":           totalPhysicalReads - lastCounters["totalPhysicalReads"],
				"totalLogicalWrites":           totalLogicalWrites - lastCounters["totalLogicalWrites"],
				"totalLogicalReads":            totalLogicalReads - lastCounters["totalLogicalReads"],
				"totalClrTime":                 totalClrTime - lastCounters["totalClrTime"],
				"totalElapsedTime":             totalElapsedTime - lastCounters["totalElapsedTime"],
				"totalRows":                    totalRows - lastCounters["totalRows"],
				"totalDop":                     totalDop - lastCounters["totalDop"],
				"totalGrantKb":                 totalGrantKb - lastCounters["totalGrantKb"],
				"totalUsedGrantKb":             totalUsedGrantKb - lastCounters["totalUsedGrantKb"],
				"totalIdealGrantKb":            totalIdealGrantKb - lastCounters["totalIdealGrantKb"],
				"totalReservedThreads":         totalReservedThreads - lastCounters["totalReservedThreads"],
				"totalUsedThreads":             totalUsedThreads - lastCounters["totalUsedThreads"],
				"totalColumnstoreSegmentReads": totalColumnstoreSegmentReads - lastCounters["totalColumnstoreSegmentReads"],
				"totalColumnstoreSegmentSkips": totalColumnstoreSegmentSkips - lastCounters["totalColumnstoreSegmentSkips"],
				"totalSpills":                  totalSpills - lastCounters["totalSpills"],
			}
		} else {

			counters = map[string]int64{
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
		}
		ret = append(ret, &common_domain.QueryMetric{
			QueryHash:         queryHash,
			Text:              text,
			Database:          common_domain.DataBaseMetadata{},
			LastExecutionTime: lastExecutionTime,
			LastElapsedTime:   time.Duration(lastElapsedTime) * time.Microsecond,
			Counters:          counters,
			Rates:             nil,
		})
		S.lastQueryCounters[string(queryHash)] = counters
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("collecting metrics - rows.Err: %w", err)
	}
	return ret, nil
}
