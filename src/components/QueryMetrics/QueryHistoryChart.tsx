import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { css } from '@emotion/css';
import {
  createDataFrame,
  DataFrame,
  EventBusSrv,
  FieldType,
  GrafanaTheme2,
  LoadingState,
  TimeRange,
} from '@grafana/data';
import { GraphDrawStyle } from '@grafana/schema';
import { FetchResponse, getBackendSrv } from '@grafana/runtime';
import {
  Alert,
  Button,
  Combobox,
  ComboboxOption,
  LoadingPlaceholder,
  useStyles2,
} from '@grafana/ui';
import { lastValueFrom, Observable } from 'rxjs';
import { BarChart3, RefreshCw } from 'lucide-react';
import { QueryMetricsSeriesPoint } from './types';
import { MyGraph } from '../../pages/graph';
import { formatDurationMs } from '../../utils/sqlHelpers';

export type HistoryChartMetric =
  | 'execution_count'
  | 'duration_cpu'
  | 'logical_reads'
  | 'physical_writes'
  | 'spills'
  | 'memory_grant';

const CHART_METRIC_OPTIONS: Array<ComboboxOption<HistoryChartMetric>> = [
  { label: 'Executions per Bucket', value: 'execution_count' },
  { label: 'Avg Duration vs Avg CPU (ms)', value: 'duration_cpu' },
  { label: 'Avg Logical Reads', value: 'logical_reads' },
  { label: 'Avg Physical Reads & Writes', value: 'physical_writes' },
  { label: 'TempDB Spills', value: 'spills' },
  { label: 'Memory Grant vs Used (KB)', value: 'memory_grant' },
];

const getStyles = (theme: GrafanaTheme2) => ({
  toolbar: css`
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: ${theme.spacing(1)};
    margin-bottom: ${theme.spacing(1.5)};
  `,
  toolbarLeft: css`
    display: flex;
    align-items: center;
    gap: ${theme.spacing(1)};
    flex-wrap: wrap;
  `,
  kpiStrip: css`
    display: flex;
    flex-wrap: wrap;
    gap: ${theme.spacing(1)};
    margin-bottom: ${theme.spacing(1.5)};
  `,
  kpiPill: css`
    padding: ${theme.spacing(0.75, 1.25)};
    background: ${theme.colors.background.secondary};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    min-width: 110px;
  `,
  kpiLabel: css`
    font-size: 11px;
    color: ${theme.colors.text.secondary};
  `,
  kpiValue: css`
    font-size: ${theme.typography.body.fontSize};
    font-weight: ${theme.typography.fontWeightMedium};
    font-family: ${theme.typography.fontFamilyMonospace};
  `,
});

export function buildQueryHistoryDataFrames(
  points: QueryMetricsSeriesPoint[],
  metric: HistoryChartMetric
): { frames: DataFrame[]; unit: string; drawStyle: GraphDrawStyle } {
  if (points.length === 0) {
    return { frames: [], unit: 'short', drawStyle: GraphDrawStyle.Line };
  }

  const timestamps = points.map((p) => new Date(p.timestamp).getTime());

  switch (metric) {
    case 'duration_cpu':
      return {
        unit: 'ms',
        drawStyle: GraphDrawStyle.Line,
        frames: [
          createDataFrame({
            name: 'duration_cpu',
            fields: [
              { name: 'Time', type: FieldType.time, values: timestamps },
              { name: 'Avg Elapsed (ms)', type: FieldType.number, values: points.map((p) => p.avg_elapsed_ms) },
              { name: 'Avg CPU (ms)', type: FieldType.number, values: points.map((p) => p.avg_worker_ms) },
            ],
          }),
        ],
      };
    case 'logical_reads':
      return {
        unit: 'short',
        drawStyle: GraphDrawStyle.Line,
        frames: [
          createDataFrame({
            name: 'logical_reads',
            fields: [
              { name: 'Time', type: FieldType.time, values: timestamps },
              { name: 'Avg Logical Reads', type: FieldType.number, values: points.map((p) => p.avg_logical_reads) },
            ],
          }),
        ],
      };
    case 'physical_writes':
      return {
        unit: 'short',
        drawStyle: GraphDrawStyle.Line,
        frames: [
          createDataFrame({
            name: 'physical_writes',
            fields: [
              { name: 'Time', type: FieldType.time, values: timestamps },
              { name: 'Avg Physical Reads', type: FieldType.number, values: points.map((p) => p.avg_physical_reads) },
              { name: 'Avg Logical Writes', type: FieldType.number, values: points.map((p) => p.avg_logical_writes) },
            ],
          }),
        ],
      };
    case 'spills':
      return {
        unit: 'short',
        drawStyle: GraphDrawStyle.Bars,
        frames: [
          createDataFrame({
            name: 'spills',
            fields: [
              { name: 'Time', type: FieldType.time, values: timestamps },
              { name: 'Total Spills', type: FieldType.number, values: points.map((p) => p.total_spills) },
            ],
          }),
        ],
      };
    case 'memory_grant':
      return {
        unit: 'deckbytes',
        drawStyle: GraphDrawStyle.Line,
        frames: [
          createDataFrame({
            name: 'memory_grant',
            fields: [
              { name: 'Time', type: FieldType.time, values: timestamps },
              { name: 'Avg Grant (KB)', type: FieldType.number, values: points.map((p) => p.avg_grant_kb) },
              { name: 'Avg Used Grant (KB)', type: FieldType.number, values: points.map((p) => p.avg_used_grant_kb) },
            ],
          }),
        ],
      };
    case 'execution_count':
    default:
      return {
        unit: 'short',
        drawStyle: GraphDrawStyle.Bars,
        frames: [
          createDataFrame({
            name: 'execution_count',
            fields: [
              { name: 'Time', type: FieldType.time, values: timestamps },
              { name: 'Executions', type: FieldType.number, values: points.map((p) => p.execution_count) },
            ],
          }),
        ],
      };
  }
}

