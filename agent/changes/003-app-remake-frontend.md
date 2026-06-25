# Application Remake Frontend

- task is renamed to change
- some columns were dropped
- name is renamed to title
- description is renamed to body
- epic and epic_history are added
- change is no longer hierarchycal - it has fixed structure: change can be standalone or have reference to epic
- using new endpoints schemas as well as naming convention defined in `backend/internal/dto` - rename entire frontend to the new naming convention
- fix all frontend naming - refactor to the new database schema - rename `task` to `change` everywhere - the frontend must no longer contain word `task`
- fix all frontend tests to follow new naming convention
