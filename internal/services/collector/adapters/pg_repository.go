package adapters

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/guilhermearpassos/database-monitoring/internal/common/custom_errors"
	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain/converters"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"slices"
	"time"
)

type PostgresRepo struct {
	db *sqlx.DB
}

func NewPostgresRepo(db *sqlx.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

var _ domain.SampleRepository = (*PostgresRepo)(nil)
var _ domain.QueryMetricsRepository = (*PostgresRepo)(nil)

func (p *PostgresRepo) StoreSnapshot(ctx context.Context, snapshot common_domain.DataBaseSnapshot) (err error) {
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
	// Prepare the COPY statement
	stmt, err := tx.PrepareContext(ctx, pq.CopyIn("query_samples", "f_id", "snap_id", "sql_handle",
		"blocked", "blocker", "plan_handle", "data", "wait_event", "wait_time",
		"sid", "connection_id", "transaction_id", "block_ms", "block_count",
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
		_, err = stmt.ExecContext(ctx, sample.Id, snapId, base64.StdEncoding.EncodeToString(sample.SqlHandle),
			sample.IsBlocked, sample.IsBlocker, base64.StdEncoding.EncodeToString(sample.PlanHandle), protoBytes, waitType,
			sample.Wait.WaitTime, sample.Session.SessionID, sample.Session.ConnectionId, sample.CommandMetadata.TransactionId, -1, len(sample.Block.BlockedSessions))
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

			encodedHandle := base64.StdEncoding.EncodeToString(data.PlanHandle)
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

func (p *PostgresRepo) GetKnownPlanHandles(ctx context.Context, server *common_domain.ServerMeta) ([][]byte, error) {
	//language=SQL
	query := `

select plan_handle from query_plans where target_id = $1;

`
	tId, err := p.getTargetID(ctx, p.db, *server)
	if err != nil {
		if errors.As(err, &custom_errors.NotFoundErr{}) {
			return [][]byte{}, nil
		}
		return nil, fmt.Errorf("get target id: %w", err)
	}

	rows, err := p.db.QueryContext(ctx, query, tId)
	if err != nil {
		return nil, fmt.Errorf("GetKnownPlanHandles: %w", err)
	}
	ret := make([][]byte, 0)
	for rows.Next() {
		var planHandle string
		err = rows.Scan(&planHandle)
		if err != nil {
			return nil, fmt.Errorf("GetKnownPlanHandles scan: %w", err)
		}
		var bytePlanHandle []byte
		bytePlanHandle, err = base64.StdEncoding.DecodeString(planHandle)
		if err != nil {
			return nil, fmt.Errorf("GetKnownPlanHandles base64 decode: %w", err)
		}

		ret = append(ret, bytePlanHandle)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("GetKnownPlanHandles rowserr: %w", err)
	}
	return ret, nil
}

func (p *PostgresRepo) getOrCreateTargetID(ctx context.Context, tx *sqlx.Tx, server common_domain.ServerMeta) (int, error) {
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
	q := "insert into target (host, type_id, agent_version) values ($1, $2, $3);"
	_, err := tx.ExecContext(ctx, q, server.Host, 1, "-")
	if err != nil {
		return 0, fmt.Errorf("createTargetID: %w", err)
	}
	return p.getTargetID(ctx, tx, server)

}

func (p *PostgresRepo) StoreQueryMetrics(ctx context.Context, metrics []*common_domain.QueryMetric, serverMeta common_domain.ServerMeta, timestamp time.Time) error {

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

		_, err = stmt.ExecContext(ctx, snapId, base64.StdEncoding.EncodeToString(sample.QueryHash), protoBytes)
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
