package app

import (
	"context"
	"strings"
	"time"

	"mch/internal/agent"
	"mch/internal/changes"
	"mch/internal/dto"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) beginAgentNewChange() (tea.Model, tea.Cmd) {
	if _, err := currentProjectNumericID(m.currentProject.ID); err != nil {
		m.err = err.Error()
		m.status = "validation failed"
		return m, nil
	}
	m.agentFlow = agent.NewModelWithWorkspace(m.agentWorkspace)
	if err := m.agentFlow.Workspace.Ensure(); err != nil {
		m.err = err.Error()
		m.status = "agent failed"
		return m, nil
	}
	exists, err := m.agentFlow.Workspace.IdeaExists()
	if err != nil {
		m.err = err.Error()
		m.status = "agent failed"
		return m, nil
	}
	if exists {
		content, err := m.agentFlow.Workspace.ReadIdea()
		if err != nil {
			m.err = err.Error()
			m.status = "agent failed"
			return m, nil
		}
		if strings.TrimSpace(content) == "" {
			return m.openAgentIdeaEditor(false, "")
		}
		m.agentFlow.IdeaEntryContent = content
		m.openDropdown(CreateIdeaState, dropdownAgent, ChangesListState, CreateIdeaState, "Resume idea?", []dto.Option{
			{ID: "/resume", Label: "/resume"},
			{ID: "/clear", Label: "/clear"},
			{ID: "/cancel", Label: "/cancel"},
		}, false)
		return m, nil
	}
	return m.openAgentIdeaEditor(true, "")
}

func (m Model) openAgentIdeaEditor(replace bool, initialContent string) (tea.Model, tea.Cmd) {
	if replace {
		if err := m.agentFlow.Workspace.ResetIdea(); err != nil {
			m.err = err.Error()
			m.status = "agent failed"
			return m, nil
		}
	}
	m.agentFlow.IdeaEntryContent = initialContent
	m.agentFlow.Stage = agent.StageIdeaEntry
	m.state = CreateIdeaState
	m.status = "agent idea"
	return m.openPersistentEditor(CreateIdeaState, m.agentFlow.Workspace.IdeaPath())
}

func (m Model) handleAgentEditorFinished(msg editorFinishedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err.Error()
		m.status = "editor failed"
		return m, tea.ClearScreen
	}
	switch m.agentFlow.Stage {
	case agent.StageIdeaEntry:
		if err := m.agentFlow.Workspace.WriteIdea(msg.content); err != nil {
			m.err = err.Error()
			m.status = "agent failed"
			return m, tea.ClearScreen
		}
		if _, err := changes.ParseIdeaStructure(msg.content); err != nil {
			m.err = "error parsing title"
			m.agentFlow.IdeaEntryContent = msg.content
			m.openDropdown(CreateIdeaState, dropdownIdea, CreateIdeaState, CreateIdeaState, "error parsing title:", []dto.Option{
				{ID: "/edit", Label: "/edit"},
				{ID: "/cancel", Label: "/cancel"},
			}, false)
			return m, tea.ClearScreen
		}
		m.agentFlow.IdeaEntryContent = msg.content
		m.agentFlow.Stage = agent.StageCreateConfirmation
		m.openDropdown(CreateIdeaState, dropdownIdea, CreateIdeaState, CreateIdeaState, "Create Change?", []dto.Option{
			{ID: "/yes", Label: "/yes"},
			{ID: "/no", Label: "/no"},
		}, false)
		return m, tea.ClearScreen
	default:
		return m, tea.ClearScreen
	}
}

func (m Model) handleUpdateIdeaEditorFinished(msg editorFinishedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err.Error()
		m.status = "editor failed"
		return m, tea.ClearScreen
	}
	if err := m.agentFlow.Workspace.WriteIdea(msg.content); err != nil {
		m.err = err.Error()
		m.status = "agent failed"
		return m, tea.ClearScreen
	}
	m.agentFlow.IdeaEntryContent = msg.content
	if _, err := changes.ParseIdeaStructure(msg.content); err != nil {
		m.err = "error parsing title"
		m.openDropdown(UpdateIdeaState, dropdownIdea, UpdateIdeaState, UpdateIdeaState, "error parsing title:", []dto.Option{
			{ID: "/edit", Label: "/edit"},
			{ID: "/cancel", Label: "/cancel"},
		}, false)
		return m, tea.ClearScreen
	}
	id, err := changeNumericID(m.changeList.Detail)
	if err != nil {
		m.err = err.Error()
		m.status = "validation failed"
		return m, tea.ClearScreen
	}
	m.state = UpdateIdeaState
	m.status = "saving idea"
	return m, tea.Sequence(tea.ClearScreen, changeIdeaUpdateForRewriteCommand(m.client, id, msg.content))
}

func (m Model) discardAgentIdea(target State, status string, clear bool) (tea.Model, tea.Cmd) {
	if clear {
		if err := m.agentFlow.Workspace.RemoveIdea(); err != nil {
			m.err = err.Error()
			m.status = "agent failed"
			return m, nil
		}
	}
	m.agentFlow = agent.NewModelWithWorkspace(m.agentWorkspace)
	m.state = target
	m.status = status
	return m, nil
}

