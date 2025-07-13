
alter table query_samples add sid varchar(50);
alter table query_samples add connection_id varchar(50);
alter table query_samples add transaction_id varchar(100);
alter table query_samples add block_ms bigint;
alter table query_samples add block_count int;
