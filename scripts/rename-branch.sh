#!/usr/bin/env bash
set -euo pipefail

usage() {
	echo "usage: scripts/rename-branch.sh <new-branch-name>" >&2
}

fail() {
	echo "rename-branch: $*" >&2
	exit 1
}

if [[ $# -ne 1 ]]; then
	usage
	exit 2
fi

new_branch=$1

git rev-parse --git-dir >/dev/null 2>&1 || fail "not inside a git repository"
git check-ref-format --branch "$new_branch" >/dev/null || fail "invalid branch name: $new_branch"

old_branch=$(git branch --show-current)
[[ -n "$old_branch" ]] || fail "cannot rename from detached HEAD"
[[ "$old_branch" != "$new_branch" ]] || fail "current branch is already named $new_branch"

if git show-ref --verify --quiet "refs/heads/$new_branch"; then
	fail "local branch already exists: $new_branch"
fi

remote=$(git config --get "branch.$old_branch.remote" || true)
remote_branch=$(git config --get "branch.$old_branch.merge" || true)

if [[ -n "$remote_branch" ]]; then
	remote_branch=${remote_branch#refs/heads/}
else
	remote_branch=$old_branch
fi

if [[ -z "$remote" ]]; then
	if git remote get-url origin >/dev/null 2>&1; then
		remote=origin
	else
		fail "current branch has no upstream remote and origin does not exist"
	fi
fi

if ! git ls-remote --exit-code --heads "$remote" "$remote_branch" >/dev/null; then
	fail "remote branch does not exist: $remote/$remote_branch"
fi

if git ls-remote --exit-code --heads "$remote" "$new_branch" >/dev/null; then
	fail "remote branch already exists: $remote/$new_branch"
fi

echo "Renaming $old_branch to $new_branch on $remote."

git push "$remote" "HEAD:refs/heads/$new_branch"
git branch -m "$new_branch"
git branch --set-upstream-to="$remote/$new_branch" "$new_branch"
git push "$remote" ":refs/heads/$remote_branch"

echo "Renamed local branch $old_branch to $new_branch."
echo "Renamed remote branch $remote/$remote_branch to $remote/$new_branch."
