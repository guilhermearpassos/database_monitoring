package collector

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/guilhermearpassos/database-monitoring/internal/services/common_domain"
	"github.com/prometheus/client_golang/prometheus"
)

func TestParseWaitResource(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantOK   bool
		wantType string
		wantDB   int
		wantHoBT int64
		wantObj  int64
		wantAU   int64
		wantFile int
		wantPage int64
		wantSlot int
	}{
		{
			name:     "key lock with hash",
			raw:      "KEY: 6:72057594041335808 (8194443284a0)",
			wantOK:   true,
			wantType: "KEY",
			wantDB:   6,
			wantHoBT: 72057594041335808,
		},
		{
			name:     "hobt lock",
			raw:      "HOBT: 6:72057594041335808",
			wantOK:   true,
			wantType: "HOBT",
			wantDB:   6,
			wantHoBT: 72057594041335808,
		},
		{
			name:     "object lock",
			raw:      "OBJECT: 6:2105058535:0",
			wantOK:   true,
			wantType: "OBJECT",
			wantDB:   6,
			wantObj:  2105058535,
		},
		{
			name:     "object compile lock",
			raw:      "OBJECT: 6:2105058535:0 [COMPILE]",
			wantOK:   true,
			wantType: "OBJECT",
			wantDB:   6,
			wantObj:  2105058535,
		},
		{
			name:     "tab lock normalized to OBJECT",
			raw:      "TAB: 6:2105058535:0",
			wantOK:   true,
			wantType: "OBJECT",
			wantDB:   6,
			wantObj:  2105058535,
		},
		{
			name:     "page lock",
			raw:      "PAGE: 6:1:441920",
			wantOK:   true,
			wantType: "PAGE",
			wantDB:   6,
			wantFile: 1,
			wantPage: 441920,
		},
		{
			name:     "rid lock",
			raw:      "RID: 6:1:441920:12",
			wantOK:   true,
			wantType: "RID",
			wantDB:   6,
			wantFile: 1,
			wantPage: 441920,
			wantSlot: 12,
		},
		{
			name:     "allocation unit lock",
			raw:      "ALLOCATION_UNIT: 6:72057594041335808",
			wantOK:   true,
			wantType: "ALLOCATION_UNIT",
			wantDB:   6,
			wantAU:   72057594041335808,
		},
		{
			name:     "database lock",
			raw:      "DATABASE: 6",
			wantOK:   true,
			wantType: "DATABASE",
			wantDB:   6,
		},
		{
			name:     "application lock",
			raw:      "APPLICATION: 6:[my_app_lock]:(deadbeef)",
			wantOK:   true,
			wantType: "APPLICATION",
			wantDB:   6,
		},
		{
			name:   "empty string",
			raw:    "   ",
			wantOK: false,
		},
		{
			name:     "no colon",
			raw:      "something_weird",
			wantOK:   false,
			wantType: "OTHER",
		},
		{
			name:     "malformed key",
			raw:      "KEY: notanumber:123",
			wantOK:   false,
			wantType: "KEY",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseWaitResource(tc.raw)
			if ok != tc.wantOK {
				t.Fatalf("ParseWaitResource(%q) ok = %v, want %v (got=%+v)", tc.raw, ok, tc.wantOK, got)
			}
			if got.ResourceType != tc.wantType {
				t.Errorf("ResourceType = %q, want %q", got.ResourceType, tc.wantType)
			}
			if !tc.wantOK {
				return
			}
			if got.DBID != tc.wantDB || got.HoBTID != tc.wantHoBT || got.ObjectID != tc.wantObj ||
				got.AllocUnitID != tc.wantAU || got.FileID != tc.wantFile || got.PageID != tc.wantPage || got.Slot != tc.wantSlot {
				t.Errorf("ParseWaitResource(%q) = %+v, unexpected parsed fields", tc.raw, got)
			}
		})
	}
}

func TestQuoteSQLIdentifier(t *testing.T) {
	if got := quoteSQLIdentifier("OrdersDB"); got != "[OrdersDB]" {
		t.Fatalf("got %q, want [OrdersDB]", got)
	}
	if got := quoteSQLIdentifier("Weird]DB"); got != "[Weird]]DB]" {
		t.Fatalf("got %q, want [Weird]]DB]", got)
	}
}

