#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 0 ]]; then
  printf '%s\n' 'usage: init-idea-file.sh' >&2
  exit 2
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
change_slug="$("${script_dir}/extract-slug.sh")"

repo="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -z "${repo}" ]]; then
  printf '%s\n' 'cannot extract repo root' >&2
  exit 1
fi

idea_dir="${repo}/agent/ideas"
if [[ ! -d "${idea_dir}" ]]; then
  printf 'missing ideas directory: %s\n' "${idea_dir}" >&2
  exit 1
fi

idea_path="${idea_dir}/${change_slug}.md"
temp_path="$(mktemp "${idea_dir}/.${change_slug}.md.XXXXXX")"
trap 'rm -f "${temp_path}"' EXIT

cat > "${temp_path}"
if [[ ! -s "${temp_path}" ]]; then
  printf '%s\n' 'idea input is empty' >&2
  exit 1
fi

chmod 0644 "${temp_path}"
if [[ -e "${idea_path}" || -L "${idea_path}" ]]; then
  backup_path="${idea_path}.$(date '+%Y-%m-%d_%H_%M').md"
  if [[ -e "${backup_path}" || -L "${backup_path}" ]]; then
    printf 'idea backup already exists: %s\n' "${backup_path}" >&2
    exit 1
  fi
  cp -p -- "${idea_path}" "${backup_path}"
fi

mv "${temp_path}" "${idea_path}"
trap - EXIT
