alter table query_plans add analyzed bool default FALSE not null;

alter table query_plans add missing_indexes int ;
alter table query_plans add implicit_conversions int ;
alter table query_plans add large_table_scans int ;
alter table query_plans add has_problems bool default FALSE;
