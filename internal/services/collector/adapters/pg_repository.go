package adapters

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/guilhermearpassos/database-monitoring/internal/common/custom_errors"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"math"
	"slices"
	"time"
)

type PostgresRepo struct {
	db     *sqlx.DB
	tracer trace.Tracer
}

func NewPostgresRepo(db *sqlx.DB) *PostgresRepo {
	return &PostgresRepo{db: db, tracer: otel.Tracer("postgres-repo")}
}

var _ domain.SampleRepository = (*PostgresRepo)(nil)
var _ domain.QueryMetricsRepository = (*PostgresRepo)(nil)

func (p *PostgresRepo) StoreSnapshotSamples(ctx context.Context, snapID string, samples []*common_domain.QuerySample) error {
	ctx, span := p.tracer.Start(ctx, "StoreSnapshotSamples")
	defer span.End()
	span.SetAttributes(attribute.String("snapshot_id", snapID),
		attribute.Int64("num_samples", int64(len(samples))))
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				err = errors.Join(err, err2)
			}
			return
		}
		err2 := tx.Commit()
		if err2 != nil {
			err = errors.Join(err, err2)
			return
		}
	}()
	querySnapPK := `select id from snapshot where f_id = $1`
	var snapPK int
	row := tx.QueryRowContext(ctx, querySnapPK, snapID)
	err = row.Err()
	if err != nil {
		return fmt.Errorf("query snapshot pk: %w", err)
	}
	err = row.Scan(&snapPK)
	if err != nil {
		return fmt.Errorf("query snapshot pk scan: %w", err)
	}
	err = p.bulkInsertSamples(ctx, tx, samples, snapPK)
	if err != nil {
		return fmt.Errorf("insert samples: %w", err)
	}
	return nil
}

func (p *PostgresRepo) StoreSnapshot(ctx context.Context, snapshot common_domain.DataBaseSnapshot) (err error) {
	ctx, span := p.tracer.Start(ctx, "StoreSnapshot")
	defer span.End()
	span.SetAttributes(attribute.String("snapshot_id", snapshot.SnapInfo.ID),
		attribute.Int("num_samples", len(snapshot.Samples)))
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				err = errors.Join(err, err2)
			}
			return
		}
		err2 := tx.Commit()
		if err2 != nil {
			err = errors.Join(err, err2)
			return
		}
	}()
	var snapId int
	snapId, err = p.insertSnapshot(ctx, tx, &snapshot.SnapInfo)
	if err != nil {
		return fmt.Errorf("insert snapshot: %w", err)
	}
	err = p.bulkInsertSamples(ctx, tx, snapshot.Samples, snapId)
	if err != nil {
		return fmt.Errorf("insert samples: %w", err)
	}
	return nil

}

