package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/v2/querier/domain"
	"github.com/lib/pq"
)

var _ domain.DeadlockRepository = (*PostgresRepo)(nil)

func (p *PostgresRepo) ListDeadlocks(
	ctx context.Context,
	start time.Time,
	end time.Time,
	host string,
	databases []string,
) ([]*common_domain.DeadlockReport, error) {
	ctx, span := p.tracer.Start(ctx, "ListDeadlocks")
	defer span.End()

	args := []any{host, start.UTC(), end.UTC()}
	query := `
		SELECT
			d.f_id::text,
			t.host,
			d.occurred_at,
			d.xml_report,
			d.xml_hash,
			d.fingerprint,
			d.pattern,
			d.databases,
			d.victim_spid,
			d.max_wait_ms,
			d.process_count,
			d.resource_count
		FROM deadlocks d
		INNER JOIN public.target t ON t.id = d.target_id
		WHERE t.host = $1
		  AND d.occurred_at BETWEEN $2 AND $3
	`
	if len(databases) > 0 {
		query += ` AND d.databases && $4`
		args = append(args, pq.Array(databases))
	}
	query += ` ORDER BY d.occurred_at DESC LIMIT 500`

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing deadlocks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	results := make([]*common_domain.DeadlockReport, 0)
	for rows.Next() {
		var (
			id            string
			srvHost       string
			occurredAt    time.Time
			xmlReport     string
			xmlHash       string
			fingerprint   string
			pattern       string
			dbs           pq.StringArray
			victimSPID    string
			maxWaitMs     int64
			processCount  int32
			resourceCount int32
		)
		if err := rows.Scan(
			&id,
			&srvHost,
			&occurredAt,
			&xmlReport,
			&xmlHash,
			&fingerprint,
			&pattern,
			&dbs,
			&victimSPID,
			&maxWaitMs,
			&processCount,
			&resourceCount,
		); err != nil {
			return nil, fmt.Errorf("scan deadlock row: %w", err)
		}
		results = append(results, &common_domain.DeadlockReport{
			ID:            id,
			Server:        common_domain.ServerMeta{Host: srvHost, Type: "mssql"},
			OccurredAt:    occurredAt.UTC(),
			XMLReport:     xmlReport,
			XMLHash:       xmlHash,
			Fingerprint:   fingerprint,
			Pattern:       pattern,
			Databases:     []string(dbs),
			VictimSPID:    victimSPID,
			MaxWaitMs:     maxWaitMs,
			ProcessCount:  processCount,
			ResourceCount: resourceCount,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("deadlocks rows.Err: %w", err)
	}
	return results, nil
}
