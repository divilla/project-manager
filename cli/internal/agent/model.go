package agent

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
	// DefaultWorkspaceDir is the temporary workspace for agent planning files.
	DefaultWorkspaceDir = "/tmp/mch"
	// IdeaFileName is the user-authored idea markdown file.
	IdeaFileName = "initial-idea.md"
	// GeneratedFileName is the generated Change markdown file.
	GeneratedFileName = "initial-change.md"
	// CodexOutputName is the final text output file written by Codex exec.
	CodexOutputName = "codex-output.txt"
	// CodexRunLogName is the JSON event log file for the initial Codex exec.
	CodexRunLogName = "codex-run.jsonl"
	// RewritePrompt is the Codex prompt for rewriting the idea draft.
	RewritePrompt = "Use $change-idea-tmp."
	// InitPrompt is the Codex prompt for generating the final Change file.
	InitPrompt = "Use $change-spec-tmp."
	// GenericError is shown when Codex output does not satisfy the flow contract.
	GenericError = "something went wrong - please try again"
)

// Model stores agent-assisted workflow state.
type Model struct {
	Workspace        Workspace
	Stage            Stage
	SessionID        string
	RepoRoot         string
	IdeaEntryContent string
	CommandOutput    string
}

// NewModel returns an idle agent model for the default workspace.
func NewModel() Model {
	return Model{Workspace: Workspace{Dir: DefaultWorkspaceDir}}
}

// NewModelWithWorkspace returns an idle agent model for a specific workspace.
func NewModelWithWorkspace(dir string) Model {
	return Model{Workspace: Workspace{Dir: dir}}
}

// Active reports whether an agent flow is in progress.
func (m Model) Active() bool {
	return m.Stage != StageIdle
}