// insertSnapshot inserts a single snapshot and returns the generated ID
func (p *PostgresRepo) insertSnapshot(ctx context.Context, tx *sqlx.Tx, snapshot *common_domain.SnapInfo) (int, error) {
	ctx, span := p.tracer.Start(ctx, "insertSnapshot")
	defer span.End()
	query := `
		INSERT INTO snapshot (f_id, snap_time, target_id)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	targetId, err := p.getOrCreateTargetID(ctx, tx, snapshot.Server)
	if err != nil {
		return 0, fmt.Errorf("get target id: %w", err)
	}
	var id int
	err = tx.QueryRowxContext(ctx, query, snapshot.ID, snapshot.Timestamp.In(time.UTC), targetId).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// bulkInsertSamples performs bulk insert of samples using PostgreSQL COPY
func (p *PostgresRepo) bulkInsertSamples(ctx context.Context, tx *sqlx.Tx, samples []*common_domain.QuerySample, snapId int) error {
	ctx, span := p.tracer.Start(ctx, "bulkInsertSamples")
	defer span.End()
	// Prepare the COPY statement
	stmt, err := tx.PrepareContext(ctx, pq.CopyIn("query_samples", "f_id", "snap_id", "sql_handle",
		"blocked", "blocker", "plan_handle", "data", "wait_event", "wait_time",
		"sid", "connection_id", "transaction_id", "block_ms", "block_count", "query_hash",
	))
	if err != nil {
		return fmt.Errorf("failed to prepare COPY statement: %w", err)
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)

	// Execute COPY for each sample
	for _, sample := range samples {
		var protoBytes []byte
		protoBytes, err = converters.SampleToProto(sample).MarshalVT()
		if err != nil {
			return fmt.Errorf("error serializing sample: %v", sample)
		}
		var waitType string
		if sample.Wait.WaitType != nil {
			waitType = *sample.Wait.WaitType
		}
		_, err = stmt.ExecContext(ctx, sample.Id, snapId, sample.SqlHandle,
			sample.IsBlocked, sample.IsBlocker, sample.PlanHandle, protoBytes, waitType,
			sample.Wait.WaitTime, sample.Session.SessionID, sample.Session.ConnectionId,
			sample.CommandMetadata.TransactionId, -1, len(sample.Block.BlockedSessions),
			sample.QueryHash)
		if err != nil {
			return fmt.Errorf("failed to execute COPY for sample: %w", err)
		}
	}

	// Execute the final COPY command
	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to complete COPY operation: %w", err)
	}

	return nil
}

func (p *PostgresRepo) StoreExecutionPlans(ctx context.Context, snapshot []*common_domain.ExecutionPlan) error {
	ctx, span := p.tracer.Start(ctx, "StoreExecutionPlans")
	defer span.End()
	span.SetAttributes(attribute.Int("num_samples", len(snapshot)))
	q := `insert into query_plans (plan_handle, plan_xml, target_id) VALUES `
	if len(snapshot) == 0 {
		return nil
	}

	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("transaction begin: %w", err)
	}
	defer func() {
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				err = errors.Join(err, err2)
			}
			return
		}
		err2 := tx.Commit()
		if err2 != nil {
			err = errors.Join(err, err2)
			return
		}
	}()
	var targetID int
	targetID, err = p.getOrCreateTargetID(ctx, tx, snapshot[0].Server)
	if err != nil {
		return fmt.Errorf("get target id: %w", err)
	}
	chunks := slices.Chunk(snapshot, 300)
	n := 1
	for chunk := range chunks {
		currentQuery := q
		args := make([]interface{}, 0, len(chunk)*3)
		for _, data := range chunk {
			currentQuery = currentQuery + fmt.Sprintf(" ($%d, $%d, $%d),", n, n+1, n+2)

			encodedHandle := data.PlanHandle
			args = append(args, encodedHandle, data.XmlData, targetID)
			n += 3
		}
		currentQuery = currentQuery[:len(currentQuery)-1]
		res, err := tx.ExecContext(ctx, currentQuery, args...)
		if err != nil {
			return fmt.Errorf("exec query: %w", err)
		}
		_, _ = res.LastInsertId()
		_, _ = res.RowsAffected()

	}
	return nil
}

func (p *PostgresRepo) GetKnownPlanHandles(ctx context.Context, server *common_domain.ServerMeta, pageNumber int, pageSize int) ([]string, int, error) {
	ctx, span := p.tracer.Start(ctx, "GetKnownPlanHandles")
	defer span.End()
	//language=SQL
	query := `
select plan_handle, COUNT(*) OVER () AS total_count from query_plans where target_id = $1
                          order by id
offset $2 rows
limit $3

`
	tId, err := p.getTargetID(ctx, p.db, *server)
	if err != nil {
		if errors.As(err, &custom_errors.NotFoundErr{}) {
			return []string{}, 0, nil
		}
		return nil, 0, fmt.Errorf("get target id: %w", err)
	}

	rows, err := p.db.QueryContext(ctx, query, tId, (pageNumber-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("GetKnownPlanHandles: %w", err)
	}
	ret := make([]string, 0)
	var totalCount int
	for rows.Next() {
		var planHandle string
		err = rows.Scan(&planHandle, &totalCount)
		if err != nil {
			return nil, 0, fmt.Errorf("GetKnownPlanHandles scan: %w", err)
		}

		ret = append(ret, planHandle)
	}
	err = rows.Err()
	if err != nil {
		return nil, 0, fmt.Errorf("GetKnownPlanHandles rowserr: %w", err)
	}
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	return ret, totalPages, nil
}

func (p *PostgresRepo) getOrCreateTargetID(ctx context.Context, tx *sqlx.Tx, server common_domain.ServerMeta) (int, error) {
	ctx, span := p.tracer.Start(ctx, "getOrCreateTargetID")
	defer span.End()
	tId, err := p.getTargetID(ctx, tx, server)
	if err != nil {
		if errors.As(err, &custom_errors.NotFoundErr{}) {
			tId, err = p.createTargetID(ctx, tx, server)
			if err != nil {
				return 0, fmt.Errorf("create target id: %w", err)
			}
			return tId, nil

		}
		return 0, fmt.Errorf("get target id: %w", err)
	}
	return tId, nil
}

func (p *PostgresRepo) getTargetID(ctx context.Context, tx sqlx.QueryerContext, server common_domain.ServerMeta) (int, error) {
	ctx, span := p.tracer.Start(ctx, "getTargetID")
	defer span.End()
	q := "select id from target where host = $1 and type_id = 1;"
	rows := tx.QueryRowxContext(ctx, q, server.Host)
	err := rows.Err()
	if err != nil {
		return 0, fmt.Errorf("getTargetID: %w", err)
	}
	var targetID int
	err = rows.Scan(&targetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, custom_errors.NotFoundErr{Message: err.Error()}
		}
		return 0, fmt.Errorf("getTargetID scan: %w", err)
	}
	return targetID, nil

}

func (p *PostgresRepo) createTargetID(ctx context.Context, tx *sqlx.Tx, server common_domain.ServerMeta) (int, error) {
	ctx, span := p.tracer.Start(ctx, "createTargetID")
	defer span.End()
	q := "insert into target (host, type_id, agent_version) values ($1, $2, $3);"
	_, err := tx.ExecContext(ctx, q, server.Host, 1, "-")
	if err != nil {
		return 0, fmt.Errorf("createTargetID: %w", err)
	}
	return p.getTargetID(ctx, tx, server)

}

func (p *PostgresRepo) StoreQueryMetrics(ctx context.Context, metrics []*common_domain.QueryMetric, serverMeta common_domain.ServerMeta, timestamp time.Time) error {
	ctx, span := p.tracer.Start(ctx, "StoreQueryMetrics")
	defer span.End()
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				err = errors.Join(err, err2)
			}
			return
		}
		err2 := tx.Commit()
		if err2 != nil {
			err = errors.Join(err, err2)
			return
		}
	}()
	var snapId int
	snapId, err = p.insertQueryStatSnapshot(ctx, tx, serverMeta, timestamp)
	if err != nil {
		return fmt.Errorf("insert snapshot: %w", err)
	}
	err = p.bulkInsertQueryStatSamples(ctx, tx, metrics, snapId)
	if err != nil {
		return fmt.Errorf("bulk insert stat samples: %w", err)
	}
	return nil
}

// insertQueryStatSnapshot inserts a single query stat snapshot and returns the generated ID
func (p *PostgresRepo) insertQueryStatSnapshot(ctx context.Context, tx *sqlx.Tx, meta common_domain.ServerMeta, collectedAt time.Time) (int, error) {
	ctx, span := p.tracer.Start(ctx, "insertQueryStatSnapshot")
	defer span.End()
	query := `
		INSERT INTO query_stat_snapshot (f_id, target_id, collected_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int
	newUuid, err := uuid.NewUUID()
	if err != nil {
		return 0, fmt.Errorf("insertQueryStatSnapshot uuid: %w", err)
	}

	targetId, err := p.getOrCreateTargetID(ctx, tx, meta)
	if err != nil {
		return 0, fmt.Errorf("get target id: %w", err)
	}
	err = tx.QueryRowxContext(ctx, query, newUuid, targetId, collectedAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insertQueryStatSnapshot scan: %w", err)
	}

	return id, nil
}

// bulkInsertQueryStatSamples performs bulk insert of query stat samples using PostgreSQL COPY
func (p *PostgresRepo) bulkInsertQueryStatSamples(ctx context.Context, tx *sqlx.Tx, samples []*common_domain.QueryMetric, snapId int) error {
	ctx, span := p.tracer.Start(ctx, "bulkInsertQueryStatSamples")
	defer span.End()
	// Prepare the COPY statement
	stmt, err := tx.PrepareContext(ctx, pq.CopyIn("query_stat_sample", "snap_id", "sql_handle", "data"))
	if err != nil {
		return fmt.Errorf("failed to prepare COPY statement: %w", err)
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)

	// Execute COPY for each sample
	for _, sample := range samples {

		proto, err2 := converters.QueryMetricToProto(sample)
		if err2 != nil {
			return fmt.Errorf("convert to proto: %w", err2)
		}
		var protoBytes []byte
		protoBytes, err = proto.MarshalVT()
		if err != nil {
			return fmt.Errorf("marshal proto: %w", err)
		}

		_, err = stmt.ExecContext(ctx, snapId, sample.QueryHash, protoBytes)
		if err != nil {
			return fmt.Errorf("failed to execute COPY for sample: %w", err)
		}
	}

	// Execute the final COPY command
	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to complete COPY operation: %w", err)
	}

	return nil
}

