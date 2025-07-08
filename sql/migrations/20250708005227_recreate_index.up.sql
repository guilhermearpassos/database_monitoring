drop index if exists query_samples_snap_id_f_id_index;
create index if not exists query_samples_snap_id_f_id_index
    on public.query_samples (snap_id, f_id);