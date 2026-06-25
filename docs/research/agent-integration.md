Create an agent-facing command layer, for example:

./scripts/appctl project list
./scripts/appctl task list --project 1
./scripts/appctl task get --id 42
./scripts/appctl task create --project 1 --name "..."
./scripts/appctl verify-pr
./scripts/appctl seed-demo
./scripts/appctl reset-test-db

Internally, most of these should call the backend API. Some special commands may touch Postgres, but only for controlled operations like test setup, reset, backup, or diagnostics.

So the hierarchy should be:

1. Skill tells the agent what workflow to follow.
2. CLI scripts give the agent stable tools.
3. API remains the authoritative behavior boundary.
4. DB scripts are reserved for setup/inspection/reset.
5. Playwright/browser automation validates real SPA behavior.

Why CLI Over Direct API Calls

Direct API calls are possible, but a CLI gives you a better agent contract:

./scripts/appctl task get --id 42 --json

is better than asking the agent to remember:

curl -X POST http://localhost:8080/api/v1/task/get \
  -H 'Content-Type: application/json' \
  -d '{"id":42}'

The CLI can handle:

- base URL
- JSON formatting
- errors
- retries
- server readiness
- consistent exit codes
- test fixtures
- future auth
- concise output

That makes the agent more reliable.

What The Skill Should Contain

A skill should not contain all app knowledge inline. Keep it small and route to references/scripts.

Example structure:

aipm-agent/
  SKILL.md
  references/
    architecture.md
    api-contracts.md
    pr-workflow.md
    verification.md
  scripts/
    appctl

SKILL.md should say things like:

Use this skill when working on the Project Manager app.

Default interaction rules:
- Use backend API through `./scripts/appctl` for app behavior.
- Do not write directly to Postgres except through documented reset/seed/inspection commands.
- Use Playwright/browser checks for SPA workflow validation.
- For PR verification, run `./scripts/appctl verify-pr`.

DB Access Policy

I’d explicitly classify DB commands:

Allowed:
- inspect schema
- inspect test data
- reset local test DB
- seed demo data
- backup/restore local DB

Avoid:
- creating application records directly
- mutating task/project/requirement rows directly
- bypassing API validation

For your app especially, direct DB writes are risky because task history, requirement recalculation, and phase/completion logic may live behind backend/repository/procedure behavior.

Best Long-Term Shape

Start with shell scripts. If they grow, replace them with a small typed CLI.

Good progression:

Phase 1: Makefile + shell scripts
Phase 2: single `scripts/appctl` wrapper
Phase 3: Go/Node CLI if logic grows
Phase 4: MCP server only if you want rich tool integration

I would not start with MCP unless you already know you need it. A deterministic CLI plus skill instructions is simpler, debuggable, and works with any agent.

So my recommendation:

> Build a small agent-facing CLI that wraps the backend API, reserve DB scripts for controlled setup/inspection, and teach the agent through a concise skill when to use API commands, DB commands, tests, and browser checks.



## Programming language

My ranking for your case:

1. Go: best fit for repo consistency and long-term maintainability.
2. Node/TypeScript: good if the CLI mainly serves frontend/dev UX and you want to reuse TS types.
3. Python: good for quick automation, but adds another ecosystem and weaker type/contracts story here.
4. Shell: okay only as thin wrappers around stable commands.

A good architecture:

backend/cmd/appctl/main.go
backend/internal/appctl/...

or, if you want it repo-level:

cmd/appctl/main.go

The CLI should call the API over HTTP, not import backend services directly, for normal app operations.

Example commands:

go run ./cmd/appctl task get --id 42
go run ./cmd/appctl task list --project 1
go run ./cmd/appctl verify-pr

Eventually build it:

go build -o bin/appctl ./cmd/appctl
./bin/appctl task get --id 42

Use Go especially if you want:

- stable JSON output
- clear exit codes
- typed request/response structs
- easy distribution as one binary
- reuse of backend DTO shapes, if structured carefully
- low install friction for future agents

I’d still keep tiny shell aliases if useful:

./scripts/appctl task get --id 42

where scripts/appctl simply runs or builds the Go CLI. But the real logic should be Go.



## Current project as a per-user global preference

That means:

user.current_project_id

or better:

user_preferences.current_project_id

Semantically, it says:

> For this user, when no project is explicitly specified, use this project as their current working context.

That fits both browser and CLI/agent use.

Important distinction:

- Good: current project is global per user
- Bad: current project is global for the whole backend/app

A reasonable design:

users
  id
  ...

user_preferences
  user_id
  current_project_id
  ...

Or for a local single-user app before auth exists:

app_user
  id = "local"

user_preferences
  user_id = "local"
  current_project_id

Then:

- SPA reads current project from backend on startup.
- SPA updates backend when selector changes.
- CLI can read/use/update the same setting.
- Agent can share the same current project context.
- API endpoints can either accept explicit project_id or default to the user’s current project where appropriate.

I’d still keep this rule:

> Explicit project ID beats current project.

For example:

appctl task list

uses the user’s current project.

appctl task list --project 42

uses project 42 regardless of preference.

Backend API design could be:

GET /api/v1/me/preferences
POST /api/v1/me/current-project

or in your current POST-style API:

POST /api/v1/user/current-project/get
POST /api/v1/user/current-project/update

But I would avoid making every backend endpoint silently depend on current project. For mutating domain operations, explicit IDs are usually safer.

Good split:

  Operation                       Use current project?
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  List current project tasks      Yes
──────────────────────────────  ────────────────────────────────────────────
  Create root task from UI/CLI    Yes if no explicit project supplied
──────────────────────────────  ────────────────────────────────────────────
  Create child task               Use parent task’s project
──────────────────────────────  ────────────────────────────────────────────
  Update task                     Task ID is enough
──────────────────────────────  ────────────────────────────────────────────
  Delete task                     Task ID is enough
──────────────────────────────  ────────────────────────────────────────────
  Project selector state          Per-user current project
──────────────────────────────  ────────────────────────────────────────────
  Reports/dashboard               Default to current project, allow override

This gives you shared context for browser, CLI, and agents without making the backend ambiguous.
