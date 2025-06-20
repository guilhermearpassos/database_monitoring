create index if not exists query_samples_snap_id_f_id_index
    on public.query_samples (snap_id, f_id) include (data);


create unique index if not exists snapshot_f_id_uindex
    on public.snapshot (f_id);

create index if not exists snapshot_target_id_snap_time_index
    on public.snapshot (target_id, snap_time);