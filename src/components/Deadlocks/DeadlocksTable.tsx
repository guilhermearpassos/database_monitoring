import React, { useMemo, useState } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2 } from '@grafana/data';
import { Button, RadioButtonGroup, Tooltip, useStyles2 } from '@grafana/ui';
import { ArrowDown, ArrowUp, ArrowUpDown, Download } from 'lucide-react';
import { DeadlockItem, DeadlocksSortField, DeadlocksViewMode, SortDirection } from './types';
import { downloadTextFile, formatDurationMs } from '../../utils/sqlHelpers';

const VIEW_MODE_OPTIONS: Array<{ label: string; value: DeadlocksViewMode }> = [
  { label: 'Events', value: 'events' },
  { label: 'Grouped Patterns', value: 'patterns' },
];

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
      padding: ${theme.spacing(1.1, 1.25)};
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
  badge: css`
    display: inline-flex;
    align-items: center;
    padding: 2px 7px;
    border-radius: ${theme.shape.radius.default};
    font-size: 11px;
    font-weight: ${theme.typography.fontWeightMedium};
    white-space: nowrap;
    border: 1px solid transparent;
  `,
  badgeError: css`
    background: ${theme.colors.error.transparent};
    color: ${theme.colors.error.text};
    border-color: ${theme.colors.error.border};
  `,
  badgeWarning: css`
    background: ${theme.colors.warning.transparent};
    color: ${theme.colors.warning.text};
    border-color: ${theme.colors.warning.border};
  `,
  badgePurple: css`
    background: rgba(163, 82, 204, 0.16);
    color: #c78ce8;
    border-color: rgba(163, 82, 204, 0.45);
  `,
  badgeInfo: css`
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
    font-size: 11px;
    white-space: nowrap;
  `,
  sqlPreview: css`
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: 12px;
    max-width: 360px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  `,
  subLine: css`
    margin-top: 3px;
    font-size: 11px;
    color: ${theme.colors.text.secondary};
    font-family: ${theme.typography.fontFamilyMonospace};
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

function escapeCSV(val: unknown): string {
  const s = String(val ?? '');
  if (/[",\n\r]/.test(s)) {
    return `"${s.replace(/"/g, '""')}"`;
  }
  return s;
}

