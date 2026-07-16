#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 0 ]]; then
  printf '%s\n' 'usage: init-spec-file.sh' >&2
  exit 2
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
change_slug="$("${script_dir}/extract-slug.sh")"

repo="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -z "${repo}" ]]; then
  printf '%s\n' 'cannot extract repo root' >&2
  exit 1
fi

spec_dir="${repo}/specs"
if [[ ! -d "${spec_dir}" ]]; then
  printf 'missing specs directory: %s\n' "${spec_dir}" >&2
  exit 1
fi

spec_path="${spec_dir}/${change_slug}.md"
temp_path="$(mktemp "${spec_dir}/.${change_slug}.md.XXXXXX")"
trap 'rm -f "${temp_path}"' EXIT

cat > "${temp_path}"
if [[ ! -s "${temp_path}" ]]; then
  printf '%s\n' 'spec input is empty' >&2
  exit 1
fi

chmod 0644 "${temp_path}"
if [[ -e "${spec_path}" || -L "${spec_path}" ]]; then
  backup_path="${spec_path}.$(date '+%Y-%m-%d_%H_%M').md"
  if [[ -e "${backup_path}" || -L "${backup_path}" ]]; then
    printf 'spec backup already exists: %s\n' "${backup_path}" >&2
    exit 1
  fi
  cp -p -- "${spec_path}" "${backup_path}"
fi

mv "${temp_path}" "${spec_path}"
trap - EXIT
