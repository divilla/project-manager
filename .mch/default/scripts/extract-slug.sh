#!/usr/bin/env bash
set -euo pipefail

fail() {
    printf 'error: %s\n' "$1" >&2
    exit 1
}

if (($# != 0)); then
    fail "usage: scripts/extract-slug.sh"
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

if [[ ! "$branch_name" =~ ^change/([0-9]+-[a-z0-9_-]+)$ ]]; then
    fail "current branch does not match ^change/([0-9]+-[a-z0-9_-]+)$: $branch_name"
fi

change_name="${BASH_REMATCH[1]}"

printf '%s\n' "$change_name"
