# CLI Flow 1

First CLI Change in regarding Flow will be Idea Stage implementation.

## Idea Flow

- user writes the idea for software change artifact
- agent reviews idea, asks for clarification - outputs questions and suggestions
- user provides answers or accepts/rejects suggestions or rewrites idea
- agent rewrites idea improving wording and readability and outputs idea
- user accepts / edits idea introducing changes / asks agent for idea changes
- agent applies changes and outputs rewritten idea

The state model:
Create  -> Review      -> Refine
Create  -> Review      -> Preview
Preview -> Rewrite     -> Preview
Preview -> Refine      -> Preview
Preview -> Edit        -> Review
Preview -> Save/Cancel -> ChangeDetails
Review  -> Refine      -> Preview
Review  -> Stop        -> Preview
Review  -> Cancel      -> ChangeDetails
Review  -> Error       -> Preview/ChangeDetails
Rewrite -> Cancel      -> ChangeDetails
Rewrite -> Preview
Rewrite -> Stop        -> Preview
Rewrite -> Error       -> Preview/ChangeDetails

Idea screens - DocumentCreateScreen is a starting screen:
- DocumentCreateScreen(<homeScreen>, <fromScreen>, <toScreen>):
  - Process: user writes an initial idea for a software change.
  - Layout: external `code` app is running in another window
  - State: `mch` is waiting for user to exit
  - Exits:
    DocumentRewriteScreen

- DocumentEditScreen(<homeScreen>, <fromScreen>, <toScreen>):
  - Process: user edits an idea for a software change to reduce or introduce further changes.
  - Layout: external `code` app is running in another window
  - State: `mch` is waiting for user to exit
  - Exits:
    <fromScreen>(<homeScreen>) - if user made edits - input.md not equal to output.md
    DocumentAcceptScreen(<homeScreen>, <fromScreen>, <toScreen>) - if toScreen is not null and user made no edits - input.md equals output.md
    <homeScreen> - if toScreen null and user made no edits - input.md equals output.md

- DocumentAcceptScreen(<homeScreen>, <fromScreen>, <toScreen>):
  - Process: user edits an idea for a software change to reduce or introduce further changes.
  - Layout: Viewport + Prompt
  - Viewport: displays `bat --diff` output
  - Prompt:
    - State: command menu open
    - Selected: /accept
    - Commands:
      /accept - go to <toScreen>(<homeScreen>)
      /edit   - go to DocumentEditScreen(<homeScreen>, <fromScreen>, <toScreen>)
      /reject
      /cancel

- DocumentRewriteScreen(<homeScreen>):
  - Process: agent rewrites the idea for clarity and readability without changing its intent.
  - Shell: `make idea-rewrite` executed in workdir `.mch/default`
  - Layout: while running `codex exec` in background there is screen animation + seconds counter in the screen View
  - Prompt: while running user can start typing `/`
    /stop   -> stop current agent operation and keep the current/latest draft
    /cancel -> leave the idea workflow without accepting
  - State: `mch` is waiting for command to finish
  - Expected output: `Done.`
  - Exits:
    DocumentAcceptScreen(<homeScreen>, DocumentRewriteScreen, DocumentReviewScreen) - on expected output.
    DocumentEditScreen(<homeScreen>, DocumentRewriteScreen) - if user interrupted with /stop
    DocumentPreviewScreen(selected_command=/rewrite) - if user interrupted with /stop
    ChangeDetailsScreen - user interrupted with /cancel
    DocumentErrorScreen(stage=idea, step=rewrite, last_command, stdout, stderr) - on exit_code not equal to zero error or unrecognized output

