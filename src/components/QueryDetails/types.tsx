import {ParsedExecutionPlan} from "../ExecutionPlanTree/plan";
import {BlockingChain} from "./blockingChain";

export interface QuerySample {
    sid: string;
    query: string;
    status: string;
    execution_time: string;
    is_blocker: boolean;
    sample_id?: string;
    query_hash: string;
}

export interface QueryDetails {
    plan?: ParsedExecutionPlan;
    blocking_chain?: BlockingChain
    query_sample: QuerySample
}
