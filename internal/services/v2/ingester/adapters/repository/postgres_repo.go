package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
)

// PostgresRepo persists v2 ingester data into the existing v1 tables.
// It reuses tables: target, snapshot, query_samples.
// Notes:
//  - target is keyed by host and a fixed type_id (1) for now
//  - snapshot is keyed by external id in column f_id
//  - query_samples stores various denormalized fields plus the original proto bytes in data
//    to keep full fidelity for downstream consumers
//
// Idempotency: SessionStore deduplicates per (snapshot_id, seq). We still structure operations
// to be safe for retries (e.g., EnsureSnapshot upserts by f_id).
//
// Tracing is intentionally omitted here to keep the first cut simple; can be added later.
//
// This adapter purposely depends only on v2 domain + v1 proto types.
// It does not reuse the v1 collector repository code to keep coupling minimal.

type PostgresRepo struct {
	db     *sqlx.DB
	tracer trace.Tracer
}

func NewPostgresRepo(db *sqlx.DB) *PostgresRepo {
	return &PostgresRepo{
		db:     db,
		tracer: otel.Tracer("ingesterRepo"),
	}
}

var _ domain.SnapshotRepository = (*PostgresRepo)(nil)

func (p *PostgresRepo) EnsureSnapshot(ctx context.Context, header domain.SnapshotHeader) (err error) {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// Upsert/ensure target
	tid, err := p.getOrCreateTargetID(ctx, tx, header.ServerHost)
	if err != nil {
		return fmt.Errorf("getOrCreateTargetID: %w", err)
	}

	// Insert snapshot if missing by f_id
	//language=SQL
	qIns := `
		INSERT INTO snapshot (f_id, snap_time, target_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (f_id) DO NOTHING
	`
	_, err = tx.ExecContext(ctx, qIns, header.SnapshotID, time.Unix(header.TimestampUnix, 0).In(time.UTC), tid)
	if err != nil {
		return fmt.Errorf("insert snapshot: %w", err)
	}
	return nil
}

func (p *PostgresRepo) SaveSamples(ctx context.Context, snapshotID string, seq uint32, samples []*dbmv1.QuerySample) (err error) {
	if len(samples) == 0 {
		return nil
	}
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// Resolve snapshot PK
	var snapPK int
	row := tx.QueryRowContext(ctx, "select id from snapshot where f_id = $1", snapshotID)
	if row == nil {
		return fmt.Errorf("query snapshot row nil")
	}
	if err = row.Scan(&snapPK); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("unknown snapshot id %s", snapshotID)
		}
		return fmt.Errorf("scan snapshot id: %w", err)
	}

	// COPY INTO query_samples mirroring v1 columns
	stmt, err := tx.PrepareContext(ctx, pq.CopyIn(
		"query_samples",
		"f_id", "snap_id", "sql_handle",
		"blocked", "blocker", "plan_handle", "data", "wait_event", "wait_time",
		"sid", "connection_id", "transaction_id", "block_ms", "block_count", "query_hash",
	))
	if err != nil {
		return fmt.Errorf("prepare COPY: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, s := range samples {
		var protoBytes []byte
		protoBytes, err = s.MarshalVT()
		if err != nil {
			return fmt.Errorf("marshal proto: %w", err)
		}
		// Defensive nil checks across nested optional messages
		var waitType string
		var waitTime int64
		if s.GetWaitInfo() != nil {
			waitType = s.GetWaitInfo().GetWaitType()
			waitTime = s.GetWaitInfo().GetWaitTime()
		}
		var sid, connID string
		if s.GetSession() != nil {
			sid = s.GetSession().GetSessionId()
			connID = s.GetSession().GetConnectionId()
		}
		var txID string
		if s.GetCommand() != nil {
			txID = s.GetCommand().GetTransactionId()
		}
		blockCount := 0
		if s.GetBlockInfo() != nil {
			blockCount = len(s.GetBlockInfo().GetBlockedSessions())
		}

		_, err = stmt.ExecContext(ctx,
			s.GetId(),         // f_id (row external id)
			snapPK,            // snap_id (FK)
			s.GetSqlHandle(),  // sql_handle
			s.GetBlocked(),    // blocked
			s.GetBlocker(),    // blocker
			s.GetPlanHandle(), // plan_handle
			protoBytes,        // data (protobuf)
			waitType,          // wait_event
			waitTime,          // wait_time
			sid,               // sid
			connID,            // connection_id
			txID,              // transaction_id
			-1,                // block_ms (not tracked in v2; keep -1 like v1)
			blockCount,        // block_count
			s.GetQueryHash(),  // query_hash
		)
		if err != nil {
			return fmt.Errorf("copy exec: %w", err)
		}
	}
	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("copy complete: %w", err)
	}
	return nil
}

func (p *PostgresRepo) FinalizeSnapshot(ctx context.Context, _ string) error {
	// v1 schema has no finalize marker; nothing to do for now.
	return nil
}

// getOrCreateTargetID resolves a target row id by host and a fixed type (1), creating if missing.
func (p *PostgresRepo) getOrCreateTargetID(ctx context.Context, tx *sqlx.Tx, host string) (int, error) {
	id, err := p.getTargetID(ctx, tx, host)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	// create
	//language=SQL
	qIns := `insert into target (host, type_id, agent_version) values ($1, $2, $3)`
	if _, err = tx.ExecContext(ctx, qIns, host, 1, "-"); err != nil {
		return 0, fmt.Errorf("insert target: %w", err)
	}
	return p.getTargetID(ctx, tx, host)
}

func (p *PostgresRepo) getTargetID(ctx context.Context, tx sqlx.QueryerContext, host string) (int, error) {
	//language=SQL
	q := `select id from target where host = $1 and type_id = 1`
	row := tx.QueryRowxContext(ctx, q, host)
	if row == nil {
		return 0, fmt.Errorf("nil row")
	}
	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (p *PostgresRepo) SaveExecutionPlans(ctx context.Context, executionPlans []*common_domain.ExecutionPlan) error {
	ctx, span := p.tracer.Start(ctx, "StoreExecutionPlans")
	defer span.End()
	span.SetAttributes(attribute.Int("num_samples", len(executionPlans)))
	q := `insert into query_plans (plan_handle, plan_xml, target_id) VALUES `
	if len(executionPlans) == 0 {
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
	targetID, err = p.getOrCreateTargetID(ctx, tx, executionPlans[0].Server.Host)
	if err != nil {
		return fmt.Errorf("get target id: %w", err)
	}
	chunks := slices.Chunk(executionPlans, 300)
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
