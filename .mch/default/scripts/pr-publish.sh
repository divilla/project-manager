#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 0 ]]; then
  printf '%s\n' 'usage: pr-publish.sh' >&2
  exit 2
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
change_slug="$("${script_dir}/extract-slug.sh")"

repo="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -z "${repo}" ]]; then
  printf '%s\n' 'cannot extract repo root' >&2
  exit 1
fi

pr_author="divilla"
authenticated_login="$(gh api user --jq '.login')"
if [[ "${authenticated_login}" != "${pr_author}" ]]; then
  printf 'gh is authenticated as %s, expected %s\n' \
    "${authenticated_login}" "${pr_author}" >&2
  exit 1
fi

git -C "${repo}" fetch origin
if ! git -C "${repo}" merge-base --is-ancestor origin/stage HEAD; then
  printf '%s\n' 'rebase needed: origin/stage is not an ancestor of HEAD' >&2
  exit 1
fi

pr_file="${repo}/agent/prs/${change_slug}.md"
if [[ ! -f "${pr_file}" ]]; then
  printf 'cannot read PR body file: %s\n' "${pr_file}" >&2
  exit 1
fi

first_line="$(sed -n '1p' "${pr_file}")"
first_line="${first_line%$'\r'}"
if [[ ! "${first_line}" =~ ^#[[:space:]]+(.+)$ ]]; then
  printf "first line of %s must be '# <Title>'\n" "${pr_file}" >&2
  exit 1
fi
pr_title="$(printf '%s' "${BASH_REMATCH[1]}" | sed 's/[[:space:]]*$//')"
if [[ -z "${pr_title}" ]]; then
  printf "first line of %s must be '# <Title>'\n" "${pr_file}" >&2
  exit 1
fi

git -C "${repo}" add -A
git -C "${repo}" commit -m "Write PR for ${change_slug} by agent"
git -C "${repo}" push origin "change/${change_slug}"

pr_url_dir="${repo}/agent/prurls"
pr_url_file="${pr_url_dir}/${change_slug}"
mkdir -p "${pr_url_dir}"
pr_url="$(
  gh pr create \
    --base stage \
    --head "${pr_author}:change/${change_slug}" \
    --title "${pr_title}" \
    --body-file "${pr_file}"
)"
printf '%s\n' "${pr_url}" > "${pr_url_file}"

git -C "${repo}" add -A
git -C "${repo}" commit -m "Write PR URL for ${change_slug} by agent"
git -C "${repo}" push origin "change/${change_slug}"
