begin;

insert into public.change_phase (slug, priority)
values
    ('backlog', 0),
    ('progress', 1),
    ('review', 2),
    ('staging', 3),
    ('production', 4),
    ('rejected', 5)
on conflict (slug) do update
set priority = excluded.priority;

insert into public.change_type (slug, priority)
values
    ('feature', 0),
    ('fix', 0),
    ('refactor', 0),
    ('upgrade', 0),
    ('chore', 0),
    ('docs', 0),
    ('test', 0),
    ('test', 0),
    ('ci', 0),
    ('security', 0),
    ('migration', 0),
    ('revert', 0),
    ('spike', 0)
on conflict (slug) do update
set priority = excluded.priority;

commit;
