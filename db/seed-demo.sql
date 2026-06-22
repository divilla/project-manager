begin;

truncate table
    public.requirement_history,
    public.task_history,
    public.requirement,
    public.task,
    public.project
restart identity;

insert into public.project (name)
values ('demo1'), ('demo2'), ('demo3');

insert into public.task (
    project_id, name, description, task_type, task_phase, difficulty, priority
)
select
    p.id, seed.name, seed.description, 'epic', seed.phase, seed.difficulty, seed.priority
from public.project p
cross join (values
    ('Platform Foundation', 'Establish the architecture and developer workflow for the product.', 'production', 5, 10),
    ('Project Workspace', 'Provide reliable project organization and navigation.', 'staging', 4, 20),
    ('Task Planning', 'Support structured planning with hierarchical tasks.', 'review', 5, 30),
    ('Completeness Engine', 'Measure delivery through concrete verified requirements.', 'progress', 5, 40),
    ('Delivery Workflow', 'Move planned work through a visible delivery lifecycle.', 'backlog', 4, 50),
    ('Quality and Operations', 'Keep the application observable, tested, and recoverable.', 'repair', 4, 60)
) as seed(name, description, phase, difficulty, priority)
where p.name = 'demo1';

insert into public.task (
    project_id, parent_id, name, description, task_type, task_phase, difficulty, priority
)
select
    p.id, parent.id, seed.name, seed.description, 'feature', seed.phase, seed.difficulty, seed.priority
from public.project p
join (values
    ('Platform Foundation', 'Local Application Runtime', 'Run the frontend, API, and database consistently.', 'production', 3, 11),
    ('Platform Foundation', 'Database Lifecycle', 'Initialize, migrate, and seed PostgreSQL safely.', 'staging', 4, 12),
    ('Project Workspace', 'Project Management', 'Create and maintain project workspaces.', 'production', 3, 21),
    ('Project Workspace', 'Workspace Navigation', 'Make active work easy to locate and revisit.', 'review', 2, 22),
    ('Task Planning', 'Task Hierarchy', 'Break large outcomes into navigable child tasks.', 'progress', 4, 31),
    ('Task Planning', 'Planning Controls', 'Capture task metadata and workflow intent.', 'backlog', 3, 32),
    ('Completeness Engine', 'Requirement Management', 'Maintain binary definitions of done.', 'production', 4, 41),
    ('Completeness Engine', 'Progress Aggregation', 'Roll requirement progress through the task tree.', 'review', 5, 42),
    ('Delivery Workflow', 'Phase Board', 'Visualize and change the current delivery phase.', 'progress', 3, 51),
    ('Delivery Workflow', 'Release Readiness', 'Expose work that is ready for staging and production.', 'backlog', 4, 52),
    ('Quality and Operations', 'Automated Verification', 'Protect core behavior with repeatable tests.', 'repair', 4, 61),
    ('Quality and Operations', 'Runtime Diagnostics', 'Make service and database failures understandable.', 'staging', 3, 62)
) as seed(parent_name, name, description, phase, difficulty, priority)
    on true
join public.task parent
    on parent.project_id = p.id and parent.name = seed.parent_name
where p.name = 'demo1';

insert into public.task (
    project_id, parent_id, name, description, task_type, task_phase, difficulty, priority
)
select
    p.id,
    parent.id,
    seed.name,
    seed.description,
    seed.task_type,
    seed.phase,
    seed.difficulty,
    seed.priority
