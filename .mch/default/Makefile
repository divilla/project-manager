FLOW_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
GIT_ROOT := $(shell git rev-parse --show-toplevel 2>/dev/null)
GIT_BRANCH := $(shell git branch --show-current 2>/dev/null)
CHANGE_NAME := $(patsubst change/%,%,$(GIT_BRANCH))
EVENTS_LOG := /tmp/mch/codex-events.jsonl
PROGRESS_LOG := /tmp/mch/codex-progress.log
LAST_OUTPUT := /tmp/mch/codex-output.txt

.PHONY: help list git-root root require-repo-root init idea-exec idea-exit idea spec-exit commit
.PHONY: %-entry %-exec %-exit %-before %-run %-leave %-prompt

help: list

list:
	@printf '%s\n' 'Default mch Flow targets:'
	@printf '%s\n' '  <stage>-entry   run the stage entry hook'
	@printf '%s\n' '  <stage>-exec    run the stage exec hook'
	@printf '%s\n' '  <stage>-exit    run the stage exit hook'
	@printf '%s\n' '  <stage>-before  alias for <stage>-entry'
	@printf '%s\n' '  <stage>-run     alias for <stage>-exec'
	@printf '%s\n' '  <stage>-leave   alias for <stage>-exit'
	@printf '%s\n' '  <stage>-prompt  print the stage prompt'
init:
	@if [ -z "$(GIT_ROOT)" ]; then printf '%s\n' 'cannot extract repo root' >&2; exit 1; fi
	@if [ -z "$(GIT_BRANCH)" ]; then printf '%s\n' 'cannot extract git branch' >&2; exit 1; fi
	@printf '%s\n' "$(GIT_BRANCH)" | grep -Eq '^change/[0-9]+-[a-z0-9_-]+$$' || { printf '%s\n' "not a change branch ^change/([0-9]+-[a-z0-9_-]+)$$: $(GIT_BRANCH)" >&2; exit 1; }
	@mkdir -p "$(dir $(EVENTS_LOG))"

root: init
	@printf '%s\n' "$(GIT_ROOT)"
	@printf '%s\n' "$(GIT_BRANCH)"
	@printf '%s\n' "$(CHANGE_NAME)"
	@printf '%s\n' "$(GIT_ROOT)/agent/ideas/$(CHANGE_NAME).md"

script-entry:
	@cp /tmp/mch/output.md /tmp/mch/input.md

idea-exec: init
	@cp "$(GIT_ROOT)/agent/ideas/$(CHANGE_NAME).md" /tmp/mch/input.md
	@codex exec --json -C "$(GIT_ROOT)" -o "$(LAST_OUTPUT)" 'Output exactly: `Done.`' | tee "$(EVENTS_LOG)"; \
	session_id=$$(jq -r 'select(.type=="thread.started") | .thread_id' "$(JSON_LOG)" | awk 'NF && $$0 != "null" { print; exit }'); \
	codex_output=$$(cat "$(LAST_OUTPUT)"); \
	if [ -z "$$session_id" ]; then printf '%s\n' 'missing codex session id' >&2; exit 1; fi; \
	if [ "$$codex_output" != "Done." ]; then printf '%s\n' "unexpected codex output: $$codex_output" >&2; exit 1; fi; \
	printf '%s\n' "codex session-id: $$session_id"; \
	cat prompts/idea.md | codex exec -C "$(GIT_ROOT)" resume "$$session_id"

idea-exit:
	@git add -A
	@git commit -m "Idea for $(CHANGE_NAME) by user"
	@git push origin "change/$(CHANGE_NAME)"

idea: init
	@cp "$(GIT_ROOT)/agent/ideas/115-cli-config-loading.md" /tmp/mch/input.md
	@codex exec --json -C "$(GIT_ROOT)" -o "$(LAST_OUTPUT)" "" | tee "$(EVENTS_LOG)"; \
	session_id=$$(jq -r 'select(.type=="thread.started") | .thread_id' "$(JSON_LOG)" | awk 'NF && $$0 != "null" { print; exit }'); \
	codex_output=$$(cat "$(LAST_OUTPUT)"); \
	if [ -z "$$session_id" ]; then printf '%s\n' 'missing codex session id' >&2; exit 1; fi; \
	cat prompts/idea.md | codex exec -C "$(GIT_ROOT)" -o "$(LAST_OUTPUT)" resume "$$session_id" -; \
	codex resume -C "$(GIT_ROOT)" "$$session_id"

