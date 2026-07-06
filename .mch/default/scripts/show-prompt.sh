#!/usr/bin/env bash
set -euo pipefail

flow_dir="${1:-}"
stage="${2:-}"

if [[ -z "${flow_dir}" || -z "${stage}" ]]; then
  printf '%s\n' 'usage: show-prompt.sh <flow-dir> <stage>' >&2
  exit 2
fi

case "${stage}" in
  idea) prompt="prompts/change-idea.md" ;;
  spec) prompt="prompts/change-spec.md" ;;
  ready) prompt="prompts/change-verify.md" ;;
  docs) prompt="prompts/change-docs.md" ;;
  code) prompt="prompts/change-code.md" ;;
  polish) prompt="prompts/polish.md" ;;
  pr) prompt="prompts/change-pr.md" ;;
  review) prompt="prompts/change-review.md" ;;
  fix) prompt="prompts/change-fix.md" ;;
  sync) prompt="prompts/change-docs-post-code.md" ;;
  merge) prompt="prompts/merge.md" ;;
  stage) prompt="prompts/stage.md" ;;
  master) prompt="prompts/master.md" ;;
  change-idea-tmp) prompt="prompts/change-idea-tmp.md" ;;
  change-spec-tmp) prompt="prompts/change-spec-tmp.md" ;;
  *)
    printf 'unknown stage: %s\n' "${stage}" >&2
    exit 2
    ;;
esac

cat "${flow_dir}${prompt}"
