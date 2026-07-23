# Naming Convention Refactor

We are refactoring the artifact naming convention by replacing `idea` with `def`.

The artifact terminology is:

| Short name | Full name       | Verb           |
|------------|-----------------|----------------|
| `def`      | `definition`    | `define`       |
| `spec`     | `specification` | `specify`      |
| `pr`       | `pull-request`  | `pull-request` |

This terminology follows the same pattern for issues, tasks, and features:

- `define issue`, `specify issue`
- `define task`, `specify task`
- `define feature`, `specify feature`

`def`, `spec`, and `pr` are three separate artifacts. All three are edited and written in the
shared `artifact` stage.

Do not mix the artifact name with the stage name. The `idea` artifact becomes `def`, or
`definition` where appropriate, while the former `idea` stage becomes `artifact` because it handles
`def`, `spec`, and `pr` writes.

## Task

Rename `idea` to `def`, or `definition` where appropriate, everywhere, including documentation,
the frontend, the backend, the CLI, scripts, Makefiles, prompts, files, and folders.

Rename every reference to the former `idea` stage to the `artifact` stage.

## Database

The definition column, database functions, update procedure, history type, and writable artifact
types have already been renamed.

The following remaining database-file corrections are explicitly authorized as part of this
Change:

- In `db/seed.sql`, replace the former `idea` stage value, currently written as `def`, with
  `artifact`; retain `def` as the definition artifact type in `change_docs`.
- In `db/seed-demo.sql`, restore the ordinary English phrase `Different take on #2673 ideas`; it
  does not name the definition artifact.

## Backend

The rename has been applied to ChangeDTOs. Renaming the rest is pending.

## Non-Goals

Do not rename `agent/defs` (renamed from `agent/ideas`), `agent/prs`, or `specs`.
