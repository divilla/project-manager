package agent

import (
	"os"
	"path/filepath"
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

func TestFormatCommandOutputPrettyPrintsJSONLines(t *testing.T) {
	output := "{\"type\":\"thread.started\",\"thread_id\":\"thread-1\"}\n{\"type\":\"turn.completed\",\"usage\":{\"input_tokens\":10,\"output_tokens\":2}}\n{\"type\":\"error\",\"message\":\"failed\"}\nplain text\n"

	formatted := FormatCommandOutput(output)

	assert.Contains(t, formatted, "\"type\": \"thread.started\"")
	assert.Contains(t, formatted, "\"thread_id\": \"thread-1\"")
	assert.Contains(t, formatted, "\"usage\": {\n    \"input_tokens\": 10,\n    \"output_tokens\": 2\n  }")
	assert.Contains(t, formatted, "\"type\": \"error\"")
	assert.Contains(t, formatted, "\"message\": \"failed\"")
	assert.Contains(t, formatted, "plain text")
}

func TestCodexPromptsUseExpectedSkills(t *testing.T) {
	assert.Equal(t, "Use $change-idea-tmp.", RewritePrompt)
	assert.Equal(t, "Use $change-spec-tmp.", InitPrompt)
}

func generatedChangeTypes() []string {
	return []string{"feature", "test"}
}

func TestParseGeneratedChange(t *testing.T) {
	body := "\n# Generated Change\n\nTypes: feature|test\n\n## Goal\nShip it.\n\n## QA Test Cases\n\n- First scenario.\n- Second scenario spans\n  more detail.\n1. Numbered scenario.\n\n## Review Focus\n\n- Parser."

	parsed, err := ParseGeneratedChange(body, generatedChangeTypes())
	require.NoError(t, err)

	assert.Equal(t, "Generated Change", parsed.Title)
	assert.Equal(t, []string{"feature", "test"}, parsed.ChangeTypes)
	assert.Equal(t, []string{"First scenario.", "Second scenario spans more detail.", "Numbered scenario."}, parsed.TestCases)
	assert.Equal(t, "\n# Generated Change\n\nTypes: feature|test\n\n## Goal\nShip it.\n\n## QA Test Cases\n\n- First scenario.\n- Second scenario spans\n  more detail.\n1. Numbered scenario.\n\n## Review Focus\n\n- Parser.", parsed.Body)
}

func TestParseGeneratedChangeSkipsNoneQATestCase(t *testing.T) {
	body := "# Generated Change\n\nTypes: feature\n\n## QA Test Cases\n\n- None.\n\n## Review Focus\n\n- Parser."

	parsed, err := ParseGeneratedChange(body, generatedChangeTypes())
	require.NoError(t, err)

	assert.Empty(t, parsed.TestCases)
}

func TestParseGeneratedChangeExtractsQATestCasesFromGeneratedBody(t *testing.T) {
	body := `# Generated Change

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

- Generated Change parser strictness for H1 title, ` + "`Types:`" + ` metadata, type slugs, full body preservation, and ` + "`## QA Test Cases`" + ` extraction.`

	parsed, err := ParseGeneratedChange(body, generatedChangeTypes())
	require.NoError(t, err)

	assert.Equal(t, []string{
		"Start `/new-change` with no valid current project and confirm the flow stops with a recoverable project-context error and no editor, Codex, or create request runs.",
		"Start `/new-change` when `/tmp/mch` is absent and confirm the directory and blank `initial-idea.md` are created.",
		"Simulate a generated test case create failure after Change create and confirm the error is displayed above the prompt and created details are not opened.",
		"Complete a successful create and confirm `mch` reloads and renders the created Change detail using backend data.",
	}, parsed.TestCases)
}

func TestExtractQATestCasesFromBodyFragment(t *testing.T) {
	body := `## Relevant Specs

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

- Generated Change parser strictness for H1 title, ` + "`Types:`" + ` metadata, type slugs, full body preservation, and ` + "`## QA Test Cases`" + ` extraction.`

	assert.Equal(t, []string{
		"Start `/new-change` with no valid current project and confirm the flow stops with a recoverable project-context error and no editor, Codex, or create request runs.",
		"Start `/new-change` when `/tmp/mch` is absent and confirm the directory and blank `initial-idea.md` are created.",
		"Complete a successful create and confirm `mch` reloads and renders the created Change detail using backend data.",
	}, ExtractQATestCases(body))
}

func TestParseGeneratedChangeValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "missing title", body: "Types: feature\n\n## Goal\nShip it."},
		{name: "missing types", body: "# Generated Change\n\n## Goal\nShip it."},
		{name: "blank types", body: "# Generated Change\n\nTypes: \n\n## Goal\nShip it."},
		{name: "malformed types", body: "# Generated Change\n\nTypes: feature | test\n\n## Goal\nShip it."},
		{name: "empty type", body: "# Generated Change\n\nTypes: feature|\n\n## Goal\nShip it."},
		{name: "comma separator", body: "# Generated Change\n\nTypes: feature,test\n\n## Goal\nShip it."},
		{name: "unsupported type", body: "# Generated Change\n\nTypes: spike\n\n## Goal\nShip it."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseGeneratedChange(tt.body, generatedChangeTypes())
			require.Error(t, err)
		})
	}
}
