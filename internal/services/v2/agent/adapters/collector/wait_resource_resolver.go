package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/jmoiron/sqlx"
)

// ParsedWaitResource represents a structured SQL Server dm_exec_requests.wait_resource value.
type ParsedWaitResource struct {
	ResourceType string // KEY, HOBT, OBJECT, TAB, PAGE, RID, ALLOCATION_UNIT, DATABASE, APPLICATION, OTHER
	DBID         int
	HoBTID       int64
	ObjectID     int64
	AllocUnitID  int64
	FileID       int
	PageID       int64
	Slot         int
}

// ResolvedLockTarget holds the resolved table/index identity for a wait_resource.
type ResolvedLockTarget struct {
	Database     string
	Schema       string
	Table        string // schema.table (or table if schema empty)
	Index        string // index name, "HEAP", or ""
	ResourceType string
}

// ParseWaitResource parses SQL Server wait_resource strings such as:
//   - "KEY: 6:72057594041335808 (8194443284a0)"
//   - "HOBT: 6:72057594041335808"
//   - "OBJECT: 6:2105058535:0" or "OBJECT: 6:2105058535:0 [COMPILE]"
//   - "TAB: 6:2105058535:0"
//   - "PAGE: 6:1:441920"
//   - "RID: 6:1:441920:12"
//   - "ALLOCATION_UNIT: 6:72057594041335808"
//   - "DATABASE: 6"
//   - "APPLICATION: 6:[my_lock]:(hash)"
func ParseWaitResource(raw string) (ParsedWaitResource, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ParsedWaitResource{}, false
	}
	colonIdx := strings.IndexByte(trimmed, ':')
	if colonIdx <= 0 {
		return ParsedWaitResource{ResourceType: "OTHER"}, false
	}

	resType := strings.ToUpper(strings.TrimSpace(trimmed[:colonIdx]))
	rest := strings.TrimSpace(trimmed[colonIdx+1:])
	// Strip trailing parenthetical hash "(8194443284a0)" or bracketed subtype "[COMPILE]"
	if spIdx := strings.IndexAny(rest, " \t(["); spIdx >= 0 {
		rest = strings.TrimSpace(rest[:spIdx])
	}
	parts := strings.Split(rest, ":")

	parsed := ParsedWaitResource{ResourceType: resType}
	switch resType {
	case "KEY", "HOBT":
		if len(parts) < 2 {
			return parsed, false
		}
		dbID, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		hobtID, err2 := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err1 != nil || err2 != nil {
			return parsed, false
		}
		parsed.DBID = dbID
		parsed.HoBTID = hobtID
		return parsed, true

	case "OBJECT", "TAB":
		if len(parts) < 2 {
			return parsed, false
		}
		dbID, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		objID, err2 := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err1 != nil || err2 != nil {
			return parsed, false
		}
		parsed.ResourceType = "OBJECT"
		parsed.DBID = dbID
		parsed.ObjectID = objID
		return parsed, true

	case "PAGE":
		if len(parts) < 3 {
			return parsed, false
		}
		dbID, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		fileID, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		pageID, err3 := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
		if err1 != nil || err2 != nil || err3 != nil {
			return parsed, false
		}
		parsed.DBID = dbID
		parsed.FileID = fileID
		parsed.PageID = pageID
		return parsed, true

	case "RID":
		if len(parts) < 3 {
			return parsed, false
		}
		dbID, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		fileID, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		pageID, err3 := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
		if err1 != nil || err2 != nil || err3 != nil {
			return parsed, false
		}
		parsed.DBID = dbID
		parsed.FileID = fileID
		parsed.PageID = pageID
		if len(parts) >= 4 {
			if slot, err := strconv.Atoi(strings.TrimSpace(parts[3])); err == nil {
				parsed.Slot = slot
			}
		}
		return parsed, true

	case "ALLOCATION_UNIT":
		if len(parts) < 2 {
			return parsed, false
		}
		dbID, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		auID, err2 := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err1 != nil || err2 != nil {
			return parsed, false
		}
		parsed.DBID = dbID
		parsed.AllocUnitID = auID
		return parsed, true

	case "DATABASE", "APPLICATION":
		if len(parts) < 1 {
			return parsed, false
		}
		dbID, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return parsed, false
		}
		parsed.DBID = dbID
		return parsed, true

	default:
		return ParsedWaitResource{ResourceType: resType}, false
	}
}

