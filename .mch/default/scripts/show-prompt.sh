#!/usr/bin/env bash
set -euo pipefail

flow_dir="${1:-}"
stage="${2:-}"

if [[ -z "${flow_dir}" || -z "${stage}" ]]; then
  printf '%s\n' 'usage: show-prompt.sh <flow-dir> <stage>' >&2
  exit 2
fi

case "${stage}" in
  idea|idea-write|idea-refine|change-idea-tmp) prompt="prompts/idea-write.md" ;;
  idea-review) prompt="prompts/idea-review.md" ;;
  spec|spec-write|change-spec-tmp) prompt="prompts/spec-write.md" ;;
  ready|spec-review) prompt="prompts/spec-review.md" ;;
  docs) prompt="prompts/docs-update.md" ;;
  code) prompt="prompts/code-implement.md" ;;
  polish) prompt="prompts/code-polish.md" ;;
  pr|pr-write) prompt="prompts/pr-write.md" ;;
  review) prompt="prompts/code-review.md" ;;
  fix) prompt="prompts/code-fix.md" ;;
  sync) prompt="prompts/code-docs-spec-update.md" ;;
  merge) prompt="prompts/merge.md" ;;
  stage) prompt="prompts/stage.md" ;;
  master) prompt="prompts/master.md" ;;
  *)
    printf 'unknown stage: %s\n' "${stage}" >&2
    exit 2
    ;;
esac

cat "${flow_dir}${prompt}"
