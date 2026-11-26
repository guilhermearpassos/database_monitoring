import { MissingIndexWarning, PlanAffectingConvert } from '../ExecutionPlanTree/plan';

export interface PlanAnalysisItem {
  plan_handle: string;
  query_hash: string;
  sql_handle: string;
  text: string;
  database: string;
  snap_id: string;
  sample_id: string;
  last_snap_time: string;
  trace_id?: string;
  issue_count: number;
  occurrences: number;
  related_lock_count: number;
  issue_tags: string[];
  converts: PlanAffectingConvert[];
  missing_indexes: MissingIndexWarning[];
  max_impact: number;
}

export interface PlanAnalysisResponse {
  plans: PlanAnalysisItem[];
  available_databases: string[];
}