func (p ParsedWaitResource) cacheKey() string {
	switch p.ResourceType {
	case "KEY", "HOBT":
		return fmt.Sprintf("hobt:%d:%d", p.DBID, p.HoBTID)
	case "OBJECT":
		return fmt.Sprintf("obj:%d:%d", p.DBID, p.ObjectID)
	case "PAGE", "RID":
		return fmt.Sprintf("page:%d:%d:%d", p.DBID, p.FileID, p.PageID)
	case "ALLOCATION_UNIT":
		return fmt.Sprintf("au:%d:%d", p.DBID, p.AllocUnitID)
	default:
		return ""
	}
}

// lockMetadataLookup abstracts SQL Server catalog lookups for unit testing.
type lockMetadataLookup interface {
	FetchDatabases(ctx context.Context) (map[int]string, error)
	LookupByHoBT(ctx context.Context, dbName string, hobtID int64) (schema, table, index string, err error)
	LookupByObjectID(ctx context.Context, dbName string, objectID int64) (schema, table string, err error)
	LookupByAllocUnit(ctx context.Context, dbName string, allocUnitID int64) (schema, table, index string, err error)
	LookupByPage(ctx context.Context, dbID int, dbName string, fileID int, pageID int64) (schema, table, index string, err error)
}

type sqlServerLockMetadataLookup struct {
	db *sqlx.DB
}

func quoteSQLIdentifier(name string) string {
	return "[" + strings.ReplaceAll(name, "]", "]]") + "]"
}

func (l sqlServerLockMetadataLookup) FetchDatabases(ctx context.Context) (map[int]string, error) {
	if l.db == nil {
		return nil, errors.New("nil db connection")
	}
	query := monitoringSessionPrefix + `SELECT database_id, name FROM sys.databases OPTION (MAXDOP 1)`
	rows, err := l.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query sys.databases: %w", err)
	}
	defer func() { _ = rows.Close() }()

	res := make(map[int]string)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("scan sys.databases: %w", err)
		}
		res[id] = name
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows sys.databases: %w", err)
	}
	return res, nil
}

func (l sqlServerLockMetadataLookup) LookupByHoBT(ctx context.Context, dbName string, hobtID int64) (string, string, string, error) {
	if l.db == nil {
		return "", "", "", errors.New("nil db connection")
	}
	qDB := quoteSQLIdentifier(dbName)
	query := fmt.Sprintf(monitoringSessionPrefix+`
SELECT TOP 1
    ISNULL(s.name, '') AS schema_name,
    ISNULL(o.name, '') AS object_name,
    CASE WHEN p.index_id = 0 THEN 'HEAP' ELSE ISNULL(i.name, '') END AS index_name
FROM %s.sys.partitions p
INNER JOIN %s.sys.objects o ON p.object_id = o.object_id
INNER JOIN %s.sys.schemas s ON o.schema_id = s.schema_id
LEFT JOIN %s.sys.indexes i ON p.object_id = i.object_id AND p.index_id = i.index_id
WHERE p.hobt_id = ?
OPTION (MAXDOP 1)`, qDB, qDB, qDB, qDB)

	var schema, table, index string
	err := l.db.QueryRowContext(ctx, query, hobtID).Scan(&schema, &table, &index)
	return schema, table, index, err
}

func (l sqlServerLockMetadataLookup) LookupByObjectID(ctx context.Context, dbName string, objectID int64) (string, string, error) {
	if l.db == nil {
		return "", "", errors.New("nil db connection")
	}
	qDB := quoteSQLIdentifier(dbName)
	query := fmt.Sprintf(monitoringSessionPrefix+`
SELECT TOP 1
    ISNULL(s.name, '') AS schema_name,
    ISNULL(o.name, '') AS object_name
FROM %s.sys.objects o
INNER JOIN %s.sys.schemas s ON o.schema_id = s.schema_id
WHERE o.object_id = ?
OPTION (MAXDOP 1)`, qDB, qDB)

	var schema, table string
	err := l.db.QueryRowContext(ctx, query, objectID).Scan(&schema, &table)
	return schema, table, err
}

