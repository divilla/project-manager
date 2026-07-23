# Change Into Spec

- Rename the legacy Change file structure prompt to `.mch/default/prompts/spec-file-structure.md`.
- Search the entire repository for references to the legacy Change file structure prompt, including prompts, scripts, Makefiles, and Codex skills, and update them to `.mch/default/prompts/spec-file-structure.md`.
- Clarify the wording around Change, Idea, and Spec:
  - Change is the entire flow and consists of multiple artifacts, such as branch, idea, spec, docs, code, and PR.
  - A Change has a title, types, and epic.
  - A Change has a slug:
    - branches are named `change/<change-slug>`
    - ideas are named `agent/ideas/<change-slug>.md`
    - specs are named `specs/<change-slug>.md`
  - Use the placeholder `<change-slug>` everywhere to remove ambiguity.
  - Change, Idea, and Spec share the same title.
  - When the Idea title changes, Change is updated.
  - When the Spec title changes, Idea and Change are updated.

The final implementation step of this Change must include removing the legacy Change file structure prompt file.

After this Change, references to the legacy Change file structure prompt must not appear in any part of this repository or in any Codex skill.
