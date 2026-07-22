# Database Recovery

## Purpose

This runbook covers the safety checks that cannot be encoded into the repository's current backup
and restore helpers. Restore operations can replace database contents and must never be selected by
guessing from an archive name.

## Target Selection

Before a restore:

1. Identify the exact database that may be replaced and confirm that it is disposable or an
   explicitly approved recovery target.
2. Read the chosen helper under `db/backup/` and verify its hardcoded database name. `restore.sh`
   targets `changes`; `restore-test.sh` targets `changes_test`.
3. Confirm the archive path and provenance. Repository backup archives are snapshots of particular
   local states, not automatically safe or current restore points.
4. Back up any target state that must remain recoverable before starting the restore.
5. Obtain explicit authorization for the exact restore operation. Agents must not run a restore,
   migration, seed, or other database mutation from a general verification or implementation task.

Never restore a repository demo or test archive into a live or production database.

## Recovery

Run only the helper whose inspected target matches the approved database, passing the selected gzip
archive as its sole argument. Treat a partial or failed restore as an unknown database state: stop
application writes, preserve the command output, and do not retry against a different target.

After a successful restore, verify database availability through the application health diagnostic
and inspect the intended application behavior. If either check fails, keep the target isolated and
recover from the pre-restore backup or escalate with the captured error.
