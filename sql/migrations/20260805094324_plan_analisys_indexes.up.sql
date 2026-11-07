
create index query_plans_analyzed_has_problems_plan_handle_target_id_index
    on query_plans (analyzed, has_problems, plan_handle, target_id);
