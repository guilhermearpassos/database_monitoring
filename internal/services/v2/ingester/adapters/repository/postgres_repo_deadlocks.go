package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/guilhermearpassos/database-monitoring/internal/deadlock"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/ingester/domain"
	"github.com/lib/pq"
)

var _ domain.DeadlockRepository = (*PostgresRepo)(nil)

func (p *PostgresRepo) SaveDeadlocks(ctx context.Context, serverMeta common_domain.ServerMeta, deadlocks []*common_domain.DeadlockReport) (stored int, err error) {
	ctx, span := p.tracer.Start(ctx, "SaveDeadlocks")
	defer span.End()

	if len(deadlocks) == 0 {
		return 0, nil
	}

	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	targetID, err := p.getOrCreateTargetID(ctx, tx, serverMeta.Host)
	if err != nil {
		return 0, fmt.Errorf("getOrCreateTargetID: %w", err)
	}

	const qIns = `
		INSERT INTO deadlocks (
			f_id, target_id, occurred_at, xml_hash, fingerprint, pattern,
			databases, victim_spid, max_wait_ms, process_count, resource_count, xml_report
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (target_id, occurred_at, xml_hash) DO NOTHING
	`

	for _, d := range deadlocks {
		if d == nil || d.XMLReport == "" {
			continue
		}
		fid := d.ID
		if _, parseErr := uuid.Parse(fid); parseErr != nil {
			fid = uuid.NewString()
		}
		xmlHash := d.XMLHash
		if xmlHash == "" {
			xmlHash = deadlock.HashXML(d.XMLReport)
		}
		dbs := d.Databases
		if dbs == nil {
			dbs = []string{}
		}
		occurredAt := d.OccurredAt.UTC()
		if occurredAt.IsZero() {
			occurredAt = time.Now().UTC()
		}

		res, execErr := tx.ExecContext(ctx, qIns,
			fid,
			targetID,
			occurredAt,
			xmlHash,
			d.Fingerprint,
			d.Pattern,
			pq.Array(dbs),
			d.VictimSPID,
			d.MaxWaitMs,
			d.ProcessCount,
			d.ResourceCount,
			d.XMLReport,
		)
		if execErr != nil {
			err = fmt.Errorf("insert deadlock: %w", execErr)
			return 0, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			stored += int(n)
		}
	}

	return stored, nil
}

func (p *PostgresRepo) PurgeDeadlocks(ctx context.Context, start time.Time, end time.Time, batchSize int) error {
	ctx, span := p.tracer.Start(ctx, "PurgeDeadlocks")
	defer span.End()

	const q = `
		WITH rows_to_delete AS (
			SELECT CTID FROM deadlocks
			WHERE occurred_at BETWEEN $1 AND $2
			LIMIT $3
		)
		DELETE FROM deadlocks USING rows_to_delete
		WHERE deadlocks.CTID = rows_to_delete.CTID
	`
	rowsAffected := int64(1)
	for rowsAffected > 0 {
		r, err := p.db.ExecContext(ctx, q, start, end, batchSize)
		if err != nil {
			return fmt.Errorf("purgeDeadlocks: %w", err)
		}
		rowsAffected, err = r.RowsAffected()
		if err != nil {
			return fmt.Errorf("purgeDeadlocks rows affected: %w", err)
		}
	}
	return nil
}