type fakeLockMetadataLookup struct {
	databases     map[int]string
	fetchDBCalls  int
	hobtCalls     int
	objectCalls   int
	allocCalls    int
	pageCalls     int
	hobtHandler   func(dbName string, hobtID int64) (string, string, string, error)
	objectHandler func(dbName string, objectID int64) (string, string, error)
	allocHandler  func(dbName string, allocUnitID int64) (string, string, string, error)
	pageHandler   func(dbID int, dbName string, fileID int, pageID int64) (string, string, string, error)
}

func (f *fakeLockMetadataLookup) FetchDatabases(ctx context.Context) (map[int]string, error) {
	f.fetchDBCalls++
	out := make(map[int]string, len(f.databases))
	for k, v := range f.databases {
		out[k] = v
	}
	return out, nil
}

func (f *fakeLockMetadataLookup) LookupByHoBT(ctx context.Context, dbName string, hobtID int64) (string, string, string, error) {
	f.hobtCalls++
	if f.hobtHandler != nil {
		return f.hobtHandler(dbName, hobtID)
	}
	return "", "", "", sql.ErrNoRows
}

func (f *fakeLockMetadataLookup) LookupByObjectID(ctx context.Context, dbName string, objectID int64) (string, string, error) {
	f.objectCalls++
	if f.objectHandler != nil {
		return f.objectHandler(dbName, objectID)
	}
	return "", "", sql.ErrNoRows
}

func (f *fakeLockMetadataLookup) LookupByAllocUnit(ctx context.Context, dbName string, allocUnitID int64) (string, string, string, error) {
	f.allocCalls++
	if f.allocHandler != nil {
		return f.allocHandler(dbName, allocUnitID)
	}
	return "", "", "", sql.ErrNoRows
}

func (f *fakeLockMetadataLookup) LookupByPage(ctx context.Context, dbID int, dbName string, fileID int, pageID int64) (string, string, string, error) {
	f.pageCalls++
	if f.pageHandler != nil {
		return f.pageHandler(dbID, dbName, fileID, pageID)
	}
	return "", "", "", sql.ErrNoRows
}

