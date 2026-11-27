import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { FetchResponse, getBackendSrv, PluginPage } from '@grafana/runtime';
import { lastValueFrom, Observable } from 'rxjs';
import {
  Alert,
  Button,
  Combobox,
  ComboboxOption,
  Drawer,
  Input,
  LoadingPlaceholder,
  MultiCombobox,
  Switch,
  TimeRangePicker,
  useStyles2,
} from '@grafana/ui';
import { GrafanaTheme2, TimeRange } from '@grafana/data';
import { css } from '@emotion/css';
import { Activity, FileSearch, RefreshCw, Search, X } from 'lucide-react';
import {
  mapSortFieldToAggregationMode,
  METRIC_ISSUE_FLAGS,
  QueryMetricItem,
  QueryMetricsAggregationMode,
  QueryMetricsResponse,
  QueryMetricsSortField,
  SortDirection,
} from '../components/QueryMetrics/types';
import { QueryMetricsTable } from '../components/QueryMetrics/QueryMetricsTable';
import { QueryMetricDrawer } from '../components/QueryMetrics/QueryMetricDrawer';
import {
  applyTimeRangeToSearchParams,
  buildPreservedNavSearch,
  parseCSVString,
  parseTimeRangeFromSearchParams,
  setCSVParam,
  setOptionalParam,
} from '../utils/urlState';
import { ROUTES } from '../constants';
import { prefixRoute } from '../utils/utils.routing';
import { formatDurationMs } from '../utils/sqlHelpers';

