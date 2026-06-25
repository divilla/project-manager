# Specs - specifications

- Specifications are all contained in folder `specs`
- Specifications precisely define the desired external behavior and constraints
- Specifications are single-source-of-truth for developers - they support every decision relevant to project
- Specifications must not be overly detailed and a single spec file has a maximum of 300 lines

```
specs/
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
    smart-history.md
    pr-integration.md
    agent-interaction.md
    current-project-context.md
  guidelines/
  operations/
    verification.md
    local-development.md
  plans/
    01-project-roadmap.md
  research/
```

- Functionality specs define product behavior.
- Architecture specs define technical structure.
- Operations specs define how to run, test, verify, deploy.
- Concept specs define vocabulary and domain model.
- Plans define project phases
- Guidelines provide general considerations for product development
- Research is dump of various document spikes