func (l sqlServerLockMetadataLookup) LookupByAllocUnit(ctx context.Context, dbName string, allocUnitID int64) (string, string, string, error) {
	if l.db == nil {
		return "", "", "", errors.New("nil db connection")
	}
	qDB := quoteSQLIdentifier(dbName)
	query := fmt.Sprintf(monitoringSessionPrefix+`
SELECT TOP 1
    ISNULL(s.name, '') AS schema_name,
    ISNULL(o.name, '') AS object_name,
    CASE WHEN p.index_id = 0 THEN 'HEAP' ELSE ISNULL(i.name, '') END AS index_name
FROM %s.sys.allocation_units au
INNER JOIN %s.sys.partitions p
    ON (au.type IN (1, 3) AND au.container_id = p.hobt_id)
    OR (au.type = 2 AND au.container_id = p.partition_id)
INNER JOIN %s.sys.objects o ON p.object_id = o.object_id
INNER JOIN %s.sys.schemas s ON o.schema_id = s.schema_id
LEFT JOIN %s.sys.indexes i ON p.object_id = i.object_id AND p.index_id = i.index_id
WHERE au.allocation_unit_id = ?
OPTION (MAXDOP 1)`, qDB, qDB, qDB, qDB, qDB)

	var schema, table, index string
	err := l.db.QueryRowContext(ctx, query, allocUnitID).Scan(&schema, &table, &index)
	return schema, table, index, err
}

func (l sqlServerLockMetadataLookup) LookupByPage(ctx context.Context, dbID int, dbName string, fileID int, pageID int64) (string, string, string, error) {
	if l.db == nil {
		return "", "", "", errors.New("nil db connection")
	}
	qDB := quoteSQLIdentifier(dbName)
	query := fmt.Sprintf(monitoringSessionPrefix+`
SELECT TOP 1
    ISNULL(s.name, '') AS schema_name,
    ISNULL(o.name, '') AS object_name,
    CASE WHEN pi.index_id = 0 THEN 'HEAP' ELSE ISNULL(i.name, '') END AS index_name
FROM sys.dm_db_page_info(?, ?, ?, 'LIMITED') pi
INNER JOIN %s.sys.objects o ON pi.object_id = o.object_id
INNER JOIN %s.sys.schemas s ON o.schema_id = s.schema_id
LEFT JOIN %s.sys.indexes i ON pi.object_id = i.object_id AND pi.index_id = i.index_id
WHERE pi.object_id IS NOT NULL
OPTION (MAXDOP 1)`, qDB, qDB, qDB)

	var schema, table, index string
	err := l.db.QueryRowContext(ctx, query, dbID, fileID, pageID).Scan(&schema, &table, &index)
	return schema, table, index, err
}

// WaitResourceResolver caches sys.databases (id -> name) and resolved table/index lock
// targets in memory (no expiration; locked objects are typically a small set of repeat offenders).
type WaitResourceResolver struct {
	lookup lockMetadataLookup

	mu            sync.RWMutex
	dbCache       map[int]string
	resourceCache map[string]ResolvedLockTarget
}

func NewWaitResourceResolver(lookup lockMetadataLookup) *WaitResourceResolver {
	return &WaitResourceResolver{
		lookup:        lookup,
		dbCache:       make(map[int]string),
		resourceCache: make(map[string]ResolvedLockTarget),
	}
}

// GetDatabaseMetadataMap returns the cached database_id -> DataBaseMetadata map,
// loading from sys.databases on first call.
func (r *WaitResourceResolver) GetDatabaseMetadataMap(ctx context.Context) (map[string]common_domain.DataBaseMetadata, error) {
	r.mu.RLock()
	if len(r.dbCache) > 0 {
		out := make(map[string]common_domain.DataBaseMetadata, len(r.dbCache))
		for id, name := range r.dbCache {
			idStr := strconv.Itoa(id)
			out[idStr] = common_domain.DataBaseMetadata{DatabaseID: idStr, DatabaseName: name}
		}
		r.mu.RUnlock()
		return out, nil
	}
	r.mu.RUnlock()

	if err := r.refreshDatabases(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]common_domain.DataBaseMetadata, len(r.dbCache))
	for id, name := range r.dbCache {
		idStr := strconv.Itoa(id)
		out[idStr] = common_domain.DataBaseMetadata{DatabaseID: idStr, DatabaseName: name}
	}
	return out, nil
}

