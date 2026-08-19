alter table query_samples
    add trace_id varchar(32);

create index query_samples_trace_id_snap_id_id_index
    on query_samples (trace_id, snap_id, id);

