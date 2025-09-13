package adapters

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/guilhermearpassos/database-monitoring/internal/common/custom_errors"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"google.golang.org/protobuf/proto"
	"slices"
	"time"
)

func (p *PostgresRepo) ListServers(ctx context.Context, start time.Time, end time.Time) ([]domain.ServerSummary, error) {
	q := `select distinct t.host, t.type_id  from snapshot s 
    inner join public.target t on t.id = s.target_id where snap_time between $1 and $2`
	rows, err := p.db.QueryContext(ctx, q, start, end)
	if err != nil {
		return nil, fmt.Errorf("listing servers: %w", err)
	}
	defer rows.Close()
	servers := make([]domain.ServerSummary, 0)
	for rows.Next() {
		var name string
		var typeID int
		err = rows.Scan(&name, &typeID)
		if err != nil {
			return nil, fmt.Errorf("listing servers scan: %w", err)
		}
		servers = append(servers, domain.ServerSummary{
			Name:             name,
			Type:             "mssql",
			Connections:      0,
			RequestRate:      0,
			ConnsByWaitGroup: make(map[string]int32),
		})
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("listing servers rows: %w", err)
	}
	return servers, nil
}

func (p *PostgresRepo) ListSnapshots(ctx context.Context, databaseID string, start time.Time, end time.Time, pageNumber int, pageSize int, serverID string) ([]common_domain.DataBaseSnapshot, int, error) {
	//language=SQL
	q := fmt.Sprintf(`
with snapinfos as (
	select s.id, s.f_id, s.snap_time, t.host, t.type_id, count(*) OVER() AS full_count from snapshot s
	inner join public.target t on t.id = s.target_id
	where s.snap_time between $1 and $2 and t.host = $3
	order by s.snap_time desc
	offset %d rows limit %d
)

select si.f_id, si.snap_time, si.host, si.type_id, qs.f_id as qfid, qs.data, full_count from snapinfos si
inner join query_samples qs on qs.snap_id = si.id


`, pageSize*(pageNumber-1), pageSize)
	rows, err := p.db.QueryContext(ctx, q, start, end, serverID)
	if err != nil {
		return nil, 0, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	fullCount, snapshots, err2 := parseSnapshotRows(rows)
	if err2 != nil {
		return snapshots, fullCount, fmt.Errorf("parsing snapshots: %w", err2)
	}
	slices.SortFunc(snapshots, func(a, b common_domain.DataBaseSnapshot) int {
		if a.SnapInfo.Timestamp.Before(b.SnapInfo.Timestamp) {
			return 1
		}
		return -1
	})
	return snapshots, fullCount, nil

}

func parseSnapshotRows(rows *sql.Rows) (int, []common_domain.DataBaseSnapshot, error) {
	var err error
	queriesBySnapId := make(map[string][]*common_domain.QuerySample)
	snapInfos := make(map[string]common_domain.SnapInfo)
	var fullCount int
	for rows.Next() {
		var sId string
		var qId string
		var snapTime time.Time
		var host string
		var typeID int
		var queryData []byte
		err = rows.Scan(&sId, &snapTime, &host, &typeID, &qId, &queryData, &fullCount)
		if err != nil {
			return 0, nil, fmt.Errorf("listing snapshots: %w", err)
		}
		snapInfos[sId] = common_domain.SnapInfo{
			ID:        sId,
			Timestamp: snapTime,
			Server: common_domain.ServerMeta{
				Host: host,
				Type: "mssql",
			},
		}
		proto := dbmv1.QuerySample{}
		err = proto.UnmarshalVT(queryData)
		if err != nil {
			return 0, nil, fmt.Errorf("listing snapshots unmarshal proto: %w", err)
		}
		proto.Id = qId
		toDomain := converters.SampleToDomain(&proto)
		_, ok := queriesBySnapId[sId]
		if !ok {
			queriesBySnapId[sId] = []*common_domain.QuerySample{toDomain}
		} else {
			queriesBySnapId[sId] = append(queriesBySnapId[sId], toDomain)
		}

	}
	err = rows.Err()
	if err != nil {
		return 0, nil, fmt.Errorf("listing snapshots rows: %w", err)
	}
	ret := make([]common_domain.DataBaseSnapshot, 0, len(snapInfos))
	for k, v := range snapInfos {
		samples := queriesBySnapId[k]
		ret = append(ret, common_domain.DataBaseSnapshot{
			SnapInfo: v,
			Samples:  samples,
		})
	}
	return fullCount, ret, nil
}

func (p *PostgresRepo) GetSnapshot(ctx context.Context, id string) (common_domain.DataBaseSnapshot, error) {
	q := `select s.f_id, s.snap_time, t.host, t.type_id, qs.f_id as sid, qs.data, count(*) OVER() AS full_count from snapshot s
inner join public.target t on t.id = s.target_id
inner join public.query_samples qs on s.id = qs.snap_id
where s.f_id = $1`
	rows, err := p.db.QueryContext(ctx, q, id)
	if err != nil {
		return common_domain.DataBaseSnapshot{}, fmt.Errorf("getting snapshot %s: %w", id, err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)
	_, snapshots, err2 := parseSnapshotRows(rows)

	if err2 != nil {
		return common_domain.DataBaseSnapshot{}, fmt.Errorf("getting snapshot %s: %w", id, err2)
	}
	if len(snapshots) == 0 {
		return common_domain.DataBaseSnapshot{}, custom_errors.NotFoundErr{Message: fmt.Sprintf("snapshot %s not found", id)}
	}
	return snapshots[0], nil
}

func (p *PostgresRepo) GetExecutionPlan(ctx context.Context, planHandle string, server *common_domain.ServerMeta) (*common_domain.ExecutionPlan, error) {
	targetId, err := p.getTargetID(ctx, p.db, *server)
	if err != nil {
		return nil, fmt.Errorf("getting target id: %w", err)
	}
	q := "select plan_xml from query_plans where plan_handle = $1 and target_id = $2"
	result := p.db.QueryRowContext(ctx, q, planHandle, targetId)
	err = result.Err()
	if err != nil {
		return nil, fmt.Errorf("getting query plan: %w", err)
	}
	var planXML string
	err = result.Scan(&planXML)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, custom_errors.NotFoundErr{Message: "plan not found"}
		}
		return nil, fmt.Errorf("scanning query plan: %w", err)
	}
	return &common_domain.ExecutionPlan{
		PlanHandle: planHandle,
		Server:     *server,
		XmlData:    planXML,
	}, nil

}

