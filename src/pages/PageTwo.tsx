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
import { Activity, BarChart3, RefreshCw, Search, X } from 'lucide-react';
import { QueryDetailsComponent } from '../components/QueryDetails/Details';
import { PlanAnalysisItem, PlanAnalysisResponse } from '../components/PlanAnalysis/types';
import { PlanAnalysisTable, PlanSortField, SortDirection } from '../components/PlanAnalysis/PlanAnalysisTable';
import {
  applyTimeRangeToSearchParams,
  buildPreservedNavSearch,
  parseCSVParam,
  parseTimeRangeFromSearchParams,
  setCSVParam,
  setOptionalParam,
} from '../utils/urlState';
import { ROUTES } from '../constants';
import { prefixRoute } from '../utils/utils.routing';

const ISSUE_TYPE_OPTIONS: Array<ComboboxOption<string>> = [
  { label: 'Implicit Conversion (Seek Plan)', value: 'seek_plan_convert' },
  { label: 'Missing Index', value: 'missing_index' },
  { label: 'Implicit Conversion (Cardinality)', value: 'cardinality_convert' },
  { label: 'Implicit Conversion (Any)', value: 'implicit_conversion' },
  { label: 'Large Table Scan', value: 'large_table_scan' },
];

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
    min-width: 180px;
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

