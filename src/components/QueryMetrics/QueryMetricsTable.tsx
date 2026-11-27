import React, { useMemo, useState } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2 } from '@grafana/data';
import { Button, RadioButtonGroup, Tooltip, useStyles2 } from '@grafana/ui';
import {
  Activity,
  ArrowDown,
  ArrowUp,
  ArrowUpDown,
  Download,
  FileSearch,
  Flame,
} from 'lucide-react';
import {
  getMetricIssueMeta,
  QueryMetricItem,
  QueryMetricsAggregationMode,
  QueryMetricsSortField,
  SortDirection,
} from './types';
import { downloadTextFile, formatDurationMs } from '../../utils/sqlHelpers';

function escapeCSVCell(value: unknown): string {
  const s = String(value ?? '');
  if (/[",\n\r]/.test(s)) {
    return `"${s.replace(/"/g, '""')}"`;
  }
  return s;
}

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
      vertical-align: top;
    }

    tbody tr {
      cursor: pointer;
      transition: background-color 0.12s ease;

      &:hover {
        background: ${theme.colors.action.hover};
      }
    }
  `,
  sortHeaderBtn: css`
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
  tagsCell: css`
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    max-width: 190px;
  `,
  tagBadge: css`
    display: inline-flex;
    align-items: center;
    gap: 3px;
    padding: 2px 6px;
    border-radius: ${theme.shape.radius.default};
    font-size: 11px;
    font-weight: ${theme.typography.fontWeightMedium};
    white-space: nowrap;
    border: 1px solid transparent;
  `,
  tagError: css`
    background: ${theme.colors.error.transparent};
    color: ${theme.colors.error.text};
    border-color: ${theme.colors.error.border};
  `,
  tagWarning: css`
    background: ${theme.colors.warning.transparent};
    color: ${theme.colors.warning.text};
    border-color: ${theme.colors.warning.border};
  `,
  tagInfo: css`
    background: ${theme.colors.info.transparent};
    color: ${theme.colors.info.text};
    border-color: ${theme.colors.info.border};
  `,
  dbBadge: css`
    display: inline-block;
    padding: 2px 7px;
    border-radius: ${theme.shape.radius.default};
    background: ${theme.colors.background.secondary};
    border: 1px solid ${theme.colors.border.weak};
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: 12px;
    white-space: nowrap;
  `,
  queryPreview: css`
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: 12px;
    max-width: 420px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  `,
  subLine: css`
    margin-top: 3px;
    font-size: 11px;
    color: ${theme.colors.text.secondary};
    font-family: ${theme.typography.fontFamilyMonospace};
    white-space: nowrap;
  `,
  numericCell: css`
    font-family: ${theme.typography.fontFamilyMonospace};
    white-space: nowrap;
  `,
  shareBarTrack: css`
    width: 72px;
    height: 4px;
    border-radius: 2px;
    background: ${theme.colors.background.secondary};
    margin-top: 4px;
    overflow: hidden;
  `,
  shareBarFill: css`
    height: 100%;
    background: ${theme.colors.primary.main};
  `,
  spillPill: css`
    display: inline-flex;
    align-items: center;
    gap: 3px;
    padding: 2px 6px;
    border-radius: ${theme.shape.radius.default};
    background: ${theme.colors.error.transparent};
    color: ${theme.colors.error.text};
    border: 1px solid ${theme.colors.error.border};
    font-weight: ${theme.typography.fontWeightMedium};
  `,
  actionsCell: css`
    display: flex;
    align-items: center;
    gap: 4px;
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

function formatCompactNumber(n: number): string {
  if (n >= 1_000_000) {
    return `${(n / 1_000_000).toFixed(1)}M`;
  }
  if (n >= 10_000) {
    return `${(n / 1_000).toFixed(1)}K`;
  }
  return Math.round(n).toLocaleString();
}

function formatAvgCount(n: number): string {
  if (n >= 10_000) {
    return formatCompactNumber(n);
  }
  if (n > 0 && n < 10) {
    return Number.isInteger(n) ? n.toLocaleString() : n.toFixed(2);
  }
  return Math.round(n).toLocaleString();
}

const AGG_MODE_OPTIONS: Array<{ label: string; value: QueryMetricsAggregationMode }> = [
  { label: 'Totals', value: 'total' },
  { label: 'Avg / Exec', value: 'avg' },
];

export const QueryMetricsTable: React.FC<{
  metrics: QueryMetricItem[];
  aggMode: QueryMetricsAggregationMode;
  onAggModeChange: (mode: QueryMetricsAggregationMode) => void;
  sortField: QueryMetricsSortField;
  sortDir: SortDirection;
  onSortChange: (field: QueryMetricsSortField) => void;
  onSelectMetric: (item: QueryMetricItem) => void;
  onOpenPlanAnalysis: (item: QueryMetricItem) => void;
  onOpenSnapshots: (item: QueryMetricItem) => void;
}> = ({
  metrics,
  aggMode,
  onAggModeChange,
  sortField,
  sortDir,
  onSortChange,
  onSelectMetric,
  onOpenPlanAnalysis,
  onOpenSnapshots,
}) => {
  const styles = useStyles2(getStyles);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(25);
  const isAvg = aggMode === 'avg';

  const durationSortField: QueryMetricsSortField = isAvg ? 'avg_elapsed_ms' : 'total_elapsed_ms';
  const cpuSortField: QueryMetricsSortField = isAvg ? 'avg_worker_ms' : 'total_worker_ms';
  const readsSortField: QueryMetricsSortField = isAvg ? 'avg_logical_reads' : 'total_logical_reads';
  const spillsSortField: QueryMetricsSortField = isAvg ? 'avg_spills' : 'total_spills';
  const grantSortField: QueryMetricsSortField = isAvg ? 'avg_grant_kb' : 'total_grant_kb';

  const sorted = useMemo(() => {
    const copy = [...metrics];
    const dir = sortDir === 'asc' ? 1 : -1;
    copy.sort((a, b) => {
      switch (sortField) {
        case 'avg_elapsed_ms':
          return (a.avg_elapsed_ms - b.avg_elapsed_ms) * dir;
        case 'total_worker_ms':
          return (a.total_worker_ms - b.total_worker_ms) * dir;
        case 'avg_worker_ms':
          return (a.avg_worker_ms - b.avg_worker_ms) * dir;
        case 'execution_count':
          return (a.execution_count - b.execution_count) * dir;
        case 'total_logical_reads':
          return (a.total_logical_reads - b.total_logical_reads) * dir;
        case 'avg_logical_reads':
          return (a.avg_logical_reads - b.avg_logical_reads) * dir;
        case 'total_physical_reads':
          return (a.total_physical_reads - b.total_physical_reads) * dir;
        case 'total_spills':
          return (a.total_spills - b.total_spills) * dir;
        case 'avg_spills':
          return ((a.avg_spills ?? 0) - (b.avg_spills ?? 0)) * dir;
        case 'total_grant_kb':
          return ((a.total_grant_kb ?? 0) - (b.total_grant_kb ?? 0)) * dir;
        case 'avg_grant_kb':
          return (a.avg_grant_kb - b.avg_grant_kb) * dir;
        case 'avg_rows':
          return (a.avg_rows - b.avg_rows) * dir;
        case 'last_execution_time':
          return (new Date(a.last_execution_time).getTime() - new Date(b.last_execution_time).getTime()) * dir;
        case 'total_elapsed_ms':
        default:
          return (a.total_elapsed_ms - b.total_elapsed_ms) * dir;
      }
    });
    return copy;
  }, [metrics, sortField, sortDir]);

  const totalPages = Math.max(1, Math.ceil(sorted.length / pageSize));
  const safePage = Math.min(page, totalPages);
  const pageItems = useMemo(() => {
    const start = (safePage - 1) * pageSize;
    return sorted.slice(start, start + pageSize);
  }, [sorted, safePage, pageSize]);

  const renderSortIcon = (field: QueryMetricsSortField) => {
    if (sortField !== field) {
      return <ArrowUpDown size={12} />;
    }
    return sortDir === 'asc' ? <ArrowUp size={12} /> : <ArrowDown size={12} />;
  };

  const handleExportCSV = () => {
    const headers = [
      'Database',
      'Query Hash',
      'Issue Flags',
      'Executions',
      'Total Duration (ms)',
      'Avg Duration (ms)',
      'Total CPU (ms)',
      'Avg CPU (ms)',
      'Total Logical Reads',
      'Avg Logical Reads',
      'Total Physical Reads',
      'Total Spills',
      'Avg Spills',
      'Total Grant (KB)',
      'Avg Grant (KB)',
      'Avg Used Grant (KB)',
      'Avg Rows',
      'Last Execution',
      'SQL Text',
    ];
    const rows = sorted.map((m) => [
      m.database,
      m.query_hash,
      (m.issue_flags ?? []).join('|'),
      m.execution_count,
      m.total_elapsed_ms.toFixed(2),
      m.avg_elapsed_ms.toFixed(2),
      m.total_worker_ms.toFixed(2),
      m.avg_worker_ms.toFixed(2),
      m.total_logical_reads,
      m.avg_logical_reads.toFixed(1),
      m.total_physical_reads,
      m.total_spills,
      (m.avg_spills ?? 0).toFixed(2),
      m.total_grant_kb ?? 0,
      m.avg_grant_kb.toFixed(1),
      m.avg_used_grant_kb.toFixed(1),
      m.avg_rows.toFixed(1),
      m.last_execution_time,
      m.text,
    ]);
    const csv = [headers, ...rows].map((row) => row.map(escapeCSVCell).join(',')).join('\n');
    downloadTextFile(`sqlsights_query_metrics_${new Date().toISOString().slice(0, 19)}.csv`, csv, 'text/csv');
  };

  if (metrics.length === 0) {
    return (
      <div className={styles.wrapper}>
        <div className={styles.emptyState}>
          No query metrics match the current filters and time range.
        </div>
      </div>
    );
  }

  return (
    <div className={styles.wrapper}>
      <div className={styles.toolbar}>
        <span>
          Showing <strong>{(safePage - 1) * pageSize + 1}</strong>–
          <strong>{Math.min(safePage * pageSize, sorted.length)}</strong> of <strong>{sorted.length}</strong> queries
        </span>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexWrap: 'wrap' }}>
          <RadioButtonGroup<QueryMetricsAggregationMode>
            size="sm"
            options={AGG_MODE_OPTIONS}
            value={aggMode}
            onChange={onAggModeChange}
          />
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
          <Button size="sm" variant="secondary" fill="outline" onClick={handleExportCSV}>
            <Download size={13} />
            <span style={{ marginLeft: 4 }}>Export CSV</span>
          </Button>
        </div>
      </div>

      <div className={styles.tableScroll}>
        <table className={styles.table}>
          <thead>
            <tr>
              <th>Issues</th>
              <th>Database</th>
              <th>SQL Query</th>
              <th>
                <button type="button" className={styles.sortHeaderBtn} onClick={() => onSortChange('execution_count')}>
                  Execs {renderSortIcon('execution_count')}
                </button>
              </th>
              <th>
                <button type="button" className={styles.sortHeaderBtn} onClick={() => onSortChange(durationSortField)}>
                  {isAvg ? 'Avg Duration' : 'Duration'} {renderSortIcon(durationSortField)}
                </button>
              </th>
              <th>
                <button type="button" className={styles.sortHeaderBtn} onClick={() => onSortChange(cpuSortField)}>
                  {isAvg ? 'Avg CPU' : 'CPU'} {renderSortIcon(cpuSortField)}
                </button>
              </th>
              <th>
                <button type="button" className={styles.sortHeaderBtn} onClick={() => onSortChange(readsSortField)}>
                  {isAvg ? 'Avg Reads' : 'Logical Reads'} {renderSortIcon(readsSortField)}
                </button>
              </th>
              <th>
                <button type="button" className={styles.sortHeaderBtn} onClick={() => onSortChange(spillsSortField)}>
                  {isAvg ? 'Avg Spills' : 'Spills'} {renderSortIcon(spillsSortField)}
                </button>
              </th>
              <th>
                <button type="button" className={styles.sortHeaderBtn} onClick={() => onSortChange(grantSortField)}>
                  {isAvg ? 'Avg Mem Grant' : 'Mem Grant'} {renderSortIcon(grantSortField)}
                </button>
              </th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {pageItems.map((item) => (
              <tr key={`${item.database}-${item.query_hash}`} onClick={() => onSelectMetric(item)}>
                <td>
                  <div className={styles.tagsCell}>
                    {(item.issue_flags ?? []).length === 0 ? (
                      <span style={{ opacity: 0.45 }}>-</span>
                    ) : (
                      item.issue_flags.map((flag) => {
                        const meta = getMetricIssueMeta(flag);
                        const cls =
                          meta.severity === 'error'
                            ? styles.tagError
                            : meta.severity === 'warning'
                              ? styles.tagWarning
                              : styles.tagInfo;
                        return (
                          <Tooltip key={flag} content={meta.description}>
                            <span className={`${styles.tagBadge} ${cls}`}>{meta.shortLabel}</span>
                          </Tooltip>
                        );
                      })
                    )}
                  </div>
                </td>
                <td>
                  <span className={styles.dbBadge}>{item.database || '-'}</span>
                </td>
                <td>
                  <div className={styles.queryPreview} title={item.text}>
                    {item.text}
                  </div>
                  <div className={styles.subLine}>
                    {isAvg
                      ? `avg ${formatCompactNumber(item.avg_rows)} rows`
                      : `total ${formatCompactNumber(item.total_rows ?? item.avg_rows * item.execution_count)} rows`}
                    {' · '}last {formatDurationMs(item.last_elapsed_ms)}
                    {item.avg_dop > 1 ? ` · DOP ${item.avg_dop.toFixed(1)}` : ''}
                  </div>
                </td>
                <td className={styles.numericCell}>{item.execution_count.toLocaleString()}</td>
                <td className={styles.numericCell}>
                  <div>{formatDurationMs(isAvg ? item.avg_elapsed_ms : item.total_elapsed_ms)}</div>
                  <div className={styles.subLine}>
                    {isAvg
                      ? `total ${formatDurationMs(item.total_elapsed_ms)}`
                      : `avg ${formatDurationMs(item.avg_elapsed_ms)}`}{' '}
                    ({item.elapsed_share_pct.toFixed(1)}%)
                  </div>
                  <div className={styles.shareBarTrack}>
                    <div
                      className={styles.shareBarFill}
                      style={{ width: `${Math.min(100, item.elapsed_share_pct)}%` }}
                    />
                  </div>
                </td>
                <td className={styles.numericCell}>
                  <div>{formatDurationMs(isAvg ? item.avg_worker_ms : item.total_worker_ms)}</div>
                  <div className={styles.subLine}>
                    {isAvg
                      ? `total ${formatDurationMs(item.total_worker_ms)}`
                      : `avg ${formatDurationMs(item.avg_worker_ms)}`}{' '}
                    ({item.cpu_share_pct.toFixed(1)}%)
                  </div>
                </td>
                <td className={styles.numericCell}>
                  <div>{formatCompactNumber(isAvg ? item.avg_logical_reads : item.total_logical_reads)}</div>
                  <div className={styles.subLine}>
                    {isAvg
                      ? `total ${formatCompactNumber(item.total_logical_reads)}`
                      : `avg ${formatCompactNumber(item.avg_logical_reads)}`}
                    {isAvg
                      ? (item.avg_physical_reads ?? 0) > 0
                        ? ` · ${formatAvgCount(item.avg_physical_reads)} phys`
                        : ''
                      : item.total_physical_reads > 0
                        ? ` · ${formatCompactNumber(item.total_physical_reads)} phys`
                        : ''}
                  </div>
                </td>
                <td className={styles.numericCell}>
                  {item.total_spills > 0 ? (
                    <>
                      <span className={styles.spillPill}>
                        <Flame size={11} />
                        {isAvg ? formatAvgCount(item.avg_spills ?? 0) : item.total_spills.toLocaleString()}
                      </span>
                      <div className={styles.subLine}>
                        {isAvg
                          ? `total ${item.total_spills.toLocaleString()}`
                          : `avg ${formatAvgCount(item.avg_spills ?? 0)}`}
                      </div>
                    </>
                  ) : (
                    '0'
                  )}
                </td>
                <td className={styles.numericCell}>
                  {item.avg_grant_kb > 0 || (item.total_grant_kb ?? 0) > 0 ? (
                    <>
                      <div>
                        {formatCompactNumber(isAvg ? item.avg_grant_kb : (item.total_grant_kb ?? 0))} KB
                      </div>
                      <div className={styles.subLine}>
                        {isAvg
                          ? `total ${formatCompactNumber(item.total_grant_kb ?? 0)} KB · used ${formatCompactNumber(item.avg_used_grant_kb)} KB`
                          : `avg ${formatCompactNumber(item.avg_grant_kb)} KB · used ${formatCompactNumber(item.avg_used_grant_kb)} KB`}
                        {item.memory_grant_waste_pct >= 50 ? ` (${item.memory_grant_waste_pct.toFixed(0)}% waste)` : ''}
                      </div>
                    </>
                  ) : (
                    '-'
                  )}
                </td>
                <td>
                  <div className={styles.actionsCell} onClick={(e) => e.stopPropagation()}>
                    <Tooltip content="Search matching plans in Plan Analysis">
                      <Button size="sm" variant="secondary" fill="outline" onClick={() => onOpenPlanAnalysis(item)}>
                        <FileSearch size={13} />
                      </Button>
                    </Tooltip>
                    <Tooltip content="Open Snapshots for this time window">
                      <Button size="sm" variant="secondary" fill="outline" onClick={() => onOpenSnapshots(item)}>
                        <Activity size={13} />
                      </Button>
                    </Tooltip>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className={styles.pagination}>
          <span>
            Page {safePage} of {totalPages}
          </span>
          <div style={{ display: 'flex', gap: 8 }}>
            <Button
              size="sm"
              variant="secondary"
              disabled={safePage <= 1}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
            >
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
