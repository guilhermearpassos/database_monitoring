package adapters

import (
	"context"
	"database_monitoring/internal/services/dbm/domain"
	"github.com/jmoiron/sqlx"
	_ "github.com/microsoft/go-mssqldb"
	"strconv"
	"time"
)

type SQLServerDataReader struct {
	db *sqlx.DB
}

func NewSQLServerDataReader(db *sqlx.DB) SQLServerDataReader {
	return SQLServerDataReader{db: db}
}

func (S SQLServerDataReader) TakeSnapshot(ctx context.Context) ([]*domain.DataBaseSnapshot, error) {
	query := `
SELECT s.session_id,
       s.login_time,
       s.host_name,
       s.program_name,
       s.login_name,
       s.status,
       s.cpu_time,
       s.memory_usage,
       s.total_elapsed_time,
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
	defer rows.Close()
	querySamplesByDB := make(map[string][]*domain.QuerySample, 0)
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
		qs := domain.QuerySample{
			Status:    pStatus,
			Cmd:       "",
			SqlHandle: sqlHandle,
			Text:      text,
			IsBlocked: blockingSessionId != 0,
			IsBlocker: false,
			Session: domain.SessionMetadata{
				SessionID:            strconv.Itoa(sessionID),
				LoginTime:            loginTime,
				HostName:             hostName,
				ProgramName:          programName,
				LoginName:            loginName,
				Status:               status,
				LastRequestStartTime: lastRequestStartTime,
				LastRequestEndTime:   lastRequestEndTime,
			},
			Database: domain.DataBaseMetadata{DatabaseID: strconv.Itoa(databaseId)},
			Block: domain.BlockMetadata{
				BlockedBy:       blockedBy,
				BlockedSessions: make([]string, 0),
			},
			Wait: domain.WaitMetadata{
				WaitType:     waitType,
				WaitTime:     waitTime,
				LastWaitType: lastWaitType,
				WaitResource: waitResource,
			},
		}
		if _, ok := querySamplesByDB[strconv.Itoa(databaseId)]; !ok {
			querySamplesByDB[strconv.Itoa(databaseId)] = make([]*domain.QuerySample, 0)
		}
		querySamplesByDB[strconv.Itoa(databaseId)] = append(querySamplesByDB[strconv.Itoa(databaseId)], &qs)

	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	snapshots := make([]*domain.DataBaseSnapshot, 0)
	for dbID, querySamples := range querySamplesByDB {
		for idx, qs := range querySamples {
			var sessionID int
			sessionID, err = strconv.Atoi(qs.Session.SessionID)
			if bl, ok := blockingMap[sessionID]; ok {
				qs.SetBlockedIds(bl)
			}
		}
		snapshots = append(snapshots, &domain.DataBaseSnapshot{
			Timestamp: time.Now().In(time.UTC),
			Samples:   querySamples,
			Server: domain.ServerMeta{
				Host: "localhost",
				Type: "sqlserver",
			},
			Database: domain.DataBaseMetadata{DatabaseID: dbID},
		})
	}

	return snapshots, nil
}

var _ domain.DataBaseReader = (*SQLServerDataReader)(nil)
