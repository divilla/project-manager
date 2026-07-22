Read the software change idea from `/stg-tmp-dir/input.md`.

Rewrite the idea for clarity, structure, grammar, and readability without changing its intent, scope, or product
decisions.

Write the rewritten idea to `/stg-tmp-dir/output.md`.

Rules:

- Preserve all meaningful user intent.
- Preserve the original level of detail.
- Keep the document an idea draft.
- Fix grammar, wording, ordering, and readability.
- Improve wording, organization, headings, and bullet structure where useful.
- Wrap non-code Markdown lines at 100 characters or fewer.
- Remove repetition when it does not remove meaning.
- Clarify vague wording only when the intended meaning is obvious.
- Preserve fenced code blocks exactly, including language tags and contents.
- Preserve concrete labels, names, paths, commands, API shapes, examples, and quoted text unless the input clearly
  contains a typo.
- Do not invent product decisions, implementation details, requirements, verification steps, or
  new scope.
- Do not add Change spec sections or turn the idea into a requirements document, implementation plan, or Change spec.
- Do not implement code.
- Do not edit any files except `/stg-tmp-dir/output.md`.

After writing the file, print exactly:

`Done.`
