# CLI Config Loading

Types: feature|docs|test

## Goal

`mch` loads its CLI and default Flow configuration from the repository-root `.mch` tree from any working directory inside the repository, and users can inspect the resolved in-memory configuration through a read-only `/config` view from the Main screen.

## Scope

- Update the `cli/` `mch` configuration loader to resolve the Git repository root and load runtime CLI config from `<git-root>/.mch/config.yaml`.
- Load the default Flow profile from `<git-root>/.mch/default/*`, including `flow.yaml`, `help.yaml`, prompt paths, Makefile path, and hook command strings.
- Load and use `temp_dir` from `.mch/config.yaml` for CLI temporary planning files instead of a hardcoded `/tmp/mch` value.
- Move current project selection into the repository-root `.mch/config.yaml` by persisting numeric `project_id` there.
- Add `/config` to the Main screen command menu and route it to a read-only configuration inspection screen.
- Update focused CLI documentation for the new `.mch` repository-root config layout and the `/config` command.
- Add focused CLI tests for config path resolution, config parsing, Flow config parsing, `temp_dir` use, `/config` navigation, and read-only rendering.

## Requirements

- `mch` must resolve the repository root with Git before loading CLI configuration. Running `mch` from the repository root or from any nested directory inside the repository must use the same `<git-root>/.mch/config.yaml` file and `<git-root>/.mch/default` Flow directory.
- `.mch` is committed project configuration. The implementation must not add `.mch` or `.mch/config.yaml` to gitignore rules and must not treat `.mch` as a secrets or user-local directory.
- Runtime CLI project config must be loaded from `.mch/config.yaml`. The file fully replaces `cli/.config/config.yaml` and owns `backend_url`, `temp_dir`, and numeric `project_id`.
- Agent rewrite prompts for CLI temporary planning files must operate on the resolved temporary planning workspace supplied by the runner, including the exact `initial-idea.md` path under the configured `temp_dir`, instead of relying on shipped utility skill instructions that hardcode `/tmp/mch`.
- The runtime CLI must not define or fall back to a built-in temp workspace path. Automated tests, test fixtures, and verification build commands may still use disposable system temporary directories or `t.TempDir()` paths; those paths are test infrastructure and do not define runtime workspace behavior.
- `project_id` in `.mch/config.yaml` is intentionally committed, repository-root, branch-scoped project selection state. It is expected that all users of a repository branch share the same backend URL and current project ID for that branch instead of keeping a user-local CLI project selection.
- Missing `project_id` and `project_id: 0` in `.mch/config.yaml` are valid no-selection states. They must behave the same way as the existing no-current-project startup path instead of being treated as malformed config.
- Existing current project selection behavior must continue to work by reading and writing `project_id` in `.mch/config.yaml`. `/select-project` must persist the selected numeric `project_id` to the repository-root `.mch/config.yaml` and must not create, read, or write `cli/.config/config.yaml`.
- The CLI must always use `.mch/default` as the active Flow profile in this Change. It must not implement profile selection, overrides, local config, fallback layering, or `.mch/local`.
- The default Flow loader must parse `.mch/default/flow.yaml` into typed in-memory structs instead of leaving callers to inspect YAML files directly.
- The Flow loader must preserve hook command strings exactly as configured. Hook command strings are plain command strings.
- The Flow loader must parse default Flow Steps in order and expose each Step slug, help text, mode, prompt path, and `entry`, `exec`, and `exit` hook command string to the in-memory resolved configuration.
- Each Flow Step slug in `.mch/default/flow.yaml` must be non-empty and unique.
- Each Flow Step mode in `.mch/default/flow.yaml` must be defined and must be one of the default stage modes: `skip`, `prompt`, or `exec`.
- The Flow loader must parse `.mch/default/help.yaml` and expose configured stage modes, task statuses, and task steps to the resolved configuration without requiring a built-in vocabulary.
- The loaded Flow stage slugs must support the default ordered Flow stages: `idea`, `spec`, `ready`, `docs`, `code`, `polish`, `pr`, `review`, `fix`, `sync`, `merge`, `stage`, and `master`.
- Flow help option groups are data, not validation lists. Listed stage modes, task statuses, and task steps only require non-empty slugs; missing groups, empty groups, custom slugs, and reordered slugs are valid. This does not expand the default Flow Step mode vocabulary for `.mch/default/flow.yaml`, where Step mode remains mandatory and limited to `skip`, `prompt`, or `exec`.
- `/config` must be available from `MainState` in the Main screen command menu. Selecting it must open a read-only config view without calling backend APIs or mutating local files.
- The `/config` view must dump the resolved CLI and Flow configuration from in-memory structs. It must not render by reading or dumping raw YAML files directly.
- The `/config` view must show enough resolved data to verify the active config source, including repository root, config file path, `temp_dir`, active Flow directory, Flow metadata, Flow Steps, hook command strings, stage modes, task statuses, and task steps.
- The `/config` view must provide the normal read-only return behavior for a Main screen child view: `/return`, Esc, or Ctrl+C on an empty prompt returns to `MainState` without changing configuration.
- Missing, unreadable, or malformed `.mch/config.yaml`, `.mch/default/flow.yaml`, or `.mch/default/help.yaml` must produce a verbose startup config error and must not silently fall back to `cli/.config/config.yaml` or any other config source.
- When startup config loading fails, `mch` must print or return enough path-specific error detail for the user to identify and fix the failing file, then exit without starting the interactive app.
- This Change must not execute Flow hooks, start Flow Runs, claim Runs, update backend Flow snapshot fields, or write Flow configuration into backend `public.config`.
- This Change must not change backend or frontend behavior.

