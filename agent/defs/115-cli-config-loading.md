# CLI Config Loading

Limit this Change to CLI support for the new repository-root `.mch` configuration layout.

The CLI should load configuration from `.mch` and expose the resolved configuration through a `/config` view in the Main screen menu. This Change should not implement Flow execution yet.

## Scope

- Use the repository-root `.mch` directory for CLI configuration.
- Treat `.mch` as committed project configuration. It contains no secrets and should not be gitignored.
- Load runtime CLI config from `.mch/config.yaml`.
- Load and use `temp_dir` from the new `config.yaml`.
- Load the default Flow profile from `.mch/default/*`.
- Always use `.mch/default` for now. Do not implement profile selection, overrides, local config, or fallback layering.
- Add a `/config` command to the Main screen menu.
- `/config` should open a read-only view that dumps the loaded CLI and Flow configuration so users can inspect the resolved values.
- The `/config` view should dump configuration from in-memory structs, not by dumping YAML files directly.
- No matter where `mch` is executed inside the repository, it must use the correct repository-root config.

## Flow Config Shape

`.mch/default` is the default Flow directory. Hook commands in `flow.yaml` are plain command strings.

Example:

```yaml
  entry: make idea-entry
  exec: codex review origin/stage
  exit: ls
```

When hook execution is implemented later, every hook command should be executed exactly as written with the working directory set to:

```text
  <git-root>/.mch/default/
```

This means `make idea-entry` naturally uses:

```text
  <git-root>/.mch/default/Makefile
```

No special command rewriting is needed.

## Out of Scope

- Do not execute Flow hooks in this Change.
- Do not implement Flow profile selection.
- Do not implement overrides.
- Do not implement `.mch/local` or user-local config.
- Do not sync Flow config into backend `public.config`.
- Do not change backend behavior.
- Do not change frontend behavior.

## Flow Options - Short Wiki

`run_stage`:

- `idea` - capture or refine the initial change idea.
- `spec` - create the implementation-ready spec.
- `ready` - mark the change as ready for execution.
- `docs` - update required documentation before implementation.
- `code` - implement the change.
- `polish` - user-guided refinement after coding.
- `pr` - create or update the pull request.
- `review` - review the PR/change against the spec.
- `fix` - address review findings.
- `sync` - automatically align the spec, QA cases, and docs with final
  behavior.
- `merge` - merge the change branch.
- `stage` - promote/merge into stage.
- `master` - promote/merge into master.

`task_step`:

- `none` - task has not started yet.
- `entry` - entry script is executing.
- `prompt` - interactive session is running.
- `agent` - automated agent is executing.
- `exit` - exit script is executing.
- `done` - task has finished.

`task_status`:

- `queued` - task is waiting to start.
- `running` - task is actively executing.
- `paused` - task is temporarily paused.
- `stopped` - task was manually stopped.
- `waiting` - task is waiting for input.
- `completed` - task finished successfully.
- `failed` - task finished with an error.

`stage_mode`:

- `skip` - stage will not execute.
- `prompt` - stage will run an interactive session.
- `exec` - stage will run an automated agent.

## Flow Model

Flow:

- has many Steps.
- has many Runs.

Step:

- belongs to Flow.

Run:

- belongs to Flow.
- has many Tasks.

Task:

- belongs to Run.
- belongs to Step.
- uses Worker.

Worker:

- can perform many Tasks.

Plain meaning:

- Flow = reusable automation definition.
- Step = one named stage inside the Flow.
- Run = one execution attempt of a Flow.
- Task = one unit of work inside a Run for a specific Step.
- Worker = executor/tool/process that performs a Task.

Concrete example:

- Flow: Change Automation.
- Step: code.
- Run: Run #42 for `change/add-project-selector`.
- Task: execute `codex_exec` for step `code` in Run #42.
- Worker: `codex_exec`

Concurrency still fits:

- One Flow can have many active Runs at the same time.
- Each Run progresses independently through the Flow's Steps.
- Each Run creates Tasks for its Steps.
- Workers perform Tasks.

Shortest correct sentence:

A Flow defines Steps; a Run executes a Flow; a Task performs one Step within a
Run; a Worker executes the Task.
