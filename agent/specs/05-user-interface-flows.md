# Spec 05: User Interface & Flows

The prototype frontend is built using **Vue 3** and the **Quasar Framework**. Quasar provides ready-to-use, professional Material-style components that accelerate UI development.

---

## 1. Global Navigation & Layout
The application features a single-page layout utilizing a top navigation bar (Top Menu) to switch between the four primary operational modules.

### ASCII Layout Outline
```
+-----------------------------------------------------------------------------+
| [Logo] AI-Project-Manager     [HOME]    [PLANNING]    [PROJECTS]    [HELP]  |
+-----------------------------------------------------------------------------+
|                                                                             |
|                               Active Page                                   |
|                                                                             |
+-----------------------------------------------------------------------------+
```

### Core Quasar Structural Components
- `<q-layout view="hHh Lpr lFf">`: Standard shell responsive layout.
- `<q-header>`: Contains the primary navigation toolbar.
- `<q-tabs>`: Standard route tabs displaying the options: Home, Planning, Projects, Help.
- `<q-page-container>`: Dynamic window rendering the current router path.

---

## 2. Page Breakdowns & Interactive Flows

### A. Home Dashboard Screen (Route: `/`)
The analytical hub displaying **Completeness Grouped by Phase** instead of standard burndown charts.

#### UI Mockup Conceptual Layout:
```
+-----------------------------------------------------------------------------+
|  PROJECT COMPLETENESS OVERVIEW                                              |
|  [Active Project Select: v "Project Phoenix"]                               |
|                                                                             |
|  Overall Completeness: [████████████████████████░░░░░░░░░░] 70%             |
|                                                                             |
|  +--------------------+  +--------------------+  +--------------------+     |
|  | PHASE: DB LABEL    |  | PHASE: DB LABEL    |  | PHASE: DB LABEL    |     |
|  | Completeness: 0%   |  | Completeness: 45%  |  | Completeness: 85%  |     |
|  | [██░░░░░░░░░░]      |  | [█████░░░░░]        |  | [█████████░]       |     |
|  | Tasks count: 5     |  | Tasks count: 3     |  | Tasks count: 2     |     |
|  +--------------------+  +--------------------+  +--------------------+     |
|                                                                             |
|  +-----------------------------------------------------------------------+  |
|  | BOTTLENECK WATCH (Tasks stuck in configured review/progress phases)    |  |
|  | - "Refactor Auth Middleware" (review-like phase, 1/3 criteria)        |  |
|  +-----------------------------------------------------------------------+  |
+-----------------------------------------------------------------------------+
```
#### Quasar Components Used:
- `<q-card>` & `<q-card-section>`: For structuring layout blocks.
- `<q-linear-progress>`: Rich progress indicators colored according to progress (e.g., Red for low, Orange for mid, Green for high).
- `<q-select>`: For switching the context project instantly.

---

### B. Planning Copilot Screen (Route: `/planning`)
The AI engine room where natural language ideas are transformed into structured tasks.

#### Page Sections:
1. **Interactive Prompt Area:** A chat-like layout where developers input directives.
2. **AI Suggested Actions Pane:** Displays the decomposed roadmap suggested by the LLM before saving it.
3. **Control Actions:** Buttons to check/uncheck suggestions, adjust target phases, and commit to the database.

#### Key Flows:
1. User types: *"Add multi-file upload capabilities to our storage driver."*
2. System displays loading spinner (`q-inner-loading`).
3. System shows the LLM-derived suggestions grouped by phases loaded from the database:
   - **Phase: Example DB Phase**
     - [x] Research AWS S3 multi-part limits
   - **Phase: Example DB Phase**
     - [x] Design DB table for multi-file attachments
   - **Phase: Example DB Phase**
     - [x] Build backend Echo attachment upload handler
     - [x] Design Vue drag-and-drop component
4. User clicks "Commit Plan". The system creates the tasks and nested requirements, redirecting the user to the Projects page.

---

### C. Projects & Tasks Board (Route: `/projects`)
The operational workflow view where task phases and requirements are managed.

#### UI Mockup Conceptual Layout:
```
+-----------------------------------------------------------------------------+
| PROJECTS / TASKS BOARD                                                      |
| [Create Project Button (+)]                                                 |
|                                                                             |
| +---------------------+   +---------------------+   +---------------------+ |
| | DB PHASE            |   | DB PHASE            |   | DB PHASE            | |
| +---------------------+   +---------------------+   +---------------------+ |
| | [Task #1]           |   | [Task #2]           |   | [Task #3]           | |
| | Setup Database      |   | Implement API       |   | Design CSS Layout   | |
| | Compl: 0%           |   | Compl: 50%          |   | Compl: 90%          | |
| | [░░░░░░░░░░]        |   | [█████░░░░░]        |   | [█████████░]        | |
| +---------------------+   +---------------------+   +---------------------+ |
+-----------------------------------------------------------------------------+
```

#### Detailed Task View (Dialog Drawer):
Clicking a task opens a dedicated modal containing:
- Task title and markdown-enabled description.
- **Interactive Requirements List (The Progress Driver):** 
  - Clicking a checkbox immediately triggers the backend `POST` request.
  - Upon server response, the local task's progress bar animates to reflect the updated value.
- **Phase Selector:** Quick dropdown button (`q-btn-dropdown`) populated from the existing `task_phase` table.

---

### D. Help screen (Route: `/help`)
An embedded onboarding page containing guidelines for the "Completeness" project management framework.

#### Layout:
- Left-hand side: Anchored navigation tree list (`q-list`).
- Right-hand side: Rendered Markdown body (`v-html` powered by a library like `marked` or native CSS typography).
- Includes prompt template guidelines showing users how to get the best task breakdowns from the AI Planning module.

---

## 3. Responsive & Aesthetic Guidelines
1. **Visual States (Feedback):** Ensure immediate visual response for user gestures. Checkboxes should trigger state changes instantly; backend saving indicators should show a small, non-intrusive header loader (`q-ajax-bar`) to avoid blocking user interactions.
2. **Color Palette:**
   - **Primary:** Deep slate blue (professional, technical).
   - **Secondary:** Teal (representing action).
   - **Completeness colors:** Dynamic transitions from Warm Red (0-30%) to Soft Orange (31-70%) to Vibrant Green (71-100%).
3. **Typography:** Clear sans-serif typeface (Inter, Roboto) utilizing Quasar standard text classes (`text-h5`, `text-subtitle2`, `text-body1`).
