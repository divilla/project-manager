Do not use any skills for this request.

Rewrite a raw idea draft from the configured temp directory into clear Markdown.

## Preconditions

Read `.mch/config.yaml` and extract the top-level `temp_dir` value into `<temp-dir>`.

Stop with one concise error if:

- `.mch/config.yaml` does not exist or cannot be read
- `temp_dir` is missing, empty, or not a single path value
- `<temp-dir>/input.md` does not exist or cannot be read
- `<temp-dir>/input.md` is empty or contains only whitespace

## Instructions

Read `<temp-dir>/input.md`.

Rewrite the input idea into clear Markdown and save it to `<temp-dir>/output.md`. If `<temp-dir>/output.md` already exists, overwrite it intentionally.

- Preserve all meaningful user intent.
- Preserve the original level of detail.
- Keep the document an idea draft.
- Fix grammar, wording, ordering, and readability.
- Remove repetition when it does not remove meaning.
- Clarify vague wording only when the intended meaning is obvious.
- Preserve fenced code blocks exactly, including language tags and contents.
- Preserve concrete labels, names, paths, commands, API shapes, examples, and quoted text unless the input clearly contains a typo.
- Do not invent product decisions, implementation details, requirements, acceptance criteria, verification steps, or new scope.
- Do not add Change spec sections or turn the idea into a requirements document, implementation plan, or Change spec.

Only write `<temp-dir>/output.md`. Do not modify any other file.

## Final Response

After successfully saving `<temp-dir>/output.md`, output exactly:

Done.
