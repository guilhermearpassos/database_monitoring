import React from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2, TimeRange } from '@grafana/data';
import { Button, Card, Grid, useStyles2 } from '@grafana/ui';
import { Activity, FileSearch } from 'lucide-react';
import { getMetricIssueMeta, QueryMetricItem } from './types';
import { QueryHistoryChart } from './QueryHistoryChart';
import { formatDurationMs } from '../../utils/sqlHelpers';
import { SqlQueryCard } from '../SqlQueryCard';

const getStyles = (theme: GrafanaTheme2) => ({
  headerActions: css`
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: ${theme.spacing(1)};
    margin-bottom: ${theme.spacing(2)};
  `,
  section: css`
    margin-bottom: ${theme.spacing(2)};
  `,
  kpiGrid: css`
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: ${theme.spacing(1)};
    margin-bottom: ${theme.spacing(2)};
  `,
  kpiBox: css`
    padding: ${theme.spacing(1, 1.5)};
    background: ${theme.colors.background.secondary};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
  `,
  metaLabel: css`
    color: ${theme.colors.text.secondary};
    font-size: ${theme.typography.bodySmall.fontSize};
    margin: 0 0 2px 0;
  `,
  metaValue: css`
    color: ${theme.colors.text.primary};
    font-family: ${theme.typography.fontFamilyMonospace};
    font-size: ${theme.typography.bodySmall.fontSize};
    word-break: break-all;
    margin: 0;
  `,
  issueList: css`
    display: flex;
    flex-direction: column;
    gap: ${theme.spacing(1)};
  `,
  issueItem: css`
    padding: ${theme.spacing(1, 1.25)};
    border-radius: ${theme.shape.radius.default};
    border: 1px solid ${theme.colors.border.weak};
    background: ${theme.colors.background.secondary};
    font-size: ${theme.typography.bodySmall.fontSize};
  `,
});

