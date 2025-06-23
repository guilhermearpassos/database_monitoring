
drop index ix_samples_summary;
alter table query_samples drop column wait_event;
alter table query_samples drop column wait_time;