func (p *PostgresRepo) ListQueryMetrics(ctx context.Context, start time.Time, end time.Time, serverID string) ([]*common_domain.QueryMetric, error) {
	q := `select qss.sql_handle, data from query_stat_sample qss
inner join public.query_stat_snapshot q on q.id = qss.snap_id
         inner join target t on q.target_id = t.id
where q.collected_at between $1 and $2 and t.host = $3
order by q.collected_at desc
`
	rows, err := p.db.QueryContext(ctx, q, start, end, serverID)
	if err != nil {
		return nil, fmt.Errorf("getting query stats: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)
	ret := make(map[string][]*common_domain.QueryMetric, 0)
	for rows.Next() {
		var sqlHandle string
		var protoBytes []byte
		err = rows.Scan(&sqlHandle, &protoBytes)
		if err != nil {
			return nil, fmt.Errorf("scanning query stats: %w", err)
		}
		protoMetric := dbmv1.QueryMetric{}
		err = proto.Unmarshal(protoBytes, &protoMetric)
		if err != nil {
			return nil, fmt.Errorf("unmarshal query stat: %w", err)
		}
		queryMetric, err2 := converters.QueryMetricToDomain(&protoMetric)
		if err2 != nil {
			return nil, fmt.Errorf("converting query stat: %w", err)
		}
		if _, ok := ret[sqlHandle]; !ok {
			ret[sqlHandle] = make([]*common_domain.QueryMetric, 0)
		}
		ret[sqlHandle] = append(ret[sqlHandle], queryMetric)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("getting query stats rowserr: %w", err)
	}
	retList := make([]*common_domain.QueryMetric, 0)
	for _, v := range ret {
		base := v[0]
		for i, m := range v {
			if i == 0 {
				continue
			}
			rates := base.Counters
			for k1, v1 := range m.Counters {
				rates[k1] += v1
			}
		}
		retList = append(retList, base)
	}
	return retList, nil
}

func (p *PostgresRepo) GetQuerySample(ctx context.Context, snapID string, sampleID string) (*common_domain.QuerySample, error) {

	q := `select qs.data from query_samples qs
         inner join public.snapshot s on s.id = qs.snap_id
         where s.f_id = $1 and qs.f_id = $2`
	row := p.db.QueryRowContext(ctx, q, snapID, sampleID)
	err := row.Err()
	if err != nil {
		return nil, fmt.Errorf("getting query sample %s: %w", snapID, err)
	}
	var protoBytes []byte
	err = row.Scan(&protoBytes)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, custom_errors.NotFoundErr{Message: fmt.Sprintf("query sample %s not found", snapID)}
		}
		return nil, fmt.Errorf("scanning query sample %s: %w", snapID, err)
	}
	protoSample := dbmv1.QuerySample{}
	err = proto.Unmarshal(protoBytes, &protoSample)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling query sample %s: %w", snapID, err)
	}
	domainSample := converters.SampleToDomain(&protoSample)
	return domainSample, nil
}

