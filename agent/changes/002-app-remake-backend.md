# Application Remake Backend

- task is renamed to change
- some columns were dropped
- name is renamed to title
- description is renamed to body
- epic and epic_history are added
- change is no longer hierarchycal - it has fixed structure: change can be standalone or have reference to epic
- you will find `backend/internal/dto` already renamed - the rename was done project wide - use this new naming convention accross entire backend
- fix all backend naming - refactor to the new database schema - rename `task` to `change` everywhere - the backend must no longer contain word `task`
- fix all backend unit and integration test to follow new naming convention
