package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"mch/internal/dto"
	"mch/internal/flow"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) beginIdeaCreate() (tea.Model, tea.Cmd) {
	if err := m.cleanupIdeaCreateAttempt(); err != nil {
		return m.enterFlowError(err, ChangesListState)
	}
	if _, err := currentProjectNumericID(m.currentProject.ID); err != nil {
		return m.enterFlowError(err, ChangesListState)
	}
	attemptUUID, path, err := flow.CreateIdeaWorkspace(m.appConfig.RepositoryRoot)
	if err != nil {
		return m.enterFlowError(err, ChangesListState)
	}
	m.ideaCreateAttempt = ideaCreateAttempt{uuid: attemptUUID, path: path}
	m.ideaCreateBytes = nil
	m.ideaCreateTitle = ""
	m.state = CreateIdeaState
	m.status = "idea entry"
	return m.openPersistentEditor(CreateIdeaState, path)
}

func (m Model) handleIdeaCreateEditorFinished(message editorFinishedMsg) (tea.Model, tea.Cmd) {
	if message.err != nil {
		updated, _ := m.enterIdeaCreateError(fmt.Errorf("editor failed: %w", message.err))
		return updated, tea.ClearScreen
	}
	if len(message.content) == 0 {
		return m.cancelIdeaCreate()
	}
	if m.ideaCreateAttempt.uuid == "" || m.ideaCreateAttempt.path == "" {
		return m.enterIdeaCreateError(fmt.Errorf("IdeaCreate attempt workspace is required"))
	}
	m.stopIdeaValidation()
	validationContext, cancel := context.WithCancel(context.Background())
	m.ideaCreateAttempt.cancel = cancel
	m.state = IdeaProcessingState
	m.status = "processing idea"
	return m, tea.Batch(
		tea.ClearScreen,
		m.spinner.Tick,
		validateIdeaCreateCommand(validationContext, m.ideaCreateAttempt.uuid, m.client, m.currentProject.ID, m.ideaCreateAttempt.path, []byte(message.content)),
	)
}

func validateIdeaCreateCommand(ctx context.Context, attemptUUID string, client appClient, projectID, path string, content []byte) tea.Cmd {
	return func() tea.Msg {
		canonical, err := flow.CanonicalizeDocument(content, appDocumentOptions{client: client, projectID: projectID})
		if err != nil {
			return ideaCreateValidatedMsg{attemptUUID: attemptUUID, err: err}
		}
		if err := ctx.Err(); err != nil {
			return ideaCreateValidatedMsg{attemptUUID: attemptUUID, err: err}
		}
		if err := os.WriteFile(path, canonical.Bytes, 0o644); err != nil {
			return ideaCreateValidatedMsg{attemptUUID: attemptUUID, err: fmt.Errorf("write canonical new-idea.md: %w", err)}
		}
		return ideaCreateValidatedMsg{attemptUUID: attemptUUID, content: canonical.Bytes, title: canonical.Title}
	}
}

func (m *Model) stopIdeaValidation() {
	if m.ideaCreateAttempt.cancel != nil {
		m.ideaCreateAttempt.cancel()
		m.ideaCreateAttempt.cancel = nil
	}
}

func (m *Model) cleanupIdeaCreateAttempt() error {
	m.stopIdeaValidation()
	attemptUUID := m.ideaCreateAttempt.uuid
	if attemptUUID == "" {
		return nil
	}
	if err := flow.CleanupIdeaWorkspace(m.appConfig.RepositoryRoot, attemptUUID); err != nil {
		return err
	}
	m.ideaCreateAttempt = ideaCreateAttempt{}
	return nil
}

func (m Model) cancelIdeaCreate() (tea.Model, tea.Cmd) {
	if err := m.cleanupIdeaCreateAttempt(); err != nil {
		return m.enterFlowError(err, ChangesListState)
	}
	m.ideaCreateBytes = nil
	m.ideaCreateTitle = ""
	m.dropdown = dropdownModel{}
	m.detailEditField = ""
	m.activeTestCase = dto.TestCase{}
	m = m.setPromptValue("")
	return m.arrive(ChangesListState, "cancel")
}

func (m Model) enterIdeaCreateError(err error) (tea.Model, tea.Cmd) {
	if cleanupErr := m.cleanupIdeaCreateAttempt(); cleanupErr != nil {
		err = errors.Join(err, cleanupErr)
	}
	return m.enterFlowError(err, ChangesListState)
}

func (m Model) handleIdeaCreateValidated(message ideaCreateValidatedMsg) (tea.Model, tea.Cmd) {
	if message.err != nil {
		var validation flow.ValidationError
		if errors.As(message.err, &validation) {
			m.err = ""
			m.openDropdown(CreateIdeaState, dropdownIdea, ChangesListState, CreateIdeaState, validation.Error(), []dto.Option{
				{ID: "/fix", Label: "/fix"},
				{ID: "/cancel", Label: "/cancel"},
			}, false)
			return m, nil
		}
		return m.enterIdeaCreateError(message.err)
	}
	m.ideaCreateBytes = append([]byte(nil), message.content...)
	m.ideaCreateTitle = message.title
	m.openDropdown(CreateIdeaState, dropdownIdea, ChangesListState, CreateIdeaState, "Create Change?", []dto.Option{
		{ID: "/yes", Label: "/yes"},
		{ID: "/no", Label: "/no"},
	}, false)
	return m, nil
}

