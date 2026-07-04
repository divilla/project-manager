package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceEnsureReplacesRegularFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mch")
	require.NoError(t, os.WriteFile(path, []byte("not a directory"), 0o644))

	workspace := Workspace{Dir: path}
	require.NoError(t, workspace.Ensure())

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestWorkspaceIdeaFileLifecycle(t *testing.T) {
	workspace := Workspace{Dir: t.TempDir()}

	exists, err := workspace.IdeaExists()
	require.NoError(t, err)
	assert.False(t, exists)

	require.NoError(t, workspace.ResetIdea())
	exists, err = workspace.IdeaExists()
	require.NoError(t, err)
	assert.True(t, exists)

	content, err := workspace.ReadIdea()
	require.NoError(t, err)
	assert.Empty(t, content)
}

func TestExtractSessionID(t *testing.T) {
	tests := []struct {
		name string
		log  string
		want string
	}{
		{name: "thread started thread id", log: `{"type":"thread.started","thread_id":"thread-1"}`, want: "thread-1"},
		{name: "thread started wins over earlier event id", log: "{\"type\":\"item.started\",\"id\":\"item-1\"}\n{\"type\":\"thread.started\",\"thread_id\":\"thread-1\"}", want: "thread-1"},
		{name: "top level session_id", log: `{"session_id":"abc"}`, want: "abc"},
		{name: "nested session id", log: `{"session":{"id":"nested"}}`, want: "nested"},
		{name: "top level id", log: `{"id":"top"}`, want: "top"},
		{name: "first valid line", log: "not-json\n{\"session_id\":\"later\"}", want: "later"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ExtractSessionID(tt.log))
		})
	}
}

func TestFormatCommandOutputHumanizesJSONLines(t *testing.T) {
	output := strings.Join([]string{
		`{"type":"thread.started","thread_id":"thread-1"}`,
		`{"type":"turn.started"}`,
		`{"type":"item.completed","item":{"id":"item_0","type":"agent_message","text":"Reading idea."}}`,
		`{"type":"item.started","item":{"id":"item_1","type":"command_execution","command":"sed -n '1,40p' /tmp/mch/initial-idea.md","aggregated_output":"","exit_code":null,"status":"in_progress"}}`,
		`{"type":"item.completed","item":{"id":"item_1","type":"command_execution","command":"sed -n '1,40p' /tmp/mch/initial-idea.md","aggregated_output":"# Idea\n\nBody.","exit_code":0,"status":"completed"}}`,
		`{"type":"item.completed","item":{"id":"item_2","type":"file_change","changes":[{"path":"/tmp/mch/initial-idea.md","kind":"update"}],"status":"completed"}}`,
		`{"type":"turn.completed","usage":{"input_tokens":10,"cached_input_tokens":5,"output_tokens":2,"reasoning_output_tokens":1}}`,
		`{"type":"error","message":"failed"}`,
		`plain text`,
		``,
	}, "\n")

	formatted := FormatCommandOutput(output)

	assert.Contains(t, formatted, "thread started: thread-1")
	assert.Contains(t, formatted, "turn started")
	assert.Contains(t, formatted, "assistant: Reading idea.")
	assert.Contains(t, formatted, "running command: sed -n '1,40p' /tmp/mch/initial-idea.md")
	assert.Contains(t, formatted, "command completed (completed): exit 0")
	assert.Contains(t, formatted, "output:\n  # Idea\n  \n  Body.")
	assert.Contains(t, formatted, "file change completed:\n  update: /tmp/mch/initial-idea.md")
	assert.Contains(t, formatted, "turn completed (input=10, cached_input=5, output=2, reasoning_output=1)")
	assert.Contains(t, formatted, "error: failed")
	assert.Contains(t, formatted, "plain text")
	assert.NotContains(t, formatted, `"type":`)
}

func TestCodexPromptsUseExpectedSkills(t *testing.T) {
	assert.Equal(t, "Use $change-idea-tmp.", RewritePrompt)
	assert.Equal(t, "Use $change-spec-tmp.", InitPrompt)
}

func generatedChangeTypes() []string {
	return []string{"feature", "test"}
}

func TestParseGeneratedChange(t *testing.T) {
	spec := "\n# Generated Change\n\nTypes: feature|test\n\n## Goal\nShip it.\n\n## QA Test Cases\n\n- First scenario.\n- Second scenario spans\n  more detail.\n1. Numbered scenario.\n\n## Review Focus\n\n- Parser."

	parsed, err := ParseGeneratedChange(spec, generatedChangeTypes())
	require.NoError(t, err)

	assert.Equal(t, "Generated Change", parsed.Title)
	assert.Equal(t, []string{"feature", "test"}, parsed.ChangeTypes)
	assert.Equal(t, []string{"First scenario.", "Second scenario spans more detail.", "Numbered scenario."}, parsed.TestCases)
	assert.Equal(t, "\n# Generated Change\n\nTypes: feature|test\n\n## Goal\nShip it.\n\n## QA Test Cases\n\n- First scenario.\n- Second scenario spans\n  more detail.\n1. Numbered scenario.\n\n## Review Focus\n\n- Parser.", parsed.Spec)
}