func (p *PostgresRepo) PurgeQueryMetrics(ctx context.Context, start time.Time, end time.Time, batchSize int) error {
	ctx, span := p.tracer.Start(ctx, "PurgeQueryMetrics")
	defer span.End()
	// language=SQL
	query := `
with rows_to_delete as (
    select qss.CTID from query_stat_sample qss
inner join query_stat_snapshot qsnap on qss.snap_id = qsnap.id
where collected_at between  $1 and $2
limit $3
)
delete from query_stat_sample using rows_to_delete where query_stat_sample.CTID = rows_to_delete.CTID`
	rowsAffected := int64(1)
	for rowsAffected > 0 {
		r, err := p.db.ExecContext(ctx, query, start, end, batchSize)
		if err != nil {
			return fmt.Errorf("purgeQueryMetrics: %w", err)
		}
		rowsAffected, _ = r.RowsAffected()
	}
	// language=SQL
	querySnap := `
with rows_to_delete as (
    select CTID from query_stat_snapshot 
where collected_at between  $1 and $2
limit $3
)
delete from query_stat_snapshot using rows_to_delete where query_stat_snapshot.CTID = rows_to_delete.CTID`
	rowsAffected = int64(1)
	for rowsAffected > 0 {
		r, err := p.db.ExecContext(ctx, querySnap, start, end, batchSize)
		if err != nil {
			return fmt.Errorf("purgeQueryMetrics: %w", err)
		}
		rowsAffected, _ = r.RowsAffected()
	}
	return nil
}
func (p *PostgresRepo) PurgeAllQueryMetrics(ctx context.Context) error {
	ctx, span := p.tracer.Start(ctx, "PurgeAllQueryMetrics")
	defer span.End()
	// language=SQL
	query := `
truncate table query_stat_snapshot cascade`
	r, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("purgeQueryMetrics: %w", err)
	}
	rowsAffected, _ := r.RowsAffected()
	span.SetAttributes(attribute.Int64("rows_affected", rowsAffected))
	return nil
}

