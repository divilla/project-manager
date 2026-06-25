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

The backend reads configuration from environment variables and local config files. Health endpoints should report API and database availability.

## Frontend
From `frontend`:

```sh
pnpm dev
```

The frontend uses Quasar and talks to the local backend.

## Database
Use the supplied PostgreSQL contract. Application work must not mutate schema outside explicitly approved database changes.

Backup helpers live under `db/backup`. Restore commands can replace database contents, so use the correct target database.

## Failure States
If the backend or database is unavailable, the frontend should show a clear non-blocking error state instead of an empty page.
