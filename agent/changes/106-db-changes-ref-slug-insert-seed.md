# DB Changes ref, slug, sp_insert_change and db/seed-demo.sql

## Apply DB changes
- ref and slug are added to change table, ref is used as a per/project sequencer
- last_ref was added project project
- ref, slug or last_ref cannot be inserted or updated by user
- all inserts to change table must be done via fn_change_insert postgres function returning new change id
- read other changes by comparing db/init.sql
- backend/internal/dto changes must be used as reference
- ProjectListRequest is removed
- apply changes across backend, frontend and cli

## Update and scrape seed demo data
- Use scraped data to update `db/seed-demo.sql`
- Scrape PRs from https://github.com/labstack/echo/pulls?q=is%3Apr+is%3Aclosed 
- Scrape 200 PRs, use data to generate seed data for inserting 100 change rows
- Use project demo1 for all imports
- Set project demo1 last_ref to 200
- Use one echo PR to create arguments and call fn_change_insert for inserting change
- Use one PR to update pull_request_body

## Add DB backup to repo
- db backup is added to this repo