## Acceptance Criteria

- Starting `mch` from the repository root loads `.mch/config.yaml`, uses its `temp_dir`, loads `.mch/default/flow.yaml` and `.mch/default/help.yaml`, and reaches the normal startup flow when the config is valid, including the no-current-project path when `project_id` is missing or `0`.
- Starting `mch` from a nested directory under the repository loads the same repository-root `.mch` config and Flow files as starting from the repository root.
- A valid `.mch/config.yaml` with `temp_dir: /tmp/custom-mch` causes new CLI temporary planning workspace behavior to use `/tmp/custom-mch` instead of a hardcoded `/tmp/mch`.
- CLI tests may set `temp_dir` to paths under the system temp directory, including paths returned by `t.TempDir()`, as long as production code obtains the path only from loaded config.
- The default Flow utility prompts used by CLI temporary planning files must use the resolved `temp_dir` supplied by the runner instead of hardcoding `/tmp/mch`; for idea rewrites, the agent prompt must cause Codex to read and replace the `initial-idea.md` file in the configured workspace.
- Selecting a current project persists the selected numeric `project_id` in the committed repository-root `.mch/config.yaml` as the shared branch-scoped project selection, and does not create, read, or write `cli/.config/config.yaml`.
- The Main screen command list includes `/config` and selecting it opens a read-only `Config` view.
- The `/config` view displays resolved values from structs, including `backend_url`, `temp_dir`, `project_id`, repository root, config path, active Flow directory, Flow Step data, hook commands, stage modes, task statuses, and task steps.
- The `/config` view includes hook command strings exactly as loaded from Flow config.
- Returning from `/config` with `/return`, Esc, or Ctrl+C on an empty prompt returns to `MainState` and does not save or rewrite any config file.
- Removing or corrupting `.mch/config.yaml`, `.mch/default/flow.yaml`, or `.mch/default/help.yaml` in a test fixture produces a verbose startup config error and does not load legacy `cli/.config/config.yaml`.
- Automated CLI tests prove config loading, Flow parsing, `temp_dir` use, `/config` command routing, read-only config rendering, and failure behavior.
- Documentation under `docs/architecture/cli.md`, `docs/architecture/mch.md`, and `docs/operations/verification.md` describes the repository-root `.mch` config layout, default Flow profile loading, and `/config` verification behavior.

## Non-Goals

