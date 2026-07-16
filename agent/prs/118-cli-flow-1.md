# CLI Flow 1: Reusable Flow Runtime

## Summary
- Adds an uncomposed, reusable CLI Flow runtime with YAML-representable typed definitions, runtime-only context, shared validation, and composition-supplied terminal navigation.
- Implements generic Editor, Exec, Interactive, Preview, and Error Screens over a shared Step lifecycle and artifact workspace.
- Adds fakeable persistence and external-process boundaries, focused Change API storage for Idea, Spec, and PR artifacts, automated coverage, and supporting architecture and verification documentation.

## Behavior
- Validates definition, Step, task, Screen, command, artifact, task-sequence, field, and typed destination contracts before runtime execution.
- Creates each Step baseline under the configured `<temp_dir>/<artifact>/`, loads exact persisted bytes into `input.md` and `output.md`, preserves `input.md`, and saves changed output before entering Preview.
- Skips persistence for unchanged output and prevents stopped, cancelled, failed, or unsuccessfully saved Steps from reaching Preview.
- Supports Exec exact final-line evaluation and cancellation, same-Step Exec-to-Interactive and Interactive-to-Editor continuation, Preview/Diff toggling, fresh loads for `step` destinations, direct terminal navigation for `screen` destinations, and concrete Error return behavior.
- Keeps the runtime behind its composition boundary so existing CLI routes and product flows do not start it.

## Persistence and External Operations
- Introduces an `ArtifactStore` boundary that exposes only plain artifact bytes to the runtime.
- Maps Idea, Spec, and PR loads to the existing Change get API and maps saves to the matching focused update endpoint with `agent_edit: false`; unsupported artifacts are rejected.
- Routes editor, Exec, Interactive, Preview, Diff, and cancellation work through fakeable Bubble Tea command boundaries with configured Flow and artifact working directories.

## Tests
- Covers fresh and invalid definitions, validation-before-execution, artifact workspace resolution, exact baselines, save ordering, and immutable `input.md` enforcement.
- Covers Editor, Exec, Interactive, Preview, Diff, Error, cancellation, process-group termination, typed navigation, and missing-resource failures using fake stores and operations.
- Covers focused Change API endpoint selection, user-edit provenance, unsupported artifacts, and load/save failures without contacting a backend.

## Verification
- Passed with package-pattern `matched no packages` warnings: `cd cli && make lint`.
- Passed: `cd cli && go test ./...`.
- Passed: `cd cli && make race`.
- Passed: `cd cli && go build -o /tmp/mch ./cmd/mch`.
- Passed: `wc -l docs/architecture/cli.md docs/architecture/mch.md docs/operations/verification.md` (41, 249, and 68 lines respectively).
- Passed: `rg -n "Flow runtime|Flow definition|Flow context|ArtifactStore|temp_dir|input.md|output.md|Preview|destination|/return" docs/architecture/cli.md docs/architecture/mch.md docs/operations/verification.md`.

## References
- `specs/118-cli-flow-1.md`
- `docs/architecture/cli.md`
- `docs/architecture/mch.md`
- `docs/architecture/backend-api.md`
- `docs/operations/verification.md`
- `docs/docs-rules.md`