export const QueryMetricDrawer: React.FC<{
  metric: QueryMetricItem;
  server: string;
  timeRange: TimeRange;
  onOpenPlanAnalysis: (item: QueryMetricItem) => void;
  onOpenSnapshots: (item: QueryMetricItem) => void;
}> = ({ metric, server, timeRange, onOpenPlanAnalysis, onOpenSnapshots }) => {
  const styles = useStyles2(getStyles);

  return (
    <div>
      <div className={styles.headerActions}>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
          <Button size="sm" variant="secondary" fill="outline" onClick={() => onOpenPlanAnalysis(metric)}>
            <FileSearch size={13} />
            <span style={{ marginLeft: 6 }}>View in Plan Analysis</span>
          </Button>
          <Button size="sm" variant="secondary" fill="outline" onClick={() => onOpenSnapshots(metric)}>
            <Activity size={13} />
            <span style={{ marginLeft: 6 }}>View Snapshots</span>
          </Button>
        </div>
      </div>

      <div className={styles.kpiGrid}>
        <div className={styles.kpiBox}>
          <div className={styles.metaLabel}>Executions</div>
          <div className={styles.metaValue}>{metric.execution_count.toLocaleString()}</div>
        </div>
        <div className={styles.kpiBox}>
          <div className={styles.metaLabel}>Avg / Total Duration</div>
          <div className={styles.metaValue}>
            {formatDurationMs(metric.avg_elapsed_ms)} / {formatDurationMs(metric.total_elapsed_ms)}
          </div>
        </div>
        <div className={styles.kpiBox}>
          <div className={styles.metaLabel}>Avg / Total CPU</div>
          <div className={styles.metaValue}>
            {formatDurationMs(metric.avg_worker_ms)} / {formatDurationMs(metric.total_worker_ms)}
          </div>
        </div>
        <div className={styles.kpiBox}>
          <div className={styles.metaLabel}>Avg Logical Reads</div>
          <div className={styles.metaValue}>{Math.round(metric.avg_logical_reads).toLocaleString()}</div>
        </div>
        <div className={styles.kpiBox}>
          <div className={styles.metaLabel}>TempDB Spills</div>
          <div className={styles.metaValue} style={{ color: metric.total_spills > 0 ? '#e02f44' : undefined }}>
            {metric.total_spills.toLocaleString()}
          </div>
        </div>
        <div className={styles.kpiBox}>
          <div className={styles.metaLabel}>Workload Share</div>
          <div className={styles.metaValue}>
            CPU {metric.cpu_share_pct.toFixed(1)}% · Reads {metric.reads_share_pct.toFixed(1)}%
          </div>
        </div>
      </div>

      {(metric.issue_flags ?? []).length > 0 && (
        <div className={styles.section}>
          <Card>
            <Card.Heading>Detected Issues ({metric.issue_flags.length})</Card.Heading>
            <Card.Description>
              <div className={styles.issueList}>
                {metric.issue_flags.map((flag) => {
                  const meta = getMetricIssueMeta(flag);
                  return (
                    <div key={flag} className={styles.issueItem}>
                      <strong>{meta.label}:</strong> {meta.description}
                    </div>
                  );
                })}
              </div>
            </Card.Description>
          </Card>
        </div>
      )}

      <div className={styles.section}>
        <Card>
          <Card.Heading>Execution History Over Time</Card.Heading>
          <Card.Description>
            <QueryHistoryChart
              server={server}
              queryHash={metric.query_hash}
              timeRange={timeRange}
              initialMetric="duration_cpu"
              height={260}
            />
          </Card.Description>
        </Card>
      </div>

      <div className={styles.section}>
        <SqlQueryCard sql={metric.text} heading="SQL Statement" />
      </div>

      <Card>
        <Card.Heading>Resource &amp; Execution Breakdown</Card.Heading>
        <Card.Description>
          <Grid columns={3}>
            <div>
              <p className={styles.metaLabel}>Database</p>
              <p className={styles.metaValue}>{metric.database || '-'}</p>
            </div>
            <div>
              <p className={styles.metaLabel}>Query Hash</p>
              <p className={styles.metaValue}>{metric.query_hash}</p>
            </div>
            <div>
              <p className={styles.metaLabel}>Last Executed</p>
              <p className={styles.metaValue}>
                {metric.last_execution_time ? new Date(metric.last_execution_time).toLocaleString() : '-'}
              </p>
            </div>
            <div>
              <p className={styles.metaLabel}>Total Logical / Physical Reads</p>
              <p className={styles.metaValue}>
                {metric.total_logical_reads.toLocaleString()} / {metric.total_physical_reads.toLocaleString()}
              </p>
            </div>
            <div>
              <p className={styles.metaLabel}>Total / Avg Logical Writes</p>
              <p className={styles.metaValue}>
                {metric.total_logical_writes.toLocaleString()} / {metric.avg_logical_writes.toFixed(1)}
              </p>
            </div>
            <div>
              <p className={styles.metaLabel}>Avg Rows / Avg DOP</p>
              <p className={styles.metaValue}>
                {metric.avg_rows.toFixed(1)} / {metric.avg_dop.toFixed(1)}
              </p>
            </div>
            <div>
              <p className={styles.metaLabel}>Avg Memory Grant (KB)</p>
              <p className={styles.metaValue}>{metric.avg_grant_kb.toFixed(1)}</p>
            </div>
            <div>
              <p className={styles.metaLabel}>Avg Used / Ideal Grant (KB)</p>
              <p className={styles.metaValue}>
                {metric.avg_used_grant_kb.toFixed(1)} / {metric.avg_ideal_grant_kb.toFixed(1)}
              </p>
            </div>
            <div>
              <p className={styles.metaLabel}>Memory Grant Waste</p>
              <p className={styles.metaValue}>{metric.memory_grant_waste_pct.toFixed(1)}%</p>
            </div>
          </Grid>
        </Card.Description>
      </Card>
    </div>
  );
};