export const DeadlocksTable: React.FC<{
  deadlocks: DeadlockItem[];
  viewMode: DeadlocksViewMode;
  onViewModeChange: (mode: DeadlocksViewMode) => void;
  sortField: DeadlocksSortField;
  sortDir: SortDirection;
  onSortChange: (field: DeadlocksSortField) => void;
  onSelectDeadlock: (item: DeadlockItem) => void;
}> = ({
  deadlocks,
  viewMode,
  onViewModeChange,
  sortField,
  sortDir,
  onSortChange,
  onSelectDeadlock,
}) => {
  const styles = useStyles2(getStyles);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(25);

  const displayedItems = useMemo(() => {
    if (viewMode !== 'patterns') {
      return deadlocks;
    }
    const byFp = new Map<string, DeadlockItem>();
    for (const dl of deadlocks) {
      const key = dl.fingerprint || dl.id;
      const existing = byFp.get(key);
      if (!existing) {
        byFp.set(key, { ...dl, pattern_occurrences: 1 });
      } else {
        existing.pattern_occurrences += 1;
        if (new Date(dl.timestamp).getTime() > new Date(existing.timestamp).getTime()) {
          byFp.set(key, { ...dl, pattern_occurrences: existing.pattern_occurrences });
        }
      }
    }
    return Array.from(byFp.values());
  }, [deadlocks, viewMode]);

  const sorted = useMemo(() => {
    const copy = [...displayedItems];
    const dir = sortDir === 'asc' ? 1 : -1;
    copy.sort((a, b) => {
      switch (sortField) {
        case 'occurrences':
          return (a.pattern_occurrences - b.pattern_occurrences) * dir;
        case 'max_wait_ms':
          return (a.max_wait_ms - b.max_wait_ms) * dir;
        case 'pattern':
          return (a.diagnosis?.pattern_label || '').localeCompare(b.diagnosis?.pattern_label || '') * dir;
        case 'timestamp':
        default:
          return (new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()) * dir;
      }
    });
    return copy;
  }, [displayedItems, sortField, sortDir]);

  const totalPages = Math.max(1, Math.ceil(sorted.length / pageSize));
  const safePage = Math.min(page, totalPages);
  const pageItems = useMemo(() => {
    const start = (safePage - 1) * pageSize;
    return sorted.slice(start, start + pageSize);
  }, [sorted, safePage, pageSize]);

  const renderSortIcon = (field: DeadlocksSortField) => {
    if (sortField !== field) {
      return <ArrowUpDown size={12} />;
    }
    return sortDir === 'asc' ? <ArrowUp size={12} /> : <ArrowDown size={12} />;
  };

  const severityClass = (sev?: string) => {
    switch (sev) {
      case 'warning':
        return styles.badgeWarning;
      case 'purple':
        return styles.badgePurple;
      case 'info':
        return styles.badgeInfo;
      case 'error':
      default:
        return styles.badgeError;
    }
  };

  const handleExportCSV = () => {
    const header = ['Timestamp', 'Pattern', 'Occurrences', 'Databases', 'VictimSPID', 'MaxWaitMs', 'Resources', 'VictimSQL'];
    const rows = sorted.map((d) => {
      const victim = (d.processes ?? []).find((p) => p.is_victim) ?? d.processes?.[0];
      const resList = (d.resources ?? [])
        .map((r) => (r.index_name ? `${r.short_object}[${r.index_name}]` : r.short_object))
        .join(' | ');
      return [
        d.timestamp,
        d.diagnosis?.pattern_label || d.diagnosis?.pattern || '',
        d.pattern_occurrences,
        (d.databases ?? []).join(';'),
        d.victim_spid,
        d.max_wait_ms,
        resList,
        (victim?.active_sql || '').replace(/\s+/g, ' ').trim(),
      ]
        .map(escapeCSV)
        .join(',');
    });
    downloadTextFile(
      `sqlsights-deadlocks-${new Date().toISOString().slice(0, 19)}.csv`,
      [header.join(','), ...rows].join('\n'),
      'text/csv;charset=utf-8'
    );
  };

  if (deadlocks.length === 0) {
    return (
      <div className={styles.wrapper}>
        <div className={styles.emptyState}>
          No deadlocks captured for the selected server, filters, and time range.
        </div>
      </div>
    );
  }

  return (
    <div className={styles.wrapper}>
      <div className={styles.toolbar}>
        <span>
          Showing <strong>{(safePage - 1) * pageSize + 1}</strong>–
          <strong>{Math.min(safePage * pageSize, sorted.length)}</strong> of <strong>{sorted.length}</strong>{' '}
          {viewMode === 'patterns' ? 'distinct patterns' : 'deadlocks'}
        </span>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexWrap: 'wrap' }}>
          <RadioButtonGroup<DeadlocksViewMode>
            size="sm"
            options={VIEW_MODE_OPTIONS}
            value={viewMode}
            onChange={(mode) => {
              onViewModeChange(mode);
              setPage(1);
            }}
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
              <th>
                <button
                  type="button"
                  className={styles.sortHeaderBtn}
                  onClick={() => onSortChange(viewMode === 'patterns' ? 'occurrences' : 'timestamp')}
                >
                  {viewMode === 'patterns' ? 'Occurrences' : 'Timestamp'}{' '}
                  {renderSortIcon(viewMode === 'patterns' ? 'occurrences' : 'timestamp')}
                </button>
              </th>
              <th>
                <button type="button" className={styles.sortHeaderBtn} onClick={() => onSortChange('pattern')}>
                  Pattern {renderSortIcon('pattern')}
                </button>
              </th>
              <th>Database</th>
              <th>Contested Resources</th>
              <th>Victim vs Survivor(s)</th>
              <th>Victim SQL</th>
              <th>
                <button type="button" className={styles.sortHeaderBtn} onClick={() => onSortChange('max_wait_ms')}>
                  Max Wait {renderSortIcon('max_wait_ms')}
                </button>
              </th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {pageItems.map((item) => {
              const victim = (item.processes ?? []).find((p) => p.is_victim) ?? item.processes?.[0];
              const survivors = (item.processes ?? []).filter((p) => !p.is_victim);
              return (
                <tr key={item.id} onClick={() => onSelectDeadlock(item)}>
                  <td>
                    {viewMode === 'patterns' ? (
                      <>
                        <span className={`${styles.badge} ${styles.badgeError}`}>{item.pattern_occurrences}×</span>
                        <div className={styles.subLine}>Last: {new Date(item.timestamp).toLocaleString()}</div>
                      </>
                    ) : (
                      <>
                        <div className={styles.mono}>{new Date(item.timestamp).toLocaleString()}</div>
                        {item.pattern_occurrences > 1 && (
                          <div className={styles.subLine}>{item.pattern_occurrences}× same pattern</div>
                        )}
                      </>
                    )}
                  </td>
                  <td>
                    <span className={`${styles.badge} ${severityClass(item.diagnosis?.severity)}`}>
                      {item.diagnosis?.pattern_label || item.diagnosis?.pattern || 'Deadlock'}
                    </span>
                  </td>
                  <td>
                    <span className={styles.dbBadge}>{(item.databases ?? []).join(', ') || '-'}</span>
                  </td>
                  <td>
                    {(item.resources ?? []).slice(0, 3).map((r) => (
                      <div key={r.id} className={styles.mono} style={{ fontSize: 11.5 }}>
                        {r.short_object || r.object_name}
                        {r.index_name ? <span style={{ opacity: 0.7 }}> [{r.index_name}]</span> : null}
                      </div>
                    ))}
                  </td>
                  <td>
                    <div>
                      <span className={`${styles.badge} ${styles.badgeError}`}>
                        Victim SPID {victim?.spid || item.victim_spid || '?'}
                      </span>{' '}
                      <span style={{ fontSize: 11, opacity: 0.8 }}>{victim?.client_app}</span>
                    </div>
                    <div className={styles.subLine}>
                      vs {survivors.map((s) => `SPID ${s.spid} (${s.client_app})`).join(', ') || '-'}
                    </div>
                  </td>
                  <td>
                    <div className={styles.sqlPreview} title={victim?.active_sql}>
                      {victim?.active_sql || victim?.input_buf || '-'}
                    </div>
                    <div className={styles.subLine}>{victim?.isolation_level || ''}</div>
                  </td>
                  <td className={styles.mono}>{formatDurationMs(item.max_wait_ms)}</td>
                  <td onClick={(e) => e.stopPropagation()}>
                    <Tooltip content="Download SSMS-compatible .xdl file">
                      <Button
                        size="sm"
                        variant="secondary"
                        fill="outline"
                        onClick={() =>
                          downloadTextFile(
                            `deadlock-${item.server || 'sql'}-${item.id.slice(0, 8)}.xdl`,
                            item.xml_report,
                            'application/xml;charset=utf-8'
                          )
                        }
                      >
                        <Download size={12} />
                        <span style={{ marginLeft: 4 }}>.xdl</span>
                      </Button>
                    </Tooltip>
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