func (p *PostgresRepo) PurgeSnapshots(ctx context.Context, start time.Time, end time.Time, batchSize int) error {
	ctx, span := p.tracer.Start(ctx, "PurgeSnapshots")
	defer span.End()
	q := `with rows_to_delete as (
    select CTID from snapshot
where snap_time between  $1 and $2
limit $3
)
delete from snapshot using rows_to_delete where snapshot.CTID = rows_to_delete.CTID`
	rowsAffected := int64(1)
	for rowsAffected > 0 {
		r, err := p.db.ExecContext(ctx, q, start, end, batchSize)
		if err != nil {
			return fmt.Errorf("purgeSnapshots: %w", err)
		}
		rowsAffected, _ = r.RowsAffected()
	}
	return nil
}

func (p *PostgresRepo) PurgeAllSnapshots(ctx context.Context) error {
	ctx, span := p.tracer.Start(ctx, "PurgeAllSnapshots")
	defer span.End()
	// language=SQL
	query := `
truncate table snapshot cascade`
	r, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("purgeSnapshots: %w", err)
	}
	rowsAffected, _ := r.RowsAffected()
	span.SetAttributes(attribute.Int64("rows_affected", rowsAffected))
	return nil
}
func (p *PostgresRepo) PurgeQueryPlans(ctx context.Context, batchSize int) error {
	ctx, span := p.tracer.Start(ctx, "PurgeQueryPlans")
	defer span.End()
	query := `with rows_to_delete as (
    select qp.CTID from query_plans qp
                left join query_samples qs on qs.plan_handle = qp.plan_handle
                left join snapshot snap on qs.snap_id = snap.id
                where qs.f_id is null
limit $1
)
delete from query_plans using rows_to_delete where query_plans.CTID = rows_to_delete.CTID`
	rowsAffected := int64(1)
	for rowsAffected > 0 {
		r, err := p.db.ExecContext(ctx, query, batchSize)
		if err != nil {
			return fmt.Errorf("PurgeQueryPlans: %w", err)
		}
		rowsAffected, _ = r.RowsAffected()
	}
	return nil
}

func (p *PostgresRepo) PurgeAllQueryPlans(ctx context.Context) error {
	ctx, span := p.tracer.Start(ctx, "PurgeAllQueryPlans")
	defer span.End()
	query := `truncate table query_plans `
	rowsAffected := int64(1)
	for rowsAffected > 0 {
		r, err := p.db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("PurgeAllQueryPlans: %w", err)
		}
		rowsAffected, _ = r.RowsAffected()
	}
	return nil
}
