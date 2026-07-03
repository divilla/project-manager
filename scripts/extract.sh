#!/usr/bin/env bash
set -euo pipefail

fail() {
    printf 'error: %s\n' "$1" >&2
    exit 1
}

if (($# != 0)); then
    fail "usage: scripts/extract.sh"
fi

if ! repo_root=$(git rev-parse --show-toplevel 2>/dev/null); then
    fail "cannot extract repo root"
fi

if [[ -z "$repo_root" ]]; then
    fail "cannot extract repo root"
fi

if ! branch_name=$(git branch --show-current 2>/dev/null); then
    fail "cannot read current branch"
fi

if [[ -z "$branch_name" ]]; then
    fail "cannot read current branch"
fi

if [[ ! "$branch_name" =~ ^changes/([0-9]+-[a-z0-9-]+)$ ]]; then
    fail "current branch does not match ^changes/([0-9]+-[a-z0-9-]+)$: $branch_name"
fi

change_name="${BASH_REMATCH[1]}"
change_path="$repo_root/agent/changes/$change_name.md"
rel_path="agent/changes/$change_name.md"

if [[ ! -f "$change_path" ]]; then
    fail "change file does not exist: $change_path"
fi

printf '%s\n' "$rel_path"
