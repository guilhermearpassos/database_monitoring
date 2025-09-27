alter table query_samples
    drop constraint query_samples_snap_id_fkey,
        add constraint query_samples_snap_id_fkey
        foreign key (snap_id) references snapshot on delete cascade;

alter table query_stat_sample
    drop constraint query_stat_sample_snap_id_fkey,
        add constraint query_stat_sample_snap_id_fkey
        foreign key (snap_id) references query_stat_snapshot on delete cascade;