alter table query_plans drop column analyzed;

alter table query_plans drop column missing_indexes;
alter table query_plans drop column implicit_conversions;
alter table query_plans drop column large_table_scans;
alter table query_plans drop  column has_problems ;
