drop procedure if exists public.sp_epic_to_history;
drop procedure if exists public.sp_change_to_history;
drop procedure if exists public.sp_requirement_to_history;
drop procedure if exists public.sp_change_requirement_recalculate;
drop procedure if exists public.sp_epic_requirement_recalculate;

drop view if exists public.vw_project;
drop view if exists public.vw_change;

drop table if exists public.requirement_history cascade;
drop table if exists public.requirement cascade;
drop table if exists public.change_history cascade;
drop table if exists public.change cascade;
drop table if exists public.change_phase cascade;
drop table if exists public.change_type cascade;
drop table if exists public.epic_history cascade;
drop table if exists public.epic cascade;
drop table if exists public.project cascade;

create table public.project
(
    id       bigint generated always as identity primary key,
    name     text                                   not null,
    created  timestamp with time zone default now() not null,
    modified timestamp with time zone default now() not null
);

create table public.epic
(
    id         bigint generated always as identity primary key,
    version    smallint                 default 0     not null,
    project_id bigint                                 not null references public.project (id),
    name       text                                   not null,
    done_req   smallint                 default 0     not null,
    total_req  smallint                 default 0     not null,
    completed  smallint generated always as
        (coalesce((100 * done_req / nullif(total_req, 0))::smallint, 0)) stored,
    created    timestamp with time zone default now() not null,
    modified   timestamp with time zone default now() not null
);

create table public.epic_history
(
    id         bigint                   not null,
    version    smallint                 not null,
    project_id bigint                   not null,
    name       text                     not null,
    modified   timestamp with time zone not null,
    deleted    boolean default false    not null,
    primary key (id, version)
);

create table public.change_phase
(
    slug     text    default 'backlog' not null primary key,
    priority integer default 0         not null
);

create table public.change_type
(
    slug     text    default 'feature' not null primary key,
    priority integer default 0         not null
);

create table public.change
(
    id           bigint generated always as identity primary key,
    version      smallint                 default 0               not null,
    project_id   bigint                                           not null references public.project (id),
    epic_id      bigint references public.epic (id),
    change_phase text                     default 'backlog'       not null references public.change_phase (slug),
    change_types text[]                                           not null,
    title        text                                             not null,
    body         text,
    codex_session_id text,
    closed       boolean                  default false           not null,
    done_req     smallint                 default 0               not null,
    total_req    smallint                 default 0               not null,
    completed    smallint generated always as
        (coalesce((100 * done_req / nullif(total_req, 0))::smallint, 0)) stored,
    created      timestamp with time zone default now()           not null,
    modified     timestamp with time zone default now()           not null
);

create table public.change_history
(
    id           bigint                   not null,
    version      smallint                 not null,
    project_id   bigint                   not null,
    epic_id      bigint,
    change_phase text                     not null,
    change_types text[]                   not null,
    title        text                     not null,
    body         text,
    codex_session_id text,
    closed       boolean                  not null,
    modified     timestamp with time zone not null,
    deleted      boolean default false    not null,
    primary key (id, version)
);

create table public.requirement
(
    id         bigint generated always as identity primary key,
    version    smallint                 default 0     not null,
    definition text                                   not null,
    done       boolean                  default false not null,
    change_id  bigint                                 not null references public.change (id),
    created    timestamp with time zone default now() not null,
    modified   timestamp with time zone default now() not null
);

create table public.requirement_history
(
    id         bigint                   not null,
    version    smallint                 not null,
    change_id  bigint                   not null,
    definition text                     not null,
    modified   timestamp with time zone not null,
    deleted    boolean                  not null,
    primary key (id, version)
);

create view public.vw_change as
select
    id,
    version,
    project_id,
    epic_id,
    change_phase,
    change_types,
    title,
    body,
    codex_session_id,
    closed,
    done_req,
    total_req,
    completed,
    created,
    modified
from public.change;

create view public.vw_project as
select
    p.id,
    p.name,
    p.created,
    p.modified,
    count(c.*)::integer as change_count
from public.project p
left join public.change c on p.id = c.project_id
group by p.id, p.name, p.created, p.modified
order by p.id;

create view public.vw_epic as
select e.id,
       e.version,
       e.project_id,
       e.name,
       e.done_req,
       e.total_req,
       e.completed,
       count(c.*)::integer as change_count,
       e.created,
       e.modified
from public.epic e
left join public.change c on e.id = c.epic_id
group by e.id, e.version, e.project_id, e.name, e.done_req, e.total_req, e.completed, e.created, e.modified
order by e.name;

create procedure public.sp_epic_to_history(IN _id bigint, IN _deleted boolean)
    language plpgsql
as
$$
begin
    insert into public.epic_history (
        id, version, project_id, name, modified, deleted
    )
    select
        id, version, project_id, name, modified, _deleted
    from public.epic
    where id = _id;
end;
$$;

create procedure public.sp_change_to_history(in _id bigint, in _deleted boolean)
    language plpgsql
as
$$
begin
    insert into public.change_history (
        id, version, project_id, epic_id, change_phase, change_types, title, body, codex_session_id, closed, modified, deleted
    )
    select
        id, version, project_id, epic_id, change_phase, change_types, title, body, codex_session_id, closed, modified, _deleted
    from public.change
    where id = _id;
end;
$$;

create procedure public.sp_requirement_to_history(in _id bigint, in _deleted boolean)
    language plpgsql
as
$$
begin
    insert into public.requirement_history (
        id, version, change_id, definition, modified, deleted
    )
    select
        id, version, change_id, definition, modified, _deleted
    from public.requirement
    where id = _id;
end;
$$;

create procedure public.sp_epic_requirement_recalculate(IN _epic_id bigint)
    language plpgsql
as
$$
declare
    _done_req smallint;
    _total_req smallint;
begin
    if _epic_id is null then return; end if;

    select
        coalesce(sum(done_req), 0),
        coalesce(sum(total_req), 0)
    into
        _done_req,
        _total_req
    from public.change
    where epic_id=_epic_id;

    update public.epic
    set
        done_req=_done_req,
        total_req=_total_req
    where id=_epic_id;
end;
$$;

create procedure public.sp_change_requirement_recalculate(in _change_id bigint)
    language plpgsql
as
$$
declare
    _epic_id bigint;
    _done_req smallint;
    _total_req smallint;
begin
    select
        count(*) filter (where done)::smallint,
        count(*)::smallint
    into
        _done_req,
        _total_req
    from public.requirement
    where change_id = _change_id;

    update public.change
    set
        done_req = _done_req,
        total_req = _total_req,
        modified = now()
    where id = _change_id;

    select epic_id into _epic_id
    from public.change
    where id = _change_id;

    call public.sp_epic_requirement_recalculate(_epic_id);
end;
$$;
