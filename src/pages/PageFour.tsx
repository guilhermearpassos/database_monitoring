import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { FetchResponse, getBackendSrv, PluginPage } from '@grafana/runtime';
import { Observable } from 'rxjs';
import {
  Alert,
  Button,
  Combobox,
  ComboboxOption,
  Drawer,
  Input,
  LoadingPlaceholder,
  MultiCombobox,
  TimeRangePicker,
  useStyles2,
} from '@grafana/ui';
import { GrafanaTheme2, TimeRange } from '@grafana/data';
import { css } from '@emotion/css';
import { Activity, BarChart3, FileSearch, RefreshCw, Search, X } from 'lucide-react';
import {
  DEADLOCK_PATTERN_OPTIONS,
  DeadlockItem,
  DeadlocksResponse,
  DeadlocksSortField,
  DeadlocksViewMode,
  SortDirection,
} from '../components/Deadlocks/types';
import { DeadlocksTable } from '../components/Deadlocks/DeadlocksTable';
import { DeadlockDrawer } from '../components/Deadlocks/DeadlockDrawer';
import { DeadlocksTimeline } from '../components/Deadlocks/DeadlocksTimeline';
import {
  applyTimeRangeToSearchParams,
  buildPreservedNavSearch,
  extractErrorMessage,
  parseCSVString,
  parseTimeRangeFromSearchParams,
  resolveTimeRange,
  setCSVParam,
  setOptionalParam,
  shiftTimeRange,
  zoomOutTimeRange,
} from '../utils/urlState';
import { ROUTES } from '../constants';
import { prefixRoute } from '../utils/utils.routing';
import { formatDurationMs } from '../utils/sqlHelpers';

const PATTERN_OPTIONS: Array<ComboboxOption<string>> = DEADLOCK_PATTERN_OPTIONS.map((p) => ({
  label: p.label,
  value: p.value,
}));

const getStyles = (theme: GrafanaTheme2) => ({
  headerRow: css`
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: ${theme.spacing(2)};
    flex-wrap: wrap;
    margin-bottom: ${theme.spacing(2)};
  `,
  headerRight: css`
    display: flex;
    align-items: center;
    gap: ${theme.spacing(1)};
    flex-wrap: wrap;
  `,
  filterCard: css`
    background: ${theme.colors.background.secondary};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    padding: ${theme.spacing(2)};
    margin-bottom: ${theme.spacing(2)};
  `,
  filterGrid: css`
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    gap: ${theme.spacing(1.5)};
  `,
  filterField: css`
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 170px;
  `,
  filterLabel: css`
    font-size: ${theme.typography.bodySmall.fontSize};
    color: ${theme.colors.text.secondary};
    font-weight: ${theme.typography.fontWeightMedium};
  `,
  kpiGrid: css`
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
    gap: ${theme.spacing(1.5)};
    margin-bottom: ${theme.spacing(2)};
  `,
  kpiCard: css`
    background: ${theme.colors.background.primary};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    padding: ${theme.spacing(1.5, 2)};
  `,
  kpiLabel: css`
    font-size: ${theme.typography.bodySmall.fontSize};
    color: ${theme.colors.text.secondary};
    margin-bottom: 4px;
  `,
  kpiValue: css`
    font-size: ${theme.typography.h2.fontSize};
    font-weight: ${theme.typography.fontWeightBold};
    color: ${theme.colors.text.primary};
  `,
  kpiSub: css`
    font-size: 11px;
    color: ${theme.colors.text.secondary};
    margin-top: 2px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  `,
  section: css`
    margin-bottom: ${theme.spacing(2)};
  `,
  drawerContent: css`
    overflow-y: auto;
    padding: ${theme.spacing(2)};
    box-sizing: border-box;
  `,
  emptyPrompt: css`
    padding: ${theme.spacing(5)};
    text-align: center;
    background: ${theme.colors.background.secondary};
    border: 1px dashed ${theme.colors.border.medium};
    border-radius: ${theme.shape.radius.default};
    color: ${theme.colors.text.secondary};
  `,
});