idea-highlight: init
	@cp "$(GIT_ROOT)/agent/ideas/115-cli-config-loading.md" /tmp/mch/input.md
	@codex exec --json -C "$(GIT_ROOT)" -o "$(LAST_OUTPUT)" "" | tee "$(EVENTS_LOG)"; \
	session_id=$$(jq -r 'select(.type=="thread.started") | .thread_id' "$(JSON_LOG)" | awk 'NF && $$0 != "null" { print; exit }'); \
	codex_output=$$(cat "$(LAST_OUTPUT)"); \
	if [ -z "$$session_id" ]; then printf '%s\n' 'missing codex session id' >&2; exit 1; fi; \
	cat prompts/idea.md | codex exec -C "$(GIT_ROOT)" -o "$(LAST_OUTPUT)" resume "$$session_id" -; \
	bat -pp --theme 'Coldark-Dark' /tmp/mch/input.md

spec-entry: init
	@cp "$(GIT_ROOT)/agent/ideas/$(CHANGE_NAME).md" /tmp/mch/input.md

spec-exec: init
	@cat prompts/spec.md | codex exec --json -C "$(GIT_ROOT)" -o "$(LAST_OUTPUT)" - > "$(EVENTS_LOG)" 2> "$(PROGRESS_LOG)"; \
	session_id=$$(jq -r 'select(.type=="thread.started") | .thread_id' "$(EVENTS_LOG)" | awk 'NF && $$0 != "null" { print; exit }'); \
	codex_output=$$(cat "$(LAST_OUTPUT)"); \
	if [ -z "$$session_id" ]; then printf '%s\n' 'missing codex session id' >&2; exit 1; fi; \
	codex -C "$(GIT_ROOT)" resume "$$session_id"; \
	bat -pp --theme 'Coldark-Dark' "$(GIT_ROOT)/agent/ideas/$(CHANGE_NAME).md"

spec-old: init
	@cat prompts/spec.md | codex exec --json -C "$(GIT_ROOT)" -o "$(LAST_OUTPUT)" - > "$(EVENTS_LOG)" 2> "$(PROGRESS_LOG)"; \
	session_id=$$(jq -r 'select(.type=="thread.started") | .thread_id' "$(EVENTS_LOG)" | awk 'NF && $$0 != "null" { print; exit }'); \
	codex_output=$$(cat "$(LAST_OUTPUT)"); \
	if [ -z "$$session_id" ]; then printf '%s\n' 'missing codex session id' >&2; exit 1; fi; \
	if [ "$$codex_output" != "Done." ]; then printf '%s\n' "unexpected codex output: $$codex_output" >&2; exit 1; fi; \
	bat -pp --theme 'Coldark-Dark' "$(GIT_ROOT)/agent/ideas/$(CHANGE_NAME).md"

spec-exit: init
	@git add -A
	@git commit -m "Spec for $(CHANGE_NAME) by agent"
	@git push origin "change/$(CHANGE_NAME)"

verify-exit: init
	@git add -A
	@git commit -m "Spec fix for $(CHANGE_NAME) by agent" || true
	@git push origin "change/$(CHANGE_NAME)"

fix-exit: init
	@git add -A
	@git commit -m "Code fix for $(CHANGE_NAME) by agent" || true
	@git push origin "change/$(CHANGE_NAME)"

commit:
	@git add -A
	@git commit -m "Update for $(CHANGE_NAME) by user"
	@git push origin "change/$(CHANGE_NAME)"

%-entry:
	@$(FLOW_DIR)scripts/demo-entry.sh "$*"

%-exec:
	@$(FLOW_DIR)scripts/demo-exec.sh "$*"

%-exit:
	@$(FLOW_DIR)scripts/demo-exit.sh "$*"

%-before:
	@$(MAKE) --no-print-directory -f $(lastword $(MAKEFILE_LIST)) "$*-entry"

%-run:
	@$(MAKE) --no-print-directory -f $(lastword $(MAKEFILE_LIST)) "$*-exec"

%-leave:
	@$(MAKE) --no-print-directory -f $(lastword $(MAKEFILE_LIST)) "$*-exit"

%-prompt:
	@$(FLOW_DIR)scripts/show-prompt.sh "$(FLOW_DIR)" "$*"
