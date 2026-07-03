package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"mch/internal/agent"
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
		m.openDropdown(ChangesListState, dropdownAgent, ChangesListState, ChangesListState, "Resume idea?", []dto.Option{
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
	m.state = ChangesListState
	m.status = "agent idea"
	return m.openPersistentEditor(ChangesListState, m.agentFlow.Workspace.IdeaPath())
}

func (m Model) handleAgentEditorFinished(msg editorFinishedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err.Error()
		m.status = "editor failed"
		return m, tea.ClearScreen
	}
	switch m.agentFlow.Stage {
	case agent.StageIdeaEntry:
		if msg.content == m.agentFlow.IdeaEntryContent && strings.TrimSpace(m.agentFlow.IdeaEntryContent) != "" {
			m.agentFlow = agent.NewModelWithWorkspace(m.agentWorkspace)
			m.status = string(ChangesListState)
			return m, tea.ClearScreen
		}
		if strings.TrimSpace(msg.content) == "" {
			m.agentFlow = agent.NewModelWithWorkspace(m.agentWorkspace)
			m.status = string(ChangesListState)
			return m, tea.ClearScreen
		}
		return m.startAgentRewrite("")
	case agent.StageReview:
		if msg.content != m.agentFlow.ReviewContent {
			return m.startAgentRewrite(m.agentFlow.SessionID)
		}
		return m.openAgentInit()
	default:
		return m, tea.ClearScreen
	}
}

func (m Model) startAgentRewrite(sessionID string) (tea.Model, tea.Cmd) {
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
	content, err := m.agentFlow.Workspace.ReadIdea()
	if err != nil {
		m.err = err.Error()
		m.status = "agent failed"
		return m, nil
	}
	m.agentFlow.ReviewContent = content
	m.agentFlow.Stage = agent.StageReview
	m.status = "agent review"
	return m.openPersistentEditor(ChangesListState, m.agentFlow.Workspace.IdeaPath())
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

func (m Model) openAgentInit() (tea.Model, tea.Cmd) {
	if m.agentRunner == nil {
		m.agentRunner = agent.NewProcessRunner()
	}
	m.agentFlow.Stage = agent.StageAIRunning
	m.agentFlow.CommandOutput = ""
	m.agentElapsed = 0
	m.status = string(agent.StageAIRunning)
	cmd := tea.ExecProcess(m.agentRunner.InitCommand(m.agentFlow.RepoRoot, m.agentFlow.SessionID), func(err error) tea.Msg {
		return agentInitFinishedMsg{err: err}
	})
	return m, tea.Batch(cmd, m.agentSpinner.Tick, agentElapsedTick())
}

func (m Model) handleAgentInitFinished(msg agentInitFinishedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err.Error()
		m.status = "agent failed"
		return m, tea.ClearScreen
	}
	projectID, err := currentProjectNumericID(m.currentProject.ID)
	if err != nil {
		m.err = err.Error()
		m.status = "validation failed"
		return m, tea.ClearScreen
	}
	m.status = "saving"
	return m, tea.Sequence(tea.ClearScreen, agentChangeCreateCommand(m.client, projectID, m.agentFlow.Workspace, optionIDs(m.optionCatalog.types)))
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

func agentCommandOutputCommand(updates <-chan string) tea.Cmd {
	return func() tea.Msg {
		output, ok := <-updates
		if !ok {
			return agentCommandOutputMsg{updates: updates, done: true}
		}
		return agentCommandOutputMsg{output: output, updates: updates}
	}
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

func agentChangeCreateCommand(client appClient, projectID int, workspace agent.Workspace, validTypes []string) tea.Cmd {
	return func() tea.Msg {
		body, err := workspace.ReadGenerated()
		if err != nil {
			return changeSavedMsg{source: ChangesListState, err: err}
		}
		parsed, err := agent.ParseGeneratedChange(body, validTypes)
		if err != nil {
			return changeSavedMsg{source: ChangesListState, err: err}
		}
		created, err := client.CreateChange(dto.ChangeCreateInput{
			ProjectID:   projectID,
			Title:       parsed.Title,
			Body:        parsed.Body,
			ChangeTypes: parsed.ChangeTypes,
		})
		if err != nil {
			return changeSavedMsg{source: ChangesListState, err: err}
		}
		id, err := changeNumericID(created)
		if err != nil {
			return changeSavedMsg{source: ChangesListState, err: err}
		}
		for i, testCase := range parsed.TestCases {
			if _, err := client.CreateTestCase(id, testCase); err != nil {
				return changeSavedMsg{source: ChangesListState, err: fmt.Errorf("create QA test case %d: %w", i+1, err)}
			}
		}
		change, err := client.GetChange(id)
		if err != nil {
			return changeSavedMsg{source: ChangesListState, err: err}
		}
		return changeSavedMsg{source: ChangesListState, change: change}
	}
}