func TestWaitResourceResolver_CachesDatabasesAndTargets(t *testing.T) {
	ctx := context.Background()
	fake := &fakeLockMetadataLookup{
		databases: map[int]string{
			6: "OrdersDB",
		},
		hobtHandler: func(dbName string, hobtID int64) (string, string, string, error) {
			if dbName == "OrdersDB" && hobtID == 72057594041335808 {
				return "dbo", "Orders", "PK_Orders", nil
			}
			if dbName == "BillingDB" && hobtID == 555 {
				return "billing", "Invoices", "IX_Invoices_Date", nil
			}
			return "", "", "", sql.ErrNoRows
		},
		objectHandler: func(dbName string, objectID int64) (string, string, error) {
			if dbName == "OrdersDB" && objectID == 2105058535 {
				return "dbo", "Customers", nil
			}
			return "", "", sql.ErrNoRows
		},
		allocHandler: func(dbName string, allocUnitID int64) (string, string, string, error) {
			if dbName == "OrdersDB" && allocUnitID == 99999 {
				return "dbo", "AuditLog", "HEAP", nil
			}
			return "", "", "", sql.ErrNoRows
		},
		pageHandler: func(dbID int, dbName string, fileID int, pageID int64) (string, string, string, error) {
			if dbID == 6 && dbName == "OrdersDB" && fileID == 1 && pageID == 441920 {
				return "sales", "LineItems", "IX_LineItems_OrderId", nil
			}
			return "", "", "", sql.ErrNoRows
		},
	}

	resolver := NewWaitResourceResolver(fake)

	// Database map should load once and then be served from cache.
	dbMap, err := resolver.GetDatabaseMetadataMap(ctx)
	if err != nil {
		t.Fatalf("GetDatabaseMetadataMap error: %v", err)
	}
	if dbMap["6"].DatabaseName != "OrdersDB" {
		t.Fatalf("expected OrdersDB, got %+v", dbMap["6"])
	}
	if _, err := resolver.GetDatabaseMetadataMap(ctx); err != nil {
		t.Fatalf("second GetDatabaseMetadataMap error: %v", err)
	}
	if fake.fetchDBCalls != 1 {
		t.Fatalf("FetchDatabases calls = %d, want 1", fake.fetchDBCalls)
	}

	// KEY (two different row hashes) and HOBT share the same HoBT cache key -> 1 lookup total.
	t1, _ := resolver.Resolve(ctx, "KEY: 6:72057594041335808 (8194443284a0)", "")
	t2, _ := resolver.Resolve(ctx, "KEY: 6:72057594041335808 (deadbeefcafe)", "")
	t3, _ := resolver.Resolve(ctx, "HOBT: 6:72057594041335808", "")
	if t1.Database != "OrdersDB" || t1.Table != "dbo.Orders" || t1.Index != "PK_Orders" || t1.ResourceType != "KEY" {
		t.Fatalf("unexpected t1: %+v", t1)
	}
	if t2.Table != "dbo.Orders" || t2.ResourceType != "KEY" {
		t.Fatalf("unexpected t2: %+v", t2)
	}
	if t3.Table != "dbo.Orders" || t3.ResourceType != "HOBT" {
		t.Fatalf("unexpected t3: %+v", t3)
	}
	if fake.hobtCalls != 1 {
		t.Fatalf("LookupByHoBT calls = %d, want 1", fake.hobtCalls)
	}

	// OBJECT and TAB share the same object cache key -> 1 lookup total.
	o1, _ := resolver.Resolve(ctx, "OBJECT: 6:2105058535:0", "")
	o2, _ := resolver.Resolve(ctx, "TAB: 6:2105058535:0", "")
	if o1.Table != "dbo.Customers" || o1.Index != "" || o2.Table != "dbo.Customers" {
		t.Fatalf("unexpected object targets: o1=%+v o2=%+v", o1, o2)
	}
	if fake.objectCalls != 1 {
		t.Fatalf("LookupByObjectID calls = %d, want 1", fake.objectCalls)
	}

	// PAGE and RID on same page share the same page cache key -> 1 lookup total.
	p1, _ := resolver.Resolve(ctx, "PAGE: 6:1:441920", "")
	p2, _ := resolver.Resolve(ctx, "RID: 6:1:441920:12", "")
	if p1.Table != "sales.LineItems" || p1.Index != "IX_LineItems_OrderId" || p1.ResourceType != "PAGE" {
		t.Fatalf("unexpected p1: %+v", p1)
	}
	if p2.Table != "sales.LineItems" || p2.ResourceType != "RID" {
		t.Fatalf("unexpected p2: %+v", p2)
	}
	if fake.pageCalls != 1 {
		t.Fatalf("LookupByPage calls = %d, want 1", fake.pageCalls)
	}

	// ALLOCATION_UNIT cached after first call.
	a1, _ := resolver.Resolve(ctx, "ALLOCATION_UNIT: 6:99999", "")
	a2, _ := resolver.Resolve(ctx, "ALLOCATION_UNIT: 6:99999", "")
	if a1.Table != "dbo.AuditLog" || a1.Index != "HEAP" || a2.Table != "dbo.AuditLog" {
		t.Fatalf("unexpected alloc unit targets: a1=%+v a2=%+v", a1, a2)
	}
	if fake.allocCalls != 1 {
		t.Fatalf("LookupByAllocUnit calls = %d, want 1", fake.allocCalls)
	}

	// Unknown DB ID triggers a refresh of sys.databases and resolves the newly added database.
	fake.databases[9] = "BillingDB"
	b1, _ := resolver.Resolve(ctx, "KEY: 9:555 (001122)", "")
	if b1.Database != "BillingDB" || b1.Table != "billing.Invoices" || b1.Index != "IX_Invoices_Date" {
		t.Fatalf("unexpected new DB target: %+v", b1)
	}
	if fake.fetchDBCalls != 2 {
		t.Fatalf("FetchDatabases calls = %d, want 2 after DB miss", fake.fetchDBCalls)
	}
}

