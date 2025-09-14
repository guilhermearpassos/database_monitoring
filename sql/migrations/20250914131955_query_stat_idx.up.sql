
create index idx_stat_snap on public.query_stat_snapshot (target_id, collected_at, id);
create index idx_stat_sample on public.query_stat_sample (snap_id, sql_handle);