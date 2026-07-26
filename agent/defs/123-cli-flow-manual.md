# CLI Flow Manual

- def, spec, and pr are artifacts.
- input identifies the artifact loaded from the database.
- output identifies the artifact saved to the database after success.
- input.md and output.md are temporary workspace files representing artifacts.
- “Docs” or “documentation” remains reserved for repository documentation.

## input and output setting in flow.yaml

1. Load the input document once from the backend.
2. Write those exact bytes to both input.md and output.md.
3. Treat input.md as the immutable baseline.
4. Let the component modify only output.md.
5. On success, save output.md to the document named by output.

I’d implement it by writing the fetched byte slice to both files directly, rather than performing a filesystem copy:

```go
content := loadDocument(step.Input)

writeFile("input.md", content)
writeFile("output.md", content)
```

That works even for:

```yaml
input: def
output: spec
```

## input-path and output-path setting in flow.yaml

- `input-path` and `output-path` only come into play when `branch-init` was previously ran and the slug
  can be extracted from the branch name using regex `^change/([0-9]+-[0-9A-Za-z_-]+)$` and slug matches
  `slug` from change.slug field
- path mode is active when the current branch matches the Change branch pattern and its extracted slug exactly
  equals change.slug.
- if all scenarios where the above is not true, fallback to `input.md` and `output.md` in temp dir mechanism
- that means when slug is matched you still need to write to both files on step start, but target files are
  now `input-path` and `output-path`
- and read `output-path` file to update DB after step was executed

In Path mode, the filesystem is the source of truth. The database is updated from filesystem content.

The only exception is initial file creation:

- If a configured path exists, its content wins and may update DB.
- If it does not exist, create it from the corresponding DB artifact, even when that artifact is empty.
- After creation, the filesystem becomes authoritative.

## Temporary mode

When no matching Change branch is active:

1. Load the input artifact from DB.
2. Write it to both temporary input.md and output.md.
3. Run the step.
4. After success, save temporary output.md to the DB output artifact.

## Path mode

When the current branch slug matches change.slug:

1. Read input-path.
2. Compare the exact file content with the exact database artifact content.
3. If they differ, update the DB input artifact with the file content.
4. Do not overwrite either configured path from DB.
5. Run the component using input-path and output-path.
6. After success, read output-path and update the DB output artifact when their exact contents differ.

A few resulting rules:

- Never copy input-path over output-path in path mode.
- A missing input-path is an error; falling back to DB would violate filesystem authority.
- output-path may initially be absent if the step is expected to create it, but it must exist after successful execution.
- On failure or cancellation, do not restore files from DB. The filesystem remains authoritative.
- Compare exact content so the DB artifact remains byte-for-byte identical to its filesystem source.
- When no path is configured, use temporary `input.md` and `output.md` and initialize both with the
  same input artifact content.
- `path` configures only the input file. It does not configure the output file.
- When `path` is configured without `output-path`, read input from `path`, initialize temporary
  `output.md` with the same content, and use temporary `output.md` as the output file.

So the concise rule is:

> Temporary mode synchronizes DB → workspace → DB. Path mode reconciles filesystem → DB before and after execution.

In Path mode, missing files are initialized from their corresponding DB artifacts:

- Missing input-path → create it using the DB artifact named by input.
- Missing output-path → create it using the DB artifact named by output.
- An empty DB artifact produces an empty, zero-byte file.
- Create missing parent directories as needed.
- Never overwrite a path that already exists; existing filesystem content remains authoritative.

After creation, normal Path-mode reconciliation applies.

For example:

```yaml
input: def
output: spec
input-path: agent/defs/$slug.md
output-path: specs/$slug.md
```

If neither file exists:

- agent/defs/$slug.md is created from the DB def artifact.
- specs/$slug.md is created from the DB spec artifact, including as an empty file when spec is empty.

If input-path and output-path resolve to the same path, create it once.

## Optional commit field

A step in flow.yaml may define an optional commit field. When present, the Flow must commit and push completed work after a successful editor, exec, chat, or script operation.

The Flow parses variables in the configured commit message, then invokes the script relative to the Flow directory:

```bash
./scripts/commit.sh "<slug>" "<parsed-message>"
```

The step is complete only when the commit and push succeed. If the script fails, the Flow must stop and report the error.

