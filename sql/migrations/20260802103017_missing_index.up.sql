
create index query_plans_plan_handle_index
    on query_plans (plan_handle);
create index query_samples_snap_id_plan_handle_index
    on query_samples (snap_id, plan_handle);
create index snapshot_target_id_snap_time_id_index
    on snapshot (target_id, snap_time, id);