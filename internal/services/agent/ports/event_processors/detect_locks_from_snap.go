package event_processors

import (
	"context"
	"fmt"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/adapters/metrics"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/app"
	"github.com/guilhermearpassos/database-monitoring/internal/services/agent/domain/events"
	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"regexp"
	"strings"
)

type MetricsDetector struct {
	app                  app.Application
	in                   chan events.Event
	trace                trace.Tracer
	knownHandlesByServer map[string]map[string]struct{}
	lockThresholdSeconds float64
}

func NewMetricsDetector(app app.Application) *MetricsDetector {
	return &MetricsDetector{
		app:                  app,
		in:                   make(chan events.Event, 200),
		trace:                otel.Tracer("MetricsDetector"),
		knownHandlesByServer: make(map[string]map[string]struct{}),
	}
}

func (f MetricsDetector) Run(ctx context.Context) {
	for ev := range f.in {
		snapTakenEvent, ok := ev.(events.SampleSnapshotTaken)
		if !ok {
			continue
		}
		_, span := f.trace.Start(ctx, "ExtractMetricsFromSnap")

		f.processSnapshot(snapTakenEvent.Snap)
		span.End()
	}
}

func (f MetricsDetector) Register(router *events.EventRouter) {
	router.Register(events.SampleSnapshotTaken{}.EventName(), f.in)
}

// generateLockKey creates a unique key for a lock
func generateLockKey(server, database, sessionID string) string {
	return fmt.Sprintf("%s:%s:%s", server, database, sessionID)
}

// processSnapshot extracts metrics from a database snapshot
func (f MetricsDetector) processSnapshot(snapshot *common_domain.DataBaseSnapshot) {
	server := snapshot.SnapInfo.Server.Host

	// Track the current set of active locks in this snapshot
	currentLocks := make(map[string]bool)
	longRunningLocks := make(map[string]int) // database -> count

	// Process each sample to update metrics and track active locks
	for _, sample := range snapshot.Samples {
		if sample.IsBlocked {
			// Create a unique key for this lock
			lockKey := generateLockKey(server, sample.Database.DatabaseName, sample.Session.SessionID)
			tables, err := ExtractTablesFromQuery(sample.Text)
			if err != nil {
				fmt.Println(fmt.Errorf("Error extracting tables: %s", err))
			}
			// Mark this lock as currently active
			currentLocks[lockKey] = true

			// Get wait time in seconds
			waitTimeSeconds := float64(sample.Wait.WaitTime) / 1000.0

			// Update lock duration metric
			var waitType string
			if sample.Wait.WaitType != nil {
				waitType = *sample.Wait.WaitType
			} else {
				waitType = "unknown"
			}
			for _, t := range tables {

				metrics.DatabaseLockDuration.With(prometheus.Labels{
					"server":    server,
					"database":  sample.Database.DatabaseName,
					"wait_type": waitType,
					"table":     t,
				}).Observe(float64(sample.Wait.WaitTime) / 1000.0)
			}

			// Increment total locks counter
			metrics.DatabaseLocksTotal.With(prometheus.Labels{
				"server":   server,
				"database": sample.Database.DatabaseName,
			}).Inc()

			// Check if this is a long-running lock
			if waitTimeSeconds > f.lockThresholdSeconds {
				longRunningLocks[sample.Database.DatabaseName]++
			}
		}
	}

	// Clear databases with no long-running locks
	// This ensures metrics are properly zeroed when locks clear
	databasesWithLocks := make(map[string]bool)
	for database := range longRunningLocks {
		databasesWithLocks[database] = true
	}

	// Get databases from this server that don't have long-running locks
	// and set their metrics to zero

}

// ExtractTablesFromQuery parses a SQL query and returns a slice of table names
// that are involved in the query.
func ExtractTablesFromQuery(batchQuery string) ([]string, error) {
	// Parse the SQL query
	batchQuery = strings.ReplaceAll(batchQuery, "\r\n", "\n")

	// First, handle the case where the query begins with parameter declarations
	// and we need to skip the first line
	if strings.HasPrefix(batchQuery, "(@p") {
		lines := strings.SplitN(batchQuery, "\n", 2)
		if len(lines) > 1 {
			batchQuery = lines[1]
		}
	}

	// Split into individual statements (separated by semicolons)
	statements := splitBatchIntoStatements(batchQuery)

	tableSet := make(map[string]struct{})

	// Process each statement
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Extract tables based on statement type
		if strings.HasPrefix(strings.ToUpper(stmt), "INSERT INTO") {
			extractTablesFromInsert(stmt, tableSet)
		} else if strings.HasPrefix(strings.ToUpper(stmt), "UPDATE") {
			extractTablesFromUpdate(stmt, tableSet)
		} else if strings.HasPrefix(strings.ToUpper(stmt), "DELETE") {
			extractTablesFromDelete(stmt, tableSet)
		} else if strings.HasPrefix(strings.ToUpper(stmt), "SELECT") {
			extractTablesFromSelect(stmt, tableSet)
		} else if strings.Contains(strings.ToUpper(stmt), "JOIN") {
			extractTablesFromJoin(stmt, tableSet)
		}
	}

	// Convert map to slice
	tables := make([]string, 0, len(tableSet))
	for table := range tableSet {
		tables = append(tables, table)
	}

	return tables, nil
}

