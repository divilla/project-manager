# Spec 01: Product Requirements Document (PRD)

## 1. Executive Summary & Vision
Traditional software project management tools (such as Jira, Linear, or Trello) focus heavily on **time or effort estimations** (hours, story points, velocity). However, human estimation is notoriously inaccurate, causing missed deadlines, bloated sprints, and team friction.

This project introduces an **AI-Powered Project Management Tool** built on a fundamentally different paradigm: **Completeness over Estimation**. 

Rather than asking "How long will this take?", the tool focuses on **"What does 100% complete look like?"** and **"What verified progress has been made?"** It uses artificial intelligence to help break down high-level features into discrete, verifiable components of completeness and validate status, creating an objective, progress-based dashboard.

The strategy begins with a **Developer Prototype**—a lightweight, single-user/trusted-team web application designed to run locally. This prototype will validate the core completeness tracking and AI planning concepts before expanding into a multi-tenant, enterprise-ready platform.

---

## 2. Core Philosophy: Completeness vs. Estimation

### The Estimation Problem
- **Inherent Bias:** Developers naturally underestimate complexity (optimism bias).
- **Gamification:** Story points are often gamified, losing correlation with actual product progress.
- **Micro-management:** Tracking hours shifts focus from output to input.

### The Completeness Solution
- **Discrete Milestones (Definition of Done):** Every task must have a clear, binary set of completeness requirements (e.g., "Database table created", "API endpoint unit tested", "Frontend routing configured"). Each single Definition of Done (DoD) item is defined as a 'requirement'.
- **Weighted Progress:** Progress is measured by the completion of these explicit requirements rather than arbitrary percentage guesses or time spent.
- **AI-Driven Decomposition:** High-level goals are transformed by an LLM into explicit requirements, which are categorized by phases.
- **Empirical Status:** The "Dashboard" aggregates requirement progress directly to provide a real-time, objective completeness score for the entire project, grouped by functional phases.

---

## 3. Product Scope: Prototype vs. Full-Scale (V2)

To ensure high velocity and quick validation during the prototype phase, we strictly separate the prototype scope from the future production application.

| Feature Area | Prototype Scope (Current) | Full-Scale Solution (V2) |
| :--- | :--- | :--- |
| **Authentication & Auth** | None. Single trust domain, no login screens. | Multi-tenant RBAC, OAuth2, and team permissions. |
| **User & Team Management**| None. Single-user mode or shared team session. | Team workspaces, user profiles, invitations. |
| **Task Assignment** | No assignments. Tasks are shared in a single pool.| Explicit task assignees, watchers, notifications. |
| **Core UI / Layout** | Top menu navigation: Home, Planning, Projects, Help. | Side navigation, personalized user dashboard. |
| **AI Integration** | Basic prompt workflows for planning and breakdown. | Background AI agents, code integrations, chat. |
| **Database** | Existing local PostgreSQL database used as-is. No prototype migrations or lookup-table changes. | Future database changes only through explicit human-owned schema design. |

---

## 4. Key Functional Areas (Prototype)

The prototype centers around four main user screens accessible via a top navigation menu:

### A. Home Dashboard
- **Goal:** Provide an immediate, visual snapshot of the project portfolio's overall health and completeness.
- **Features:**
  - High-level progress charts showing project completion percentages.
  - Kanban or list style view of task phases loaded from the existing `task_phase` table, indicating the aggregated completeness of tasks within each phase.
  - Highlighted bottlenecks using the existing phase semantics from `task_phase` (e.g., review-like work below 100% completeness or active work with no completed requirements).

### B. Planning Copilot
- **Goal:** Leverage generative AI to brainstorm, scope, and break down project initiatives.
- **Features:**
  - Conversational interface (Planning Chat) where the developer can write a high-level goal (e.g., "Implement markdown-based documentation generator").
  - AI suggests a list of discrete requirements grouped by valid development phases loaded from the existing database.
  - The developer can edit, add, or reject items before pushing them directly to the active project task board.

### C. Projects & Tasks
- **Goal:** Manage projects and their associated tasks, allowing granular progress tracking.
- **Features:**
  - Project listing and creation screens.
  - Task management: create, read, update, delete (CRUD) tasks.
  - Interactive requirements list for each task where checking off a requirement recalculates the task's completeness percentage instantly.
  - Manual phase transitions using options loaded from the existing `task_phase` table.

### D. Help & Documentation
- **Goal:** Instruct developers on how the tool operates, how to write prompt templates, and how to structure project completeness criteria.
- **Features:**
  - Standard markdown-rendered documentation page.
  - Guidelines on utilizing the "Completeness" framework effectively.

---

## 5. Non-Functional Requirements (Prototype)
- **Local First:** Easy setup and execution on developer machines via simple commands.
- **Responsive Interface:** Responsive UI built with Quasar to support both standard monitors and tablet displays.
- **Auditability:** The existing `task_history` and `requirement_history` tables must capture the current task/requirement version before every user or AI update/delete. Delete history rows must be marked with `deleted = true`.
- **Performance:** DB queries and AI prompts must return responses fast enough to maintain interactive dashboard fluidness (sub-second UI updates for data, <10s for LLM interactions).