const PageTwo = () => {
  const styles = useStyles2(getStyles);
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  // URL-derived state
  const selectedServer = searchParams.get('server') || '';
  const selectedDatabases = useMemo(() => parseCSVParam(searchParams, 'db'), [searchParams]);
  const selectedIssues = useMemo(() => parseCSVParam(searchParams, 'issues'), [searchParams]);
  const onlyWithLocks = searchParams.get('onlyLocks') === '1';
  const searchQuery = searchParams.get('q') || '';
  const sortField = (searchParams.get('sort') as PlanSortField) || 'related_lock_count';
  const sortDir: SortDirection = searchParams.get('dir') === 'asc' ? 'asc' : 'desc';
  const drawerSnapId = searchParams.get('snapId') || '';
  const drawerSampleId = searchParams.get('sampleId') || '';

  const [timeRange, setTimeRange] = useState<TimeRange>(() => parseTimeRangeFromSearchParams(searchParams));
  const [servers, setServers] = useState<Array<ComboboxOption<string>>>([]);
  const [availableDatabases, setAvailableDatabases] = useState<string[]>([]);
  const [plans, setPlans] = useState<PlanAnalysisItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [refreshCounter, setRefreshCounter] = useState(0);

  const updateParams = useCallback(
    (mutator: (next: URLSearchParams) => void) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          mutator(next);
          return next;
        },
        { replace: true }
      );
    },
    [setSearchParams]
  );

  // Load servers for time range
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
      const opts = (result.data || []).map((s) => ({ label: s.label, value: s.value }));
      setServers(opts);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load servers');
    }
  }, []);

  useEffect(() => {
    loadServers(timeRange);
  }, [timeRange, loadServers, refreshCounter]);

  // Auto-select single server if none is chosen
  useEffect(() => {
    if (!selectedServer && servers.length === 1) {
      updateParams((p) => p.set('server', servers[0].value));
    }
  }, [selectedServer, servers, updateParams]);

  // Load plan analysis items
  const loadPlanAnalysis = useCallback(async () => {
    if (!selectedServer) {
      setPlans([]);
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
      if (selectedIssues.length > 0) {
        params.set('issueTags', selectedIssues.join(','));
      }
      const response: Observable<FetchResponse<PlanAnalysisResponse>> = getBackendSrv().fetch({
        url: `/api/plugins/guilhermearpassos-sqlsights-app/resources/planAnalysis?${params.toString()}`,
      });
      const result = await lastValueFrom(response);
      setPlans(result.data?.plans || []);
      setAvailableDatabases(result.data?.available_databases || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load execution plans');
      setPlans([]);
    } finally {
      setLoading(false);
    }
  }, [selectedServer, timeRange, selectedDatabases, selectedIssues]);

  useEffect(() => {
    loadPlanAnalysis();
  }, [loadPlanAnalysis, refreshCounter]);

  // Client-side refinements (locks toggle + search across SQL / convert expression / missing index table)
  const filteredPlans = useMemo(() => {
    const q = searchQuery.trim().toLowerCase();
    return plans.filter((item) => {
      if (onlyWithLocks && item.related_lock_count <= 0) {
        return false;
      }
      if (q) {
        const haystacks = [
          item.text,
          item.database,
          item.query_hash,
          item.trace_id || '',
          ...(item.converts ?? []).map((c) => `${c.convert_issue} ${c.expression}`),
          ...(item.missing_indexes ?? []).map((m) => `${m.schema}.${m.table} ${(m.equality_columns ?? []).join(' ')}`),
        ];
        if (!haystacks.some((h) => h.toLowerCase().includes(q))) {
          return false;
        }
      }
      return true;
    });
  }, [plans, onlyWithLocks, searchQuery]);

  // KPI summary over currently filtered plans
  const kpis = useMemo(() => {
    let seekPlanCount = 0;
    let missingIndexCount = 0;
    let totalLocks = 0;
    for (const p of filteredPlans) {
      if ((p.issue_tags ?? []).includes('seek_plan_convert')) {
        seekPlanCount++;
      }
      if ((p.issue_tags ?? []).includes('missing_index')) {
        missingIndexCount++;
      }
      totalLocks += p.related_lock_count || 0;
    }
    return {
      total: filteredPlans.length,
      seekPlanCount,
      missingIndexCount,
      totalLocks,
    };
  }, [filteredPlans]);

  const dbOptions = useMemo<Array<ComboboxOption<string>>>(() => {
    const merged = new Set([...availableDatabases, ...selectedDatabases]);
    return Array.from(merged)
      .sort()
      .map((db) => ({ label: db, value: db }));
  }, [availableDatabases, selectedDatabases]);

  const handleTimeRangeChange = (next: TimeRange) => {
    setTimeRange(next);
    updateParams((p) => applyTimeRangeToSearchParams(p, next));
  };

  const handleSortChange = (field: PlanSortField) => {
    updateParams((p) => {
      if (sortField === field) {
        p.set('dir', sortDir === 'desc' ? 'asc' : 'desc');
      } else {
        p.set('sort', field);
        p.set('dir', 'desc');
      }
    });
  };

  const hasActiveFilters =
    selectedDatabases.length > 0 || selectedIssues.length > 0 || onlyWithLocks || searchQuery.length > 0;

  return (
    <PluginPage>
      <div className={styles.headerRow}>
        <div>
          <h2 style={{ margin: '0 0 4px 0' }}>Plan Analysis Explorer</h2>
          <p style={{ margin: 0 }}>
            Filter problematic execution plans by database, Seek Plan implicit conversions, missing indexes, and lock impact.
          </p>
        </div>

        <div className={styles.headerRight}>
          <Button
            variant="secondary"
            fill="outline"
            onClick={() => navigate(`${prefixRoute(ROUTES.Snapshots)}${buildPreservedNavSearch(searchParams)}`)}
          >
            <Activity size={14} />
            <span style={{ marginLeft: 6 }}>Snapshots View</span>
          </Button>
          <Button
            variant="secondary"
            fill="outline"
            onClick={() => navigate(`${prefixRoute(ROUTES.QueryMetrics)}${buildPreservedNavSearch(searchParams)}`)}
          >
            <BarChart3 size={14} />
            <span style={{ marginLeft: 6 }}>Query Metrics</span>
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
              width={28}
              options={servers}
              value={selectedServer || null}
              onChange={(opt) =>
                updateParams((p) => {
                  setOptionalParam(p, 'server', opt?.value);
                  p.delete('db');
                })
              }
              placeholder="Select a server..."
              isClearable
            />
          </div>

          <div className={styles.filterField}>
            <span className={styles.filterLabel}>Database(s)</span>
            <MultiCombobox<string>
              width={32}
              options={dbOptions}
              value={selectedDatabases}
              onChange={(opts) => updateParams((p) => setCSVParam(p, 'db', opts.map((o) => o.value)))}
              placeholder={selectedServer ? 'All databases' : 'Select server first'}
              disabled={!selectedServer}
            />
          </div>

          <div className={styles.filterField}>
            <span className={styles.filterLabel}>Issue Types</span>
            <MultiCombobox<string>
              width={36}
              options={ISSUE_TYPE_OPTIONS}
              value={selectedIssues}
              onChange={(opts) => updateParams((p) => setCSVParam(p, 'issues', opts.map((o) => o.value)))}
              placeholder="All issue types"
            />
          </div>

          <div className={styles.filterField} style={{ flex: 1, minWidth: 220 }}>
            <span className={styles.filterLabel}>Search SQL / Expression / Table</span>
            <Input
              prefix={<Search size={14} />}
              value={searchQuery}
              onChange={(e) => updateParams((p) => setOptionalParam(p, 'q', e.currentTarget.value))}
              placeholder="Filter by SQL text, CONVERT_IMPLICIT, table..."
            />
          </div>
        </div>

        <div className={styles.quickRow}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
            <label style={{ display: 'inline-flex', alignItems: 'center', gap: 8, cursor: 'pointer' }}>
              <Switch
                value={onlyWithLocks}
                onChange={(e) => updateParams((p) => setOptionalParam(p, 'onlyLocks', e.currentTarget.checked ? '1' : null))}
              />
              <span>Only plans with related locks</span>
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
                  p.delete('onlyLocks');
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
          Select a SQL Server instance above to analyze execution plans.
        </div>
      ) : loading ? (
        <LoadingPlaceholder text="Analyzing execution plans..." />
      ) : (
        <>
          <div className={styles.kpiGrid}>
            <div className={styles.kpiCard}>
              <div className={styles.kpiLabel}>Problematic Plans</div>
              <div className={styles.kpiValue}>{kpis.total}</div>
            </div>
            <div className={styles.kpiCard}>
              <div className={styles.kpiLabel}>Seek Plan Conversions</div>
              <div className={styles.kpiValue} style={{ color: kpis.seekPlanCount > 0 ? '#e02f44' : undefined }}>
                {kpis.seekPlanCount}
              </div>
            </div>
            <div className={styles.kpiCard}>
              <div className={styles.kpiLabel}>Missing Indexes</div>
              <div className={styles.kpiValue} style={{ color: kpis.missingIndexCount > 0 ? '#ff9830' : undefined }}>
                {kpis.missingIndexCount}
              </div>
            </div>
            <div className={styles.kpiCard}>
              <div className={styles.kpiLabel}>Related Lock Events</div>
              <div className={styles.kpiValue} style={{ color: kpis.totalLocks > 0 ? '#e02f44' : undefined }}>
                {kpis.totalLocks}
              </div>
            </div>
          </div>

          <PlanAnalysisTable
            plans={filteredPlans}
            sortField={sortField}
            sortDir={sortDir}
            onSortChange={handleSortChange}
            onSelectPlan={(item) =>
              updateParams((p) => {
                p.set('snapId', item.snap_id);
                p.set('sampleId', item.sample_id);
              })
            }
            onOpenSnapshot={(snapId, sampleId) =>
              navigate(
                `${prefixRoute(ROUTES.Snapshots)}${buildPreservedNavSearch(searchParams, {
                  snapId,
                  sampleId,
                })}`
              )
            }
            onOpenQueryMetrics={(queryHash) =>
              navigate(`${prefixRoute(ROUTES.QueryMetrics)}${buildPreservedNavSearch(searchParams, { queryHash })}`)
            }
            timeRange={timeRange}
          />
        </>
      )}

      {drawerSnapId && drawerSampleId && (
        <Drawer
          title="Query & Execution Plan Details"
          size="lg"
          onClose={() =>
            updateParams((p) => {
              p.delete('snapId');
              p.delete('sampleId');
            })
          }
        >
          <div className={styles.drawerContent}>
            <QueryDetailsComponent
              snapID={drawerSnapId}
              sampleID={drawerSampleId}
              timeRange={timeRange}
              server={selectedServer}
              initialTab="plan"
              onSelectSample={(snapId, sampleId) =>
                updateParams((p) => {
                  p.set('snapId', snapId);
                  p.set('sampleId', sampleId);
                })
              }
              onOpenSnapshot={(snapId) =>
                navigate(`${prefixRoute(ROUTES.Snapshots)}${buildPreservedNavSearch(searchParams, { snapId })}`)
              }
              onOpenQueryMetrics={(queryHash) =>
                navigate(`${prefixRoute(ROUTES.QueryMetrics)}${buildPreservedNavSearch(searchParams, { queryHash })}`)
              }
            />
          </div>
        </Drawer>
      )}
    </PluginPage>
  );
};

export default PageTwo;
