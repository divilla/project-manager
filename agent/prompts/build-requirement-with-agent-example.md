You are helping turn a rough software idea into a clear, testable requirement specification.

The user will provide an initial idea below. Treat it as raw intent, not as a complete requirement.

Initial idea:
I'd like to wire enable application to use codex for change (requirements) building.
I'd like to use following flow:
- ask user for starting prompt roughly defining requirement
- parse user entry through `agent/prompts/build-requirement-with-agent.md`
- the starting entry will start new sessions so we need to capture codex_session_id
- the agent then asks for next prompt or answer agent's question
- the app must continue a session `printf <next-prompt> | codex exec -C /home/vito/go/src/project-manager/ resume <codex-session-id> -`
- final output must be 

Work in phases:

1. Inspect relevant repository files and documentation when that helps clarify current product behavior, API contracts, architecture, terminology, or constraints.
2. Ask the smallest useful set of clarifying questions before drafting if important product intent, scope, target user, persistence behavior, API/UI boundary, or acceptance criteria are unclear.
3. Challenge weak assumptions directly. If a request is ambiguous, risky, too broad, or conflicts with existing documentation, say so plainly and explain what must be decided.
4. Draft the requirement only when there is enough information to make it actionable.

Hard boundaries:

- Do not implement code, edit files, create commits, run migrations, or mutate databases in this session, even if asked later.
- Do not silently invent product decisions. Mark unresolved decisions as open questions.
- Do not produce vague acceptance criteria. Every acceptance criterion must be observable and testable.
- Do not use markdown tables unless the user explicitly asks for them.

Requirement types list must be retrieved from :

Final output contract:

- The first non-blank line must be an H1 requirement title.
- The H1 title must be concise enough to reuse as a planning item title.
- The first non-blank line after the H1 title must be the requirement type line.
- The requirement type line must contain only selected types joined by `|`, with no spaces.
- Example type line: `feature|test`
- Do not include any preamble before the H1 title.
- Do not wrap the final requirement in a code block.

Final requirement structure:

# Requirement Title

Types: feature|test|docs

Epic: xxx (optional)

## Problem Statement

State the problem, user need, and expected outcome in concrete terms.

## Primary Workflows

Describe the main user or system workflows that must work.

## Acceptance Criteria

List binary, testable outcomes. Each item should be independently verifiable.

## Edge Cases

List relevant failure states, empty states, invalid input, concurrency, persistence, permissions, integration, or recovery cases.

## Non-Goals

List related work intentionally excluded from this requirement.

## Dependencies And Risks

List technical dependencies, external tools, data contracts, operational risks, security/privacy concerns, and assumptions that could affect implementation.

## Open Questions

List unresolved product or technical decisions. Use `None.` only if there are no open questions.

Quality bar:

- Use the repository's product vocabulary.
- Prefer practical, implementation-ready language.
- Optimize for a requirement an engineer can implement without re-litigating scope.
- Keep the requirement concise, but make it detailed enough to serve as a strong foundation for high-quality documentation, implementation, tests, and review.
