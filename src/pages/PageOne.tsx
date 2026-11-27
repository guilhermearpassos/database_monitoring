import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { FetchResponse, getBackendSrv, PluginPage } from '@grafana/runtime';
import { lastValueFrom, Observable } from 'rxjs';
import {
  Alert,
  Button,
  Card,
  Combobox,
  ComboboxOption,
  Drawer,
  LoadingPlaceholder,
  RadioButtonGroup,
  Switch,
  TimeRangePicker,
  useStyles2,
} from '@grafana/ui';
import {
  DataFrame,
  EventBusSrv,
  FieldType,
  GrafanaTheme2,
  LoadingState,
  MutableDataFrame,
  TimeRange,
} from '@grafana/data';
import { css } from '@emotion/css';
import { BarChart3, FileSearch, RefreshCw, ShieldAlert, X } from 'lucide-react';
import { MyGraph } from './graph';
import { QueryDetailsComponent } from '../components/QueryDetails/Details';
import { SnapshotSummaryItem } from '../components/Snapshots/types';
import { SnapshotSortField, SnapshotsTable, SortDirection } from '../components/Snapshots/SnapshotsTable';
import {
  applyTimeRangeToSearchParams,
  buildPreservedNavSearch,
  parseTimeRangeFromSearchParams,
  setOptionalParam,
} from '../utils/urlState';
import { ROUTES } from '../constants';
import { prefixRoute } from '../utils/utils.routing';

