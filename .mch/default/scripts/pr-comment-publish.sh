#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 0 ]]; then
  printf '%s\n' 'usage: pr-comment-publish.sh' >&2
  exit 2
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
change_slug="$("${script_dir}/extract-slug.sh")"
: "${change_slug}"

#TBD