func TestParseGeneratedChangeSkipsNoneQATestCase(t *testing.T) {
	spec := "# Generated Change\n\nTypes: feature\n\n## QA Test Cases\n\n- None.\n\n## Review Focus\n\n- Parser."

	parsed, err := ParseGeneratedChange(spec, generatedChangeTypes())
	require.NoError(t, err)

	assert.Empty(t, parsed.TestCases)
}

func TestParseGeneratedChangeExtractsQATestCasesFromGeneratedBody(t *testing.T) {
	spec := `# Generated Change

Types: feature|test

## Relevant Specs

- ` + "`/tmp/mch/initial-change.md`" + `
- ` + "`agent/prompts/change-file-structure.md`" + `
- ` + "`docs/architecture/mch.md`" + `

## Verification

- From ` + "`cli`" + `: ` + "`make lint`" + `
- From ` + "`cli`" + `: ` + "`go test ./...`" + `
- From ` + "`cli`" + `: ` + "`go build -o /tmp/mch ./cmd/mch`" + `

## QA Test Cases

- Start ` + "`/new-change`" + ` with no valid current project and confirm the flow stops with a recoverable project-context error and no editor, Codex, or create request runs.
- Start ` + "`/new-change`" + ` when ` + "`/tmp/mch`" + ` is absent and confirm the directory and blank ` + "`initial-idea.md`" + ` are created.
- Simulate a generated test case create failure after Change create and confirm the error is displayed above the prompt and created details are not opened.
- Complete a successful create and confirm ` + "`mch`" + ` reloads and renders the created Change detail using backend data.

## Review Focus

- Generated Change parser strictness for H1 title, ` + "`Types:`" + ` metadata, type slugs, full spec preservation, and ` + "`## QA Test Cases`" + ` extraction.`

	parsed, err := ParseGeneratedChange(spec, generatedChangeTypes())
	require.NoError(t, err)

	assert.Equal(t, []string{
		"Start `/new-change` with no valid current project and confirm the flow stops with a recoverable project-context error and no editor, Codex, or create request runs.",
		"Start `/new-change` when `/tmp/mch` is absent and confirm the directory and blank `initial-idea.md` are created.",
		"Simulate a generated test case create failure after Change create and confirm the error is displayed above the prompt and created details are not opened.",
		"Complete a successful create and confirm `mch` reloads and renders the created Change detail using backend data.",
	}, parsed.TestCases)
}

func TestExtractQATestCasesFromSpecFragment(t *testing.T) {
	spec := `## Relevant Specs

- ` + "`/tmp/mch/initial-change.md`" + `
- ` + "`agent/prompts/change-file-structure.md`" + `

## Verification

- From ` + "`cli`" + `: ` + "`make lint`" + `
- From ` + "`cli`" + `: ` + "`go test ./...`" + `

## QA Test Cases

- Start ` + "`/new-change`" + ` with no valid current project and confirm the flow stops with a recoverable project-context error and no editor, Codex, or create request runs.
- Start ` + "`/new-change`" + ` when ` + "`/tmp/mch`" + ` is absent and confirm the directory and blank ` + "`initial-idea.md`" + ` are created.
- Complete a successful create and confirm ` + "`mch`" + ` reloads and renders the created Change detail using backend data.

## Review Focus

- Generated Change parser strictness for H1 title, ` + "`Types:`" + ` metadata, type slugs, full spec preservation, and ` + "`## QA Test Cases`" + ` extraction.`

	assert.Equal(t, []string{
		"Start `/new-change` with no valid current project and confirm the flow stops with a recoverable project-context error and no editor, Codex, or create request runs.",
		"Start `/new-change` when `/tmp/mch` is absent and confirm the directory and blank `initial-idea.md` are created.",
		"Complete a successful create and confirm `mch` reloads and renders the created Change detail using backend data.",
	}, ExtractQATestCases(spec))
}

func TestParseGeneratedChangeValidation(t *testing.T) {
	tests := []struct {
		name string
		spec string
	}{
		{name: "missing title", spec: "Types: feature\n\n## Goal\nShip it."},
		{name: "missing types", spec: "# Generated Change\n\n## Goal\nShip it."},
		{name: "blank types", spec: "# Generated Change\n\nTypes: \n\n## Goal\nShip it."},
		{name: "malformed types", spec: "# Generated Change\n\nTypes: feature | test\n\n## Goal\nShip it."},
		{name: "empty type", spec: "# Generated Change\n\nTypes: feature|\n\n## Goal\nShip it."},
		{name: "comma separator", spec: "# Generated Change\n\nTypes: feature,test\n\n## Goal\nShip it."},
		{name: "unsupported type", spec: "# Generated Change\n\nTypes: spike\n\n## Goal\nShip it."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseGeneratedChange(tt.spec, generatedChangeTypes())
			require.Error(t, err)
		})
	}
}
