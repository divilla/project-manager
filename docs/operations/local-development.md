# Local Development

## Services
Run the app locally with:

- PostgreSQL database
- Go backend on `http://localhost:8080`
- Vite frontend on `http://localhost:8000`

## Backend
From `backend`:

```sh
go run ./cmd/server
```

The backend reads configuration from environment variables and local config files. For temporary overrides, pass `-port` and `-db` to the server binary:

```sh
go run ./cmd/server -port 18080 -db postgres://postgres:postgres@localhost:5432/changes_test
```

Health endpoints should report API and database availability.

## Frontend
From `frontend`:

```sh
pnpm dev
```

The frontend uses Quasar and talks to the local backend.

## Database
Use the supplied PostgreSQL contract. Application work must not mutate schema outside explicitly approved database changes.

The repository-root `db` folder is database-owner territory. Agents must not edit, create, delete, move, or regenerate files under `db` or any of its subfolders unless the user explicitly requests that exact file change.

Agents must not run PostgreSQL commands that alter database structure, including create, alter, drop, truncate, grant, revoke, migration, or restore operations, unless the user explicitly requests that exact structural change.

Backup helpers live under `db/backup`. Restore commands can replace database contents, so use the correct target database.

Demo seed data is intended for disposable local databases. The demo dataset may include generated changes based on captured public pull request data, but running seed or restore commands must target an explicitly chosen disposable database.

Repository backup artifacts document a known local demo database state. They are not an automatic restore mechanism and must not be applied to a live or production database.

## Failure States
If the backend or database is unavailable, the frontend should show a clear non-blocking error state instead of an empty page.