- DocumentReviewScreen:
  - Process: agent reviews the initial change idea, identifies gaps or ambiguities, and outputs clarifying questions plus suggestions to help refine it.
  - Shell: `make idea-review` executed in workdir `.mch/default`
  - Layout: while running `codex exec` in the background - main View displays screen animation + seconds counter
  - State: `mch` is waiting for command to finish
  - Prompt: while running user can start typing `/`
    /stop   -> stop current agent operation and keep the current/latest draft
    /cancel -> leave the idea workflow without accepting
  - Expected output: `Done.`
  - Exits: `No questions or suggestions.`
    DocumentRefineScreen - if command finished with questions or suggestions - session-id is passed forward
    DocumentPreviewScreen - if command finished with  or user interrupted with /stop
    ChangeDetailsScreen - if user interrupted with /cancel
    DocumentErrorScreen - on `codex exec` error

- InteractiveAgentScreen:
  - Process: user refines the idea by answering any questions, accepting or rejecting any suggestions.
  - Shell: `make agent-resume` executed in workdir `.mch/default`
  - Layout: `codex` CLI is running full screen,
  - State: `mch` is waiting for user to exit
  - Exits:
    DocumentPreviewScreen

- DocumentPreviewScreen
  - Process: The user either accepts the revised idea, edits it to introduce further changes, or asks the agent to apply additional changes.
  - Layout: screen View displays latest idea version using `bat` utility.
  - Prompt: user must pick direction: /save /edit /rewrite /refine /cancel.
  - Commands:
    /save     -> persist accepted idea, then ChangeDetailsScreen
    /edit     -> external editor
    /rewrite  -> agent rewrites for clarity
    /refine   -> conversational refinement
    /cancel   -> discard or return to ChangeDetailsScreen without saving
  - Exits:
    DocumentRewriteScreen on /rewrite
    DocumentRefineScreen on /refine
    DocumentEditScreen on /edit
    ChangeDetailsScreen on /save or /cancel


## Idea Flow

- user writes the idea for software change artifact
- agent reviews idea, asks for clarification - outputs questions and suggestions
- user provides answers or accepts/rejects suggestions or rewrites idea via codex cli with prompt
- user accepts / edits idea introducing changes / asks agent for idea changes
- agent applies changes and outputs rewritten idea

The state model:
Create  -> Review      -> Refine
Create  -> Review      -> Preview
Preview -> Refine      -> Preview
Preview -> Edit        -> Review
Preview -> Save/Cancel -> ChangeDetails
Review  -> Refine      -> Preview
Review  -> Stop        -> Preview
Review  -> Cancel      -> ChangeDetails
Review  -> Error       -> Preview/ChangeDetails

Idea screens - DocumentCreateScreen is a starting screen:
- DocumentCreateScreen: 
  - Process: user writes an initial idea for a software change.
  - Layout: external `code` app is running in another window  
  - State: `mch` is waiting for user to exit
  - Exits:
    DocumentReviewScreen

- DocumentEditScreen: 
  - Process: user edits an idea for a software change to reduce or introduce further changes.
  - Layout: external `code` app is running in another window
  - State: `mch` is waiting for user to exit
  - Exits:
    DocumentReviewScreen

- DocumentReviewScreen:
  - Process: agent reviews the initial change idea, identifies gaps or ambiguities, and outputs clarifying questions plus suggestions to help refine it.
  - Shell: `make idea-review` executed in workdir `.mch/default`
  - Layout: while running `codex exec` in the background - main View displays screen animation + seconds counter
  - State: `mch` is waiting for command to finish
  - Prompt: while running user can start typing `/`
    /stop   -> stop current agent operation and keep the current/latest draft
    /cancel -> leave the idea workflow without accepting
  - Exits:
    DocumentRefineScreen - if command finished with questions or suggestions - session-id is passed forward
    DocumentPreviewScreen - if command finished with `No questions or suggestions.` or user interrupted with /stop
    ChangeDetailsScreen - if user interrupted with /cancel
    DocumentErrorScreen - on `codex exec` error

- DocumentRefineScreen: 
  - Process: user refines the idea by answering any questions, accepting or rejecting any suggestions.
  - Shell: `make idea-refine` executed in workdir `.mch/default`
  - Layout: `codex` CLI is running full screen, 
  - State: `mch` is waiting for user to exit
  - Exits:
    DocumentPreviewScreen

