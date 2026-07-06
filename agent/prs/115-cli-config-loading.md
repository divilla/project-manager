# CLI Config Loading

## Summary
- Moves `mch` runtime configuration to committed repository-root `.mch/config.yaml`, loaded after resolving the Git repository root.
- Adds the committed `.mch/default` Flow profile with `flow.yaml`, `help.yaml`, prompt files, demo hook scripts, and hook command strings for the default ordered Change workflow.
- Reworks CLI temporary planning files and current project selection to use loaded config values instead of the legacy `cli/.config/config.yaml` path or a built-in temp workspace.
- Adds a read-only Main screen `/config` view that renders resolved CLI and Flow config from in-memory structs.

## Behavior
- Starting `mch` from the repository root or a nested directory resolves the same Git root, reads `.mch/config.yaml`, and loads `.mch/default/flow.yaml` plus `.mch/default/help.yaml`.
- Missing or malformed `.mch/config.yaml`, `flow.yaml`, or `help.yaml` now produces a path-specific startup config error instead of falling back to legacy config.
- Missing `project_id` and `project_id: 0` remain valid no-current-project states; `/select-project` persists the selected numeric ID back to `.mch/config.yaml`.
- The `/new-change` planning workspace uses loaded `temp_dir`, and the Codex idea rewrite prompt includes the configured `initial-idea.md` path.
- `/config` appears in the Main command menu, does not call backend APIs, does not mutate files, and returns to `MainState` with `/return`, Esc, or Ctrl+C on an empty prompt.
- Flow hook strings are parsed and rendered exactly as loaded, but hooks are not executed in this Change.

## CLI
- Adds typed parsing for Flow metadata, ordered Steps, Step modes, prompt paths, entry/exec/exit hook strings, stage modes, task statuses, and task steps.
- Rejects duplicate Flow Step slugs, missing Step modes, unsupported Step modes, and empty Flow help option slugs.
- Keeps Flow help option groups as configurable data, allowing missing, empty, custom, and reordered help groups where slugs are non-empty.
- Adds focused tests for repository-root resolution, `.mch` config parsing, Flow parsing, configured `temp_dir`, `/config` routing/rendering, read-only return behavior, and legacy config non-fallback.

## Docs
- Updates CLI architecture, `mch` architecture, current project context, concepts, and verification docs for repository-root `.mch` config, branch-scoped `project_id`, configured `temp_dir`, `.mch/default` Flow loading, and `/config`.

## Verification
- Passed: `awk 'FNR > 300 { print FILENAME " exceeds 300 lines"; failed = 1; nextfile } END { exit failed }' docs/architecture/cli.md docs/architecture/mch.md docs/operations/verification.md docs/docs-rules.md`
- Passed: `rg -n "cli/\\.config/config\\.yaml|\\.mch|/config|temp_dir|Flow profile" docs/architecture/cli.md docs/architecture/mch.md docs/operations/verification.md`
- Passed: `! git check-ignore -q .mch/config.yaml`
- Not run while drafting this PR body: `(cd cli && make lint)`.
- Not run while drafting this PR body: `(cd cli && go test ./...)`.
- Not run while drafting this PR body: `(cd cli && go build -o /tmp/mch ./cmd/mch)`.

## Risks
- Scope override accepted before drafting: the branch also includes workflow idea/script artifacts outside the strict `115-cli-config-loading` contract, including `scripts/change-idea.pl` and additional files under `agent/ideas/`.

## References
- `specs/115-cli-config-loading.md`
- `docs/architecture/cli.md`
- `docs/architecture/mch.md`
- `docs/concepts.md`
- `docs/functionality/current-project-context.md`
- `docs/operations/verification.md`
