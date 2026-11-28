import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2, TimeRange } from '@grafana/data';
import { FetchResponse, getBackendSrv } from '@grafana/runtime';
import {
  Alert,
  Button,
  Combobox,
  ComboboxOption,
  Input,
  LoadingPlaceholder,
  Switch,
  Tab,
  TabsBar,
  Tooltip,
  useStyles2,
} from '@grafana/ui';
import { lastValueFrom, Observable } from 'rxjs';
import { ArrowDown, ArrowUp, ArrowUpDown, ChevronDown, ChevronRight, Clock, Lock, Search } from 'lucide-react';
import { SnapshotSampleItem, SnapshotSummaryItem } from './types';
import { buildBlockingTrees, SnapshotBlockingTree } from './SnapshotBlockingTree';
import { formatDurationMs } from '../../utils/sqlHelpers';
import { getWaitEventColor } from '../../utils/wait_event_colors';
import { TempoTraceLink } from '../TempoTraceLink';
import { extractErrorMessage } from '../../utils/urlState';

export type SnapshotSortField = 'timestamp' | 'connections' | 'blockers' | 'waiters' | 'wait_duration_ms';
export type SortDirection = 'asc' | 'desc';

const getStyles = (theme: GrafanaTheme2) => ({
  wrapper: css`
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    background: ${theme.colors.background.primary};
    overflow: hidden;
  `,
  toolbar: css`
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: ${theme.spacing(1)};
    padding: ${theme.spacing(1.25, 1.5)};
    border-bottom: 1px solid ${theme.colors.border.weak};
    background: ${theme.colors.background.secondary};
    font-size: ${theme.typography.bodySmall.fontSize};
    color: ${theme.colors.text.secondary};
  `,
  tableScroll: css`
    width: 100%;
    overflow-x: auto;
  `,
  table: css`
    width: 100%;
    border-collapse: collapse;
    font-size: ${theme.typography.bodySmall.fontSize};

    th {
      text-align: left;
      padding: ${theme.spacing(1, 1.25)};
      background: ${theme.colors.background.secondary};
      color: ${theme.colors.text.secondary};
      font-weight: ${theme.typography.fontWeightMedium};
      border-bottom: 1px solid ${theme.colors.border.medium};
      white-space: nowrap;
      user-select: none;
    }

    td {
      padding: ${theme.spacing(1, 1.25)};
      border-bottom: 1px solid ${theme.colors.border.weak};
      color: ${theme.colors.text.primary};
      vertical-align: middle;
    }
  `,
  summaryRow: css`
    cursor: pointer;
    transition: background-color 0.12s ease;

    &:hover {
      background: ${theme.colors.action.hover};
    }
  `,
  summaryRowExpanded: css`
    background: ${theme.colors.background.secondary};
  `,
  detailRowCell: css`
    padding: 0 !important;
    background: ${theme.colors.background.canvas};
  `,
  detailPanel: css`
    border-left: 3px solid ${theme.colors.primary.border};
    padding: ${theme.spacing(1.5, 2)};
  `,
  sampleFilterBar: css`
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: ${theme.spacing(1.5)};
    margin: ${theme.spacing(1.25, 0)};
  `,
  samplesTable: css`
    width: 100%;
    border-collapse: collapse;
    font-size: ${theme.typography.bodySmall.fontSize};
    background: ${theme.colors.background.primary};
    border: 1px solid ${theme.colors.border.weak};

    th {
      text-align: left;
      padding: ${theme.spacing(0.75, 1)};
      background: ${theme.colors.background.secondary};
      color: ${theme.colors.text.secondary};
      border-bottom: 1px solid ${theme.colors.border.medium};
      white-space: nowrap;
    }

    td {
      padding: ${theme.spacing(0.75, 1)};
      border-bottom: 1px solid ${theme.colors.border.weak};
      vertical-align: middle;
    }

    tbody tr {
      cursor: pointer;
      &:hover {
        background: ${theme.colors.action.hover};
      }
    }
  `,
  sortBtn: css`
    display: inline-flex;
    align-items: center;
    gap: 4px;
    background: none;
    border: none;
    color: inherit;
    font: inherit;
    cursor: pointer;
    padding: 0;

    &:hover {
      color: ${theme.colors.text.primary};
    }
  `,
  waitBar: css`
    display: flex;
    height: 16px;
    min-width: 160px;
    max-width: 280px;
    border-radius: 3px;
    overflow: hidden;
    background: ${theme.colors.background.secondary};
    border: 1px solid ${theme.colors.border.weak};
  `,
  blockerPill: css`
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 6px;
    border-radius: ${theme.shape.radius.default};
    background: ${theme.colors.error.transparent};
    color: ${theme.colors.error.text};
    border: 1px solid ${theme.colors.error.border};
    font-weight: ${theme.typography.fontWeightMedium};
    white-space: nowrap;
  `,
  waiterPill: css`
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 6px;
    border-radius: ${theme.shape.radius.default};
    background: ${theme.colors.warning.transparent};
    color: ${theme.colors.warning.text};
    border: 1px solid ${theme.colors.warning.border};
    font-weight: ${theme.typography.fontWeightMedium};
    white-space: nowrap;
  `,
  sqlCell: css`
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: 12px;
    max-width: 420px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  `,
  mono: css`
    font-family: ${theme.typography.fontFamilyMonospace};
    white-space: nowrap;
  `,
  pagination: css`
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: ${theme.spacing(1)};
    padding: ${theme.spacing(1.25, 1.5)};
    border-top: 1px solid ${theme.colors.border.weak};
    background: ${theme.colors.background.secondary};
    font-size: ${theme.typography.bodySmall.fontSize};
  `,
  emptyState: css`
    padding: ${theme.spacing(5)};
    text-align: center;
    color: ${theme.colors.text.secondary};
  `,
});

