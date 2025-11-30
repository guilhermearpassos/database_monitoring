create table public.warnings
(
    id           integer generated always as identity,
    name         varchar(100) not null,
    target_id int not null,
    warning_type varchar(30)  not null,
    warning_data bytea
);
create unique index ix_warning_name on public.warnings (target_id, name);
create index ix_warning_type on public.warnings (target_id, warning_type, name);

