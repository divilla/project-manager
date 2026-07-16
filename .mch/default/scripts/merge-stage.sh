#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 0 ]]; then
  printf '%s\n' 'usage: merge-stage.sh' >&2
  exit 2
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
change_slug="$("${script_dir}/extract-slug.sh")"

repo="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -z "${repo}" ]]; then
  printf '%s\n' 'cannot extract repo root' >&2
  exit 1
fi

change_branch="change/${change_slug}"

if [[ -n "$(git -C "${repo}" status --short)" ]]; then
  printf '%s\n' 'uncommitted changes' >&2
  exit 1
fi

remote_head_commit() {
  local branch_name="${1}"
  local output commit ref extra

  output="$(git -C "${repo}" ls-remote --heads origin "${branch_name}")"
  read -r commit ref extra <<< "${output}"
  if [[ ! "${commit}" =~ ^[0-9a-f]{40}$ || "${ref}" != "refs/heads/${branch_name}" || -n "${extra:-}" ]]; then
    printf 'cannot verify origin/%s\n' "${branch_name}" >&2
    return 1
  fi
  printf '%s\n' "${commit}"
}

ensure_pr_exists() {
  local state
  state="$(cd "${repo}" && gh pr view "${change_branch}" --json state --jq '.state')"
  if [[ "${state}" != "OPEN" ]]; then
    printf 'PR for %s is not open\n' "${change_branch}" >&2
    return 1
  fi
}

is_squashed_change_commit() {
  local commit="${1}"
  local parents stage_commit subject
  local -a fields

  parents="$(git -C "${repo}" rev-list --parents -n 1 "${commit}")"
  read -r -a fields <<< "${parents}"
  [[ ${#fields[@]} -eq 2 ]] || return 1

  stage_commit="$(git -C "${repo}" rev-parse origin/stage)"
  [[ "${fields[1]}" == "${stage_commit}" ]] || return 1

  subject="$(git -C "${repo}" log -1 --format=%s "${commit}")"
  [[ "${subject}" == "Implement change ${change_slug}" ]]
}

create_squash_commit() {
  local tree
  tree="$(git -C "${repo}" rev-parse 'HEAD^{tree}')"
  git -C "${repo}" commit-tree \
    "${tree}" \
    -p origin/stage \
    -m "Implement change ${change_slug}"
}

git -C "${repo}" fetch origin
original_commit="$(git -C "${repo}" rev-parse HEAD)"
origin_change_commit="$(remote_head_commit "${change_branch}")"
if [[ "${original_commit}" != "${origin_change_commit}" ]]; then
  printf 'local %s %s does not match origin/%s %s\n' \
    "${change_branch}" "${original_commit}" "${change_branch}" "${origin_change_commit}" >&2
  exit 1
fi

ensure_pr_exists
if ! git -C "${repo}" merge-base --is-ancestor origin/stage HEAD; then
  printf '%s\n' 'rebase needed: origin/stage is not an ancestor of HEAD' >&2
  exit 1
fi

squashed_commit="${original_commit}"
if ! is_squashed_change_commit "${original_commit}"; then
  squashed_commit="$(create_squash_commit)"
  git -C "${repo}" push \
    "--force-with-lease=refs/heads/${change_branch}:${origin_change_commit}" \
    origin \
    "${squashed_commit}:refs/heads/${change_branch}"
  git -C "${repo}" update-ref \
    "refs/heads/${change_branch}" \
    "${squashed_commit}" \
    "${original_commit}"
fi

ensure_pr_exists
git -C "${repo}" checkout stage
git -C "${repo}" pull --ff-only origin stage
git -C "${repo}" merge --ff-only "${change_branch}"
stage_commit="$(git -C "${repo}" rev-parse HEAD)"
if [[ "${stage_commit}" != "${squashed_commit}" ]]; then
  printf 'stage HEAD %s does not match squashed change commit %s\n' \
    "${stage_commit}" "${squashed_commit}" >&2
  exit 1
fi

git -C "${repo}" push -u origin stage
origin_stage_commit="$(remote_head_commit stage)"
if [[ "${origin_stage_commit}" != "${squashed_commit}" ]]; then
  printf 'origin/stage %s does not match squashed change commit %s\n' \
    "${origin_stage_commit}" "${squashed_commit}" >&2
  exit 1
fi
git -C "${repo}" push origin --delete "${change_branch}"