func createChangeForIdeaFlowCommand(client appClient, attemptUUID string, projectID int, title string, idea []byte) tea.Cmd {
	return func() tea.Msg {
		change, err := client.CreateChange(dto.ChangeCreateInput{ProjectID: projectID, Title: title, Idea: string(idea)})
		if err != nil {
			return changeCreatedForIdeaFlowMsg{attemptUUID: attemptUUID, err: err}
		}
		if _, err := changeNumericID(change); err != nil {
			return changeCreatedForIdeaFlowMsg{attemptUUID: attemptUUID, err: err}
		}
		if _, err := (flow.Workspace{Root: ".", ChangeRef: change.RefUUID, Artifact: flow.ArtifactIdea}).Directory(); err != nil {
			return changeCreatedForIdeaFlowMsg{attemptUUID: attemptUUID, err: fmt.Errorf("created Change ref_uuid must be a valid UUID")}
		}
		return changeCreatedForIdeaFlowMsg{attemptUUID: attemptUUID, change: change}
	}
}

func (m Model) startIdeaFlow(change dto.Change, origin State, step flow.StepID, editorCaller flow.ScreenID) (tea.Model, tea.Cmd) {
	changeID, err := changeNumericID(change)
	if err != nil {
		return m.enterFlowError(err, origin)
	}
	model := flow.Compose(flow.Composition{
		Definition: flow.IdeaDefinition(),
		Context: flow.Context{
			Root:         m.appConfig.RepositoryRoot,
			FlowDir:      m.appConfig.FlowDir,
			ChangeID:     changeID,
			ChangeRef:    change.RefUUID,
			Origin:       flow.ScreenID(origin),
			Step:         step,
			EditorCaller: editorCaller,
		},
		TerminalScreens: []flow.ScreenID{flow.MainTerminal, flow.ChangesListTerminal, flow.ChangeDetailsTerminal},
		Store:           flow.NewChangeArtifactStore(m.client),
		Options:         appDocumentOptions{client: m.client, projectID: m.currentProject.ID},
		Operations:      m.flowOperations,
	})
	m.ideaFlow = model
	m.ideaFlowActive = true
	m.err = ""
	m.changeList = m.changeList.WithDetail(change)
	m.state = State(step)
	m.status = "idea flow"
	return m.syncIdeaFlow(model.Init())
}

func (m Model) enterFlowError(err error, origin State) (tea.Model, tea.Cmd) {
	m.ideaFlowActive = false
	m.dropdown = dropdownModel{}
	m.flowErrorOrigin = origin
	m.state = FlowErrorState
	m.status = string(FlowErrorState)
	m.err = err.Error()
	return m, nil
}

func (m Model) syncIdeaFlow(command tea.Cmd) (tea.Model, tea.Cmd) {
	if !m.ideaFlowActive {
		return m, command
	}
	if m.ideaFlow.Done() {
		terminal := State(m.ideaFlow.TerminalScreen())
		m.ideaFlowActive = false
		m.state = terminal
		m.status = "idea flow complete"
		m = m.setPromptValue("")
		switch terminal {
		case ChangeDetailsState:
			return m.arrive(ChangeDetailsState, "loading change")
		case ChangesListState:
			return m.arrive(ChangesListState, "loading changes")
		case MainState:
			return m, command
		default:
			m.err = fmt.Sprintf("unsupported Flow terminal Screen %q", terminal)
			return m, nil
		}
	}
	if m.ideaFlow.Screen() != "" {
		m.state = State(m.ideaFlow.Screen())
		m.status = string(m.ideaFlow.Screen())
	}
	return m, command
}

func (m Model) updateIdeaFlow(message tea.Msg) (tea.Model, tea.Cmd) {
	updated, command := m.ideaFlow.Update(message)
	m.ideaFlow = updated.(flow.Model)
	return m.syncIdeaFlow(command)
}

func (m Model) beginIdeaEdit() (tea.Model, tea.Cmd) {
	return m.startIdeaFlow(m.changeList.Detail, ChangeDetailsState, flow.IdeaEdit, flow.ChangeDetailsTerminal)
}

func (m Model) executeIdeaFlowCommand(command string) (tea.Model, tea.Cmd) {
	allowed := false
	for _, candidate := range m.ideaFlow.Commands() {
		if string(candidate) == command {
			allowed = true
			break
		}
	}
	if !allowed {
		m.err = "unknown command: " + command
		return m, nil
	}
	return m.updateIdeaFlow(flow.CommandMsg{ID: flow.CommandID(command)})
}

type appDocumentOptions struct {
	client    appClient
	projectID string
}

func (o appDocumentOptions) ChangeTypes() ([]flow.TypeOption, error) {
	options, err := o.client.ListTypes()
	if err != nil {
		return nil, err
	}
	result := make([]flow.TypeOption, 0, len(options))
	for _, option := range options {
		slug := strings.TrimSpace(option.ID)
		if slug == "" {
			slug = strings.TrimSpace(option.Label)
		}
		result = append(result, flow.TypeOption{Slug: slug})
	}
	return result, nil
}

func (o appDocumentOptions) Epics() ([]flow.EpicOption, error) {
	options, err := o.client.ListEpics(o.projectID)
	if err != nil {
		return nil, err
	}
	result := make([]flow.EpicOption, 0, len(options))
	for _, option := range options {
		id, err := strconv.Atoi(strings.TrimSpace(option.ID))
		if err != nil {
			return nil, fmt.Errorf("epic option ID %q is not numeric", option.ID)
		}
		result = append(result, flow.EpicOption{ID: id, Title: option.Label})
	}
	return result, nil
}