func (m Model) startAgentRewrite(sessionID string) (tea.Model, tea.Cmd) {
	m.state = RewriteIdeaState
	m.agentFlow.Stage = agent.StageAIRunning
	m.agentFlow.CommandOutput = ""
	m.agentElapsed = 0
	m.status = string(agent.StageAIRunning)
	updates := make(chan string)
	return m, tea.Batch(
		tea.ClearScreen,
		agentRewriteCommand(m.agentRunner, m.agentFlow.Workspace, sessionID, updates),
		agentCommandOutputCommand(updates),
		m.agentSpinner.Tick,
		agentElapsedTick(),
	)
}

func (m Model) handleAgentRewriteFinished(msg agentRewriteFinishedMsg) (tea.Model, tea.Cmd) {
	m.agentFlow.CommandOutput = agentRewriteDisplayOutput(msg.result)
	if msg.err != nil {
		m.agentFlow.Stage = agent.StageAIRunning
		m.err = msg.err.Error()
		m.status = "agent failed"
		return m, nil
	}
	if strings.TrimSpace(msg.result.SessionID) == "" || msg.result.Output != "Done." {
		m.agentFlow.Stage = agent.StageAIRunning
		m.err = agentRewriteContractError(msg.result)
		m.status = "agent failed"
		return m, nil
	}
	m.agentFlow.SessionID = msg.result.SessionID
	m.agentFlow.RepoRoot = msg.result.RepoRoot
	id, err := changeNumericID(m.changeList.Detail)
	if err != nil {
		m.err = err.Error()
		m.status = "validation failed"
		return m, nil
	}
	m.status = "saving idea"
	return m, changeIdeaAgentEditSaveCommand(m.client, id, m.agentFlow.Workspace)
}

//lint:ignore U1000 preserved for the future Write Spec with Agent flow.
func (m Model) openAgentSpecInit() (tea.Model, tea.Cmd) {
	m.state = RewriteIdeaState
	m.agentFlow.Stage = agent.StageAIRunning
	m.agentFlow.CommandOutput = ""
	m.agentElapsed = 0
	m.status = string(agent.StageAIRunning)
	return m, tea.Batch(
		tea.ClearScreen,
		agentSpecInitCommand(m.agentRunner, m.agentFlow.RepoRoot, m.agentFlow.SessionID),
		m.agentSpinner.Tick,
		agentElapsedTick(),
	)
}

func (m Model) handleAgentInitFinished(msg agentInitFinishedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.agentFlow.Stage = agent.StageAIRunning
		m.err = msg.err.Error()
		m.status = "agent failed"
		return m, nil
	}
	m.agentFlow.RepoRoot = msg.repoRoot
	projectID, err := currentProjectNumericID(m.currentProject.ID)
	if err != nil {
		m.err = err.Error()
		m.status = "validation failed"
		return m, nil
	}
	m.status = "saving change"
	return m, agentSpecCreateCommand(m.client, projectID, m.agentFlow.Workspace, optionIDs(m.optionCatalog.types))
}

func (m Model) handleAgentCommandOutput(msg agentCommandOutputMsg) (tea.Model, tea.Cmd) {
	if msg.output != "" {
		output := agent.FormatCommandOutput(msg.output)
		if strings.TrimSpace(output) != "" {
			m.agentFlow.CommandOutput = appendCommandOutput(m.agentFlow.CommandOutput, output)
		}
	}
	if msg.done {
		return m, nil
	}
	return m, agentCommandOutputCommand(msg.updates)
}

func agentRewriteCommand(runner agent.Runner, workspace agent.Workspace, sessionID string, updates chan<- string) tea.Cmd {
	return func() tea.Msg {
		defer close(updates)
		if runner == nil {
			runner = agent.NewProcessRunner()
		}
		ctx := context.Background()
		repoRoot, err := runner.ResolveRepoRoot(ctx)
		if err != nil {
			return agentRewriteFinishedMsg{err: err}
		}
		result, err := runner.Rewrite(ctx, repoRoot, sessionID, workspace, func(output string) {
			updates <- output
		})
		if err != nil {
			return agentRewriteFinishedMsg{result: result, err: err}
		}
		return agentRewriteFinishedMsg{result: result}
	}
}

//lint:ignore U1000 preserved for the future Write Spec with Agent flow.
func agentSpecInitCommand(runner agent.Runner, repoRoot string, sessionID string) tea.Cmd {
	return func() tea.Msg {
		if runner == nil {
			runner = agent.NewProcessRunner()
		}
		if strings.TrimSpace(repoRoot) == "" {
			var err error
			repoRoot, err = runner.ResolveRepoRoot(context.Background())
			if err != nil {
				return agentInitFinishedMsg{err: err}
			}
		}
		cmd := runner.InitCommand(repoRoot, sessionID)
		if err := cmd.Run(); err != nil {
			return agentInitFinishedMsg{repoRoot: repoRoot, err: err}
		}
		return agentInitFinishedMsg{repoRoot: repoRoot}
	}
}

