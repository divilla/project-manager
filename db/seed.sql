begin;

insert into public.task_phase (slug, priority)
values
    ('backlog', 0),
    ('progress', 1),
    ('review', 2),
    ('staging', 3),
    ('production', 4),
    ('repair', 5)
on conflict (slug) do update
set priority = excluded.priority;

insert into public.task_type (slug, priority)
values
    ('epic', 0),
    ('feature', 0),
    ('group', 0),
    ('issue', 0),
    ('spike', 0),
    ('task', 0),
    ('upgrade', 0)
on conflict (slug) do update
set priority = excluded.priority;

commit;