- DocumentPreviewScreen
  - Process: The user either accepts the revised idea, edits it to introduce further changes, or asks the agent to apply additional changes.
  - Layout: screen View displays latest idea version using `bat` utility.
  - Prompt: user must pick direction: /save /edit /refine /cancel.
  - Commands:
    /save     -> persist accepted idea, then ChangeDetailsScreen
    /edit     -> external editor
    /refine   -> conversational refinement
    /cancel   -> discard or return to ChangeDetailsScreen without saving
  - Exits:
    DocumentRefineScreen on /refine
    DocumentEditScreen on /edit
    ChangeDetailsScreen on /save or /cancel

- DocumentErrorScreen
    - Layout: screen View displays verbose Error.
    - Prompt: user must pick direction: /continue /cancel.
    - Commands:
      /continue -> keep the current/latest draft and go to DocumentPreviewScreen
      /cancel   -> discard and return to ChangeDetailsScreen without saving
    - Exits:
      DocumentPreviewScreen on /continue
      ChangeDetailsScreen on /cancel

Every manual edit triggers agent review again, including edits launched from preview - that is intentional.

Exact prose checks like `Done.` and `No questions or suggestions.` will be replaced in the future Change.

Only `review` can result in error or produce unexpected output, because they are run with `codex exec`.

Latest draft can change during create, edit, refine, and rewrite. Accepted idea changes only on /save.

---

## CLI Temp Directory

To prevent future data races we'll generate one uuid in Go, export it to the Make/Codex process as an env var, and make the prompt explicitly tell Codex to read/write the UUID-scoped files.

SessionID in Go is kept alive as long as app lives. There must be one dummy `codex exec --json` call in Makefile that has sole purpose to get session_id that can be reused later.

This is new example layout:

.mch/tmp/<uuid>/input.md
.mch/tmp/<uuid>/output.md
.mch/tmp/<uuid>/session_id
.mch/tmp/<uuid>/events.jsonl
.mch/tmp/<uuid>/error.log
.mch/tmp/<uuid>/result.json

In Go:

```go
  cmd := exec.Command("make", "idea-review")
  cmd.Dir = filepath.Join(repoRoot, ".mch", "default")
  cmd.Env = append(os.Environ(),
    "MCH_TEMP_UUID="+flowID,
  )
```

In Makefile:

```shell
idea-review:
    @if [ -z "$$MCH_TEMP_UUID" ]; then printf '%s\n' 'missing MCH_TEMP_UUID' >&2; exit 1; fi
	@repo="$$(git rev-parse --show-toplevel)"; \
	temp_dir="$$repo/.mch/tmp/$$MCH_TEMP_UUID"; \
	mkdir -p "$$temp_dir"; \
	prompt="$$(sed 's|/tmp-dir/|'"$$temp_dir"'/|g' "$$repo/.mch/default/prompts/def-review.md")"; \
	codex exec -C "$$repo" --json "$$prompt" > "$$temp_dir/events.jsonl" 2> "$$temp_dir/error.log"; \
```

Runtime files:
- Temp dir: $$temp_dir
- Input idea: $$temp_dir/input.md
- Output idea: $$temp_dir/output.md

Only read and write files inside $$temp_dir for this flow.

---

## Default Editor

Default Editor used in past implementation is now replaced with VSCode. Run examples:

```shell
# create
code --new-window --wait . --goto .mch/vscode/output.md:1:1

# edit
code --new-window --wait . --diff .mch/vscode/input.md .mch/tmp/output.md 

# must be executed in .mch/default workdir
codex exec -C "$repo" --json $(cat prompts/idea.md) > events.jsonl 2> error.log
jq -r 'select(.type=="thread.started") | (.thread_id // .session_id // .session.id // .id // empty)' events.jsonl | head -n 1 | tr -d '\n' > session.txt
jq -r 'select(.type=="agent_message" or .type=="message") | (.message // .text // .output // .content // empty)' events.jsonl | tail -n 1 | tr -d '\n' > output.txt

```