func agentCommandOutputCommand(updates <-chan string) tea.Cmd {
	return func() tea.Msg {
		output, ok := <-updates
		if !ok {
			return agentCommandOutputMsg{updates: updates, done: true}
		}
		return agentCommandOutputMsg{output: output, updates: updates}
	}
}

func agentSpecCreateCommand(client appClient, projectID int, workspace agent.Workspace, validTypes []string) tea.Cmd {
	return func() tea.Msg {
		spec, err := workspace.ReadGenerated()
		if err != nil {
			return agentSpecCreatedMsg{err: err}
		}
		parsed, err := agent.ParseGeneratedChange(spec, validTypes)
		if err != nil {
			return agentSpecCreatedMsg{err: err}
		}
		created, err := client.CreateChange(dto.ChangeCreateInput{
			ProjectID: projectID,
			Title:     parsed.Title,
			Idea:      parsed.Spec,
		})
		if err != nil {
			return agentSpecCreatedMsg{err: err}
		}
		id, err := changeNumericID(created)
		if err != nil {
			return agentSpecCreatedMsg{err: err}
		}
		if _, err := client.UpdateChangeSpec(id, parsed.Spec, true); err != nil {
			return agentSpecCreatedMsg{err: err}
		}
		if _, err := client.UpdateChangeTypes(id, parsed.ChangeTypes); err != nil {
			return agentSpecCreatedMsg{err: err}
		}
		for _, scenario := range parsed.TestCases {
			if _, err := client.CreateTestCase(id, scenario); err != nil {
				return agentSpecCreatedMsg{err: err}
			}
		}
		change, err := client.GetChange(id)
		if err != nil {
			return agentSpecCreatedMsg{change: created, reloadErr: err}
		}
		return agentSpecCreatedMsg{change: change}
	}
}

func optionIDs(options []dto.Option) []string {
	ids := make([]string, 0, len(options))
	for _, option := range options {
		id := strings.TrimSpace(option.ID)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func agentRewriteDisplayOutput(result agent.RewriteResult) string {
	parts := make([]string, 0, 2)
	if strings.TrimSpace(result.CommandOutput) != "" {
		parts = append(parts, agent.FormatCommandOutput(result.CommandOutput))
	}
	if strings.TrimSpace(result.Output) != "" {
		parts = append(parts, "final output:\n"+agent.FormatCommandOutput(result.Output))
	}
	return strings.Join(parts, "\n")
}

func appendCommandOutput(existing string, next string) string {
	if strings.TrimSpace(existing) == "" {
		return next
	}
	if strings.TrimSpace(next) == "" {
		return existing
	}
	return strings.TrimRight(existing, "\n") + "\n" + strings.TrimLeft(next, "\n")
}

func agentRewriteContractError(result agent.RewriteResult) string {
	return agent.GenericError
}

func (m Model) agentRunningActive() bool {
	return m.agentFlow.Stage == agent.StageAIRunning && m.status != "agent failed"
}

func agentElapsedTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return agentElapsedMsg(t)
	})
}

func agentChangeCreateForRewriteCommand(client appClient, projectID int, workspace agent.Workspace) tea.Cmd {
	return func() tea.Msg {
		idea, err := workspace.ReadIdea()
		if err != nil {
			return changeCreatedForRewriteMsg{err: err}
		}
		parsed, err := changes.ParseIdeaStructure(idea)
		if err != nil {
			return changeCreatedForRewriteMsg{err: err}
		}
		created, err := client.CreateChange(dto.ChangeCreateInput{
			ProjectID: projectID,
			Title:     parsed.Title,
			Idea:      parsed.Idea,
		})
		if err != nil {
			return changeCreatedForRewriteMsg{err: err}
		}
		if _, err := changeNumericID(created); err != nil {
			return changeCreatedForRewriteMsg{err: err}
		}
		return changeCreatedForRewriteMsg{change: created}
	}
}

func changeIdeaUpdateForRewriteCommand(client appClient, id int, idea string) tea.Cmd {
	return func() tea.Msg {
		change, err := client.UpdateChangeIdea(id, idea, false)
		if err != nil {
			return changeIdeaUpdatedForRewriteMsg{err: err}
		}
		return changeIdeaUpdatedForRewriteMsg{change: change}
	}
}

func changeIdeaAgentEditSaveCommand(client appClient, id int, workspace agent.Workspace) tea.Cmd {
	return func() tea.Msg {
		idea, err := workspace.ReadIdea()
		if err != nil {
			return changeIdeaAgentEditSavedMsg{err: err}
		}
		updated, err := client.UpdateChangeIdea(id, idea, true)
		if err != nil {
			return changeIdeaAgentEditSavedMsg{err: err}
		}
		if err := workspace.RemoveIdea(); err != nil {
			return changeIdeaAgentEditSavedMsg{change: updated, err: err}
		}
		change, err := client.GetChange(id)
		if err != nil {
			return changeIdeaAgentEditSavedMsg{change: updated, reloadErr: err}
		}
		return changeIdeaAgentEditSavedMsg{change: change}
	}
}