const WaitBreakdownBar: React.FC<{ counts: Record<string, number> }> = ({ counts }) => {
  const styles = useStyles2(getStyles);
  const entries = Object.entries(counts || {}).sort((a, b) => b[1] - a[1]);
  const total = entries.reduce((sum, [, c]) => sum + c, 0);

  if (total === 0) {
    return <span style={{ opacity: 0.6 }}>-</span>;
  }

  return (
    <div className={styles.waitBar}>
      {entries.map(([rawWait, count]) => {
        const pct = (count / total) * 100;
        if (pct < 0.5) {
          return null;
        }
        const label = rawWait === '' ? 'cpu' : rawWait;
        const color = getWaitEventColor(label);
        return (
          <Tooltip key={label} content={`${label}: ${count} (${pct.toFixed(1)}%)`}>
            <div style={{ width: `${pct}%`, backgroundColor: `rgba(${color})`, height: '100%' }} />
          </Tooltip>
        );
      })}
    </div>
  );
};

const ExpandedSnapshotPanel: React.FC<{
  snapId: string;
  samples?: SnapshotSampleItem[];
  loading: boolean;
  error?: string;
  onRetry: () => void;
  timeRange: TimeRange;
  onSelectSample: (snapId: string, sampleId: string) => void;
}> = ({ samples, loading, error, onRetry, timeRange, onSelectSample }) => {
  const styles = useStyles2(getStyles);
  const [activeTab, setActiveTab] = useState<'samples' | 'blocking'>('samples');
  const [dbFilter, setDbFilter] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [onlyLocking, setOnlyLocking] = useState(false);
  const [textFilter, setTextFilter] = useState('');

  const dbOptions = useMemo<Array<ComboboxOption<string>>>(() => {
    const set = new Set<string>();
    (samples ?? []).forEach((s) => s.database && set.add(s.database));
    return Array.from(set)
      .sort()
      .map((d) => ({ label: d, value: d }));
  }, [samples]);

  const statusOptions = useMemo<Array<ComboboxOption<string>>>(() => {
    const set = new Set<string>();
    (samples ?? []).forEach((s) => s.status && set.add(s.status));
    return Array.from(set)
      .sort()
      .map((s) => ({ label: s, value: s }));
  }, [samples]);

  const blockingChainsCount = useMemo(() => buildBlockingTrees(samples ?? []).length, [samples]);

  const filteredSamples = useMemo(() => {
    const q = textFilter.trim().toLowerCase();
    return (samples ?? [])
      .filter((s) => {
        if (dbFilter && s.database !== dbFilter) {
          return false;
        }
        if (statusFilter && s.status !== statusFilter) {
          return false;
        }
        if (onlyLocking && !s.is_blocked && !s.is_blocker) {
          return false;
        }
        if (q) {
          const hay = `${s.text} ${s.session_id} ${s.user} ${s.wait_event} ${s.database} ${s.trace_id || ''} ${s.program_name} ${s.client_ip || ''}`.toLowerCase();
          if (!hay.includes(q)) {
            return false;
          }
        }
        return true;
      })
      .sort((a, b) => {
        // Surface blockers first, then blocked, then longest elapsed
        const rankA = a.is_blocker ? 2 : a.is_blocked ? 1 : 0;
        const rankB = b.is_blocker ? 2 : b.is_blocked ? 1 : 0;
        if (rankA !== rankB) {
          return rankB - rankA;
        }
        return b.elapsed_ms - a.elapsed_ms;
      });
  }, [samples, dbFilter, statusFilter, onlyLocking, textFilter]);

  if (loading) {
    return (
      <div className={styles.detailPanel}>
        <LoadingPlaceholder text="Loading snapshot samples..." />
      </div>
    );
  }

  if (error) {
    return (
      <div className={styles.detailPanel}>
        <Alert title="Failed to load snapshot samples" severity="error">
          <div>{error}</div>
          <Button size="sm" variant="secondary" onClick={onRetry} style={{ marginTop: 8 }}>
            Retry
          </Button>
        </Alert>
      </div>
    );
  }

  return (
    <div className={styles.detailPanel}>
      <TabsBar>
        <Tab
          label="Samples"
          counter={(samples ?? []).length}
          active={activeTab === 'samples'}
          onChangeTab={() => setActiveTab('samples')}
        />
        <Tab
          label="Blocking Chains"
          counter={blockingChainsCount}
          active={activeTab === 'blocking'}
          onChangeTab={() => setActiveTab('blocking')}
        />
      </TabsBar>

      {activeTab === 'blocking' ? (
        <SnapshotBlockingTree samples={samples ?? []} timeRange={timeRange} onSelectSample={onSelectSample} />
      ) : (
        <>
          <div className={styles.sampleFilterBar}>
            <Combobox<string>
              width={22}
              options={dbOptions}
              value={dbFilter || null}
              onChange={(o) => setDbFilter(o?.value || '')}
              placeholder="All databases"
              isClearable
            />
            <Combobox<string>
              width={18}
              options={statusOptions}
              value={statusFilter || null}
              onChange={(o) => setStatusFilter(o?.value || '')}
              placeholder="All statuses"
              isClearable
            />
            <div style={{ minWidth: 220, flex: 1 }}>
              <Input
                prefix={<Search size={13} />}
                value={textFilter}
                onChange={(e) => setTextFilter(e.currentTarget.value)}
                placeholder="Search SQL, session, user, wait, traceId..."
              />
            </div>
            <label style={{ display: 'inline-flex', alignItems: 'center', gap: 6, cursor: 'pointer' }}>
              <Switch value={onlyLocking} onChange={(e) => setOnlyLocking(e.currentTarget.checked)} />
              <span>Blocking / Blocked only</span>
            </label>
          </div>

          <div className={styles.tableScroll}>
            <table className={styles.samplesTable}>
              <thead>
                <tr>
                  <th>Role</th>
                  <th>SID</th>
                  <th>Database</th>
                  <th>Elapsed</th>
                  <th>Wait Event</th>
                  <th>Status</th>
                  <th>Trace</th>
                  <th>User</th>
                  <th>SQL Query</th>
                </tr>
              </thead>
              <tbody>
                {filteredSamples.length === 0 ? (
                  <tr>
                    <td colSpan={9} style={{ textAlign: 'center', padding: 20, opacity: 0.7 }}>
                      No samples match the current filter.
                    </td>
                  </tr>
                ) : (
                  filteredSamples.map((s) => (
                    <tr key={`${s.session_id}-${s.sample_id}`} onClick={() => onSelectSample(s.snap_id, s.sample_id)}>
                      <td>
                        {s.is_blocker ? (
                          <span className={styles.blockerPill}>
                            <Lock size={11} />
                            {s.blocked_sessions.length} waiting
                          </span>
                        ) : s.is_blocked ? (
                          <span className={styles.waiterPill}>
                            <Clock size={11} />
                            by {s.blocked_by}
                          </span>
                        ) : (
                          <span style={{ opacity: 0.4 }}>-</span>
                        )}
                      </td>
                      <td className={styles.mono}>{s.session_id}</td>
                      <td className={styles.mono}>{s.database || '-'}</td>
                      <td className={styles.mono}>{formatDurationMs(s.elapsed_ms)}</td>
                      <td className={styles.mono}>{s.wait_event || 'CPU'}</td>
                      <td>{s.status}</td>
                      <td>
                        {s.trace_id ? <TempoTraceLink traceId={s.trace_id} timeRange={timeRange} compact /> : '-'}
                      </td>
                      <td>{s.user || '-'}</td>
                      <td>
                        <div className={styles.sqlCell} title={s.text}>
                          {s.text}
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  );
};

export const SnapshotsTable: React.FC<{
  summaries: SnapshotSummaryItem[];
  sortField: SnapshotSortField;
  sortDir: SortDirection;
  onSortChange: (field: SnapshotSortField) => void;
  expandedSnapId: string | null;
  onToggleExpand: (snapId: string) => void;
  onSelectSample: (snapId: string, sampleId: string) => void;
  timeRange: TimeRange;
}> = ({
  summaries,
  sortField,
  sortDir,
  onSortChange,
  expandedSnapId,
  onToggleExpand,
  onSelectSample,
  timeRange,
}) => {
  const styles = useStyles2(getStyles);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(25);
  const [samplesCache, setSamplesCache] = useState<Record<string, SnapshotSampleItem[]>>({});
  const [samplesLoading, setSamplesLoading] = useState<Record<string, boolean>>({});
  const [samplesError, setSamplesError] = useState<Record<string, string>>({});
  const samplesCacheRef = useRef(samplesCache);
  const samplesLoadingRef = useRef(samplesLoading);
  const samplesErrorRef = useRef(samplesError);
  samplesCacheRef.current = samplesCache;
  samplesLoadingRef.current = samplesLoading;
  samplesErrorRef.current = samplesError;

  const mountedRef = useRef(true);
  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  const fetchSamples = useCallback(async (snapId: string, force = false) => {
    if (
      !snapId ||
      (!force && (samplesCacheRef.current[snapId] || samplesLoadingRef.current[snapId] || samplesErrorRef.current[snapId]))
    ) {
      return;
    }
    samplesLoadingRef.current = { ...samplesLoadingRef.current, [snapId]: true };
    const nextErrors = { ...samplesErrorRef.current };
    delete nextErrors[snapId];
    samplesErrorRef.current = nextErrors;
    setSamplesLoading((prev) => ({ ...prev, [snapId]: true }));
    setSamplesError((prev) => {
      const next = { ...prev };
      delete next[snapId];
      return next;
    });
    try {
      const params = new URLSearchParams({ snapId });
      const response: Observable<FetchResponse<SnapshotSampleItem[]>> = getBackendSrv().fetch({
        url: `/api/plugins/guilhermearpassos-sqlsights-app/resources/snapshotSamples?${params.toString()}`,
      });
      const result = await lastValueFrom(response);
      if (!mountedRef.current) {
        return;
      }
      const data = result.data || [];
      samplesCacheRef.current = { ...samplesCacheRef.current, [snapId]: data };
      setSamplesCache((prev) => ({ ...prev, [snapId]: data }));
    } catch (err) {
      if (!mountedRef.current) {
        return;
      }
      const message = extractErrorMessage(err, 'Failed to load samples');
      samplesErrorRef.current = { ...samplesErrorRef.current, [snapId]: message };
      setSamplesError((prev) => ({
        ...prev,
        [snapId]: message,
      }));
    } finally {
      samplesLoadingRef.current = { ...samplesLoadingRef.current, [snapId]: false };
      if (mountedRef.current) {
        setSamplesLoading((prev) => ({ ...prev, [snapId]: false }));
      }
    }
  }, []);

  useEffect(() => {
    if (expandedSnapId) {
      fetchSamples(expandedSnapId);
    }
  }, [expandedSnapId, fetchSamples]);

  const sortedSummaries = useMemo(() => {
    const copy = [...summaries];
    const dir = sortDir === 'asc' ? 1 : -1;
    copy.sort((a, b) => {
      switch (sortField) {
        case 'connections':
          return (a.connections - b.connections) * dir;
        case 'blockers':
          return (a.blockers - b.blockers) * dir;
        case 'waiters':
          return (a.waiters - b.waiters) * dir;
        case 'wait_duration_ms':
          return (a.wait_duration_ms - b.wait_duration_ms) * dir;
        case 'timestamp':
        default:
          return (new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()) * dir;
      }
    });
    return copy;
  }, [summaries, sortField, sortDir]);

  // If URL expandedSnapId changes and is on another page, jump to that page once (don't yank on auto-refresh)
  const lastJumpedSnapIdRef = useRef<string | null>(null);
  useEffect(() => {
    if (!expandedSnapId) {
      lastJumpedSnapIdRef.current = null;
      return;
    }
    if (lastJumpedSnapIdRef.current === expandedSnapId) {
      return;
    }
    const idx = sortedSummaries.findIndex((s) => s.id === expandedSnapId);
    if (idx >= 0) {
      lastJumpedSnapIdRef.current = expandedSnapId;
      const targetPage = Math.floor(idx / pageSize) + 1;
      setPage(targetPage);
    }
  }, [expandedSnapId, sortedSummaries, pageSize]);

  const totalPages = Math.max(1, Math.ceil(sortedSummaries.length / pageSize));
  const safePage = Math.min(page, totalPages);
  const pageItems = useMemo(() => {
    const start = (safePage - 1) * pageSize;
    return sortedSummaries.slice(start, start + pageSize);
  }, [sortedSummaries, safePage, pageSize]);

  const renderSortIcon = (field: SnapshotSortField) => {
    if (sortField !== field) {
      return <ArrowUpDown size={12} />;
    }
    return sortDir === 'asc' ? <ArrowUp size={12} /> : <ArrowDown size={12} />;
  };

  if (summaries.length === 0) {
    return (
      <div className={styles.wrapper}>
        <div className={styles.emptyState}>No snapshots match the selected filters and time range.</div>
      </div>
    );
  }

  return (
    <div className={styles.wrapper}>
      <div className={styles.toolbar}>
        <span>
          Showing <strong>{(safePage - 1) * pageSize + 1}</strong>–
          <strong>{Math.min(safePage * pageSize, sortedSummaries.length)}</strong> of{' '}
          <strong>{sortedSummaries.length}</strong> snapshots
        </span>
        <label style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
          Rows:
          <select
            value={pageSize}
            onChange={(e) => {
              setPageSize(Number(e.target.value));
              setPage(1);
            }}
            style={{ padding: '2px 6px', borderRadius: 3 }}
          >
            <option value={25}>25</option>
            <option value={50}>50</option>
            <option value={100}>100</option>
          </select>
        </label>
      </div>

      <div className={styles.tableScroll}>
        <table className={styles.table}>
          <thead>
            <tr>
              <th style={{ width: 32 }} />
              <th>
                <button type="button" className={styles.sortBtn} onClick={() => onSortChange('timestamp')}>
                  Timestamp {renderSortIcon('timestamp')}
                </button>
              </th>
              <th>
                <button type="button" className={styles.sortBtn} onClick={() => onSortChange('connections')}>
                  Connections {renderSortIcon('connections')}
                </button>
              </th>
              <th>
                <button type="button" className={styles.sortBtn} onClick={() => onSortChange('blockers')}>
                  Blockers {renderSortIcon('blockers')}
                </button>
              </th>
              <th>
                <button type="button" className={styles.sortBtn} onClick={() => onSortChange('waiters')}>
                  Waiters {renderSortIcon('waiters')}
                </button>
              </th>
              <th>
                <button type="button" className={styles.sortBtn} onClick={() => onSortChange('wait_duration_ms')}>
                  Wait Duration {renderSortIcon('wait_duration_ms')}
                </button>
              </th>
              <th>Wait Breakdown</th>
            </tr>
          </thead>
          <tbody>
            {pageItems.map((snap) => {
              const isExpanded = expandedSnapId === snap.id;
              return (
                <React.Fragment key={snap.id}>
                  <tr
                    className={`${styles.summaryRow} ${isExpanded ? styles.summaryRowExpanded : ''}`}
                    onClick={() => onToggleExpand(snap.id)}
                  >
                    <td>{isExpanded ? <ChevronDown size={15} /> : <ChevronRight size={15} />}</td>
                    <td className={styles.mono}>{new Date(snap.timestamp).toLocaleString()}</td>
                    <td className={styles.mono}>{snap.connections}</td>
                    <td>
                      {snap.blockers > 0 ? (
                        <span className={styles.blockerPill}>
                          <Lock size={11} />
                          {snap.blockers}
                        </span>
                      ) : (
                        <span className={styles.mono}>0</span>
                      )}
                    </td>
                    <td>
                      {snap.waiters > 0 ? (
                        <span className={styles.waiterPill}>
                          <Clock size={11} />
                          {snap.waiters}
                        </span>
                      ) : (
                        <span className={styles.mono}>0</span>
                      )}
                    </td>
                    <td className={styles.mono}>{formatDurationMs(snap.wait_duration_ms)}</td>
                    <td>
                      <WaitBreakdownBar counts={snap.connections_by_wait_event} />
                    </td>
                  </tr>
                  {isExpanded && (
                    <tr>
                      <td colSpan={7} className={styles.detailRowCell}>
                        <ExpandedSnapshotPanel
                          snapId={snap.id}
                          samples={samplesCache[snap.id]}
                          loading={Boolean(samplesLoading[snap.id])}
                          error={samplesError[snap.id]}
                          onRetry={() => fetchSamples(snap.id, true)}
                          timeRange={timeRange}
                          onSelectSample={onSelectSample}
                        />
                      </td>
                    </tr>
                  )}
                </React.Fragment>
              );
            })}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className={styles.pagination}>
          <span>
            Page {safePage} of {totalPages}
          </span>
          <div style={{ display: 'flex', gap: 8 }}>
            <Button size="sm" variant="secondary" disabled={safePage <= 1} onClick={() => setPage((p) => Math.max(1, p - 1))}>
              Previous
            </Button>
            <Button
              size="sm"
              variant="secondary"
              disabled={safePage >= totalPages}
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            >
              Next
            </Button>
          </div>
        </div>
      )}
    </div>
  );
};
