# CLI Flow Redesign

stage -> step -> task is a new hierarchy:

- Stage: lifecycle grouping such as initialize, implement, review, deploy
- Step: ordered workflow outcome such as spec, docs, code, production
- Task: executable unit such as spec-write, code-review, deploy-master

task slugs are globally unique

The CLI Flow loader must be updated as part of this Change. Replace the current
top-level `steps` parsing with typed parsing for the new `stage -> step -> task`
hierarchy. The typed task configuration must load `slug`, `help`, `mode`, `prompt`,
`script`, `args`, `stdin`, `expect`, `to`, and `until`. Validation must reject empty
required fields, unsupported task modes, and duplicate task slugs across the entire
Flow. The committed default Flow must load successfully, and automated tests must
cover nested parsing, every task field, global task-slug uniqueness, validation,
and application startup with the new format.

This loader and validation work is mandatory implementation scope for the current
Change, not a follow-up. The Change is not complete until the CLI code and tests use
the new hierarchy and successfully load the committed default Flow and task modes.

Task modes execute as follows:

- `editor` opens `<tmp-dir>/output.md` in the configured editor.
- `exec` executes the configured prompt using
  `scripts/codex-exec-restore-session.sh`. After the script finishes, the task
  response is read from `<tmp-dir>/agent-output.md`; script stdout is not the task
  response.
- `prompt` opens an interactive agent session using
  `scripts/codex-resume-session.sh`.
- `script` executes the configured script.
- `loop` evaluates the last response line produced by the task identified by the
  globally unique task slug in `to`. When it equals `until`, the loop finishes.
  When it does not match, the Flow pauses and returns control to the user for
  remediation. The loop jumps back to the `to` task only after the user explicitly
  retries the paused Flow.

Task execution fields have these semantics:

- `prompt` is the prompt file for an `exec` or `prompt` task. Its path is relative
  to the Flow directory.
- `script` is the executable for a `script` task. Its path is relative to the Flow
  directory.
- `args` is an ordered list of arguments passed to the configured script. Runtime
  references such as `change.slug` are resolved from the active Flow context
  before execution.
- `stdin` identifies the runtime value, such as `change.idea`, `change.spec`, or a
  review `finding`, that is passed to the process through standard input without
  shell interpolation.
- `expect` defines the successful final response for a task. When configured, the
  last line of the task response must equal `expect`.
- `to` is required for a `loop` task and contains the globally unique slug of the
  task to jump back to.
- `until` is required for a `loop` task. The loop finishes when the last response
  line produced by the `to` task equals `until`.

Task field validation is mode-specific:

- Every task requires a non-empty, globally unique `slug` and a supported `mode`.
- `help` is optional for every task.
- `editor` requires no additional fields because it always opens
  `<tmp-dir>/output.md`.
- `exec` requires a non-empty `prompt`. `expect` is optional.
- `prompt` requires a non-empty `prompt`.
- `script` requires a non-empty `script`. `args` and `stdin` are optional.
- `loop` requires non-empty `to` and `until` values. `to` must resolve to an
  existing task slug.

Please review new Flow:
- plan-edit / plan-create
- plan-sequence - single session
  - idea-rewrite
  - idea-review
  - idea-fix - interactive (skip on output==`No questions or suggestions.`)
  - spec-write
- spec-review-loop - new session for each iteration
  - spec-review (exit-loop-on output=`No blocking issues found.`)
  - spec-fix-directions - interactive
  - spec-fix
- docs-sequence - single session
  - docs-update
- code-sequence - single session
  - code-implement
  - code-polish - interactive
  - pr-write
  - pr-publish
- code-review-loop - new session for each iteration
  - code-review
  - pr-comment (exit-loop-on comment_body=`No blocking issues found.`)
  - code-fix-directions - interactive
  - code-fix
  - pr-fix
- finalize-sequence - single session
  - code-docs-spec
  - merge
  - deploy
