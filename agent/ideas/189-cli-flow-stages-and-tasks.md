# CLI Agent Stages and Tasks Config

- Use new repo root .mch dir for config files.
- Dir has no secrets and it's commited entirely, no .gitignore
- .mch/default is used for default config files
- do not implement overrides yet, let's just always use .mch/default/* for the time being
- 

---

Options

run_stage:
- idea   - capture or refine the initial change idea
- spec   - create the implementation-ready spec
- ready  - mark the change as ready for execution
- docs   - update required documentation before implementation
- code   - implement the change
- polish - user-guided refinement after coding
- pr     - create or update the pull request
- review - review the PR/change against the spec
- fix    - address review findings
- sync   - automatically align spec, QA cases, and docs with final behavior
- merge  - merge the change branch
- stage  - promote/merge into stage
- master - promote/merge into master

task_step:
- none   - task has not started yet
- entry  - entry script is executing
- prompt - interactive session is running
- agent  - automated agent is executing
- exit   - exit script is executing
- done   - task has finished

task_status
- queued    - task is waiting to start
- running   - task is actively executing
- paused    - task is temporarily paused
- stopped   - task was manually stopped
- waiting   - task is waiting for input
- completed - task finished successfully
- failed    - task finished with an error

stage_mode
- skip   - stage will not execute
- prompt - stage will run an interactive session
- exec   - stage will run an automated agent

---

Spec

Flow
has many Steps
has many Runs

Step
belongs to Flow

Run
belongs to Flow
has many Tasks

Task
belongs to Run
belongs to Step
uses Worker

Worker
can perform many Tasks

Plain meaning:

Flow = reusable automation definition
Step = one named stage inside the flow
Run = one execution attempt of a flow
Task = one unit of work inside a run for a specific step
Worker = executor/tool/process that performs a task

Concrete example:

Flow: Change Automation
Step: code
Run: Run #42 for change/add-project-selector
Task: execute codex_exec for step code in Run #42
Worker: codex_exec

Concurrency still fits:

One Flow can have many Runs active at the same time.
Each Run progresses independently through the Flow’s Steps.
Each Run creates Tasks for its Steps.
Workers perform Tasks.

So the shortest correct sentence is:

A Flow defines Steps; a Run executes a Flow; a Task performs one Step within a Run; a Worker executes the Task.