- Do not execute Flow hooks in this Change.
- Do not implement Flow profile selection.
- Do not implement config overrides, fallback layering, `.mch/local`, or user-local config.
- Do not sync Flow config into backend `public.config`.
- Do not add backend Flow assignment, Run claiming, Run update, claim reset, branch reconciliation, or per-Change stage mode controls to the CLI.
- Do not change backend behavior, frontend behavior, database schema, seed data, migrations, or files under `db/**`.
- Do not replace existing Change list/detail/create/update behavior except where it must read the new config path or `temp_dir`.
- Do not keep `cli/.config/config.yaml` as a supported config source, compatibility fallback, or current-project persistence target.
- Do not preserve current project selection as user-local CLI state in this Change.

## Design Notes

- `docs/architecture/mch.md` currently documents `cli/.config/config.yaml`; this Change intentionally moves the reference `mch` config contract to repository-root `.mch/config.yaml` and should update docs in the documentation cycle for this Change.
- Documentation and code currently tie `project_id` persistence to the old local config file and describe current project selection as user-specific state. This Change intentionally supersedes that CLI behavior: `.mch/config.yaml` is the sole owner of `backend_url`, `temp_dir`, and `project_id`, and its `project_id` is shared repository/branch configuration.
- The intended collaboration model is one backend URL and one current project ID per repository branch. Different backend users working in the same branch should use the same committed `project_id` through `mch`.
- Existing `mch` architecture still applies: Bubble Tea owns app state, commands do asynchronous work, rendering must use model data only, and package boundaries should stay focused.
- The committed repository already contains `.mch/config.yaml` and `.mch/default`, and this Change should treat those files as the default project configuration source.
- The `/config` screen can render deterministic YAML-like or text output generated from resolved structs, but the source of truth for the view must be loaded in-memory config, not direct file contents.
- The Flow config example from the idea is:

```yaml
  entry: make idea-entry
  exec: codex review origin/stage
  exit: ls
```

- When hook execution is implemented later, every hook command should be executed exactly as written with the working directory set to:

```text
  <git-root>/.mch/default/
```

- This means `make idea-entry` naturally uses:

```text
  <git-root>/.mch/default/Makefile
```

- No special command rewriting is needed.
- The Flow model vocabulary is: Flow = reusable automation definition; Step = one named stage inside the Flow; Run = one execution attempt of a Flow; Task = one unit of work inside a Run for a specific Step; Worker = executor/tool/process that performs a Task.
- A concrete example is: Flow `Change Automation`, Step `code`, Run `Run #42 for change/add-project-selector`, Task `execute codex_exec for step code in Run #42`, Worker `codex_exec`.
- Concurrency remains compatible with this model: one Flow can have many active Runs, each Run progresses independently through the Flow's Steps, each Run creates Tasks for its Steps, and Workers perform Tasks.
- Shortest correct sentence: A Flow defines Steps; a Run executes a Flow; a Task performs one Step within a Run; a Worker executes the Task.
- `run_stage` in the idea maps to loaded Flow Step slugs. `stage_mode`, `task_status`, and `task_step` map to loaded help/config option groups and should be represented as config data, not hardcoded application behavior.
- Valid Change type slugs are sourced from repository seed data; this spec uses `feature|docs|test`.

## Relevant Specs

- `specs/115-cli-config-loading.md`
- `docs/architecture/cli.md`
- `docs/architecture/mch.md`
- `docs/concepts.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/agent-interaction.md`
- `docs/functionality/current-project-context.md`
- `docs/operations/verification.md`

## Verification

- `(cd cli && make lint)`
- `(cd cli && go test ./...)`
- `(cd cli && go build -o /tmp/mch ./cmd/mch)`
- `awk 'FNR > 300 { print FILENAME " exceeds 300 lines"; failed = 1; nextfile } END { exit failed }' docs/architecture/cli.md docs/architecture/mch.md docs/operations/verification.md docs/docs-rules.md`
- `rg -n "cli/\\.config/config\\.yaml|\\.mch|/config|temp_dir|Flow profile" docs/architecture/cli.md docs/architecture/mch.md docs/operations/verification.md`
- `! git check-ignore -q .mch/config.yaml`