func (r *WaitResourceResolver) refreshDatabases(ctx context.Context) error {
	if r.lookup == nil {
		return errors.New("nil lock metadata lookup")
	}
	dbs, err := r.lookup.FetchDatabases(ctx)
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.dbCache = dbs
	r.mu.Unlock()
	return nil
}

func (r *WaitResourceResolver) resolveDBName(ctx context.Context, dbID int) string {
	r.mu.RLock()
	name, ok := r.dbCache[dbID]
	r.mu.RUnlock()
	if ok && name != "" {
		return name
	}
	// Cache miss (e.g. cold cache or newly created database): refresh sys.databases once.
	if err := r.refreshDatabases(ctx); err != nil {
		slog.WarnContext(ctx, "failed to refresh sys.databases cache", "db_id", dbID, "error", err)
		return ""
	}
	r.mu.RLock()
	name = r.dbCache[dbID]
	r.mu.RUnlock()
	return name
}

// Resolve parses a wait_resource string and resolves its database, schema.table, and index
// using the in-memory caches and querying SQL Server only on cache misses.
func (r *WaitResourceResolver) Resolve(ctx context.Context, rawWaitResource string, fallbackDB string) (ResolvedLockTarget, error) {
	parsed, ok := ParseWaitResource(rawWaitResource)
	if !ok {
		resType := parsed.ResourceType
		if resType == "" {
			resType = "UNKNOWN"
		}
		return ResolvedLockTarget{
			Database:     fallbackDB,
			ResourceType: resType,
		}, nil
	}

	dbName := r.resolveDBName(ctx, parsed.DBID)
	if dbName == "" {
		dbName = fallbackDB
	}

	cacheKey := parsed.cacheKey()
	if cacheKey == "" {
		return ResolvedLockTarget{
			Database:     dbName,
			ResourceType: parsed.ResourceType,
		}, nil
	}

	r.mu.RLock()
	if cached, hit := r.resourceCache[cacheKey]; hit {
		r.mu.RUnlock()
		cached.ResourceType = parsed.ResourceType
		return cached, nil
	}
	r.mu.RUnlock()

	if dbName == "" || r.lookup == nil {
		return ResolvedLockTarget{
			Database:     dbName,
			ResourceType: parsed.ResourceType,
		}, nil
	}

	var schema, table, index string
	var err error
	switch parsed.ResourceType {
	case "KEY", "HOBT":
		schema, table, index, err = r.lookup.LookupByHoBT(ctx, dbName, parsed.HoBTID)
	case "OBJECT":
		schema, table, err = r.lookup.LookupByObjectID(ctx, dbName, parsed.ObjectID)
	case "ALLOCATION_UNIT":
		schema, table, index, err = r.lookup.LookupByAllocUnit(ctx, dbName, parsed.AllocUnitID)
	case "PAGE", "RID":
		schema, table, index, err = r.lookup.LookupByPage(ctx, parsed.DBID, dbName, parsed.FileID, parsed.PageID)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ResolvedLockTarget{
				Database:     dbName,
				ResourceType: parsed.ResourceType,
			}, nil
		}
		return ResolvedLockTarget{
			Database:     dbName,
			ResourceType: parsed.ResourceType,
		}, fmt.Errorf("resolving lock target: %w", err)
	}

	fullTable := formatTableName(schema, table)
	target := ResolvedLockTarget{
		Database:     dbName,
		Schema:       schema,
		Table:        fullTable,
		Index:        index,
		ResourceType: parsed.ResourceType,
	}

	if fullTable != "" {
		r.mu.Lock()
		r.resourceCache[cacheKey] = target
		r.mu.Unlock()
	}
	return target, nil
}

func formatTableName(schema, table string) string {
	schema = strings.TrimSpace(schema)
	table = strings.TrimSpace(table)
	if table == "" {
		return ""
	}
	if schema == "" {
		return table
	}
	return schema + "." + table
}
