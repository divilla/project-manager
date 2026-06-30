begin;

truncate table
    public.test_case_history,
    public.test_case,
    public.change_history,
    public.change,
    public.epic_history,
    public.epic,
    public.project
restart identity;

insert into public.project (name)
values ('demo1'), ('demo2'), ('demo3');

insert into public.epic (project_id, name)
select p.id, seed.name
from public.project p
cross join (values
    ('Platform Foundation'),
    ('Project Workspace'),
    ('Completeness Engine')
) as seed(name)
where p.name = 'demo1';

insert into public.change (
    project_id, epic_id, change_phase, change_types, title, requirement_body
)
select
    p.id,
    e.id,
    seed.phase,
    seed.types,
    seed.title,
    seed.requirement_body
from public.project p
join (values
    ('Platform Foundation', 'production', array['feature'], 'Configure local API startup', 'Load configuration and start the Echo service on the expected port.'),
    ('Platform Foundation', 'staging', array['chore'], 'Define clean database bootstrap', 'Create the complete schema from a single authoritative script.'),
    ('Project Workspace', 'review', array['feature'], 'Create project endpoint', 'Persist a named project and return its generated identifier.'),
    ('Project Workspace', 'progress', array['feature'], 'Select active project context', 'Load project-scoped changes when the user changes workspaces.'),
    ('Completeness Engine', 'review', array['feature'], 'Create binary test case', 'Attach a concrete Definition of Done item to a change.'),
    ('Completeness Engine', 'progress', array['feature'], 'Recalculate change counters', 'Count completed and total test cases for a change.')
) as seed(epic_name, phase, types, title, requirement_body)
    on true
join public.epic e on e.project_id = p.id and e.name = seed.epic_name
where p.name = 'demo1';

insert into public.test_case (change_id, scenario, done)
select
    c.id,
    seed.scenario,
    seed.done
from public.change c
join (values
    ('Configure local API startup', 'Health endpoint reports API availability.', true),
    ('Configure local API startup', 'Database health check reports availability.', true),
    ('Define clean database bootstrap', 'Schema can be rebuilt from init and seed scripts.', false),
    ('Create project endpoint', 'Create returns the generated project identifier.', true),
    ('Create binary test case', 'Test case create returns recalculated change data.', false),
    ('Recalculate change counters', 'Done toggles update change completeness.', false)
) as seed(title, scenario, done)
    on seed.title = c.title;

do
$$
declare
    _change record;
begin
    for _change in select id from public.change loop
        call public.sp_change_test_case_recalculate(_change.id);
    end loop;
end;
$$;

commit;