func TestWaitResourceResolver_TransientErrorNotCached(t *testing.T) {
	ctx := context.Background()
	failOnce := true
	fake := &fakeLockMetadataLookup{
		databases: map[int]string{6: "OrdersDB"},
		hobtHandler: func(dbName string, hobtID int64) (string, string, string, error) {
			if failOnce {
				failOnce = false
				return "", "", "", errors.New("lock timeout")
			}
			return "dbo", "Orders", "PK_Orders", nil
		},
	}

	resolver := NewWaitResourceResolver(fake)

	first, _ := resolver.Resolve(ctx, "KEY: 6:100 (aaa)", "FallbackDB")
	if first.Database != "OrdersDB" || first.Table != "" || first.ResourceType != "KEY" {
		t.Fatalf("expected graceful fallback on error, got %+v", first)
	}

	second, _ := resolver.Resolve(ctx, "KEY: 6:100 (aaa)", "FallbackDB")
	if second.Table != "dbo.Orders" || second.Index != "PK_Orders" {
		t.Fatalf("expected retry to succeed after transient error, got %+v", second)
	}

	third, _ := resolver.Resolve(ctx, "KEY: 6:100 (bbb)", "FallbackDB")
	if third.Table != "dbo.Orders" {
		t.Fatalf("expected cached hit on third call, got %+v", third)
	}
	if fake.hobtCalls != 2 {
		t.Fatalf("LookupByHoBT calls = %d, want 2 (1 failed + 1 succeeded/cached)", fake.hobtCalls)
	}
}

