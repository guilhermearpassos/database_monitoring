import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export type SqlsightsQueryType =
  | 'chart'
  | 'blocking_timeseries'
  | 'wait_breakdown'
  | 'snapshot-list'
  | 'snapshot'
  | 'plan_analysis'
  | 'plan_analysis_summary'
  | 'servers'
  | 'metrics'
  | 'metrics_series';

export interface MyQuery extends DataQuery {
  queryType?: SqlsightsQueryType | string;
  /** Monitored SQL Server instance alias (kept as `database` for backwards compatibility with saved dashboards). */
  database?: string;
  /** Optional list of SQL Server databases inside the selected instance. */
  databases?: string[];
  /** Optional list of plan issue tags (seek_plan_convert, missing_index, cardinality_convert, etc.). */
  issueTags?: string[];
  /** Plan analysis: only include plans involved in locks. */
  onlyLocks?: boolean;
  /** Plan analysis: minimum missing-index impact percentage (0-100). */
  minImpact?: number;
  /** Plan analysis: substring search across SQL, conversion expression, or table. */
  search?: string;
  /** Snapshot list / samples: only include snapshots/samples with blocking or waiting. */
  onlyBlocking?: boolean;
  /** Optional wait event filter (e.g., LCK_M_X, CXPACKET, cpu). */
  waitEvent?: string;
  /** Chart & wait breakdown metric mode. */
  metric?: 'connections' | 'wait_time_ms';
  /** Selected series for blocking_timeseries (connections, blockers, waiters, waitDurationSec). */
  blockingSeries?: string[];
  /** Snapshot ID for `snapshot` query type. */
  snapshotID?: string;
  /** Query hash for `metrics_series` query type. */
  queryHash?: string;
  /** Selected counter metrics for `metrics_series`. */
  metrics?: string[];
}

export const DEFAULT_QUERY: Partial<MyQuery> = {
  queryType: 'chart',
  database: '',
  metric: 'connections',
};

export interface DataPoint {
  Time: number;
  Value: number;
}

export interface DataSourceResponse {
  datapoints: DataPoint[];
}

export interface MyDataSourceOptions extends DataSourceJsonData {
  path?: string;
}

export interface MySecureJsonData {
  apiKey?: string;
}
