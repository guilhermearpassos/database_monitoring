export interface QueryMetricItem {
  query_hash: string;
  text: string;
  database: string;
  last_execution_time: string;
  last_elapsed_ms: number;
  execution_count: number;
  total_elapsed_ms: number;
  avg_elapsed_ms: number;
  total_worker_ms: number;
  avg_worker_ms: number;
  total_clr_ms: number;
  total_logical_reads: number;
  avg_logical_reads: number;
  total_physical_reads: number;
  avg_physical_reads: number;
  total_logical_writes: number;
  avg_logical_writes: number;
  total_rows: number;
  avg_rows: number;
  total_spills: number;
  avg_spills: number;
  total_grant_kb: number;
  avg_grant_kb: number;
  avg_used_grant_kb: number;
  avg_ideal_grant_kb: number;
  memory_grant_waste_pct: number;
  avg_dop: number;
  cpu_share_pct: number;
  elapsed_share_pct: number;
  reads_share_pct: number;
  issue_flags: string[];
}

export interface QueryMetricsResponse {
  metrics: QueryMetricItem[];
  available_databases: string[];
}

export interface QueryMetricsSeriesPoint {
  timestamp: string;
  execution_count: number;
  total_elapsed_ms: number;
  avg_elapsed_ms: number;
  total_worker_ms: number;
  avg_worker_ms: number;
  total_logical_reads: number;
  avg_logical_reads: number;
  avg_physical_reads: number;
  avg_logical_writes: number;
  avg_rows: number;
  total_spills: number;
  avg_spills: number;
  avg_grant_kb: number;
  avg_used_grant_kb: number;
}

export type QueryMetricsSortField =
  | 'total_elapsed_ms'
  | 'avg_elapsed_ms'
  | 'total_worker_ms'
  | 'avg_worker_ms'
  | 'execution_count'
  | 'total_logical_reads'
  | 'avg_logical_reads'
  | 'total_physical_reads'
  | 'total_spills'
  | 'avg_grant_kb'
  | 'avg_rows'
  | 'last_execution_time';

export type SortDirection = 'asc' | 'desc';

export interface MetricIssueMeta {
  value: string;
  label: string;
  shortLabel: string;
  severity: 'error' | 'warning' | 'info';
  description: string;
}

export const METRIC_ISSUE_FLAGS: MetricIssueMeta[] = [
  {
    value: 'spills',
    label: 'TempDB Spills',
    shortLabel: 'Spills',
    severity: 'error',
    description: 'Sort or hash operators spilled to TempDB (insufficient memory grant or bad row estimate).',
  },
  {
    value: 'memory_overgrant',
    label: 'Severe Memory Over-Grant',
    shortLabel: 'Mem Overgrant',
    severity: 'warning',
    description: 'Granted >=10 MB per execution with >=80% unused memory (severe cardinality overestimate).',
  },
];

const ISSUE_META_MAP: Record<string, MetricIssueMeta> = Object.fromEntries(
  METRIC_ISSUE_FLAGS.map((m) => [m.value, m])
);

export function getMetricIssueMeta(flag: string): MetricIssueMeta {
  return (
    ISSUE_META_MAP[flag] ?? {
      value: flag,
      label: flag,
      shortLabel: flag,
      severity: 'info',
      description: flag,
    }
  );
}