## Step execution lifecycle

A step may define optional `pre` and `post` scripts. The Flow executes every step in this order:

1. Run the optional `pre` script.
2. Run the step's main `editor`, `exec`, `chat`, or `script` operation.
3. Run the optional `post` script after the main operation completes successfully.
4. Process the optional `commit` field after all previous actions succeed.
5. Evaluate the configured transition and continue to the selected step.

If `pre`, the main operation, `post`, or `commit` fails, the Flow must stop, show the error to the
user, and skip every remaining action for that step. The `type: chat` exit-status exception is
defined under Chat completion.

## Exec session behavior

Every type: exec step must define one session value:

```yaml
session: new | restore | resume
```

Any other value is invalid and must stop the Flow with a configuration error.

The configured prompt is passed to the corresponding script:

- session: new invokes ./scripts/codex-exec-new-session.sh <prompt>. It always starts a new Codex session and replaces the saved session ID. Any existing, missing, empty, or invalid session-id file is ignored.
- session: restore invokes ./scripts/codex-exec-restore-session.sh <prompt>. It attempts to restore the saved session. If the session-id file is missing, empty, invalid, or refers to an unavailable session, it starts a new session instead.
- session: resume invokes ./scripts/codex-exec-resume-session.sh <prompt>. It requires an existing valid session and resumes it. If the session cannot be resumed, the step stops and shows the error to the user.

A non-zero script exit stops the step. Neither post nor commit may run after such a failure.

## Step transitions

A step may define next-step, switch-output, both fields, or neither:

- next-step identifies the default step to run after successful completion.
- switch-output defines conditional transitions based on lines in the main operation’s textual output.

When both fields are present, the Flow evaluates switch-output first. A case matches when any complete output line exactly equals its configured value.

```yaml
next-step: spec-write-chat
switch-output:
  - case: Done.
    goto: spec-review
```

Transition matching follows these rules:

1. Read the main operation’s output line by line.
2. Compare complete lines using exact equality, without trimming whitespace.
3. Evaluate cases in their declared order; the first matching case wins.
4. When a case matches, select its goto step as the transition target.
5. If no case matches, continue to next-step.
6. If no transition produces a target, stop the Flow at the completed step.
7. Follow the selected transition only after `post` and `commit` complete successfully.
8. Validate every referenced step slug when loading flow.yaml.

## Chat completion

For a type: chat step, the interactive Codex process is the step’s main operation.

The Flow must wait until the user exits Codex before continuing:

1. Run the optional pre script or stop on error
2. Start the interactive Codex session or stop on error
3. Wait for the user to exit Codex. Ignore exit > 0
4. Run the optional post script or stop on error
5. Process the optional commit field or stop on error
6. Continue using the configured step transition.

---

Questions:

1. When path is configured without output-path, should it represent one authoritative file used for both input and output, or an input-only file with temporary output? The current Flow uses path for in-place artifact editing, while the definition
   specifies input-only behavior. (agent/defs/123-cli-flow-manual.md:82, .mch/default/flow.yaml:18)

2. Does Path mode activate solely when the branch slug matches change.slug, or only after this Flow ran branch-init? These rules differ when a matching branch was checked out manually or resumed, changing whether the filesystem or database is
   authoritative. (agent/defs/123-cli-flow-manual.md:35)

3. Is case: '*' a catch-all or a literal line containing *? Exact matching makes it literal, but the default Flow appears to rely on catch-all behavior. (agent/defs/123-cli-flow-manual.md:167, .mch/default/flow.yaml:164)

Suggestions:

4. Unify missing-path behavior. The definition says a missing input-path is an error, then later requires creating it from the database. It also first says to write both configured paths at startup, then says never to overwrite them. (agent/defs/123-
   cli-flow-manual.md:41, agent/defs/123-cli-flow-manual.md:76, agent/defs/123-cli-flow-manual.md:90)

5. Specify session behavior for the existing exec steps or define a default. The definition requires every exec step to declare new, restore, or resume, but the shipped Flow currently declares it only for spec-write.
6. Place filesystem/database reconciliation explicitly within the pre → main → post → commit → transition lifecycle. “After success” currently leaves it unclear whether output is saved before or after post and commit, especially when committing fails.
   (agent/defs/123-cli-flow-manual.md:128)
