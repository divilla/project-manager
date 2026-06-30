# DB Change References, Slugs, Insert Function, and Demo Seed Data

## Summary

Adds backend-owned Change identity fields with project-scoped references and stable slugs. Change creation now routes through fn_change_insert, which allocates project.last_ref, creates the Change row, and returns the new Change ID.

Updates the backend API, frontend, and mch client surfaces so ref, slug, and last_ref are returned and carried as read-only state, not user-editable inputs. Change create no longer accepts phase or pull request fields; pull_request_url now has a
focused update endpoint.

Refreshes demo data with 100 generated Changes for demo1 from paired closed Echo PRs, starting from last_ref = 200 and producing refs 201..300. Adds repository-local database backup artifacts only as reviewable rebuild/reference files.

## Notable Changes

- Adds change.ref, change.slug, project.last_ref, and fn_change_insert in db/init.sql.
- Widens Change ref counters to integer-compatible DTO fields.
- Adds POST /api/v1/change/update-pull-request-url.
- Removes active ProjectListRequest usage.
- Updates Change list/detail/create behavior across backend, frontend, and CLI clients.
- Updates docs for the observable Change identity, create, and update contracts.
- Hardens backend make api-test so stale API test servers are killed before setup and the health probe cannot pass against the wrong process.
- Adds agent prompt/template files present in this branch; these do not affect runtime behavior.

## Verification

- backend: make lint
- backend: make test
- backend: make api-test
- Controlled stale-server make api-test scenario
- git diff --check
- gzip -t db/backup/changes_106_db_changes_ref_slug_insert_seed.sql.gz

## References

- Slug: 106-db-changes-ref-slug-insert-seed
- Change: agent/changes/106-db-changes-ref-slug-insert-seed.md
