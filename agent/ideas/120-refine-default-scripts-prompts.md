# Refine Default Scripts and Prompts

## DB

- One argument, `_ref_uuid`, has been added to `fn_change_insert`.

## Backend

- `dto/ChangeCreateRequest` now includes `RefUUID`. Update the Change service and repository to
  accept the argument correctly. Continue using `fn_change_insert` in the repository.

## CLI

Establish CLI constants:

```go
const (
	// pass to scripts as env var $MCH_DEFAULT_DIR
    DefaultDir = ".mch/default"

    // pass to scripts as env var $MCH_TEMP_DIR
    TempDir = ".mch/tmp"
)
```

- The CLI must always pass `DefaultDir` as the `MCH_DEFAULT_DIR` environment variable to every
  script that requires it.
- Replace all hardcoded `.mch/default` or `/default` paths throughout the `.mch` and `cli`
  directories.
- The CLI must always pass `TempDir` as the `MCH_TEMP_DIR` environment variable to every script
  that requires it.
- Replace all hardcoded `.mch/tmp` or `/tmp` paths throughout the `.mch` and `cli` directories.
- Use `github.com/gofrs/uuid/v5` to generate a UUIDv7 when creating new Change `input.md` and
  `output.md` files. Use the UUID to create `.mch/tmp/<generate-UUIDv7>` with blank `input.md` and
  `output.md` files, which will then be used to generate a new Change through the API. The API now
  accepts the `ref_uuid` parameter when creating a Change.
- Depending on the stage the current Change has reached, there is a third subfolder: `MCH_STAGE`.
- The stages are: `idea`, `spec`, `spec-review`, `docs`, `code`, `pr`, `code-review`, `code-docs`,
  and `merge`.
- The CLI must therefore supply all scripts with three values related to the temporary directory:
  `$MCH_TEMP_DIR/$MCH_REF_UUID/$MCH_STAGE`.

## Default Scripts and Prompts

- Accept the refined `/scripts` and `/prompts` in `.mch/default/`.
- Accept the new `prompts/spec-file-structure.md` and adapt the rest of the repository to it.

## Other

- Accept the already applied `AGENTS.md` fixes.
- Documentation is no longer the source of truth. Code and the PR are the single source of truth.
- Documentation must follow the code, not the other way around.
