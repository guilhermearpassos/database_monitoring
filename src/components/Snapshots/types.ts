export interface SnapshotSummaryItem {
  id: string;
  timestamp: string;
  connections: number;
  waiters: number;
  blockers: number;
  wait_duration_ms: number;
  avg_duration: number;
  max_duration: number;
  connections_by_wait_event: Record<string, number>;
  time_ms_by_wait_event: Record<string, number>;
}

export interface SnapshotSampleItem {
  sample_id: string;
  snap_id: string;
  session_id: string;
  status: string;
  text: string;
  user: string;
  elapsed_ms: number;
  wait_event: string;
  wait_time_ms: number;
  wait_resource: string;
  database: string;
  is_blocked: boolean;
  is_blocker: boolean;
  blocked_by: string;
  blocked_sessions: string[];
  trace_id?: string;
  query_hash: string;
  plan_handle: string;
  program_name: string;
  host_name: string;
  client_ip: string;
  session_status: string;
}