const ISSUE_OPTIONS: Array<ComboboxOption<string>> = METRIC_ISSUE_FLAGS.map((f) => ({
  label: f.label,
  value: f.value,
  description: f.description,
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
  quickRow: css`
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: ${theme.spacing(1.5)};
    margin-top: ${theme.spacing(1.5)};
    padding-top: ${theme.spacing(1.5)};
    border-top: 1px solid ${theme.colors.border.weak};
  `,
  kpiGrid: css`
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(170px, 1fr));
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

const PageThree = () => {
  const styles = useStyles2(getStyles);
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  // URL-synced state
  const selectedServer = searchParams.get('server') || '';
  const dbParam = searchParams.get('db') || '';
  const issuesParam = searchParams.get('issues') || '';
  const selectedDatabases = useMemo(() => parseCSVString(dbParam), [dbParam]);
  const selectedIssues = useMemo(() => parseCSVString(issuesParam), [issuesParam]);
  const onlyWithIssues = searchParams.get('onlyIssues') === '1';
  const minExecs = Number(searchParams.get('minExecs') || '0') || 0;
  const searchQuery = searchParams.get('q') || '';
  const aggMode: QueryMetricsAggregationMode = searchParams.get('agg') === 'avg' ? 'avg' : 'total';
  const defaultSortField: QueryMetricsSortField = aggMode === 'avg' ? 'avg_elapsed_ms' : 'total_elapsed_ms';
  const sortField = (searchParams.get('sort') as QueryMetricsSortField) || defaultSortField;
  const sortDir: SortDirection = searchParams.get('dir') === 'asc' ? 'asc' : 'desc';
  const selectedQueryHash = searchParams.get('queryHash') || '';

  const [timeRange, setTimeRange] = useState<TimeRange>(() => parseTimeRangeFromSearchParams(searchParams));
  const [servers, setServers] = useState<Array<ComboboxOption<string>>>([]);
  const [availableDatabases, setAvailableDatabases] = useState<string[]>([]);
  const [metrics, setMetrics] = useState<QueryMetricItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [refreshCounter, setRefreshCounter] = useState(0);

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

  const loadServers = useCallback(async (tr: TimeRange) => {
    try {
      const params = new URLSearchParams({
        start: tr.from.toISOString(),
        end: tr.to.toISOString(),
        type: 'databases',
      });
      const response: Observable<FetchResponse<Array<{ label: string; value: string }>>> = getBackendSrv().fetch({
        url: `/api/plugins/guilhermearpassos-sqlsights-app/resources/datasource-options?${params.toString()}`,
      });
      const result = await lastValueFrom(response);
      setServers((result.data || []).map((s) => ({ label: s.label, value: s.value })));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load servers');
    }
  }, []);

  useEffect(() => {
    loadServers(timeRange);
  }, [timeRange, loadServers, refreshCounter]);

  useEffect(() => {
    if (!selectedServer && servers.length === 1) {
      updateParams((p) => p.set('server', servers[0].value));
    }
  }, [selectedServer, servers, updateParams]);

  const loadQueryMetrics = useCallback(async () => {
    if (!selectedServer) {
      setMetrics([]);
      setAvailableDatabases([]);
      return;
    }
    try {
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
      const response: Observable<FetchResponse<QueryMetricsResponse>> = getBackendSrv().fetch({
        url: `/api/plugins/guilhermearpassos-sqlsights-app/resources/queryMetrics?${params.toString()}`,
      });
      const result = await lastValueFrom(response);
      setMetrics(result.data?.metrics || []);
      setAvailableDatabases(result.data?.available_databases || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load query metrics');
      setMetrics([]);
    } finally {
      setLoading(false);
    }
  }, [selectedServer, timeRange, selectedDatabases]);

  useEffect(() => {
    loadQueryMetrics();
  }, [loadQueryMetrics, refreshCounter]);

  const filteredMetrics = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();
    return metrics.filter((m) => {
      if (onlyWithIssues && (m.issue_flags ?? []).length === 0) {
        return false;
      }
      if (selectedIssues.length > 0) {
        const flags = m.issue_flags ?? [];
        if (!selectedIssues.some((want) => flags.includes(want))) {
          return false;
        }
      }
      if (minExecs > 0 && m.execution_count < minExecs) {
        return false;
      }
      if (q) {
        const hay = `${m.text} ${m.database} ${m.query_hash}`.toLowerCase();
        if (!hay.includes(q)) {
          return false;
        }
      }
      return true;
    });
  }, [metrics, onlyWithIssues, selectedIssues, minExecs, searchQuery]);

  const kpis = useMemo(() => {
    let totalExecs = 0;
    let totalCpuMs = 0;
    let totalReads = 0;
    let spillQueries = 0;
    let flaggedQueries = 0;
    for (const m of filteredMetrics) {
      totalExecs += m.execution_count;
      totalCpuMs += m.total_worker_ms;
      totalReads += m.total_logical_reads;
      if (m.total_spills > 0) {
        spillQueries++;
      }
      if ((m.issue_flags ?? []).length > 0) {
        flaggedQueries++;
      }
    }
    return {
      totalQueries: filteredMetrics.length,
      totalExecs,
      totalCpuMs,
      totalReads,
      spillQueries,
      flaggedQueries,
    };
  }, [filteredMetrics]);

  const dbOptions = useMemo<Array<ComboboxOption<string>>>(() => {
    const merged = new Set([...availableDatabases, ...selectedDatabases]);
    return Array.from(merged)
      .sort()
      .map((db) => ({ label: db, value: db }));
  }, [availableDatabases, selectedDatabases]);

  const selectedMetric = useMemo(() => {
    if (!selectedQueryHash) {
      return null;
    }
    return metrics.find((m) => m.query_hash === selectedQueryHash) ?? null;
  }, [metrics, selectedQueryHash]);

  const handleTimeRangeChange = useCallback(
    (next: TimeRange) => {
      setTimeRange(next);
      updateParams((p) => applyTimeRangeToSearchParams(p, next));
    },
    [updateParams]
  );

  const handleAggModeChange = (nextMode: QueryMetricsAggregationMode) => {
    updateParams((p) => {
      setOptionalParam(p, 'agg', nextMode === 'total' ? null : nextMode);
      const currentSort = (p.get('sort') as QueryMetricsSortField) || defaultSortField;
      const mappedSort = mapSortFieldToAggregationMode(currentSort, nextMode);
      const nextDefault: QueryMetricsSortField = nextMode === 'avg' ? 'avg_elapsed_ms' : 'total_elapsed_ms';
      if (mappedSort === nextDefault && !p.get('sort')) {
        p.delete('sort');
      } else {
        p.set('sort', mappedSort);
      }
    });
  };

  const handleSortChange = (field: QueryMetricsSortField) => {
    updateParams((p) => {
      if (sortField === field) {
        p.set('dir', sortDir === 'desc' ? 'asc' : 'desc');
      } else {
        p.set('sort', field);
        p.set('dir', 'desc');
      }
    });
  };

  const handleOpenPlanAnalysis = (item: QueryMetricItem) => {
    navigate(
      `${prefixRoute(ROUTES.PlanAnalysis)}${buildPreservedNavSearch(searchParams, {
        q: item.query_hash,
        db: item.database || null,
      })}`
    );
  };

  const handleOpenSnapshots = () => {
    navigate(`${prefixRoute(ROUTES.Snapshots)}${buildPreservedNavSearch(searchParams)}`);
  };

  const hasActiveFilters =
    selectedDatabases.length > 0 ||
    selectedIssues.length > 0 ||
    onlyWithIssues ||
    minExecs > 0 ||
    searchQuery.length > 0;

  return (
    <PluginPage>
      <div className={styles.headerRow}>
        <div>
          <h2 style={{ margin: '0 0 4px 0' }}>Query Metrics &amp; Workload Stats</h2>
          <p style={{ margin: 0 }}>
            Inspect aggregated SQL Server query performance counters (CPU, duration, logical/physical reads, TempDB
            spills, memory grants) and pinpoint resource-heavy statements.
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
          <Button variant="secondary" fill="outline" onClick={() => setRefreshCounter((c) => c + 1)} disabled={loading}>
            <RefreshCw size={14} />
          </Button>
          <TimeRangePicker
            value={timeRange}
            onChange={handleTimeRangeChange}
            timeZone="browser"
            onChangeTimeZone={() => {}}
            onZoom={() => {}}
            onMoveBackward={() => {}}
            onMoveForward={() => {}}
          />
        </div>
      </div>

      {error && (
        <Alert title="Error" severity="error" style={{ marginBottom: 16 }}>
          <div>{error}</div>
          <Button size="sm" variant="secondary" onClick={() => setRefreshCounter((c) => c + 1)} style={{ marginTop: 8 }}>
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
                  p.delete('queryHash');
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
            <span className={styles.filterLabel}>Issues</span>
            <MultiCombobox<string>
              width={30}
              options={ISSUE_OPTIONS}
              value={selectedIssues}
              onChange={(opts) => updateParams((p) => setCSVParam(p, 'issues', opts.map((o) => o.value)))}
              placeholder="All queries"
            />
          </div>

          <div className={styles.filterField} style={{ minWidth: 110 }}>
            <span className={styles.filterLabel}>Min Execs</span>
            <Input
              type="number"
              width={14}
              min={0}
              value={minExecs || ''}
              onChange={(e) => {
                const v = Number(e.currentTarget.value);
                updateParams((p) => setOptionalParam(p, 'minExecs', v > 0 ? String(v) : null));
              }}
              placeholder="0"
            />
          </div>

          <div className={styles.filterField} style={{ flex: 1, minWidth: 220 }}>
            <span className={styles.filterLabel}>Search SQL / Query Hash</span>
            <Input
              prefix={<Search size={14} />}
              value={searchQuery}
              onChange={(e) => updateParams((p) => setOptionalParam(p, 'q', e.currentTarget.value))}
              placeholder="Filter by SQL text, table, query hash..."
            />
          </div>
        </div>

        <div className={styles.quickRow}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
            <label style={{ display: 'inline-flex', alignItems: 'center', gap: 8, cursor: 'pointer' }}>
              <Switch
                value={onlyWithIssues}
                onChange={(e) => updateParams((p) => setOptionalParam(p, 'onlyIssues', e.currentTarget.checked ? '1' : null))}
              />
              <span>Only queries with issues</span>
            </label>
          </div>

          {hasActiveFilters && (
            <Button
              size="sm"
              variant="secondary"
              fill="text"
              onClick={() =>
                updateParams((p) => {
                  p.delete('db');
                  p.delete('issues');
                  p.delete('onlyIssues');
                  p.delete('minExecs');
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
        <div className={styles.emptyPrompt}>
          Select a SQL Server instance above to explore query execution metrics.
        </div>
      ) : loading && metrics.length === 0 ? (
        <LoadingPlaceholder text="Loading query metrics..." />
      ) : (
        <>
          <div className={styles.kpiGrid}>
            <div className={styles.kpiCard}>
              <div className={styles.kpiLabel}>Distinct Queries</div>
              <div className={styles.kpiValue}>{kpis.totalQueries.toLocaleString()}</div>
            </div>
            <div className={styles.kpiCard}>
              <div className={styles.kpiLabel}>Total Executions</div>
              <div className={styles.kpiValue}>{kpis.totalExecs.toLocaleString()}</div>
            </div>
            <div className={styles.kpiCard}>
              <div className={styles.kpiLabel}>Total CPU Time</div>
              <div className={styles.kpiValue}>{formatDurationMs(kpis.totalCpuMs)}</div>
            </div>
            <div className={styles.kpiCard}>
              <div className={styles.kpiLabel}>Total Logical Reads</div>
              <div className={styles.kpiValue}>{kpis.totalReads.toLocaleString()}</div>
            </div>
            <div className={styles.kpiCard}>
              <div className={styles.kpiLabel}>Flagged / Spilling Queries</div>
              <div className={styles.kpiValue} style={{ color: kpis.spillQueries > 0 ? '#e02f44' : undefined }}>
                {kpis.flaggedQueries} / {kpis.spillQueries}
              </div>
            </div>
          </div>

          <QueryMetricsTable
            metrics={filteredMetrics}
            aggMode={aggMode}
            onAggModeChange={handleAggModeChange}
            sortField={sortField}
            sortDir={sortDir}
            onSortChange={handleSortChange}
            onSelectMetric={(item) => updateParams((p) => p.set('queryHash', item.query_hash))}
            onOpenPlanAnalysis={handleOpenPlanAnalysis}
            onOpenSnapshots={handleOpenSnapshots}
          />
        </>
      )}

      {selectedMetric && (
        <Drawer
          title="Query Metrics & Execution History"
          size="lg"
          onClose={() => updateParams((p) => p.delete('queryHash'))}
        >
          <div className={styles.drawerContent}>
            <QueryMetricDrawer
              metric={selectedMetric}
              server={selectedServer}
              timeRange={timeRange}
              onOpenPlanAnalysis={handleOpenPlanAnalysis}
              onOpenSnapshots={handleOpenSnapshots}
            />
          </div>
        </Drawer>
      )}
    </PluginPage>
  );
};

export default PageThree;
