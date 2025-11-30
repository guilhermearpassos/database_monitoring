package adapters

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/guilhermearpassos/database-monitoring/internal/services/collector/domain"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	dbmv1 "github.com/guilhermearpassos/database-monitoring/proto/database_monitoring/v1"
)

var _ domain.WarningsRepository = (*PostgresRepo)(nil)

func (p *PostgresRepo) StoreWarnings(ctx context.Context, warnings []*common_domain.Warning, serverMeta common_domain.ServerMeta) error {
	if len(warnings) == 0 {
		return nil
	}

	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	tId, err := p.getTargetID(ctx, tx, serverMeta)
	if err != nil {
		return fmt.Errorf("failed to get target id: %w", err)
	}
	// Use UPSERT to handle conflicts based on the unique constraint (target_id, name)
	query := `
		INSERT INTO public.warnings (name, target_id, warning_type, warning_data)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (target_id, name) 
		DO UPDATE SET 
			warning_type = EXCLUDED.warning_type,
			warning_data = EXCLUDED.warning_data
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, warning := range warnings {
		warningData, err := proto.Marshal(warning.WarningData)
		if err != nil {
			return fmt.Errorf("failed to marshal warning data for '%s': %w", warning.Id, err)
		}

		_, err = stmt.ExecContext(ctx,
			warning.Id,
			tId,
			warning.WarningType,
			warningData,
		)
		if err != nil {
			return fmt.Errorf("failed to store warning '%s': %w", warning.Id, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (p *PostgresRepo) GetKnownWarnings(ctx context.Context, serverID string, pageSize int, pageNumber int) ([]*common_domain.Warning, error) {
	query := `
		SELECT name, warning_type, warning_data
		FROM public.warnings
		WHERE target_id = $1
		ORDER BY name
		offset $2 rows limit $3
	`
	targetID, err := p.getTargetID(ctx, p.db, common_domain.ServerMeta{Host: serverID})
	if err != nil {
		return nil, fmt.Errorf("failed to get target id: %w", err)
	}
	rows, err := p.db.QueryContext(ctx, query, targetID, pageSize*(pageNumber-1), pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to query warnings: %w", err)
	}
	defer rows.Close()

	var warnings []*common_domain.Warning

	for rows.Next() {
		var name, warningType string
		var warningData []byte

		if err := rows.Scan(&name, &warningType, &warningData); err != nil {
			return nil, fmt.Errorf("failed to scan warning row: %w", err)
		}

		var protoWarning dbmv1.Warning
		if err := proto.Unmarshal(warningData, &protoWarning); err != nil {
			return nil, fmt.Errorf("failed to unmarshal warning data for '%s': %w", name, err)
		}

		warning := &common_domain.Warning{
			Id:          name,
			WarningType: warningType,
			WarningData: &protoWarning,
		}

		warnings = append(warnings, warning)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating warning rows: %w", err)
	}

	return warnings, nil
}

// GetWarningsByType retrieves warnings of a specific type for a server
func (p *PostgresRepo) GetWarningsByType(ctx context.Context, serverID string, warningType string) ([]*common_domain.Warning, error) {
	query := `
		SELECT name, warning_type, warning_data
		FROM public.warnings
		WHERE target_id = $1 AND warning_type = $2
		ORDER BY name
	`

	rows, err := p.db.QueryContext(ctx, query, serverID, warningType)
	if err != nil {
		return nil, fmt.Errorf("failed to query warnings by type: %w", err)
	}
	defer rows.Close()

	var warnings []*common_domain.Warning

	for rows.Next() {
		var name, warningType string
		var warningData []byte

		if err := rows.Scan(&name, &warningType, &warningData); err != nil {
			return nil, fmt.Errorf("failed to scan warning row: %w", err)
		}

		var protoWarning dbmv1.Warning
		if err := proto.Unmarshal(warningData, &protoWarning); err != nil {
			return nil, fmt.Errorf("failed to unmarshal warning data for '%s': %w", name, err)
		}

		warning := &common_domain.Warning{
			Id:          name,
			WarningType: warningType,
			WarningData: &protoWarning,
		}

		warnings = append(warnings, warning)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating warning rows: %w", err)
	}

	return warnings, nil
}

// DeleteWarning removes a specific warning by name and server ID
func (p *PostgresRepo) DeleteWarning(ctx context.Context, serverID string, warningName string) error {
	query := `DELETE FROM public.warnings WHERE target_id = $1 AND name = $2`

	result, err := p.db.ExecContext(ctx, query, serverID, warningName)
	if err != nil {
		return fmt.Errorf("failed to delete warning '%s': %w", warningName, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("warning '%s' not found for server '%s'", warningName, serverID)
	}

	return nil
}

// DeleteWarningsByType removes all warnings of a specific type for a server
func (p *PostgresRepo) DeleteWarningsByType(ctx context.Context, serverID string, warningType string) error {
	query := `DELETE FROM public.warnings WHERE target_id = $1 AND warning_type = $2`

	_, err := p.db.ExecContext(ctx, query, serverID, warningType)
	if err != nil {
		return fmt.Errorf("failed to delete warnings of type '%s': %w", warningType, err)
	}

	return nil
}

// GetWarningStats returns statistics about warnings per server and type
func (p *PostgresRepo) GetWarningStats(ctx context.Context) (map[string]map[string]int, error) {
	query := `
		SELECT target_id, warning_type, COUNT(*) as count
		FROM public.warnings
		GROUP BY target_id, warning_type
		ORDER BY target_id, warning_type
	`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query warning stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]map[string]int)

	for rows.Next() {
		var targetID, warningType string
		var count int

		if err := rows.Scan(&targetID, &warningType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan warning stats row: %w", err)
		}

		if _, exists := stats[targetID]; !exists {
			stats[targetID] = make(map[string]int)
		}
		stats[targetID][warningType] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating warning stats rows: %w", err)
	}

	return stats, nil
}

// BatchStoreWarnings stores warnings in batches for better performance
func (p *PostgresRepo) BatchStoreWarnings(ctx context.Context, warnings []*common_domain.Warning, serverMeta common_domain.ServerMeta, batchSize int) error {
	if len(warnings) == 0 {
		return nil
	}

	if batchSize <= 0 {
		batchSize = 100 // default batch size
	}

	for i := 0; i < len(warnings); i += batchSize {
		end := i + batchSize
		if end > len(warnings) {
			end = len(warnings)
		}

		batch := warnings[i:end]
		if err := p.StoreWarnings(ctx, batch, serverMeta); err != nil {
			return fmt.Errorf("failed to store warning batch starting at index %d: %w", i, err)
		}
	}

	return nil
}
