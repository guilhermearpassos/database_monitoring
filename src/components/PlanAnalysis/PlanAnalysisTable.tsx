import React, { useMemo, useState } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2, TimeRange } from '@grafana/data';
import { Button, Tooltip, useStyles2 } from '@grafana/ui';
import { ArrowDown, ArrowUp, ArrowUpDown, BarChart3, Check, Copy, Download, ExternalLink, Lock } from 'lucide-react';
import { PlanAnalysisItem } from './types';
import { buildMissingIndexDDL, copyToClipboard, downloadTextFile, getIssueTagMeta } from '../../utils/sqlHelpers';
import { TempoTraceLink } from '../TempoTraceLink';

export type PlanSortField = 'related_lock_count' | 'occurrences' | 'issue_count' | 'max_impact' | 'last_snap_time' | 'database';
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
      padding: ${theme.spacing(1.25)};
      border-bottom: 1px solid ${theme.colors.border.weak};
      vertical-align: top;
      color: ${theme.colors.text.primary};
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
    max-width: 210px;
  `,
  tagBadge: css`
    display: inline-flex;
    align-items: center;
    padding: 2px 6px;
    border-radius: ${theme.shape.radius.default};
    font-size: 11px;
    font-weight: ${theme.typography.fontWeightMedium};
    white-space: nowrap;
  `,
  tagError: css`
    background: ${theme.colors.error.transparent};
    color: ${theme.colors.error.text};
    border: 1px solid ${theme.colors.error.border};
  `,
  tagWarning: css`
    background: ${theme.colors.warning.transparent};
    color: ${theme.colors.warning.text};
    border: 1px solid ${theme.colors.warning.border};
  `,
  tagInfo: css`
    background: ${theme.colors.info.transparent};
    color: ${theme.colors.info.text};
    border: 1px solid ${theme.colors.info.border};
  `,
  dbBadge: css`
    display: inline-block;
    padding: 2px 7px;
    border-radius: ${theme.shape.radius.default};
    background: ${theme.colors.background.secondary};
    border: 1px solid ${theme.colors.border.weak};
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: 11px;
    white-space: nowrap;
  `,
  queryPreview: css`
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: 12px;
    color: ${theme.colors.text.primary};
    max-width: 520px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    word-break: break-word;
  `,
  warningDetailLine: css`
    margin-top: 6px;
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: 11px;
    color: ${theme.colors.text.secondary};
    max-width: 520px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  `,
  lockPill: css`
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 7px;
    border-radius: ${theme.shape.radius.default};
    background: ${theme.colors.error.transparent};
    color: ${theme.colors.error.text};
    border: 1px solid ${theme.colors.error.border};
    font-weight: ${theme.typography.fontWeightMedium};
  `,
  numericCell: css`
    font-family: ${theme.typography.fontFamilyMonospace};
    white-space: nowrap;
  `,
  actionsCell: css`
    display: flex;
    align-items: center;
    gap: ${theme.spacing(0.75)};
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

