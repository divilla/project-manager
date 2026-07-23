# CLI Flow Stage 1

- The CLI flow will be implemented in multiple stages.
- The Change flow has three documents: idea, spec, and PR.

## Input and Output Concept

- Editing large text fields always happens in an editor.
- `idea` is used in the examples below, but the same rules apply to `spec` and `pr`.
- The user is first routed to `IdeaEditScreen` with either an empty string or existing text as the initial value.
- If initialized with an empty string, `IdeaEditScreen` enters `new` mode.
- If initialized with existing text, `IdeaEditScreen` enters `edit` mode.
- If `<temp_dir>/idea-input.md` exists, delete it. This purposely removes the old recovery scenario.
- Copy `<temp_dir>/idea-input.md` to `<temp_dir>/idea-output.md`.
- `<temp_dir>/idea-input.md` always reflects what is already in the database. If there is nothing in the database, the file is empty.
- The user or Agent always edits or updates `<temp_dir>/idea-output.md`.
- After `<temp_dir>/idea-output.md` is saved by the user or updated by Agent, the CLI compares the input and output files.
- If `<temp_dir>/idea-input.md` equals `<temp_dir>/idea-output.md`, no edit was made. Return to the previous screen.
- If `<temp_dir>/idea-input.md` is empty (`IdeaEditScreen` is in `new` mode) and `<temp_dir>/idea-output.md` is not empty, `IdeaEditScreen` must show:

```shell
bat -pp --theme 'Coldark-Dark' /tmp/mch/idea-output.md
```

- If `<temp_dir>/idea-input.md` is not empty (`IdeaEditScreen` is in `edit` mode) and `<temp_dir>/idea-output.md` is not empty, route to `IdeaEditScreen` in `edit` mode and show:

```shell
bat -pp --theme 'Coldark-Dark' --diff /tmp/mch/idea-input.md /tmp/mch/idea-output.md
```

- On `IdeaEditScreen`, the user is automatically prompted to choose `/rewrite`, `/save`, or `/cancel`. `/rewrite` is the default. The prompt appears below the `bat` command output.
- If the user selects `/rewrite`, read `<temp_dir>/idea-output.md` into a local variable, delete both files, and route to `IdeaRewriteScreen` initialized with that local variable.
- If the user selects `/save`, save `<temp_dir>/idea-output.md` to the database, delete both `<temp_dir>/idea-input.md` and `<temp_dir>/idea-output.md`, and route to `ChangeDetailsScreen`.
- If the user selects `/cancel`, delete both `<temp_dir>/idea-input.md` and `<temp_dir>/idea-output.md`, and route to `ChangeDetailsScreen`.
- On error, route to `ChangeErrorScreen` with only one argument, the message: `error saving idea to db, message: whatever-error-context-there-is`. The error is red. The user is prompted with a single `/ok` option that routes back to `ChangeDetailScreen`.

## IdeaRewriteScreen

- On entry, the input argument is saved to both `<temp_dir>/idea-input.md` and `<temp_dir>/idea-output.md`.
- The second `IdeaRewriteScreen` argument is a task constant: `EntryTask`, `ExecTask`, or `ExitTask`.
- Agent is activated, then the `idea` stage of the flow is activated. See Stage Flow below.
- On the happy path, when none of the Stage Tasks errored, the user is prompted to choose `/save`, `/edit`, or `/cancel`. `/save` is the default.
- `/save` saves `<temp_dir>/idea-output.md` to the database, deletes both `<temp_dir>/idea-input.md` and `<temp_dir>/idea-output.md`, and routes to `ChangeDetailsScreen`.
- `/edit` reads `<temp_dir>/idea-output.md` into a local variable, deletes both `<temp_dir>/idea-input.md` and `<temp_dir>/idea-output.md`, and routes to `IdeaEditScreen` with the local variable as the argument.
- `/cancel` deletes both `<temp_dir>/idea-input.md` and `<temp_dir>/idea-output.md`, and routes back to `ChangeDetailsScreen`.

## Stage Flow

- Stage Flow is the same for every stage.
- If `Config.Flow.Stage.Mode` is set to `skip`, route to `ChangeErrorScreen` with the message: `idea stage mode is set to skip - please edit flow.yaml, idea stage and set mode to exec or prompt`.
- If `Config.Flow.Stage.Entry` is set, execute the command in the `.mch/default` workdir.
- If `Config.Flow.Stage.Exec` is set, execute the command in the `.mch/default` workdir. Otherwise, start Prompt Flow.
- If `Config.Flow.Stage.Exit` is set, execute the command in the `.mch/default` workdir.
- On success or error, save data to `api/v1/change/update-run` and route to `SpecWriteScreen`.
- On error, route to `TaskErrorScreen`.