export const QueryHistoryChart: React.FC<{
  server: string;
  queryHash: string;
  timeRange: TimeRange;
  initialMetric?: HistoryChartMetric;
  height?: number;
  onOpenQueryMetrics?: (queryHash: string) => void;
}> = ({ server, queryHash, timeRange, initialMetric = 'execution_count', height = 260, onOpenQueryMetrics }) => {
  const styles = useStyles2(getStyles);
  const eventBus = useMemo(() => new EventBusSrv(), []);
  const [selectedMetric, setSelectedMetric] = useState<HistoryChartMetric>(initialMetric);
  const [points, setPoints] = useState<QueryMetricsSeriesPoint[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [refreshCounter, setRefreshCounter] = useState(0);

  const fetchSeries = useCallback(async () => {
    if (!server || !queryHash) {
      setPoints([]);
      return;
    }
    try {
      setLoading(true);
      setError(null);
      const params = new URLSearchParams({
        server,
        queryHash,
        start: timeRange.from.toISOString(),
        end: timeRange.to.toISOString(),
      });
      const response: Observable<FetchResponse<QueryMetricsSeriesPoint[]>> = getBackendSrv().fetch({
        url: `/api/plugins/guilhermearpassos-sqlsights-app/resources/queryMetricsSeries?${params.toString()}`,
      });
      const result = await lastValueFrom(response);
      setPoints(result.data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load execution history');
      setPoints([]);
    } finally {
      setLoading(false);
    }
  }, [server, queryHash, timeRange]);

  useEffect(() => {
    fetchSeries();
  }, [fetchSeries, refreshCounter]);

  const summary = useMemo(() => {
    let totalExecs = 0;
    let peakAvgElapsedMs = 0;
    let totalSpills = 0;
    let totalReads = 0;
    for (const p of points) {
      totalExecs += p.execution_count;
      if (p.avg_elapsed_ms > peakAvgElapsedMs) {
        peakAvgElapsedMs = p.avg_elapsed_ms;
      }
      totalSpills += p.total_spills;
      totalReads += p.total_logical_reads;
    }
    return { totalExecs, peakAvgElapsedMs, totalSpills, totalReads };
  }, [points]);

  const { frames, unit, drawStyle } = useMemo(
    () => buildQueryHistoryDataFrames(points, selectedMetric),
    [points, selectedMetric]
  );

  return (
    <div>
      <div className={styles.toolbar}>
        <div className={styles.toolbarLeft}>
          <Combobox<HistoryChartMetric>
            width={30}
            options={CHART_METRIC_OPTIONS}
            value={selectedMetric}
            onChange={(opt) => setSelectedMetric(opt?.value ?? 'execution_count')}
          />
          <Button size="sm" variant="secondary" fill="outline" onClick={() => setRefreshCounter((c) => c + 1)} disabled={loading}>
            <RefreshCw size={13} />
          </Button>
        </div>

        {onOpenQueryMetrics && queryHash && (
          <Button size="sm" variant="secondary" fill="outline" onClick={() => onOpenQueryMetrics(queryHash)}>
            <BarChart3 size={13} />
            <span style={{ marginLeft: 6 }}>Open in Query Metrics</span>
          </Button>
        )}
      </div>

      {error && (
        <Alert title="Failed to load execution history" severity="error" style={{ marginBottom: 12 }}>
          {error}
        </Alert>
      )}

      {loading ? (
        <LoadingPlaceholder text="Loading execution history..." />
      ) : points.length > 0 ? (
        <>
          <div className={styles.kpiStrip}>
            <div className={styles.kpiPill}>
              <div className={styles.kpiLabel}>Executions</div>
              <div className={styles.kpiValue}>{summary.totalExecs.toLocaleString()}</div>
            </div>
            <div className={styles.kpiPill}>
              <div className={styles.kpiLabel}>Peak Avg Duration</div>
              <div className={styles.kpiValue}>{formatDurationMs(summary.peakAvgElapsedMs)}</div>
            </div>
            <div className={styles.kpiPill}>
              <div className={styles.kpiLabel}>Logical Reads</div>
              <div className={styles.kpiValue}>{summary.totalReads.toLocaleString()}</div>
            </div>
            <div className={styles.kpiPill}>
              <div className={styles.kpiLabel}>TempDB Spills</div>
              <div className={styles.kpiValue} style={{ color: summary.totalSpills > 0 ? '#e02f44' : undefined }}>
                {summary.totalSpills.toLocaleString()}
              </div>
            </div>
          </div>

          <MyGraph
            data={frames}
            loadingState={LoadingState.Done}
            eventBus={eventBus}
            timeRange={timeRange}
            onTimeRangeChange={() => {}}
            height={height}
            drawStyle={drawStyle}
            stacked={false}
            unit={unit}
          />
        </>
      ) : (
        <Alert title="No Execution History Available" severity="info">
          No query metric time-series data was found for this query hash in the selected time range. Make sure the
          Sqlsights agent is running with query metrics collection enabled and wait one collection interval (30s default).
        </Alert>
      )}
    </div>
  );
};
