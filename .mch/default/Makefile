FLOW_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
GIT_ROOT := $(shell git rev-parse --show-toplevel 2>/dev/null)
GIT_BRANCH := $(shell git branch --show-current 2>/dev/null)
CHANGE_NAME := $(patsubst change/%,%,$(GIT_BRANCH))

.PHONY: help list git-root root require-repo-root init idea-init idea-exec idea-review idea-exit idea spec-exit commit
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
	@mkdir -p "$(GIT_ROOT)/.mch/tmp";

idea-init: init
	@if [ -z "$$MCH_TEMP_UUID" ]; then printf '%s\n' 'missing MCH_TEMP_UUID' >&2; exit 1; fi
	@mkdir -p "$(GIT_ROOT)/.mch/tmp/$$MCH_TEMP_UUID";

idea-review: idea-init
	@$(FLOW_DIR)scripts/codex-exec-restore-session.sh "$(FLOW_DIR)prompts/idea-review.md"

idea-refine: init
	@codex -C "$(GIT_ROOT)" resume "$$(cat ../tmp/sessions/$(CHANGE_NAME))"

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


extract-session: init
	@repo="$$(git rev-parse --show-toplevel)"; \
	stderr_file="$$(mktemp)"; \
	events="$$(codex exec -C "$$repo" --json "$$prompt" 2>"$$stderr_file")"; \
	status="$$?"; \
	stderr="$$(cat "$$stderr_file")"; \
	rm -f "$$stderr_file"; \
	session_id="$$(printf '%s\n' "$$events" | jq -rsr '\
	  [ \
		.[] \
		| select(.type=="thread.started") \
		| (.thread_id // .session_id // .session.id // .id // empty) \
	  ] \
	  | first // "" \
	')"; \
	final_output="$$(printf '%s\n' "$$events" | jq -rsr '\
	  [ \
		.[] \
		| if .type=="item.completed" and .item.type=="agent_message" then \
			.item.text \
		  elif .type=="agent_message" or .type=="message" then \
			(.message // .text // .output // .content // empty) \
		  else \
			empty \
		  end \
	  ] \
	  | last // "" \
	')"; \
	printf 'status=%s\n' "$$status"; \
	printf 'session_id=%s\n' "$$session_id"; \
	printf 'stderr=%s\n' "$$stderr"; \
	printf 'final_output=%s\n' "$$final_output"; \
	if [ "$$status" -ne 0 ]; then \
	  exit "$$status"; \
	fi; \
	if [ "$$final_output" != "No questions or suggestions." ]; then \
	  if [ -z "$$session_id" ]; then \
		printf '%s\n' 'missing session_id; cannot resume' >&2; \
		exit 1; \
	  fi; \
	  codex -C "$$repo" resume "$$session_id"; \
	fi

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