## Prompt Flow

- If `Config.Flow.Stage.Exec` is not set, then `Config.Flow.Stage.Prompt` is mandatory. It must point to a valid Markdown file, referred to below as `<prompt-file>`.
- Depending on the value of `Config.Flow.Stage.Mode`, one of two scenarios is valid.
- `prompt`: run Codex in bind mode, like the editor is run:

```shell
cat prompts/idea.md | codex -C <git-root> -
```

- `exec`: in the `.mch/default` workdir, run:

```shell
git_root="$(git rev-parse --show-toplevel)"
run_dir="$(mktemp -d <temp-dir>)"

events="$run_dir/events.jsonl"
stderr="$run_dir/stderr.log"
final="$run_dir/final.txt"
session_id="$run_dir/session-id.txt"

prompt='your prompt here'

codex exec -C "$git_root" --json \
  -o "$final" \
  "$prompt" \
  >"$events" \
  2>"$stderr"

jq -r '
    select(.type == "thread.started")
    | .thread_id // .session_id // .session.id // .id
' "$events" | head -n 1 >"$session_id"
```

- While this command is executing, there must be a character animation and a seconds counter.
- When the command finishes, first copy `$run_dir/session-id.txt` to `<temp-dir>/sessions/<change-slug>`.
- If `$run_dir/stderr.log` is not empty, route to `TaskErrorScreen` and display the contents of `$run_dir/stderr.log`.
- If `$run_dir/final.txt` is `Done.`, continue with Task Flow.
- Otherwise, run Codex in bind mode, like the editor is run:

```shell
codex -C <git-root> resume "#session_id"
```

## TaskErrorScreen

- The screen displays a view similar to `ChangeDetailsScreen` with three fields: `Stage: <stage-name>`, `Task: <one of Entry script, Exec script, or Exit script>`, and `Command: <command>`.
- The screen also displays the full execution output.
- The user is prompted with `/repeat` or `/cancel`. `/repeat` is the default.
- `/repeat` routes back to `IdeaRewriteScreen` with the `TaskName` argument.
- `/cancel` routes back to `ChangeDetailsScreen`.

## SpecWriteScreen

- `SpecWriteScreen` executes Stage Flow, the same as `IdeaRewriteScreen`, except it reuses the session ID.
- Depending on the value of `Config.Flow.Stage.Mode`, one of two scenarios is valid.
- `prompt`: run Codex in bind mode, like the editor is run:

```shell
cat <config.spec-stage.prompt> | codex -C <git-root> - resume $(cat <temp-dir>/sessions/<change-slug>)
```

- `exec`: in the `.mch/default` workdir, run:

```shell
git_root="$(git rev-parse --show-toplevel)"
run_dir="$(mktemp -d <temp-dir>)"

events="$run_dir/events.jsonl"
stderr="$run_dir/stderr.log"
final="$run_dir/final.txt"
session_id="$(cat $run_dir/sessions/<change-slug>)"
prompt="$(cat <config spec stage prompt value>)"

cat | codex exec -C "$git_root" resume \
  --json \
  -o "$final" \
  "$session_id" \
  "$prompt" \
  >"$events" \
  2>"$stderr"
```

- While this command is executing, there must be a character animation and a seconds counter.
- If `$run_dir/stderr.log` is not empty, route to `TaskErrorScreen` and display the contents of `$run_dir/stderr.log`.
- If `$run_dir/final.txt` is `Done.`, continue with Task Flow.
- Otherwise, run Codex in bind mode, like the editor is run:

```shell
codex -C <git-root> resume "#session_id"
```

## VerifyWriteScreen

- `SpecVerifyScreen` executes Stage Flow, the same as `SpecWriteScreen`, except it uses a different prompt: `<config.verify-stage.prompt>`.

## Idea Flow

Idea, in the new flow, must be treated as a `change` table field entry, not as a file. There are no Git commands around writing the idea. The previous `restore` flow is discarded.

- New Idea routes to `EditIdeaScreen` with an empty string argument.
- Edit Idea routes to `EditIdeaScreen` with the current idea from the API as the argument.