function escapeCSV(val: string | number): string {
  const s = String(val ?? '');
  if (/[",\n\r]/.test(s)) {
    return `"${s.replace(/"/g, '""')}"`;
  }
  return s;
}

export const PlanAnalysisTable: React.FC<{
  plans: PlanAnalysisItem[];
  sortField: PlanSortField;
  sortDir: SortDirection;
  onSortChange: (field: PlanSortField) => void;
  onSelectPlan: (item: PlanAnalysisItem) => void;
  onOpenSnapshot: (snapId: string, sampleId: string) => void;
  onOpenQueryMetrics?: (queryHash: string) => void;
  timeRange: TimeRange;
}> = ({ plans, sortField, sortDir, onSortChange, onSelectPlan, onOpenSnapshot, onOpenQueryMetrics, timeRange }) => {
  const styles = useStyles2(getStyles);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(25);
  const [copiedKey, setCopiedKey] = useState<string | null>(null);

  const sortedPlans = useMemo(() => {
    const copy = [...plans];
    const dir = sortDir === 'asc' ? 1 : -1;
    copy.sort((a, b) => {
      switch (sortField) {
        case 'related_lock_count':
          return (a.related_lock_count - b.related_lock_count) * dir || b.occurrences - a.occurrences;
        case 'occurrences':
          return (a.occurrences - b.occurrences) * dir;
        case 'issue_count':
          return (a.issue_count - b.issue_count) * dir;
        case 'max_impact':
          return (a.max_impact - b.max_impact) * dir;
        case 'database':
          return a.database.localeCompare(b.database) * dir;
        case 'last_snap_time':
        default:
          return (new Date(a.last_snap_time).getTime() - new Date(b.last_snap_time).getTime()) * dir;
      }
    });
    return copy;
  }, [plans, sortField, sortDir]);

  const totalPages = Math.max(1, Math.ceil(sortedPlans.length / pageSize));
  const safePage = Math.min(page, totalPages);
  const pageItems = useMemo(() => {
    const startIdx = (safePage - 1) * pageSize;
    return sortedPlans.slice(startIdx, startIdx + pageSize);
  }, [sortedPlans, safePage, pageSize]);

  const renderSortIcon = (field: PlanSortField) => {
    if (sortField !== field) {
      return <ArrowUpDown size={12} />;
    }
    return sortDir === 'asc' ? <ArrowUp size={12} /> : <ArrowDown size={12} />;
  };

  const copyTimerRef = React.useRef<number | null>(null);
  React.useEffect(() => {
    return () => {
      if (copyTimerRef.current !== null) {
        window.clearTimeout(copyTimerRef.current);
      }
    };
  }, []);

  const handleCopyDDL = async (e: React.MouseEvent, item: PlanAnalysisItem) => {
    e.stopPropagation();
    if (!item.missing_indexes || item.missing_indexes.length === 0) {
      return;
    }
    const ddl = item.missing_indexes.map(buildMissingIndexDDL).join('\n');
    if (await copyToClipboard(ddl)) {
      const rowKey = item.plan_handle || `${item.snap_id}-${item.sample_id}`;
      setCopiedKey(rowKey);
      if (copyTimerRef.current !== null) {
        window.clearTimeout(copyTimerRef.current);
      }
      copyTimerRef.current = window.setTimeout(() => setCopiedKey(null), 1800);
    }
  };

  const handleExportCSV = () => {
    const header = ['Database', 'IssueTags', 'Issues', 'Occurrences', 'RelatedLocks', 'MaxImpact', 'LastSeen', 'Details', 'SQL'];
    const rows = sortedPlans.map((p) => {
      const details = [
        ...(p.converts ?? []).map((c) => `${c.convert_issue}: ${c.expression}`),
        ...(p.missing_indexes ?? []).map((m) => `MissingIndex(${m.impact.toFixed(1)}%): ${buildMissingIndexDDL(m)}`),
      ].join(' | ');
      return [
        p.database,
        (p.issue_tags ?? []).join(';'),
        p.issue_count,
        p.occurrences,
        p.related_lock_count,
        p.max_impact.toFixed(1),
        p.last_snap_time,
        details,
        p.text.replace(/\s+/g, ' ').trim(),
      ]
        .map(escapeCSV)
        .join(',');
    });
    downloadTextFile(`sqlsights-plan-analysis-${new Date().toISOString().slice(0, 19)}.csv`, [header.join(','), ...rows].join('\n'), 'text/csv;charset=utf-8');
  };

  if (plans.length === 0) {
    return (
      <div className={styles.wrapper}>
        <div className={styles.emptyState}>
          No problematic execution plans match the selected filters and time range.
        </div>
      </div>
    );
  }

  return (
    <div className={styles.wrapper}>
      <div className={styles.toolbar}>
        <span>
          Showing <strong>{(safePage - 1) * pageSize + 1}</strong>–
          <strong>{Math.min(safePage * pageSize, sortedPlans.length)}</strong> of <strong>{sortedPlans.length}</strong>{' '}
          problematic plans
        </span>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
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
              <th>
                <button type="button" className={styles.sortHeaderBtn} onClick={() => onSortChange('database')}>
                  Database {renderSortIcon('database')}
                </button>
              </th>
              <th>Query &amp; Warning Preview</th>
              <th>
                <button type="button" className={styles.sortHeaderBtn} onClick={() => onSortChange('max_impact')}>
                  Impact {renderSortIcon('max_impact')}
                </button>
              </th>
              <th>
                <button type="button" className={styles.sortHeaderBtn} onClick={() => onSortChange('occurrences')}>
                  Occurrences {renderSortIcon('occurrences')}
                </button>
              </th>
              <th>
                <button
                  type="button"
                  className={styles.sortHeaderBtn}
                  onClick={() => onSortChange('related_lock_count')}
                >
                  Locks {renderSortIcon('related_lock_count')}
                </button>
              </th>
              <th>
                <button type="button" className={styles.sortHeaderBtn} onClick={() => onSortChange('last_snap_time')}>
                  Last Seen {renderSortIcon('last_snap_time')}
                </button>
              </th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {pageItems.map((item) => {
              const rowKey = item.plan_handle || `${item.snap_id}-${item.sample_id}`;
              const primaryConvert = item.converts?.[0];
              const primaryMissing = item.missing_indexes?.[0];
              return (
                <tr key={rowKey} onClick={() => onSelectPlan(item)}>
                  <td>
                    <div className={styles.tagsCell}>
                      {(item.issue_tags ?? []).map((tag) => {
                        const meta = getIssueTagMeta(tag);
                        const cls =
                          meta.severity === 'error'
                            ? styles.tagError
                            : meta.severity === 'warning'
                              ? styles.tagWarning
                              : styles.tagInfo;
                        return (
                          <Tooltip key={tag} content={meta.description}>
                            <span className={`${styles.tagBadge} ${cls}`}>{meta.shortLabel}</span>
                          </Tooltip>
                        );
                      })}
                    </div>
                  </td>
                  <td>
                    <span className={styles.dbBadge}>{item.database || '-'}</span>
                  </td>
                  <td>
                    <div className={styles.queryPreview} title={item.text}>
                      {item.text}
                    </div>
                    {primaryConvert && (
                      <div className={styles.warningDetailLine} title={primaryConvert.expression}>
                        <strong>{primaryConvert.convert_issue}:</strong> {primaryConvert.expression}
                      </div>
                    )}
                    {primaryMissing && (
                      <div className={styles.warningDetailLine} title={buildMissingIndexDDL(primaryMissing)}>
                        <strong>Missing Index:</strong> {primaryMissing.schema}.{primaryMissing.table} (
                        {[...(primaryMissing.equality_columns ?? []), ...(primaryMissing.inequality_columns ?? [])].join(
                          ', '
                        )}
                        )
                      </div>
                    )}
                  </td>
                  <td className={styles.numericCell}>
                    {item.max_impact > 0 ? `${item.max_impact.toFixed(1)}%` : '-'}
                  </td>
                  <td className={styles.numericCell}>{item.occurrences.toLocaleString()}</td>
                  <td className={styles.numericCell}>
                    {item.related_lock_count > 0 ? (
                      <span className={styles.lockPill}>
                        <Lock size={11} />
                        {item.related_lock_count}
                      </span>
                    ) : (
                      '0'
                    )}
                  </td>
                  <td className={styles.numericCell}>
                    {item.last_snap_time ? new Date(item.last_snap_time).toLocaleString() : '-'}
                  </td>
                  <td onClick={(e) => e.stopPropagation()}>
                    <div className={styles.actionsCell}>
                      {primaryMissing && (
                        <Tooltip content="Copy CREATE INDEX DDL">
                          <Button size="sm" variant="secondary" fill="outline" onClick={(e) => handleCopyDDL(e, item)}>
                            {copiedKey === rowKey ? <Check size={12} /> : <Copy size={12} />}
                            <span style={{ marginLeft: 4 }}>{copiedKey === rowKey ? 'Copied' : 'DDL'}</span>
                          </Button>
                        </Tooltip>
                      )}
                      {item.trace_id && <TempoTraceLink traceId={item.trace_id} timeRange={timeRange} compact />}
                      {onOpenQueryMetrics && item.query_hash && (
                        <Tooltip content="Open Query Metrics & execution history for this query">
                          <Button
                            size="sm"
                            variant="secondary"
                            fill="outline"
                            onClick={() => onOpenQueryMetrics(item.query_hash)}
                          >
                            <BarChart3 size={12} />
                          </Button>
                        </Tooltip>
                      )}
                      <Tooltip content="Open this sample inside its Snapshot view">
                        <Button
                          size="sm"
                          variant="secondary"
                          fill="outline"
                          onClick={() => onOpenSnapshot(item.snap_id, item.sample_id)}
                        >
                          <ExternalLink size={12} />
                        </Button>
                      </Tooltip>
                    </div>
                  </td>
                </tr>
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