const AUTO_REFRESH_OPTIONS: Array<ComboboxOption<number>> = [
  { label: 'Auto-refresh: Off', value: 0 },
  { label: 'Every 10s', value: 10000 },
  { label: 'Every 30s', value: 30000 },
  { label: 'Every 1m', value: 60000 },
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
    padding: ${theme.spacing(1.5, 2)};
    margin-bottom: ${theme.spacing(2)};
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: ${theme.spacing(1.5)};
  `,
  filterGroup: css`
    display: flex;
    align-items: flex-end;
    flex-wrap: wrap;
    gap: ${theme.spacing(1.5)};
  `,
  filterField: css`
    display: flex;
    flex-direction: column;
    gap: 4px;
  `,
  filterLabel: css`
    font-size: ${theme.typography.bodySmall.fontSize};
    color: ${theme.colors.text.secondary};
    font-weight: ${theme.typography.fontWeightMedium};
  `,
  section: css`
    margin-bottom: ${theme.spacing(2.5)};
  `,
  chartContainer: css`
    height: 40vh;
    min-height: 280px;
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

function buildActivityDataFrame(
  summaries: SnapshotSummaryItem[],
  timeRange: TimeRange,
  metric: 'connections' | 'wait_time_ms'
): DataFrame[] {
  if (summaries.length === 0) {
    return [];
  }
  const fromMs = timeRange.from.valueOf();
  const toMs = timeRange.to.valueOf();
  const intervalMs = 60_000;
  const points = Math.max(1, Math.floor((toMs - fromMs) / intervalMs) + 1);

  const times: number[] = [];
  for (let i = 0; i < points; i++) {
    times.push(fromMs + i * intervalMs);
  }

  const seriesByWait = new Map<string, number[]>();
  for (const sum of summaries) {
    const tsMs = new Date(sum.timestamp).getTime();
    const idx = Math.floor((tsMs - fromMs) / intervalMs);
    if (idx < 0 || idx >= points) {
      continue;
    }
    const source = metric === 'wait_time_ms' ? sum.time_ms_by_wait_event : sum.connections_by_wait_event;
    for (const [rawWait, val] of Object.entries(source || {})) {
      const key = rawWait === '' ? 'cpu' : rawWait;
      let arr = seriesByWait.get(key);
      if (!arr) {
        arr = new Array<number>(points).fill(0);
        seriesByWait.set(key, arr);
      }
      arr[idx] = Math.max(arr[idx], val);
    }
  }

  const frame = new MutableDataFrame({
    name: metric === 'wait_time_ms' ? 'Wait Time (ms) by Wait Event' : 'Active Connections by Wait Event',
    fields: [{ name: 'time', type: FieldType.time, values: times }],
  });

  for (const [waitKey, values] of seriesByWait.entries()) {
    frame.addField({ name: waitKey, type: FieldType.number, values });
  }

  return [frame];
}

const PageOne = () => {
  const styles = useStyles2(getStyles);
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const panelEventBus = useMemo(() => new EventBusSrv(), []);

  // URL-derived state
  const selectedServer = searchParams.get('server') || '';
  const expandedSnapId = searchParams.get('snapId') || null;
  const drawerSampleId = searchParams.get('sampleId') || '';
  const onlyBlocking = searchParams.get('onlyBlocking') === '1';
  const waitFilter = searchParams.get('wait') || '';
  const chartMetric: 'connections' | 'wait_time_ms' =
    searchParams.get('metric') === 'wait_time_ms' ? 'wait_time_ms' : 'connections';
  const sortField = (searchParams.get('sort') as SnapshotSortField) || 'timestamp';
  const sortDir: SortDirection = searchParams.get('dir') === 'asc' ? 'asc' : 'desc';

  const [timeRange, setTimeRange] = useState<TimeRange>(() => parseTimeRangeFromSearchParams(searchParams));
  const [autoRefreshMs, setAutoRefreshMs] = useState<number>(0);
  const [refreshCounter, setRefreshCounter] = useState(0);

  const [servers, setServers] = useState<Array<ComboboxOption<string>>>([]);
  const [summaries, setSummaries] = useState<SnapshotSummaryItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

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

  // Auto-refresh interval
  useEffect(() => {
    if (!autoRefreshMs || !selectedServer) {
      return;
    }
    const id = window.setInterval(() => {
      setRefreshCounter((c) => c + 1);
    }, autoRefreshMs);
    return () => window.clearInterval(id);
  }, [autoRefreshMs, selectedServer]);

  // Load servers
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

  // Auto-select single server if none chosen
  useEffect(() => {
    if (!selectedServer && servers.length === 1) {
      updateParams((p) => p.set('server', servers[0].value));
    }
  }, [selectedServer, servers, updateParams]);

  // Load snapshot summaries (used for both chart and table -> single round trip!)
  const loadSummaries = useCallback(async () => {
    if (!selectedServer) {
      setSummaries([]);
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
      const response: Observable<FetchResponse<SnapshotSummaryItem[]>> = getBackendSrv().fetch({
        url: `/api/plugins/guilhermearpassos-sqlsights-app/resources/snapshotSummaries?${params.toString()}`,
      });
      const result = await lastValueFrom(response);
      setSummaries(result.data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load snapshots');
      setSummaries([]);
    } finally {
      setLoading(false);
    }
  }, [selectedServer, timeRange]);

  useEffect(() => {
    loadSummaries();
  }, [loadSummaries, refreshCounter]);

  const waitEventOptions = useMemo<Array<ComboboxOption<string>>>(() => {
    const set = new Set<string>();
    for (const s of summaries) {
      for (const rawKey of Object.keys(s.connections_by_wait_event || {})) {
        set.add(rawKey === '' ? 'cpu' : rawKey);
      }
    }
    return Array.from(set)
      .sort()
      .map((w) => ({ label: w, value: w }));
  }, [summaries]);

  const filteredSummaries = useMemo(() => {
    return summaries.filter((s) => {
      if (onlyBlocking && s.blockers <= 0 && s.waiters <= 0) {
        return false;
      }
      if (waitFilter) {
        const key = waitFilter === 'cpu' ? '' : waitFilter;
        const hasWait =
          (s.connections_by_wait_event && (s.connections_by_wait_event[key] ?? s.connections_by_wait_event[waitFilter] ?? 0) > 0);
        if (!hasWait) {
          return false;
        }
      }
      return true;
    });
  }, [summaries, onlyBlocking, waitFilter]);

  const chartFrames = useMemo(
    () => buildActivityDataFrame(filteredSummaries, timeRange, chartMetric),
    [filteredSummaries, timeRange, chartMetric]
  );

  const handleTimeRangeChange = useCallback(
    (next: TimeRange) => {
      setTimeRange(next);
      updateParams((p) => applyTimeRangeToSearchParams(p, next));
    },
    [updateParams]
  );

  const handleSortChange = (field: SnapshotSortField) => {
    updateParams((p) => {
      if (sortField === field) {
        p.set('dir', sortDir === 'desc' ? 'asc' : 'desc');
      } else {
        p.set('sort', field);
        p.set('dir', 'desc');
      }
    });
  };

  const hasFilters = onlyBlocking || Boolean(waitFilter);

  return (
    <PluginPage>
      <div className={styles.headerRow}>
        <div>
          <h2 style={{ margin: '0 0 4px 0' }}>SQL Database Monitoring</h2>
          <p style={{ margin: 0 }}>
            Monitor active sessions, wait events, blocking chains, and execution plans across your SQL Server instances.
          </p>
        </div>

        <div className={styles.headerRight}>
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
          <Button
            variant="secondary"
            fill="outline"
            onClick={() => navigate(`${prefixRoute(ROUTES.Deadlocks)}${buildPreservedNavSearch(searchParams)}`)}
          >
            <ShieldAlert size={14} />
            <span style={{ marginLeft: 6 }}>Deadlocks</span>
          </Button>
          <Combobox<number>
            width={20}
            options={AUTO_REFRESH_OPTIONS}
            value={autoRefreshMs}
            onChange={(opt) => setAutoRefreshMs(opt?.value ?? 0)}
          />
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
        <Alert title="Error" severity="error" className={styles.section}>
          <div>{error}</div>
          <Button size="sm" variant="secondary" onClick={() => setRefreshCounter((c) => c + 1)} style={{ marginTop: 8 }}>
            Retry
          </Button>
        </Alert>
      )}

      <div className={styles.filterCard}>
        <div className={styles.filterGroup}>
          <div className={styles.filterField}>
            <span className={styles.filterLabel}>SQL Server Instance</span>
            <Combobox<string>
              width={28}
              options={servers}
              value={selectedServer || null}
              onChange={(opt) =>
                updateParams((p) => {
                  setOptionalParam(p, 'server', opt?.value);
                  p.delete('snapId');
                  p.delete('sampleId');
                })
              }
              placeholder="Select a server..."
              isClearable
            />
          </div>

          <div className={styles.filterField}>
            <span className={styles.filterLabel}>Wait Event</span>
            <Combobox<string>
              width={26}
              options={waitEventOptions}
              value={waitFilter || null}
              onChange={(opt) => updateParams((p) => setOptionalParam(p, 'wait', opt?.value))}
              placeholder="All wait events"
              disabled={!selectedServer}
              isClearable
            />
          </div>

          <label style={{ display: 'inline-flex', alignItems: 'center', gap: 8, cursor: 'pointer', paddingBottom: 6 }}>
            <Switch
              value={onlyBlocking}
              onChange={(e) =>
                updateParams((p) => setOptionalParam(p, 'onlyBlocking', e.currentTarget.checked ? '1' : null))
              }
            />
            <span>Only snapshots with blockers / waiters</span>
          </label>
        </div>

        {hasFilters && (
          <Button
            size="sm"
            variant="secondary"
            fill="text"
            onClick={() =>
              updateParams((p) => {
                p.delete('onlyBlocking');
                p.delete('wait');
              })
            }
          >
            <X size={13} />
            <span style={{ marginLeft: 4 }}>Clear Filters</span>
          </Button>
        )}
      </div>

      {!selectedServer ? (
        <div className={styles.emptyPrompt}>
          Select a SQL Server instance above to view database activity and snapshots.
        </div>
      ) : (
        <>
          <div className={styles.section}>
            <Card>
              <Card.Heading>Database Activity Over Time</Card.Heading>
              <Card.Actions>
                <RadioButtonGroup<'connections' | 'wait_time_ms'>
                  size="sm"
                  options={[
                    { label: 'Active Connections', value: 'connections' },
                    { label: 'Wait Time (ms)', value: 'wait_time_ms' },
                  ]}
                  value={chartMetric}
                  onChange={(val) => updateParams((p) => setOptionalParam(p, 'metric', val === 'connections' ? null : val))}
                />
              </Card.Actions>
              <Card.Description>
                <div className={styles.chartContainer}>
                  {loading && summaries.length === 0 ? (
                    <LoadingPlaceholder text="Loading activity chart..." />
                  ) : chartFrames.length > 0 ? (
                    <MyGraph
                      data={chartFrames}
                      loadingState={loading ? LoadingState.Loading : LoadingState.Done}
                      eventBus={panelEventBus}
                      timeRange={timeRange}
                      onTimeRangeChange={handleTimeRangeChange}
                    />
                  ) : (
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                      <p>No activity data available for the selected range.</p>
                    </div>
                  )}
                </div>
              </Card.Description>
            </Card>
          </div>

          <div className={styles.section}>
            {loading && summaries.length === 0 ? (
              <LoadingPlaceholder text="Loading snapshots..." />
            ) : (
              <SnapshotsTable
                summaries={filteredSummaries}
                sortField={sortField}
                sortDir={sortDir}
                onSortChange={handleSortChange}
                expandedSnapId={expandedSnapId}
                onToggleExpand={(snapId) =>
                  updateParams((p) => {
                    if (expandedSnapId === snapId) {
                      p.delete('snapId');
                    } else {
                      p.set('snapId', snapId);
                    }
                  })
                }
                onSelectSample={(snapId, sampleId) =>
                  updateParams((p) => {
                    p.set('snapId', snapId);
                    p.set('sampleId', sampleId);
                  })
                }
                timeRange={timeRange}
              />
            )}
          </div>
        </>
      )}

      {expandedSnapId && drawerSampleId && (
        <Drawer
          title="Query Sample Details"
          size="lg"
          onClose={() =>
            updateParams((p) => {
              p.delete('sampleId');
            })
          }
        >
          <div className={styles.drawerContent}>
            <QueryDetailsComponent
              snapID={expandedSnapId}
              sampleID={drawerSampleId}
              timeRange={timeRange}
              server={selectedServer}
              initialTab="overview"
              onSelectSample={(snapId, sampleId) =>
                updateParams((p) => {
                  p.set('snapId', snapId);
                  p.set('sampleId', sampleId);
                })
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

export default PageOne;
