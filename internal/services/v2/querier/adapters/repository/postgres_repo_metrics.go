package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/querier/domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
	"google.golang.org/protobuf/proto"
)

var _ domain.QueryMetricsRepository = (*PostgresRepo)(nil)

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
			base.Counters = rates
		}
		retList = append(retList, base)
	}
	return retList, nil
}

func (p *PostgresRepo) GetQueryMetricsSlice(ctx context.Context, start time.Time, end time.Time, serverID string, sampleID string) ([]*common_domain.QueryMetric, error) {
	q := `select data, q.collected_at from public.query_stat_snapshot q
inner join query_stat_sample qss on q.id = qss.snap_id
         inner join target t on q.target_id = t.id
where q.collected_at between $1 and $2 and t.host = $3
and qss.sql_handle = $4
`
	queryHash := sampleID
	rows, err := p.db.QueryContext(ctx, q, start, end, serverID, queryHash)
	if err != nil {
		return nil, fmt.Errorf("getting query stats: %w", err)
	}
	defer rows.Close()
	queryMetric := make([]*common_domain.QueryMetric, 0)
	for rows.Next() {

		var protoBytes []byte
		var collectedAt time.Time
		err = rows.Scan(&protoBytes, &collectedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning query stats: %w", err)
		}
		protoMetric := dbmv1.QueryMetric{}
		err = proto.Unmarshal(protoBytes, &protoMetric)
		if err != nil {
			return nil, fmt.Errorf("unmarshal query stat: %w", err)
		}
		qMetric, err2 := converters.QueryMetricToDomain(&protoMetric)
		if err2 != nil {
			return nil, fmt.Errorf("converting query stat: %w", err)
		}
		if qMetric.CollectionTime.IsZero() || qMetric.CollectionTime.Year() == 1970 {
			qMetric.CollectionTime = collectedAt
		}
		queryMetric = append(queryMetric, qMetric)
	}
	if err2 := rows.Err(); err2 != nil {
		return nil, fmt.Errorf("scanning query stats err: %w", err2)
	}
	return queryMetric, nil
}
