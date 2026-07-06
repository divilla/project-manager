package agent

import "fmt"

// Stage identifies where the user is in an agent-assisted workflow.
type Stage string

const (
	// StageIdle means no agent-assisted workflow is active.
	StageIdle Stage = ""
	// StageIdeaEntry means the user is editing the initial idea.
	StageIdeaEntry Stage = "idea entry"
	// StageCreateConfirmation means the user is confirming whether to create a Change.
	StageCreateConfirmation Stage = "create confirmation"
	// StageAIRunning means a Codex process is currently active.
	StageAIRunning Stage = "AI running"
)

const (
	// IdeaFileName is the user-authored idea markdown file.
	IdeaFileName = "initial-idea.md"
	// GeneratedFileName is the generated Change spec markdown file.
	GeneratedFileName = "initial-change.md"
	// CodexOutputName is the final text output file written by Codex exec.
	CodexOutputName = "codex-output.txt"
	// CodexRunLogName is the JSON event log file for the initial Codex exec.
	CodexRunLogName = "codex-run.jsonl"
	// RewritePromptSkill is the Codex skill for rewriting the idea draft.
	RewritePromptSkill = "$change-idea-tmp"
	// InitPrompt is the Codex prompt for generating the final Change spec.
	InitPrompt = "Use $change-spec-tmp."
	// GenericError is shown when Codex output does not satisfy the flow contract.
	GenericError = "something went wrong - please try again"
)

// RewritePrompt returns the Codex prompt for rewriting the configured idea file.
func RewritePrompt(workspace Workspace) string {
	return fmt.Sprintf("Use %s. The configured temp_dir is %q. Read and replace %q.", RewritePromptSkill, workspace.Dir, workspace.IdeaPath())
}

// Model stores agent-assisted workflow state.
type Model struct {
	Workspace        Workspace
	Stage            Stage
	SessionID        string
	RepoRoot         string
	IdeaEntryContent string
	CommandOutput    string
}

// NewModelWithWorkspace returns an idle agent model for a specific workspace.
func NewModelWithWorkspace(dir string) Model {
	return Model{Workspace: Workspace{Dir: dir}}
}

// Active reports whether an agent flow is in progress.
func (m Model) Active() bool {
	return m.Stage != StageIdle
}
