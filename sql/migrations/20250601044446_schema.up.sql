



create table if not exists target_type (
                                           id int primary key ,
                                           dsc_type varchar(50)
    );
do $do$
begin
        if not exists( select 1 from target_type)
            then
                insert into target_type (id, dsc_type) values (1, 'mssql');

end if;

end;
    $do$;

create table IF NOT EXISTS target (
                                      id int generated always as identity  primary key ,
                                      host varchar(100),
    type_id int references target_type (id),
    agent_version varchar(50)
    );

create table IF NOT EXISTS  snapshot (
                                         id bigint generated always as identity primary key ,
                                         f_id uuid,
                                         snap_time timestamp,
                                         target_id int references target (id)
    );


create table IF NOT EXISTS query_samples (
                                             id bigint generated always as identity primary key ,
                                             f_id varchar(200),
                                             snap_id bigint references snapshot(id),
    sql_handle varchar(100),
    blocked bool,
    blocker bool,
    plan_handle varchar(100),
    data bytea
    );
create table IF NOT EXISTS query_stat_snapshot (
                                                   id bigint generated always as identity primary key ,
                                                   f_id uuid,
                                                   target_id int references target (id),
    collected_at timestamp

    );
create table IF NOT EXISTS query_stat_sample (
                                                 id bigint generated always as identity primary key ,
                                                 snap_id bigint references query_stat_snapshot (id),
    sql_handle varchar(100),
    data bytea
    );

create table IF NOT EXISTS query_plans (
                                           id bigint generated always as identity primary key ,
                                           plan_handle varchar(100),
    plan_xml varchar,
    target_id int references target (id)
    );
