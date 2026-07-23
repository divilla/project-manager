# Ideas and Change Refactor

## Introduction

This change introduces some major shifts in naming and logic of operations.

Change consists of 4 major artifacts
- title: shared by bellow three
- idea: initial user idea rewritten until clear
- spec: specification (former change)
- pr: pull request body (description)

## New change rules and structure

- minimum requirement for change to be created is `title` and `idea`.
- `ref` and `slug` are optional and default to null
- `change_phase` is mandatory, but it defaults to `backlog`
- `types` are still mandatory, but now default to empty array
- `epic` is optional, defaults to null
- `spec` (former body) is optional, defaults to null
- `pr` and pr_url are optional and default to null

## New ref and slug flow

- `ref` and `slug` are no longer assigned with `fn_change_insert`, but in a separate procedure `sp_change_ref_update`
- `sp_change_ref_update` is independent of change update flow - updates can be done both prior and post `sp_change_ref_update`
- `fn_change_insert` now accepts only project_id, title, idea as arguments

## DB

- change.idea is added
- types are allowed empty array
- all changes are stored to init.sql - please review this file
- update seed-demo.sql to reflect new database changes

## Backend

- internal/dto/change.go is updated
- this update is single source of truth for all further updates
- the rest of the backend must be updated to accommodate changes
- frontend and cli must be updated to accommodate changes
- all systems must be thoroughly tested

## CLI

Flow change:
- User starts /new-change
- User writes Idea with Editor
- User saves and exits Editor
- CLI prompts: Create Change? Yes/No
- If No - return to ChangesListScreen
- If Yes - use agent to start new session and rewrite the skill
- Rewritten Idea is saved via API

Ideas flow:
- Every time Idea is Rewritten and Saved, user is routed ChangeIdeaScreen
- ChangeIdeaScreen loads Change data from API and displays Idea
- ChangeIdeaScreen prompts the user in the bottom
- Write Spec with Agent? Yes/No
- If No - User is routed to ChangeDetailsScreen
- If Yes - User is routed to ChangeDetailsScreen

Ref flow:
- /reference is added as a first command on ChangeDetailsScreen
- /reference executes api/v1/change/reference endpoint and refreshes the ChangeDetailsScreen
- check if app is executed in the git repo
- `git branch --list` and check whether `changes/<slug>` branch exists locally
- if yes checkout branch
- if not check whether `changes/<ref>-*` branch exists locally
- if yes checkout the branch and rename it `git branch -m changes/<slug>`
- if no check if the branch exists remote
- if yes checkout the branch
- if no check whether `changes/<ref>-*` branch exists remotely
- if yes checkout branch and rename it both locally and on remote
- if no create the branch and checkout `git checkout -b changes/<slug>`

**Important:** the above Yes flow is temporary and will be addressed in next changes. All the agent execution code must be preserved in fully.