from public.project p
join (values
    ('Local Application Runtime', 'Configure local API startup', 'Load configuration and start the Echo service on the expected port.', 'task', 'production', 2, 111),
    ('Local Application Runtime', 'Configure frontend development server', 'Serve the Quasar application with a working API base URL.', 'task', 'production', 2, 112),
    ('Local Application Runtime', 'Enable local CORS policy', 'Allow the configured frontend origin without widening production access.', 'task', 'staging', 2, 113),

    ('Database Lifecycle', 'Define clean database bootstrap', 'Create the complete schema from a single authoritative script.', 'task', 'staging', 4, 121),
    ('Database Lifecycle', 'Seed workflow reference values', 'Install canonical task phases and task types idempotently.', 'task', 'review', 2, 122),
    ('Database Lifecycle', 'Seed realistic demonstration data', 'Build a repeatable project dataset for product evaluation.', 'task', 'progress', 3, 123),
    ('Database Lifecycle', 'Document incremental migration ordering', 'Reserve an ordered path for schema changes after the baseline.', 'spike', 'backlog', 2, 124),

    ('Project Management', 'Create project endpoint', 'Persist a named project and return its generated identifier.', 'task', 'production', 2, 211),
    ('Project Management', 'List project workspaces', 'Return stable project results with paging controls.', 'task', 'production', 2, 212),
    ('Project Management', 'Rename existing project', 'Validate and persist a project name change.', 'task', 'review', 2, 213),

    ('Workspace Navigation', 'Select active project', 'Load the task board when the user changes workspaces.', 'task', 'review', 2, 221),
    ('Workspace Navigation', 'Render empty workspace state', 'Explain how to begin when a project contains no tasks.', 'task', 'staging', 1, 222),
    ('Workspace Navigation', 'Preserve navigation errors', 'Keep backend failures visible without discarding current context.', 'issue', 'repair', 3, 223),
    ('Workspace Navigation', 'Refresh workspace data', 'Reload reference, project, and task data on demand.', 'task', 'progress', 2, 224),

    ('Task Hierarchy', 'Create root task', 'Create standalone work under a selected project.', 'task', 'production', 2, 311),
    ('Task Hierarchy', 'Create child task', 'Validate that a child belongs to the same project as its parent.', 'task', 'review', 3, 312),
    ('Task Hierarchy', 'Delete task subtree', 'Archive and remove a task with all descendants transactionally.', 'task', 'progress', 5, 313),

    ('Planning Controls', 'Edit task metadata', 'Update name, description, type, difficulty, and priority safely.', 'task', 'progress', 3, 321),
    ('Planning Controls', 'Reject stale task update', 'Return a conflict when the supplied task version is outdated.', 'issue', 'review', 4, 322),
    ('Planning Controls', 'Load phase and type options', 'Read valid workflow choices from reference tables.', 'task', 'production', 2, 323),
    ('Planning Controls', 'Order task board consistently', 'Sort tasks by phase priority and task priority.', 'task', 'backlog', 2, 324),

    ('Requirement Management', 'Create binary requirement', 'Attach a concrete definition of done to a task.', 'task', 'production', 2, 411),
    ('Requirement Management', 'Toggle requirement status', 'Mark a requirement done or incomplete transactionally.', 'task', 'production', 3, 412),
    ('Requirement Management', 'Edit requirement definition', 'Preserve history while changing verification text.', 'task', 'review', 3, 413),

    ('Progress Aggregation', 'Recalculate leaf counters', 'Count completed and total requirements for a leaf task.', 'task', 'review', 3, 421),
    ('Progress Aggregation', 'Roll counters into parent tasks', 'Propagate descendant counts to every ancestor.', 'task', 'progress', 5, 422),
    ('Progress Aggregation', 'Render derived completion percentage', 'Calculate progress safely when total requirements are zero.', 'task', 'staging', 3, 423),
    ('Progress Aggregation', 'Verify multi-level aggregation', 'Prove cached counters match active descendant requirements.', 'task', 'backlog', 4, 424),

    ('Phase Board', 'Group tasks by workflow phase', 'Render one board column for each seeded phase.', 'task', 'progress', 3, 511),
    ('Phase Board', 'Move task between phases', 'Validate and persist a version-guarded phase change.', 'task', 'review', 3, 512),
    ('Phase Board', 'Display task progress card', 'Show task type and derived completeness on the board.', 'task', 'staging', 2, 513),

    ('Release Readiness', 'Identify review backlog', 'Surface review work with incomplete requirements.', 'feature', 'backlog', 3, 521),
    ('Release Readiness', 'Identify production candidates', 'Show fully complete work approaching production.', 'feature', 'backlog', 3, 522),
    ('Release Readiness', 'Represent repair work', 'Keep production repair tasks visible in a dedicated phase.', 'issue', 'repair', 3, 523),
    ('Release Readiness', 'Summarize phase completeness', 'Aggregate requirement counters for each workflow phase.', 'feature', 'backlog', 4, 524),

    ('Automated Verification', 'Test project lifecycle', 'Cover project create, read, update, and delete behavior.', 'task', 'repair', 3, 611),
    ('Automated Verification', 'Test task concurrency', 'Prove stale task writes cannot overwrite current data.', 'task', 'progress', 4, 612),
    ('Automated Verification', 'Test requirement aggregation', 'Verify requirement changes update every task ancestor.', 'task', 'review', 4, 613),

    ('Runtime Diagnostics', 'Expose health endpoint', 'Report API and database availability to local clients.', 'task', 'production', 2, 621),
    ('Runtime Diagnostics', 'Log structured request context', 'Record request path, status, and failures with zerolog.', 'task', 'staging', 2, 622),
    ('Runtime Diagnostics', 'Return consistent JSON errors', 'Map validation, missing data, and conflicts to clear responses.', 'task', 'review', 3, 623),
    ('Runtime Diagnostics', 'Document recovery workflow', 'Describe how to rebuild and reseed a local database.', 'task', 'backlog', 2, 624)
) as seed(parent_name, name, description, task_type, phase, difficulty, priority)
    on true
