import {ParsedExecutionPlan} from "../ExecutionPlanTree/plan";
import {BlockingChain} from "./blockingChain";

export interface QuerySample {
    snap_id: string;
    sid: string;
    query: string;
    status: string;
    execution_time: string;
    is_blocker: boolean;
    sample_id: string;
    query_hash: string;
    database: string;
    sql_handle: string;
    session_login_time: string;
    session_host: string;
    session_status: string;
    session_program_name: string;
    session_last_request_start: string;
    session_last_request_end: string;
    session_client_ip: string;
    user: string;
}

export interface QueryDetails {
    plan?: ParsedExecutionPlan;
    blocking_chain?: BlockingChain
    query_sample: QuerySample
}
