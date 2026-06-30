# Documentation Rules

## Purpose
Documentation in `docs` is the product source of truth for developers and agents. It records product vocabulary, behavior, architecture, local operation, and delivery plans.

## Structure
Use this structure for current documentation:

```text
docs/
  product-overview.md
  concepts.md
  architecture/
    system-architecture.md
    backend-api.md
    frontend-spa.md
    cli.md
  functionality/
    change-lifecycle.md
    requirements-and-acceptance.md
    history.md
    pr-integration.md
    agent-interaction.md
    current-project-context.md
  operations/
    verification.md
    local-development.md
  plans/
    01-project-roadmap.md
  guidelines/
  research/
```

## Rules
- Keep each Markdown file at or below 300 lines.
- Write current product documentation, not migration notes.
- Use `change`, `epic`, `test case`, and `history` as the core product vocabulary.
- Use `title`, `requirement_body`, `pull_request_body`, and `pull_request_url` for change text and PR fields.
- Do not describe obsolete hierarchy or obsolete terminology as active product behavior.
- Functionality docs define product behavior.
- Architecture docs define technical structure.
- Operations docs define how to run, test, and verify the product.
- Plans define staged delivery.
- Guidelines provide general product development considerations.
- Research stores exploratory notes and document spikes.