func (p *PostgresRepo) ListSnapshotSummaries(ctx context.Context, serverID string, start time.Time, end time.Time) ([]common_domain.SnapshotSummary, error) {
	q := `select s.snap_time, s.f_id, t.host, t.type_id, qs.wait_event, count(qs.id), sum(qs.wait_time) from snapshot s
inner join public.query_samples qs on s.id = qs.snap_id
         inner join target t on s.target_id = t.id
where t.host = $1 and snap_time between $2 and $3
group by s.snap_time, s.f_id, t.host, t.type_id, qs.wait_event`
	rows, err := p.db.QueryContext(ctx, q, serverID, start, end)
	if err != nil {
		return nil, fmt.Errorf("listing snapshot summaries: %w", err)
	}
	defer rows.Close()
	ret := make([]common_domain.SnapshotSummary, 0)
	detailsMapByID := make(map[string]struct {
		id        string
		timestamp time.Time
		server    common_domain.ServerMeta
	})
	connsMapByID := make(map[string]map[string]int64)
	timeMsMapByID := make(map[string]map[string]int64)
	for rows.Next() {
		var snapTime time.Time
		var snapID string
		var host string
		var typeID int
		var waitEvent string
		var count int64
		var waitTime int64
		err = rows.Scan(&snapTime, &snapID, &host, &typeID, &waitEvent, &count, &waitTime)
		if err != nil {
			return nil, fmt.Errorf("listing snapshot summaries scan: %w", err)
		}
		if _, ok := detailsMapByID[snapID]; !ok {
			detailsMapByID[snapID] = struct {
				id        string
				timestamp time.Time
				server    common_domain.ServerMeta
			}{
				id:        snapID,
				timestamp: snapTime,
				server: common_domain.ServerMeta{
					Host: host,
					Type: "mssql",
				},
			}
		}
		if _, ok := connsMapByID[snapID]; !ok {
			connsMapByID[snapID] = make(map[string]int64)
		}
		if _, ok := timeMsMapByID[snapID]; !ok {
			timeMsMapByID[snapID] = make(map[string]int64)
		}
		connsMapByID[snapID][waitEvent] = count
		timeMsMapByID[snapID][waitEvent] = waitTime
	}
	for k, v := range detailsMapByID {
		connMap := connsMapByID[k]
		TimeMap := timeMsMapByID[k]
		ret = append(ret, common_domain.SnapshotSummary{
			ID:               v.id,
			Timestamp:        v.timestamp,
			Server:           v.server,
			ConnsByWaitType:  connMap,
			TimeMsByWaitType: TimeMap,
		})
	}
	return ret, nil
}
func (p *PostgresRepo) GetQueryMetrics(ctx context.Context, start time.Time, end time.Time, serverID string, sampleID string) (*common_domain.QueryMetric, error) {
	q := `select data from query_stat_sample qss
inner join public.query_stat_snapshot q on q.id = qss.snap_id
         inner join target t on q.target_id = t.id
where q.collected_at between $1 and $2 and t.host = $3
and qss.sql_handle = $4
order by q.collected_at desc
`
	queryHash := sampleID
	row := p.db.QueryRowContext(ctx, q, start, end, serverID, queryHash)
	err := row.Err()
	if err != nil {
		return nil, fmt.Errorf("getting query stats: %w", err)
	}
	var protoBytes []byte
	err = row.Scan(&protoBytes)
	if err != nil {
		return nil, fmt.Errorf("scanning query stats: %w", err)
	}
	protoMetric := dbmv1.QueryMetric{}
	err = proto.Unmarshal(protoBytes, &protoMetric)
	if err != nil {
		return nil, fmt.Errorf("unmarshal query stat: %w", err)
	}
	queryMetric, err2 := converters.QueryMetricToDomain(&protoMetric)
	if err2 != nil {
		return nil, fmt.Errorf("converting query stat: %w", err)
	}
	return queryMetric, nil
}