const PageFour = () => {
  const styles = useStyles2(getStyles);
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  // URL-synced state (primitive strings for stable memo dependencies)
  const selectedServer = searchParams.get('server') || '';
  const dbParam = searchParams.get('db') || '';
  const patternsParam = searchParams.get('patterns') || '';
  const selectedDatabases = useMemo(() => parseCSVString(dbParam), [dbParam]);
  const selectedPatterns = useMemo(() => parseCSVString(patternsParam), [patternsParam]);
  const searchQuery = searchParams.get('q') || '';
  const viewMode: DeadlocksViewMode = searchParams.get('view') === 'patterns' ? 'patterns' : 'events';
  const defaultSortField: DeadlocksSortField = viewMode === 'patterns' ? 'occurrences' : 'timestamp';
  const sortField = (searchParams.get('sort') as DeadlocksSortField) || defaultSortField;
  const sortDir: SortDirection = searchParams.get('dir') === 'asc' ? 'asc' : 'desc';
  const selectedDeadlockId = searchParams.get('deadlockId') || '';

  const [timeRange, setTimeRange] = useState<TimeRange>(() => parseTimeRangeFromSearchParams(searchParams));
  const [servers, setServers] = useState<Array<ComboboxOption<string>>>([]);
  const [availableDatabases, setAvailableDatabases] = useState<string[]>([]);
  const [deadlocks, setDeadlocks] = useState<DeadlockItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [serversError, setServersError] = useState<string | null>(null);
  const [refreshCounter, setRefreshCounter] = useState(0);

  const handleRefresh = useCallback(() => {
    setTimeRange((prev) => resolveTimeRange(prev));
    setRefreshCounter((c) => c + 1);
  }, []);

  const updateParams = useCallback(
    (mutator: (next: URLSearchParams) => void) => {
      const next = new URLSearchParams(searchParams);
      mutator(next);
      if (next.toString() === searchParams.toString()) {
        return;
      }
      setSearchParams(next, { replace: true });
    },
    [searchParams, setSearchParams]
  );

  useEffect(() => {
    let cancelled = false;
    setServersError(null);
    const params = new URLSearchParams({
      start: timeRange.from.toISOString(),
      end: timeRange.to.toISOString(),
      type: 'databases',
    });
    let sub: { unsubscribe: () => void } | undefined;
    try {
      const response: Observable<FetchResponse<Array<{ label: string; value: string }>>> = getBackendSrv().fetch({
        url: `/api/plugins/guilhermearpassos-sqlsights-app/resources/datasource-options?${params.toString()}`,
      });
      sub = response.subscribe({
        next: (result) => {
          if (!cancelled) {
            setServers((result.data || []).map((s) => ({ label: s.label, value: s.value })));
          }
        },
        error: (err) => {
          if (!cancelled) {
            setServersError(extractErrorMessage(err, 'Failed to load servers'));
          }
        },
      });
    } catch (err) {
      if (!cancelled) {
        setServersError(extractErrorMessage(err, 'Failed to load servers'));
      }
    }
    return () => {
      cancelled = true;
      sub?.unsubscribe();
    };
  }, [timeRange, refreshCounter]);

  useEffect(() => {
    if (!selectedServer && servers.length === 1) {
      updateParams((p) => p.set('server', servers[0].value));
    }
  }, [selectedServer, servers, updateParams]);

  useEffect(() => {
    if (!selectedServer) {
      setDeadlocks([]);
      setAvailableDatabases([]);
      setError(null);
      setLoading(false);
      return;
    }
    let cancelled = false;
    setLoading(true);
    setError(null);
    const params = new URLSearchParams({
      server: selectedServer,
      start: timeRange.from.toISOString(),
      end: timeRange.to.toISOString(),
    });
    if (selectedDatabases.length > 0) {
      params.set('databases', selectedDatabases.join(','));
    }
    let sub: { unsubscribe: () => void } | undefined;
    try {
      const response: Observable<FetchResponse<DeadlocksResponse>> = getBackendSrv().fetch({
        url: `/api/plugins/guilhermearpassos-sqlsights-app/resources/deadlocks?${params.toString()}`,
      });
      sub = response.subscribe({
        next: (result) => {
          if (!cancelled) {
            setDeadlocks(result.data?.deadlocks || []);
            setAvailableDatabases(result.data?.available_databases || []);
            setLoading(false);
          }
        },
        error: (err) => {
          if (!cancelled) {
            setError(extractErrorMessage(err, 'Failed to load deadlocks'));
            setDeadlocks([]);
            setLoading(false);
          }
        },
      });
    } catch (err) {
      if (!cancelled) {
        setError(extractErrorMessage(err, 'Failed to load deadlocks'));
        setDeadlocks([]);
        setLoading(false);
      }
    }
    return () => {
      cancelled = true;
      sub?.unsubscribe();
    };
  }, [selectedServer, timeRange, selectedDatabases, refreshCounter]);

  const filteredDeadlocks = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();
    return deadlocks.filter((d) => {
      if (selectedPatterns.length > 0 && !selectedPatterns.includes(d.diagnosis?.pattern || '')) {
        return false;
      }
      if (q) {
        const hay = [
          d.diagnosis?.pattern_label || '',
          (d.databases ?? []).join(' '),
          ...(d.processes ?? []).map(
            (p) => `${p.spid} ${p.client_app} ${p.login_name} ${p.host_name} ${p.active_sql} ${p.input_buf}`
          ),
          ...(d.resources ?? []).map((r) => `${r.object_name} ${r.index_name} ${r.type}`),
        ]
          .join(' ')
          .toLowerCase();
        if (!hay.includes(q)) {
          return false;
        }
      }
      return true;
    });
  }, [deadlocks, selectedPatterns, searchQuery]);

  const kpis = useMemo(() => {
    const uniquePatterns = new Set(filteredDeadlocks.map((d) => d.fingerprint || d.id)).size;
    const objCounts: Record<string, number> = {};
    let maxWaitMs = 0;
    for (const d of filteredDeadlocks) {
      if (d.max_wait_ms > maxWaitMs) {
        maxWaitMs = d.max_wait_ms;
      }
      const seenLabels = new Set<string>();
      for (const r of d.resources ?? []) {
        const label = r.index_name ? `${r.short_object || r.object_name} [${r.index_name}]` : r.short_object || r.object_name;
        if (label) {
          seenLabels.add(label);
        }
      }
      for (const label of seenLabels) {
        objCounts[label] = (objCounts[label] || 0) + 1;
      }
    }
    const topObj = Object.entries(objCounts).sort((a, b) => b[1] - a[1])[0];
    return {
      total: filteredDeadlocks.length,
      uniquePatterns,
      topObject: topObj ? topObj[0] : '-',
      topObjectCount: topObj ? topObj[1] : 0,
      maxWaitMs,
    };
  }, [filteredDeadlocks]);

  const dbOptions = useMemo<Array<ComboboxOption<string>>>(() => {
    const merged = new Set([...availableDatabases, ...selectedDatabases]);
    return Array.from(merged)
      .sort()
      .map((db) => ({ label: db, value: db }));
  }, [availableDatabases, selectedDatabases]);

  const selectedDeadlock = useMemo(() => {
    if (!selectedDeadlockId) {
      return null;
    }
    return deadlocks.find((d) => d.id === selectedDeadlockId) ?? null;
  }, [deadlocks, selectedDeadlockId]);

  const handleTimeRangeChange = useCallback(
    (next: TimeRange) => {
      setTimeRange(next);
      updateParams((p) => applyTimeRangeToSearchParams(p, next));
    },
    [updateParams]
  );

  const handleSortChange = (field: DeadlocksSortField) => {
    updateParams((p) => {
      if (sortField === field) {
        p.set('dir', sortDir === 'desc' ? 'asc' : 'desc');
      } else {
        p.set('sort', field);
        p.set('dir', 'desc');
      }
    });
  };

  const handleViewModeChange = (mode: DeadlocksViewMode) => {
    updateParams((p) => {
      setOptionalParam(p, 'view', mode === 'events' ? null : mode);
      p.delete('sort');
    });
  };

  const handleOpenSnapshotsAround = (timestampIso: string) => {
    const center = new Date(timestampIso).getTime();
    if (!Number.isFinite(center)) {
      navigate(`${prefixRoute(ROUTES.Snapshots)}${buildPreservedNavSearch(searchParams)}`);
      return;
    }
    const from = new Date(center - 2 * 60_000).toISOString();
    const to = new Date(center + 2 * 60_000).toISOString();
    navigate(`${prefixRoute(ROUTES.Snapshots)}${buildPreservedNavSearch(searchParams, { from, to })}`);
  };

  const hasActiveFilters = selectedDatabases.length > 0 || selectedPatterns.length > 0 || searchQuery.length > 0;

  return (
    <PluginPage>
      <div className={styles.headerRow}>
        <div>
          <h2 style={{ margin: '0 0 4px 0' }}>Deadlocks Explorer</h2>
          <p style={{ margin: 0 }}>
            Analyze SQL Server <code>xml_deadlock_report</code> graphs collected from <code>system_health</code>,
            grouped by structural pattern with automated root-cause diagnosis.
          </p>
        </div>

        <div className={styles.headerRight}>
          <Button
            variant="secondary"
            fill="outline"
            onClick={() => navigate(`${prefixRoute(ROUTES.Snapshots)}${buildPreservedNavSearch(searchParams)}`)}
          >
            <Activity size={14} />
            <span style={{ marginLeft: 6 }}>Snapshots</span>
          </Button>
          <Button
            variant="secondary"
            fill="outline"
            onClick={() => navigate(`${prefixRoute(ROUTES.PlanAnalysis)}${buildPreservedNavSearch(searchParams)}`)}
          >
            <FileSearch size={14} />
            <span style={{ marginLeft: 6 }}>Plan Analysis</span>
          </Button>
          <Button
            variant="secondary"
            fill="outline"
            onClick={() => navigate(`${prefixRoute(ROUTES.QueryMetrics)}${buildPreservedNavSearch(searchParams)}`)}
          >
            <BarChart3 size={14} />
            <span style={{ marginLeft: 6 }}>Query Metrics</span>
          </Button>
          <Button variant="secondary" fill="outline" onClick={handleRefresh} disabled={loading}>
            <RefreshCw size={14} />
          </Button>
          <TimeRangePicker
            value={timeRange}
            onChange={handleTimeRangeChange}
            timeZone="browser"
            onChangeTimeZone={() => {}}
            onZoom={() => handleTimeRangeChange(zoomOutTimeRange(timeRange))}
            onMoveBackward={() => handleTimeRangeChange(shiftTimeRange(timeRange, -1))}
            onMoveForward={() => handleTimeRangeChange(shiftTimeRange(timeRange, 1))}
          />
        </div>
      </div>

      {(error || serversError) && (
        <Alert title="Error" severity="error" style={{ marginBottom: 16 }}>
          <div>{error || serversError}</div>
          <Button size="sm" variant="secondary" onClick={handleRefresh} style={{ marginTop: 8 }}>
            Retry
          </Button>
        </Alert>
      )}

      <div className={styles.filterCard}>
        <div className={styles.filterGrid}>
          <div className={styles.filterField}>
            <span className={styles.filterLabel}>SQL Server Instance</span>
            <Combobox<string>
              width={26}
              options={servers}
              value={selectedServer || null}
              onChange={(opt) =>
                updateParams((p) => {
                  setOptionalParam(p, 'server', opt?.value);
                  p.delete('db');
                  p.delete('deadlockId');
                })
              }
              placeholder="Select a server..."
              isClearable
            />
          </div>

          <div className={styles.filterField}>
            <span className={styles.filterLabel}>Database(s)</span>
            <MultiCombobox<string>
              width={30}
              options={dbOptions}
              value={selectedDatabases}
              onChange={(opts) => updateParams((p) => setCSVParam(p, 'db', opts.map((o) => o.value)))}
              placeholder={selectedServer ? 'All databases' : 'Select server first'}
              disabled={!selectedServer}
            />
          </div>

          <div className={styles.filterField}>
            <span className={styles.filterLabel}>Deadlock Pattern</span>
            <MultiCombobox<string>
              width={32}
              options={PATTERN_OPTIONS}
              value={selectedPatterns}
              onChange={(opts) => updateParams((p) => setCSVParam(p, 'patterns', opts.map((o) => o.value)))}
              placeholder="All patterns"
            />
          </div>

          <div className={styles.filterField} style={{ flex: 1, minWidth: 220 }}>
            <span className={styles.filterLabel}>Search SQL / Table / Index / App / SPID</span>
            <Input
              prefix={<Search size={14} />}
              value={searchQuery}
              onChange={(e) => updateParams((p) => setOptionalParam(p, 'q', e.currentTarget.value))}
              placeholder="Filter by SQL, table, index, login, SPID..."
            />
          </div>

          {hasActiveFilters && (
            <Button
              size="sm"
              variant="secondary"
              fill="text"
              onClick={() =>
                updateParams((p) => {
                  p.delete('db');
                  p.delete('patterns');
                  p.delete('q');
                })
              }
            >
              <X size={13} />
              <span style={{ marginLeft: 4 }}>Clear Filters</span>
            </Button>
          )}
        </div>
      </div>

      {!selectedServer ? (
        <div className={styles.emptyPrompt}>Select a SQL Server instance above to explore captured deadlocks.</div>
      ) : loading && deadlocks.length === 0 ? (
        <LoadingPlaceholder text="Loading deadlock reports..." />
      ) : (
        <>
          <div className={styles.kpiGrid}>
            <div className={styles.kpiCard}>
              <div className={styles.kpiLabel}>Total Deadlocks</div>
              <div className={styles.kpiValue} style={{ color: kpis.total > 0 ? '#e02f44' : undefined }}>
                {kpis.total.toLocaleString()}
              </div>
              <div className={styles.kpiSub}>Error 1205 rollbacks in window</div>
            </div>
            <div className={styles.kpiCard}>
              <div className={styles.kpiLabel}>Distinct Patterns</div>
              <div className={styles.kpiValue}>{kpis.uniquePatterns.toLocaleString()}</div>
              <div className={styles.kpiSub}>Grouped by objects, indexes &amp; cycle type</div>
            </div>
            <div className={styles.kpiCard}>
              <div className={styles.kpiLabel}>Most Contested Object</div>
              <div className={styles.kpiValue} style={{ fontSize: 15, lineHeight: 1.4, marginTop: 4 }} title={kpis.topObject}>
                {kpis.topObject}
              </div>
              <div className={styles.kpiSub}>
                {kpis.topObjectCount > 0 ? `Involved in ${kpis.topObjectCount} deadlock(s)` : 'No contention'}
              </div>
            </div>
            <div className={styles.kpiCard}>
              <div className={styles.kpiLabel}>Longest Wait Before Kill</div>
              <div className={styles.kpiValue}>{formatDurationMs(kpis.maxWaitMs)}</div>
              <div className={styles.kpiSub}>Peak victim wait duration</div>
            </div>
          </div>

          <div className={styles.section}>
            <DeadlocksTimeline deadlocks={filteredDeadlocks} timeRange={timeRange} />
          </div>

          <DeadlocksTable
            deadlocks={filteredDeadlocks}
            viewMode={viewMode}
            onViewModeChange={handleViewModeChange}
            sortField={sortField}
            sortDir={sortDir}
            onSortChange={handleSortChange}
            onSelectDeadlock={(item) => updateParams((p) => p.set('deadlockId', item.id))}
          />
        </>
      )}

      {selectedDeadlock && (
        <Drawer
          title="Deadlock Report & Analysis"
          size="lg"
          onClose={() => updateParams((p) => p.delete('deadlockId'))}
        >
          <div className={styles.drawerContent}>
            <DeadlockDrawer
              deadlock={selectedDeadlock}
              onOpenSnapshotsAround={handleOpenSnapshotsAround}
              onOpenPlanAnalysis={(db) =>
                navigate(`${prefixRoute(ROUTES.PlanAnalysis)}${buildPreservedNavSearch(searchParams, { db: db || null })}`)
              }
              onOpenQueryMetrics={(db) =>
                navigate(`${prefixRoute(ROUTES.QueryMetrics)}${buildPreservedNavSearch(searchParams, { db: db || null })}`)
              }
            />
          </div>
        </Drawer>
      )}
    </PluginPage>
  );
};

export default PageFour;
