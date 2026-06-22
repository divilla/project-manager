drop procedure if exists public.sp_task_requirement_recalculate;
drop procedure if exists public.sp_task_phase_recalculate;
drop procedure if exists public.sp_requirement_to_history;
drop procedure if exists public.sp_task_to_history;

drop table if exists public.requirement_history;
drop table if exists public.requirement;
drop table if exists public.task_history;
drop table if exists public.task;
drop table if exists public.task_phase;
drop table if exists public.task_type;
drop table if exists public.project;

create table public.project
(
    id   bigint generated always as identity primary key,
    name text not null
);

create table public.task_type
(
    slug     text primary key default 'task' not null,
    priority integer default 0 not null
);

create table public.task_phase
(
    slug     text primary key default 'backlog' not null,
    priority integer default 0 not null
);

create table public.task
(
    id          bigint generated always as identity primary key,
    version     smallint default 0 not null,
    task_type   text default 'task' not null,
    name        text not null,
    description text,
    difficulty  smallint default 1 not null,
    priority    smallint default 0 not null,
    task_phase  text default 'backlog' not null,
    parent_id   bigint,
    project_id  bigint not null,
    done_req    smallint default 0 not null,
    total_req   smallint default 0 not null,
    created     timestamp with time zone default now() not null,
    modified    timestamp with time zone default now() not null
);

create table public.task_history
(
    id          bigint not null,
    version     smallint not null,
    task_type   text not null,
    name        text not null,
    description text,
    parent_id   bigint,
    modified    timestamp with time zone not null,
    deleted     boolean default false not null,
    primary key (id, version)
);

create table public.requirement
(
    id         bigint generated always as identity primary key,
    version    smallint default 0 not null,
    definition text not null,
    done       boolean default false not null,
    task_id    bigint not null,
    created    timestamp with time zone default now() not null,
    modified   timestamp with time zone default now() not null
);

create table public.requirement_history
(
    id         bigint not null,
    version    smallint not null,
    definition text not null,
    modified   timestamp with time zone not null,
    deleted    boolean not null,
    primary key (id, version)
);

create procedure public.sp_task_to_history(IN _id bigint, IN _deleted boolean)
    language plpgsql
as
$$
begin
    insert into public.task_history (
        id, version, task_type, name, description, parent_id, modified, deleted
    )
    select
        id, version, task_type, name, description, parent_id, modified, _deleted
    from public.task
    where id = _id;
end;
$$;

create procedure public.sp_requirement_to_history(IN _id bigint, IN _deleted boolean)
    language plpgsql
as
$$
begin
    insert into public.requirement_history (
        id, version, definition, modified, deleted
    )
    select
        id, version, definition, modified, _deleted
    from public.requirement
    where id = _id;
end;
$$;

create procedure sp_task_phase_recalculate(IN _task_id bigint)
    language plpgsql
as
$$
declare
    _priority smallint;
    _slug text;
    _parent_id bigint;
begin
    loop
        select parent_id into _parent_id from task where id=_task_id;
        if _parent_id is null then
            exit;
        end if;

        select min(tp.priority) into _priority from task t join task_phase tp on t.task_phase=tp.slug where parent_id=_parent_id;
        select slug into _slug from task_phase where priority=_priority;

        update public.task
        set
            task_phase=_slug
        where id=_parent_id;

        _task_id=_parent_id;
    end loop;
end;
$$;

create procedure sp_task_requirement_recalculate(IN _task_id bigint)
    language plpgsql
as
$$
declare
    _done_req smallint;
    _total_req smallint;
    _done_task smallint;
    _total_task smallint;
begin
    loop
        select count(*) into _done_req from requirement where task_id=_task_id and done=true;
        select count(*) into _total_req from requirement where task_id=_task_id;

        select
            coalesce(sum(done_req), 0),
            coalesce(sum(total_req), 0)
        into
            _done_task,
            _total_task
        from public.task
        where parent_id=_task_id;

        update public.task
        set
            done_req=_done_req+_done_task,
            total_req=_total_req+_total_task
        where id=_task_id;

        select parent_id into _task_id from task where id=_task_id;
        if _task_id is null then
            exit;
        end if;
    end loop;
end;
$$;