join public.task parent
    on parent.project_id = p.id and parent.name = seed.parent_name
where p.name = 'demo1';

with leaf_tasks as (
    select
        t.id,
        t.name,
        row_number() over (order by t.id)::integer as leaf_number
    from public.task t
    join public.project p on p.id = t.project_id
    where p.name = 'demo1'
      and not exists (select 1 from public.task child where child.parent_id = t.id)
), requirement_templates as (
    select *
    from unnest(array[
        'Confirm the implementation matches the documented contract',
        'Add an automated test for the primary success path',
        'Add an automated test for invalid input',
        'Verify database changes are atomic on failure',
        'Verify the user-facing error identifies the corrective action',
        'Review naming and behavior against existing project conventions',
        'Confirm repeated execution does not create duplicate state',
        'Record observable evidence that the behavior works',
        'Complete peer review of the delivered behavior'
    ]) with ordinality as template(definition, requirement_number)
)
insert into public.requirement (task_id, definition, done)
select
    leaf.id,
    template.definition || ': ' || leaf.name || '.',
    case leaf.leaf_number % 3
        when 0 then true
        when 1 then false
        else template.requirement_number % 2 = 0
    end
from leaf_tasks leaf
join requirement_templates template
    on template.requirement_number <= 1 + ((leaf.leaf_number - 1) % 9);

do
$$
declare
    leaf record;
begin
    for leaf in
        select t.id
        from public.task t
        where not exists (select 1 from public.task child where child.parent_id = t.id)
        order by t.id
    loop
        call public.sp_task_requirement_recalculate(leaf.id);
    end loop;
end;
$$;

do
$$
declare
    invalid_count integer;
begin
    if (select count(*) from public.project) <> 3
        or (select count(*) from public.project where name in ('demo1', 'demo2', 'demo3')) <> 3 then
        raise exception 'demo seed must create exactly demo1, demo2, and demo3';
    end if;

    if (select count(*) from public.task t join public.project p on p.id = t.project_id where p.name = 'demo1') <> 60 then
        raise exception 'demo1 must contain exactly 60 tasks';
    end if;

    if exists (
        select 1
        from public.task t
        join public.project p on p.id = t.project_id
        where p.name in ('demo2', 'demo3')
    ) then
        raise exception 'demo2 and demo3 must not contain tasks';
    end if;

    if exists (
        select 1
        from public.task t
        where not exists (select 1 from public.task child where child.parent_id = t.id)
          and (select count(*) from public.requirement r where r.task_id = t.id) not between 1 and 9
    ) then
        raise exception 'every leaf task must have between 1 and 9 requirements';
    end if;

    if exists (
        select 1
        from public.task t
        where exists (select 1 from public.task child where child.parent_id = t.id)
          and exists (select 1 from public.requirement r where r.task_id = t.id)
    ) then
        raise exception 'non-leaf tasks must not have direct requirements';
    end if;

    with recursive descendants(root_id, task_id) as (
        select id, id
        from public.task
        union all
        select descendants.root_id, child.id
        from descendants
        join public.task child on child.parent_id = descendants.task_id
    ), expected as (
        select
            descendants.root_id,
            count(requirement.id) filter (where requirement.done) as done_req,
            count(requirement.id) as total_req
        from descendants
        left join public.requirement requirement on requirement.task_id = descendants.task_id
        group by descendants.root_id
    )
    select count(*)
    into invalid_count
    from public.task
    join expected on expected.root_id = task.id
    where task.done_req <> expected.done_req
       or task.total_req <> expected.total_req;

    if invalid_count <> 0 then
        raise exception '% seeded tasks have invalid requirement counters', invalid_count;
    end if;

    if not exists (select 1 from public.task where total_req > 0 and done_req = 0)
        or not exists (select 1 from public.task where total_req > 0 and done_req = total_req)
        or not exists (select 1 from public.task where done_req > 0 and done_req < total_req) then
        raise exception 'demo seed must include 0%%, partial, and 100%% completeness';
    end if;

    if exists (select 1 from public.task_history)
        or exists (select 1 from public.requirement_history) then
        raise exception 'fresh demo data must not create history rows';
    end if;
end;
$$;

commit;
