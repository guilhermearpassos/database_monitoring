
alter table query_samples add column wait_event varchar(50) default '';
alter table query_samples add column wait_time bigint default 0;
create index ix_samples_summary ON query_samples (wait_event);