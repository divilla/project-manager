# CLI Agent Integration

The code for this entire section is in the new group `internal/agent`

## CLI Agent Workflow
- Selecting command /new-change triggers agent mode
- User provides initial idea in the form of md file
- Agent rewrites
- User reviews
- Repeat until user applies changes
- Agent through standard codex cli builds Change file
- Change file is parsed and supplied to API to create new change

## New Change Flow 
1. Check if file exists /tmp/mch if yes, delete it 
2. Check if folder /tmp/mch/ exists if not create it
3. If there is existing /tmp/mch/inital-idea.md prompt user with command menu 
    /resume - open existing file
    /new - create new file
4. If there is no file take /new flow
5. initial-idea.md is opened using CLI full screen Editor
6. If the file is empty after user exits - route to ChangesListScreen
7. Else execute:

    ```shell
        # codex exec initial exec - captures session_id  
        repo_root=$(git rev-parse --show-toplevel)
        json_log=/tmp/mch/codex-run.jsonl
        last_output=/tmp/mch/codex-output.txt
        
        codex exec --json \
            -C "$repo_root" \
            -o "$last_output" \
            'Use $change-idea-tmp.' \
            | tee "$json_log"
        
        session_id=$(
            jq -r 'select(.type=="thread.started") | .thread_id' "$json_log" | head -n 1
        )
        
        codex_output=$(cat "$last_output")
    ```

8. If output is not `Done.` - report error: `something went wrong - please try again`
9. initial-idea.md is opened using CLI full screen Editor
10. If user changed file:

    ```shell
        # subsequent calls are simpler
        codex exec resume \
            -o "$last_output" \
            "$session_id" \
            'Use $change-idea-tmp.'
        
        codex_output=$(cat "$last_output")
    ```
11. Go back to step 7. of this flow
12. If user exited editor without save - start codex (not codex exec) in the same way editor is started in CLI:

    ```shell
        # subsequent calls are simpler
        codex resume \
            "$session_id" \
            'Use $change-spec-tmp.'
    ```

13. After user exited the try to parse /tmp/mch/initial-change.md
14. If success /api/v1/change/create
15. If it fails write errors to user