func TestLockMetricsRecorder_RecordSnapshotAndReset(t *testing.T) {
	ctx := context.Background()
	reg := prometheus.NewRegistry()
	recorder := newLockMetricsRecorder(reg)

	fake := &fakeLockMetadataLookup{
		databases: map[int]string{6: "OrdersDB"},
		hobtHandler: func(dbName string, hobtID int64) (string, string, string, error) {
			if hobtID == 100 {
				return "dbo", "Orders", "PK_Orders", nil
			}
			return "", "", "", sql.ErrNoRows
		},
		objectHandler: func(dbName string, objectID int64) (string, string, error) {
			if objectID == 200 {
				return "dbo", "Customers", nil
			}
			return "", "", sql.ErrNoRows
		},
	}
	resolver := NewWaitResourceResolver(fake)

	lckX := "LCK_M_X"
	lckSchS := "LCK_M_SCH_S"

	// Snapshot 1: 2 blocked on dbo.Orders (PK_Orders), 1 blocked on dbo.Customers, 1 non-blocked sample.
	snap1 := []*common_domain.QuerySample{
		{
			IsBlocked: true,
			Database:  common_domain.DataBaseMetadata{DatabaseName: "OrdersDB"},
			Wait:      common_domain.WaitMetadata{WaitType: &lckX, WaitTime: 1500, WaitResource: "KEY: 6:100 (aaaa)"},
		},
		{
			IsBlocked: true,
			Database:  common_domain.DataBaseMetadata{DatabaseName: "OrdersDB"},
			Wait:      common_domain.WaitMetadata{WaitType: &lckX, WaitTime: 4200, WaitResource: "KEY: 6:100 (bbbb)"},
		},
		{
			IsBlocked: true,
			Database:  common_domain.DataBaseMetadata{DatabaseName: "OrdersDB"},
			Wait:      common_domain.WaitMetadata{WaitType: &lckSchS, WaitTime: 800, WaitResource: "OBJECT: 6:200:0"},
		},
		{
			IsBlocked: false,
			Database:  common_domain.DataBaseMetadata{DatabaseName: "OrdersDB"},
			Wait:      common_domain.WaitMetadata{WaitType: &lckX, WaitTime: 9999, WaitResource: "KEY: 6:100 (cccc)"},
		},
	}

	recorder.RecordSnapshot(ctx, "sql-prod-1", snap1, resolver)

	ordersLabels := map[string]string{
		"server":        "sql-prod-1",
		"database":      "OrdersDB",
		"table":         "dbo.Orders",
		"index":         "PK_Orders",
		"wait_type":     "LCK_M_X",
		"resource_type": "KEY",
	}
	custLabels := map[string]string{
		"server":        "sql-prod-1",
		"database":      "OrdersDB",
		"table":         "dbo.Customers",
		"index":         "",
		"wait_type":     "LCK_M_SCH_S",
		"resource_type": "OBJECT",
	}

	metricValue := func(metricName string, wantLabels map[string]string) float64 {
		t.Helper()
		mfs, err := reg.Gather()
		if err != nil {
			t.Fatalf("reg.Gather error: %v", err)
		}
		for _, mf := range mfs {
			if mf.GetName() != metricName {
				continue
			}
			for _, m := range mf.GetMetric() {
				matched := len(m.GetLabel()) == len(wantLabels)
				for _, lp := range m.GetLabel() {
					if wantLabels[lp.GetName()] != lp.GetValue() {
						matched = false
						break
					}
				}
				if !matched {
					continue
				}
				if g := m.GetGauge(); g != nil {
					return g.GetValue()
				}
				if c := m.GetCounter(); c != nil {
					return c.GetValue()
				}
			}
		}
		t.Fatalf("metric %s with labels %v not found", metricName, wantLabels)
		return 0
	}

	if got := metricValue("sqlsights_database_blocked_sessions", ordersLabels); got != 2 {
		t.Fatalf("snap1 orders blockedSessions = %v, want 2", got)
	}
	if got := metricValue("sqlsights_database_blocked_max_wait_seconds", ordersLabels); got != 4.2 {
		t.Fatalf("snap1 orders blockedMaxWaitSec = %v, want 4.2", got)
	}
	if got := metricValue("sqlsights_database_locks_total", ordersLabels); got != 2 {
		t.Fatalf("snap1 orders locksTotal = %v, want 2", got)
	}

	if got := metricValue("sqlsights_database_blocked_sessions", custLabels); got != 1 {
		t.Fatalf("snap1 customers blockedSessions = %v, want 1", got)
	}
	if got := metricValue("sqlsights_database_blocked_max_wait_seconds", custLabels); got != 0.8 {
		t.Fatalf("snap1 customers blockedMaxWaitSec = %v, want 0.8", got)
	}
	if got := metricValue("sqlsights_database_locks_total", custLabels); got != 1 {
		t.Fatalf("snap1 customers locksTotal = %v, want 1", got)
	}

	// Verify histogram observations were recorded on the registry.
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("reg.Gather error: %v", err)
	}
	var histCount uint64
	for _, mf := range mfs {
		if mf.GetName() == "sqlsights_database_lock_duration_seconds" {
			for _, m := range mf.GetMetric() {
				histCount += m.GetHistogram().GetSampleCount()
			}
		}
	}
	if histCount != 3 {
		t.Fatalf("histogram sample count = %d, want 3", histCount)
	}

	// Snapshot 2: Orders lock cleared, Customers still blocked with 2.5s wait.
	snap2 := []*common_domain.QuerySample{
		{
			IsBlocked: true,
			Database:  common_domain.DataBaseMetadata{DatabaseName: "OrdersDB"},
			Wait:      common_domain.WaitMetadata{WaitType: &lckSchS, WaitTime: 2500, WaitResource: "OBJECT: 6:200:0"},
		},
	}
	recorder.RecordSnapshot(ctx, "sql-prod-1", snap2, resolver)

	if got := metricValue("sqlsights_database_blocked_sessions", ordersLabels); got != 0 {
		t.Fatalf("snap2 orders blockedSessions = %v, want 0 after clear", got)
	}
	if got := metricValue("sqlsights_database_blocked_max_wait_seconds", ordersLabels); got != 0 {
		t.Fatalf("snap2 orders blockedMaxWaitSec = %v, want 0 after clear", got)
	}
	if got := metricValue("sqlsights_database_locks_total", ordersLabels); got != 2 {
		t.Fatalf("snap2 orders locksTotal = %v, want 2 (monotonic)", got)
	}

	if got := metricValue("sqlsights_database_blocked_sessions", custLabels); got != 1 {
		t.Fatalf("snap2 customers blockedSessions = %v, want 1", got)
	}
	if got := metricValue("sqlsights_database_blocked_max_wait_seconds", custLabels); got != 2.5 {
		t.Fatalf("snap2 customers blockedMaxWaitSec = %v, want 2.5", got)
	}
	if got := metricValue("sqlsights_database_locks_total", custLabels); got != 2 {
		t.Fatalf("snap2 customers locksTotal = %v, want 2", got)
	}

	// Snapshot 3: No blocked sessions at all -> Customers gauges reset to 0 too.
	recorder.RecordSnapshot(ctx, "sql-prod-1", nil, resolver)

	if got := metricValue("sqlsights_database_blocked_sessions", custLabels); got != 0 {
		t.Fatalf("snap3 customers blockedSessions = %v, want 0 after clear", got)
	}
	if got := metricValue("sqlsights_database_blocked_max_wait_seconds", custLabels); got != 0 {
		t.Fatalf("snap3 customers blockedMaxWaitSec = %v, want 0 after clear", got)
	}
}