## Scripts

Ideally, two Bash scripts are created:

- `.mch/default/scripts/codex-exec-no-session.sh <change-slug> <temp-dir>`
- `.mch/default/scripts/codex-exec-with-session.sh <change-slug> <temp-dir>`

`<temp-dir>/*` holds all other values that can be used during script execution or afterward in Go code.

## Spec file structure update

`.mch/default/prompts/spec-file-structure.md` has been updated

## Glossary, Options, and Naming Conventions

`run_stage`:

- `idea`: capture or refine the initial change idea.
- `spec`: create the implementation-ready spec.
- `ready`: mark the change as ready for execution.
- `docs`: update required documentation before implementation.
- `code`: implement the change.
- `polish`: user-guided refinement after coding.
- `pr`: create or update the pull request.
- `review`: review the PR/change against the spec.
- `fix`: address review findings.
- `sync`: automatically align spec, QA cases, and docs with final behavior.
- `merge`: merge the change branch.
- `stage`: promote or merge into stage.
- `master`: promote or merge into master.

`task_step`:

- `none`: task has not started yet.
- `entry`: entry script is executing.
- `prompt`: interactive session is running.
- `agent`: automated agent is executing.
- `exit`: exit script is executing.
- `done`: task has finished.

`task_status`:

- `queued`: task is waiting to start.
- `running`: task is actively executing.
- `paused`: task is temporarily paused.
- `stopped`: task was manually stopped.
- `waiting`: task is waiting for input.
- `completed`: task finished successfully.
- `failed`: task finished with an error.

`stage_mode`:

- `skip`: stage will not execute.
- `prompt`: stage will run an interactive session.
- `exec`: stage will run an automated agent.

## Model Notes

Spec:

- Flow has many Steps.
- Flow has many Runs.
- Step belongs to Flow.
- Run belongs to Flow.
- Run has many Tasks.
- Task belongs to Run.
- Task belongs to Step.
- Task uses Worker.
- Worker can perform many Tasks.

Plain meaning:

- Flow = reusable automation definition.
- Step = one named stage inside the flow.
- Run = one execution attempt of a flow.
- Task = one unit of work inside a run for a specific step.
- Worker = executor/tool/process that performs a task.

Concrete example:

- Flow: Change Automation.
- Step: code.
- Run: Run #42 for `change/add-project-selector`.
- Task: execute `codex_exec` for step `code` in Run #42.
- Worker: `codex_exec`.

Concurrency still fits:

- One Flow can have many Runs active at the same time.
- Each Run progresses independently through the Flow's Steps.
- Each Run creates Tasks for its Steps.
- Workers perform Tasks.

Shortest correct sentence:

A Flow defines Steps; a Run executes a Flow; a Task performs one Step within a Run; a Worker executes the Task.

---

Review Findings:

- [P1] Define how spec output is persisted — /home/vito/go/src/project-manager/specs/117-cli-flow-stage-1-3.md:188-226
  The Goal promises that mch writes and verifies the spec, but SpecWriteScreen and VerifyWriteScreen only run Codex prompts and advance screens; they never define where generated spec content is read from, compared, saved, or persisted through POST /
  api/v1/change/update-spec or a repository artifact. Add the spec-stage input/output temp-file contract, save/cancel/error behavior, persistence endpoint, and acceptance/QA coverage so implementation can produce the promised observable end state.

- [P1] Read the saved session from the configured temp directory — /home/vito/go/src/project-manager/specs/117-cli-flow-stage-1-3.md:190-207
  The prose says SpecWriteScreen reuses the session ID saved at <temp-dir>/sessions/<change-slug>, but the exec snippet creates a fresh run_dir and reads session_id="$(cat $run_dir/sessions/<change-slug>)", a path that was never populated. Fix the
  command contract to read from the configured temp directory session path, or explicitly copy the session into run_dir before reading it.

- [P1] Clarify whether New Idea updates an existing Change or creates one — /home/vito/go/src/project-manager/specs/117-cli-flow-stage-1-3.md:46-48
  The Spec routes New Idea to IdeaEditScreen with an empty value, while the save flow later uses focused idea update behavior and routes to ChangeDetailsScreen; it never defines the existing Change ID needed for update-idea, nor a create flow with
  project_id, title, and idea for a truly new Change. Define the entry context and persistence contract so New Idea cannot reach a save or no-op path without a valid destination Change.
