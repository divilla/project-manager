# Agent Stages and Tasks Backend

## Docs

Update the docs carefully using the clarified facts from this idea.

## DB

- init.sql and seed.sql have already been updated to include all database changes required for this Change.
- Review the SQL file changes.
- `public.config` has been added.
- `public.change` has been extended with these columns:
  - flow_stages - ordered snapshot of Flow stages copied onto the Change when its ref is assigned.
  - flow_stage_modes - per-stage execution modes for this Change, matched to flow_stages by array index.
  - run_claim_id - UUID assigned when a runner claims the run. Null means the run is currently unclaimed.
  - run_flow_stage - current Flow stage of the run, such as code, review, or sync.
  - run_task_step - current execution step of the active/latest task, such as entry, prompt, agent, or exit.
  - run_task_status - current status of the active/latest task, such as queued, running, waiting, completed, or failed.
  - run_error - latest error reported by the run or active task. Empty means no current error is recorded.
  - run_is_completed - true when the run has successfully completed its copied Flow. Used for stable and fast filtering even if global Flow configuration changes later.
  - run_started_at - timestamp when the run was first started or claimed.
  - run_updated_at - timestamp when the run state was last updated.
- Empty values indicate the task has not started yet.
- Run values are updated after task execution finishes.
- Empty `run_error` indicates success.
- None of these columns are saved in `public.change_history`.

Stored procedures:

- `public.sp_change_ref_update` was renamed to `public.sp_change_assign_flow` to avoid ambiguity.
- `sp_change_assign_flow` - assigns a Flow to a Change by setting its reference and copying the current Flow stages and default stage modes onto the Change.
- `fn_change_start_run` - starts a Change run by claiming it and returning a new claim ID, or null if the run is already claimed.
- `fn_change_update_run` - updates the current state of a claimed Change run; returns null when `_id`, `_claim_id`, or both do not match an active claim.
- `fn_change_reset_claim` - clears a stale Change run claim so the run can be claimed again.
- Direct updates to `flow_*` and `run_*` columns are not allowed; these columns must be changed only through the defined procedures and functions.

## Backend

- Rename `/api/v1/change/reference` to `/api/v1/change/assign-flow`.
- Add `/api/v1/change/start-run`; it must return `claim_id`.
- Add `/api/v1/change/update-run`; it must return `change_id`.
- Add `/api/v1/change/reset-claim`; it must return `claim_id`.

`internal/dto` DTOs are updated and should be treated as the naming reference. Update the rest of the backend to match these DTOs and endpoint changes.

Build all unit tests and API integration tests to cover both the happy paths and relevant edge cases.

## Frontend and CLI

Frontend and CLI updates are out of scope of this change. They will be addressed in future Changes.

## Flow Configuration

The app has a global Flow configuration stored in `public.config`.

This configuration defines the fixed automation flow, available stage modes, task statuses, and task execution steps. The Flow stages are ordered and are not reordered or edited by users.

When a new Change is created, the current Flow configuration is copied onto that Change as its own Run configuration. Later updates to the global Flow configuration only affect newly created Changes. Existing Changes keep the configuration they were created with.

### Flow Stages

`flow_stages` defines the ordered stages of the automation flow:

- idea
- spec
- ready
- docs
- code
- polish
- pr
- review
- fix
- sync
- merge
- stage
- master

`flow_stages_help` stores short help text for each stage. Each help entry matches `flow_stages` by array index.

### Stage Modes

`stage_modes` defines the allowed execution modes for a stage:

- `skip` - stage will not execute
- `prompt` - stage will run an interactive session
- `exec` - stage will run an automated agent

`flow_stage_modes_default` stores the default mode for each Flow stage. Each mode matches `flow_stages` by array index.

Users can adjust stage modes in the Flow configuration, but those changes only apply to Changes created after the adjustment.

### Task Statuses

`task_statuses` defines the possible lifecycle statuses of a Task:

- queued
- running
- paused
- stopped
- waiting
- completed
- failed

`task_statuses_help` stores short help text for each status.

### Task Steps

`task_steps` defines the current execution step inside a Task:

- none
- entry
- prompt
- agent
- exit
- done

`task_steps_help` stores short help text for each task step.

### Flow Naming Convention

Use this naming convention as a hard rule.

Flow:

- has many Steps
- has many Runs

Step:

- belongs to Flow

Run:

- belongs to Flow
- has many Tasks

Task:

- belongs to Run
- belongs to Step
- uses Worker

Worker:

- can perform many Tasks

Plain meaning:

- Flow = reusable automation definition
- Step = one named stage inside the flow
- Run = one execution attempt of a flow
- Task = one unit of work inside a run for a specific step
- Worker = executor/tool/process that performs a task

Concrete example:

- Flow: Change Automation
- Step: code
- Run: Run #42 for change/add-project-selector
- Task: execute codex_exec for step code in Run #42
- Worker: codex_exec

Concurrency still fits:

- One Flow can have many Runs active at the same time.
- Each Run progresses independently through the Flow's Steps.
- Each Run creates Tasks for its Steps.
- Workers perform Tasks.

Shortest correct sentence:

A Flow defines Steps; a Run executes a Flow; a Task performs one Step within a Run; a Worker executes the Task.
