export interface DeadlockExecutionFrame {
  proc_name: string;
  line: string;
  sql_handle: string;
  sql: string;
}

export interface DeadlockProcess {
  id: string;
  is_victim: boolean;
  spid: string;
  ecid: string;
  status: string;
  wait_time_ms: number;
  wait_resource: string;
  lock_mode: string;
  log_used: number;
  priority: string;
  tran_count: string;
  transaction_name: string;
  isolation_level: string;
  client_app: string;
  host_name: string;
  login_name: string;
  current_db: string;
  current_db_name: string;
  frames: DeadlockExecutionFrame[];
  input_buf: string;
  active_sql: string;
}

export interface DeadlockResourceParticipant {
  process_id: string;
  mode: string;
  request_type: string;
}

export interface DeadlockResource {
  id: string;
  type: string;
  dbid: string;
  database: string;
  object_name: string;
  short_object: string;
  index_name: string;
  hobt_id: string;
  file_id: string;
  page_id: string;
  slot: string;
  mode: string;
  owners: DeadlockResourceParticipant[];
  waiters: DeadlockResourceParticipant[];
}

export interface DeadlockCycleStep {
  kind: 'hold' | 'wait' | 'victim' | string;
  text: string;
}

export interface DeadlockDiagnosis {
  pattern: string;
  pattern_label: string;
  severity: 'error' | 'warning' | 'purple' | 'info' | string;
  summary: string;
  recommendations: string[];
  steps: DeadlockCycleStep[];
}

export interface DeadlockItem {
  id: string;
  server: string;
  timestamp: string;
  xml_hash: string;
  fingerprint: string;
  pattern_occurrences: number;
  victim_spid: string;
  databases: string[];
  max_wait_ms: number;
  processes: DeadlockProcess[];
  resources: DeadlockResource[];
  diagnosis: DeadlockDiagnosis;
  xml_report: string;
}

export interface DeadlocksResponse {
  deadlocks: DeadlockItem[];
  available_databases: string[];
}

export type DeadlocksViewMode = 'events' | 'patterns';
export type DeadlocksSortField = 'timestamp' | 'occurrences' | 'max_wait_ms' | 'pattern';
export type SortDirection = 'asc' | 'desc';

export const DEADLOCK_PATTERN_OPTIONS = [
  { label: 'Reverse-Order Lock Cycle', value: 'reverse_order' },
  { label: 'Key / Bookmark Lookup', value: 'bookmark_lookup' },
  { label: 'Lock Conversion (S -> X)', value: 'lock_upgrade' },
  { label: 'Parallelism Exchange', value: 'parallel_exchange' },
  { label: 'Other Lock Cycle', value: 'generic_cycle' },
];

export const LOCK_MODE_DESCRIPTIONS: Record<string, string> = {
  S: 'Shared (read lock)',
  U: 'Update (read-for-modify lock)',
  X: 'Exclusive (write lock)',
  IS: 'Intent Shared',
  IU: 'Intent Update',
  IX: 'Intent Exclusive',
  SIX: 'Shared with Intent Exclusive',
  'Sch-S': 'Schema Stability',
  'Sch-M': 'Schema Modification',
  'RangeS-S': 'Serializable Range Shared',
  'RangeS-U': 'Serializable Range Update',
  'RangeX-X': 'Serializable Range Exclusive',
};