// splitBatchIntoStatements splits a SQL batch into individual statements
func splitBatchIntoStatements(batch string) []string {
	// This is a simplified approach - in a real implementation, you'd need to
	// handle semicolons in strings, comments, etc.
	statements := []string{}

	// Simple split by semicolon
	rawStatements := strings.Split(batch, ";")

	for _, stmt := range rawStatements {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}

// extractTablesFromInsert extracts table names from INSERT statements
func extractTablesFromInsert(stmt string, tableSet map[string]struct{}) {
	// Pattern for INSERT INTO [table]
	re := regexp.MustCompile(`(?i)INSERT\s+INTO\s+([^\s(]+)`)
	matches := re.FindStringSubmatch(stmt)

	if len(matches) > 1 {
		tableName := cleanTableName(matches[1])
		tableSet[tableName] = struct{}{}
	}

	// Also check for tables in a SELECT part of the INSERT
	if strings.Contains(strings.ToUpper(stmt), "SELECT") {
		selectPart := strings.SplitN(stmt, "SELECT", 2)
		if len(selectPart) > 1 {
			extractTablesFromSelect("SELECT"+selectPart[1], tableSet)
		}
	}
}

// extractTablesFromUpdate extracts table names from UPDATE statements
func extractTablesFromUpdate(stmt string, tableSet map[string]struct{}) {
	// Pattern for UPDATE [table]
	re := regexp.MustCompile(`(?i)UPDATE\s+([^\s]+)`)
	matches := re.FindStringSubmatch(stmt)

	if len(matches) > 1 {
		tableName := cleanTableName(matches[1])
		tableSet[tableName] = struct{}{}
	}

	// Also check for tables in FROM/JOIN parts of the UPDATE
	if strings.Contains(strings.ToUpper(stmt), "FROM") {
		fromPart := strings.SplitN(strings.ToUpper(stmt), "FROM", 2)
		if len(fromPart) > 1 {
			extractTablesFromFromClause("FROM"+fromPart[1], tableSet)
		}
	}
}

// extractTablesFromDelete extracts table names from DELETE statements
func extractTablesFromDelete(stmt string, tableSet map[string]struct{}) {
	// Pattern for DELETE FROM [table]
	re := regexp.MustCompile(`(?i)DELETE\s+FROM\s+([^\s(]+)`)
	matches := re.FindStringSubmatch(stmt)

	if len(matches) > 1 {
		tableName := cleanTableName(matches[1])
		tableSet[tableName] = struct{}{}
	}

	// Also check for tables in JOIN parts of the DELETE
	if strings.Contains(strings.ToUpper(stmt), "JOIN") {
		extractTablesFromJoin(stmt, tableSet)
	}
}

// extractTablesFromSelect extracts table names from SELECT statements
func extractTablesFromSelect(stmt string, tableSet map[string]struct{}) {
	// Extract tables from FROM clause
	if strings.Contains(strings.ToUpper(stmt), "FROM") {
		parts := strings.SplitN(strings.ToUpper(stmt), "FROM", 2)
		if len(parts) > 1 {
			extractTablesFromFromClause("FROM"+parts[1], tableSet)
		}
	}
}

// extractTablesFromJoin extracts table names from JOIN clauses
func extractTablesFromJoin(stmt string, tableSet map[string]struct{}) {
	// Pattern for all types of JOINs
	re := regexp.MustCompile(`(?i)(JOIN)\s+([^\s(]+)`)
	matches := re.FindAllStringSubmatch(stmt, -1)

	for _, match := range matches {
		if len(match) > 2 {
			tableName := cleanTableName(match[2])
			tableSet[tableName] = struct{}{}
		}
	}
}

// extractTablesFromFromClause extracts table names from FROM clauses
func extractTablesFromFromClause(fromClause string, tableSet map[string]struct{}) {
	// Handle basic FROM clause with commas
	// Cut off any WHERE, GROUP BY, etc. that might follow
	fromClauseParts := strings.Split(fromClause, " WHERE ")
	fromClause = fromClauseParts[0]

	fromClauseParts = strings.Split(fromClause, " GROUP BY ")
	fromClause = fromClauseParts[0]

	fromClauseParts = strings.Split(fromClause, " ORDER BY ")
	fromClause = fromClauseParts[0]

	// Extract the actual table list after FROM keyword
	parts := strings.SplitN(fromClause, "FROM", 2)
	if len(parts) < 2 {
		return
	}

	tableList := parts[1]

	// Handle JOIN clauses
	if strings.Contains(strings.ToUpper(tableList), "JOIN") {
		extractTablesFromJoin(tableList, tableSet)

		// Also handle tables before the first JOIN
		beforeJoin := strings.SplitN(tableList, "JOIN", 2)
		if len(beforeJoin) > 0 {
			tables := strings.Split(beforeJoin[0], ",")
			for _, table := range tables {
				table = strings.TrimSpace(table)
				if table != "" {
					// Handle table aliases (remove "AS alias" or just "alias")
					tableNameParts := strings.Fields(table)
					if len(tableNameParts) > 0 {
						tableName := cleanTableName(tableNameParts[0])
						tableSet[tableName] = struct{}{}
					}
				}
			}
		}
	} else {
		// Simple comma-separated table list
		tables := strings.Split(tableList, ",")
		for _, table := range tables {
			table = strings.TrimSpace(table)
			if table != "" {
				// Handle table aliases
				tableNameParts := strings.Fields(table)
				if len(tableNameParts) > 0 {
					tableName := cleanTableName(tableNameParts[0])
					tableSet[tableName] = struct{}{}
				}
			}
		}
	}
}

// cleanTableName removes brackets, quotes, and handles schema prefixes
func cleanTableName(tableName string) string {
	// Remove any brackets or quotes
	tableName = strings.TrimSpace(tableName)
	tableName = strings.Trim(tableName, "[]\"'`")

	// Handle table variables (they start with @)
	if strings.HasPrefix(tableName, "@") {
		return tableName
	}

	// If you want to strip schema names, uncomment:
	// parts := strings.Split(tableName, ".")
	// if len(parts) > 1 {
	//     return parts[len(parts)-1]
	// }

	return tableName
}