## QA Test Cases

- Start `mch` from the repository root with valid `.mch/config.yaml` and `.mch/default` files; verify the Main screen loads, backend URL uses repository config, and current project behavior uses `project_id` from `.mch/config.yaml`.
- Start `mch` with `project_id` missing and with `project_id: 0`; verify both configs are valid no-selection states and follow the same no-current-project startup behavior.
- Start `mch` from a nested directory inside the repository; verify the same repository-root config path and Flow directory are shown in `/config`.
- Change `temp_dir` in a test fixture config; start the `/new-change` flow and verify temporary planning files are created under the configured directory.
- Use a `temp_dir` value under a disposable system temp directory in automated tests and verify the CLI still treats it as a loaded config value, not as a built-in default.
- Select a different current project; verify the selected numeric `project_id` persists to the committed repository-root `.mch/config.yaml` as shared branch-scoped config and no `cli/.config/config.yaml` file is created or updated.
- Open the Main screen command menu; verify `/config` appears and opens the read-only config view.
- Inspect `/config`; verify it shows resolved CLI config, active Flow metadata, ordered Step slugs, prompt paths, hook command strings, stage modes, task statuses, and task steps.
- Return from `/config` with `/return`, Esc, and Ctrl+C on an empty prompt; verify each path returns to `MainState` without writing config files.
- Corrupt `.mch/config.yaml`; verify `mch` returns a verbose startup config error and does not fall back to `cli/.config/config.yaml`.
- Corrupt `.mch/default/flow.yaml`; verify `mch` returns a verbose startup Flow config error and does not start the interactive app.
- Corrupt `.mch/default/help.yaml`; verify `mch` returns a verbose startup Flow help config error and does not start the interactive app.
- Use custom, missing, empty, and reordered Flow help option groups; verify listed options with non-empty slugs load as configured and are rendered in `/config` without hardcoded vocabulary substitution.
- Configure a duplicate Flow Step slug, a missing Flow Step mode, and a Flow Step mode outside `skip`, `prompt`, and `exec`; verify startup/config loading fails with path-specific errors.
- Remove or hide `.mch/default` in a test fixture; verify startup/config loading fails clearly and does not attempt profile discovery.
- Confirm no backend request is made when opening or leaving `/config`.
- Confirm Flow hook strings displayed in `/config` match the loaded strings exactly and are not executed.
- Confirm `.mch/config.yaml` is not ignored by Git.

## Review Focus

- Repository-root detection and config path resolution from nested working directories.
- Removal of hardcoded `/tmp/mch` behavior in favor of loaded `temp_dir`.
- Clear separation between runtime temp workspace configuration and disposable temp paths used by tests or verification commands.
- Strict use of `.mch/default` with no fallback layering, profile selection, or accidental `cli/.config/config.yaml` compatibility path.
- Typed parsing of `.mch/default/flow.yaml` and `.mch/default/help.yaml`, including ordered Steps and exact hook command string preservation.
- `/config` rendering from in-memory structs without file dumping or backend calls.
- Command routing, return behavior, and tests for the new `Config` view.
- Documentation updates that reconcile the old `cli/.config/config.yaml` contract with the new repository-root `.mch` layout.
- Code updates that remove `cli/.config/config.yaml` loading and persistence in favor of `project_id` in `.mch/config.yaml`.
- Explicit replacement of user-local CLI current-project semantics with committed repository/branch-scoped `project_id` semantics.
- Preservation of backend/frontend/database boundaries and no unauthorized edits under `db/**`.

## Follow-Ups

- Implement Flow hook execution using the loaded hook commands and `.mch/default` working directory.
- Add Flow profile selection when product requirements define multiple profiles.
- Add overrides or local/user config only after the product defines a non-committed config contract.
- Connect loaded repository Flow config to backend Flow assignment or `public.config` only in a dedicated backend Change.
- Add CLI controls for Run claiming, Run updates, claim reset, per-Change stage modes, and branch reconciliation in dedicated Changes.
