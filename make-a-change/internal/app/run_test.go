package app

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"mch/internal/api"
	"mch/internal/styles"
)

type fakeClient struct {
	projects   []api.Option
	epics      []api.Option
	phases     []api.Option
	types      []api.Option
	err        error
	projectID  string
	listCalls  int
	phaseCalls int
	typeCalls  int
	epicCalls  int
}

func (f *fakeClient) ListProjects() ([]api.Option, error) {
	f.listCalls++
	return f.projects, f.err
}

func (f *fakeClient) ListEpics(projectID string) ([]api.Option, error) {
	f.epicCalls++
	f.projectID = projectID
	if f.err != nil {
		return nil, f.err
	}
	if projectID == "" {
		return nil, errors.New("current project is required")
	}
	return f.epics, nil
}

func (f *fakeClient) ListPhases() ([]api.Option, error) {
	f.phaseCalls++
	return f.phases, f.err
}

func (f *fakeClient) ListTypes() ([]api.Option, error) {
	f.typeCalls++
	return f.types, f.err
}

func TestRunVersionPrintsVersion(t *testing.T) {
	var out bytes.Buffer

	require.NoError(t, Run([]string{"--version"}, &out))

	got := out.String()
	assert.Contains(t, got, "mch")
	assert.Contains(t, got, Version)
}

func TestNewModelStartupState(t *testing.T) {
	m := NewModel()

	assert.Equal(t, MainState, m.state)
	assert.True(t, m.input.Focused())
	assert.Contains(t, m.View(), "MainScreen - Title: Main")
}

func TestStartupTriggersProjectSelectionWhenProjectIDIsUnset(t *testing.T) {
	client := &fakeClient{
		projects: []api.Option{{ID: "7", Label: "Project Seven"}},
	}
	m := NewModelWithClient(client)

	cmd := m.Init()
	require.NotNil(t, cmd)
	got := applyMsg(m, cmd())
	assert.Equal(t, SelectProjectDropDown, got.state)
	assert.Equal(t, selectorProjects, got.dropdown.source)

	load := selectorCommand(client, got.dropdown.source, got.currentProject.ID)
	got = applyMsg(got, load())

	assert.Equal(t, SelectProjectDropDown, got.state)
	assert.Equal(t, []api.Option{{ID: "7", Label: "Project Seven"}}, got.dropdown.options)
}

func TestStartupSkipsProjectSelectionWhenProjectIDIsSaved(t *testing.T) {
	m := newModelWithConfig(&fakeClient{}, appConfig{BackendURL: defaultBackendURL, ProjectID: 7}, "")

	assert.Nil(t, m.Init())
	assert.Equal(t, MainState, m.state)
	assert.Equal(t, "7", m.currentProject.ID)
}

func TestStartupProjectSelectionShowsErrorWhenNoProjectsExist(t *testing.T) {
	client := &fakeClient{}
	m := NewModelWithClient(client)

	cmd := m.Init()
	require.NotNil(t, cmd)
	got := applyMsg(m, cmd())
	load := selectorCommand(client, got.dropdown.source, got.currentProject.ID)
	got = applyMsg(got, load())

	assert.Equal(t, MainState, got.state)
	assert.Empty(t, got.dropdown.kind)
	assert.Equal(t, noProjectsToSelectError, got.err)
}

func TestInputBandUsesCliProtoFullWidthBackground(t *testing.T) {
	m := NewModel()
	m.width = 40
	assert.Equal(t, 0, m.input.Width)

	band := m.inputBand(40)
	lines := strings.Split(band, "\n")
	require.Len(t, lines, 3)
	assert.Contains(t, band, "Type / for commands")
	for i, line := range lines {
		visible := stripANSI(line)
		assert.Falsef(t, strings.TrimSpace(visible) == "" && len(visible) < 40, "blank input band line %d too short: %q", i, visible)
	}
	assert.True(t, strings.HasPrefix(stripANSI(lines[1]), "> Type / for commands"))

	m.input.SetValue("typed text")
	typedLine := stripANSI(strings.Split(m.inputBand(40), "\n")[1])
	assert.True(t, strings.HasPrefix(typedLine, "> typed text"))
	assert.Equal(t, "15", fmt.Sprint(m.input.TextStyle.GetForeground()))
	assert.Equal(t, "0", fmt.Sprint(m.input.PlaceholderStyle.GetForeground()))

	wideBand := m.inputBand(180)
	wideLines := strings.Split(wideBand, "\n")
	require.Len(t, wideLines, 3)
	assert.Len(t, stripANSI(wideLines[0]), 180)
	assert.Len(t, stripANSI(wideLines[1]), 180)
	assert.Len(t, stripANSI(wideLines[2]), 180)
}

func TestMainCommandsTransition(t *testing.T) {
	tests := []struct {
		command string
		want    State
		quit    bool
	}{
		{command: "/new-change", want: ChangeCreateState},
		{command: "/changes", want: ChangesListState},
		{command: "/epics", want: EpicsListState},
		{command: "/projects", want: ProjectsListState},
		{command: "/help", want: MainHelpState},
		{command: "/quit", want: DoneState, quit: true},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got, cmd := sendCommand(NewModel(), tt.command)
			assert.Equal(t, tt.want, got.state)
			if tt.quit && cmd == nil {
				require.NotNil(t, cmd)
			}
		})
	}
}

func TestMainNewChangeShortcutIsFirstCommand(t *testing.T) {
	commands := commandsByState[MainState]
	require.NotEmpty(t, commands)
	assert.Equal(t, "/new-change", commands[0])
	assert.Contains(t, commandsByState[ChangesListState], "/new-change")
}

func TestQuitOutsideMainIsRecoverableError(t *testing.T) {
	m := NewModel()
	m.state = ChangesListState

	got, cmd := sendCommand(m, "/quit")

	assert.Equal(t, ChangesListState, got.state)
	assert.NotEmpty(t, got.err)
	assert.Nil(t, cmd)
}

func TestUnknownCommandLeavesStateUnchanged(t *testing.T) {
	m := NewModel()
	m.state = ChangeDetailsState

	got, _ := sendCommand(m, "/bogus")

	assert.Equal(t, ChangeDetailsState, got.state)
	assert.NotEmpty(t, got.err)
}

func TestListSelectionDropdownTransitionsToDetails(t *testing.T) {
	tests := []struct {
		start State
		want  State
	}{
		{start: ChangesListState, want: ChangeDetailsState},
		{start: ChangeDetailsState, want: RequirementDetailsState},
		{start: EpicsListState, want: EpicDetailsState},
		{start: ProjectsListState, want: ProjectDetailsState},
	}

	for _, tt := range tests {
		t.Run(string(tt.start), func(t *testing.T) {
			m := NewModel()
			m.state = tt.start

			dropdown, _ := sendKey(m, tea.KeyEnter)
			require.Equal(t, ListSelectionDropDownState, dropdown.state)

			got, _ := sendKey(dropdown, tea.KeyEnter)
			assert.Equal(t, tt.want, got.state)
		})
	}
}

func TestCreateUpdateSaveCancelTransitions(t *testing.T) {
	tests := []struct {
		start   State
		command string
		want    State
	}{
		{start: ChangesListState, command: "/new-change", want: ChangeCreateState},
		{start: ChangeCreateState, command: "/save", want: ChangeDetailsState},
		{start: ChangeCreateState, command: "/cancel", want: ChangesListState},
		{start: ChangeDetailsState, command: "/edit", want: ChangeUpdateState},
		{start: ChangeUpdateState, command: "/save", want: ChangeDetailsState},
		{start: ChangeUpdateState, command: "/cancel", want: ChangeDetailsState},
		{start: ChangeDetailsState, command: "/new-requirement", want: RequirementCreateState},
		{start: RequirementCreateState, command: "/save", want: RequirementDetailsState},
		{start: RequirementUpdateState, command: "/cancel", want: RequirementDetailsState},
		{start: RequirementDetailsState, command: "/edit", want: RequirementUpdateState},
		{start: EpicsListState, command: "/new-epic", want: EpicCreateState},
		{start: EpicCreateState, command: "/save", want: EpicDetailsState},
		{start: EpicDetailsState, command: "/edit", want: EpicUpdateState},
		{start: ProjectsListState, command: "/new-project", want: ProjectCreateState},
		{start: ProjectCreateState, command: "/save", want: ProjectDetailsState},
		{start: ProjectDetailsState, command: "/edit", want: ProjectUpdateState},
	}

	for _, tt := range tests {
		t.Run(string(tt.start)+tt.command, func(t *testing.T) {
			m := NewModel()
			m.state = tt.start

			got, _ := sendCommand(m, tt.command)

			assert.Equal(t, tt.want, got.state)
		})
	}
}

func TestSlashCommandTransitionsByState(t *testing.T) {
	tests := []struct {
		start        State
		command      string
		want         State
		wantPrevious State
	}{
		{start: ChangesListState, command: "/help", want: ChangesHelpState},
		{start: ChangesListState, command: "/clear-filters", want: ChangesListState},
		{start: ChangesListState, command: "/return", want: MainState},
		{start: ChangeDetailsState, command: "/return", want: ChangesListState},
		{start: RequirementDetailsState, command: "/new-requirement", want: RequirementCreateState},
		{start: RequirementDetailsState, command: "/save", want: RequirementDetailsState},
		{start: RequirementDetailsState, command: "/cancel", want: RequirementDetailsState},
		{start: RequirementDetailsState, command: "/return", want: ChangeDetailsState},
		{start: EpicsListState, command: "/help", want: EpicsHelpState},
		{start: EpicsListState, command: "/find", want: FindInputState, wantPrevious: EpicsListState},
		{start: EpicsListState, command: "/return", want: MainState},
		{start: EpicDetailsState, command: "/help", want: EpicsHelpState},
		{start: EpicDetailsState, command: "/find", want: FindInputState, wantPrevious: EpicDetailsState},
		{start: EpicDetailsState, command: "/return", want: EpicsListState},
		{start: EpicCreateState, command: "/cancel", want: EpicsListState},
		{start: EpicUpdateState, command: "/save", want: EpicDetailsState},
		{start: EpicUpdateState, command: "/cancel", want: EpicDetailsState},
		{start: ProjectsListState, command: "/help", want: ProjectsHelpState},
		{start: ProjectsListState, command: "/find", want: FindInputState, wantPrevious: ProjectsListState},
		{start: ProjectsListState, command: "/return", want: MainState},
		{start: ProjectDetailsState, command: "/help", want: ProjectsHelpState},
		{start: ProjectDetailsState, command: "/find", want: FindInputState, wantPrevious: ProjectDetailsState},
		{start: ProjectDetailsState, command: "/return", want: ProjectsListState},
		{start: ProjectCreateState, command: "/cancel", want: ProjectsListState},
		{start: ProjectUpdateState, command: "/save", want: ProjectDetailsState},
		{start: ProjectUpdateState, command: "/cancel", want: ProjectDetailsState},
		{start: MainHelpState, command: "/return", want: MainState},
		{start: ChangesHelpState, command: "/return", want: ChangesListState},
		{start: EpicsHelpState, command: "/return", want: EpicsListState},
		{start: ProjectsHelpState, command: "/return", want: ProjectsListState},
	}

	for _, tt := range tests {
		t.Run(string(tt.start)+tt.command, func(t *testing.T) {
			m := NewModel()
			m.state = tt.start

			got, _ := sendCommand(m, tt.command)

			assert.Equal(t, tt.want, got.state)
			if tt.wantPrevious != "" {
				assert.Equal(t, tt.wantPrevious, got.previousState)
			}
		})
	}
}

func TestDeleteCommandsOpenExpectedConfirmations(t *testing.T) {
	tests := []struct {
		start State
		want  State
	}{
		{start: ChangeDetailsState, want: ChangeDeleteConfirmation},
		{start: RequirementDetailsState, want: RequirementDeleteConfirmation},
		{start: EpicDetailsState, want: EpicDeleteConfirmation},
		{start: ProjectDetailsState, want: ProjectDeleteConfirmation},
	}

	for _, tt := range tests {
		t.Run(string(tt.start), func(t *testing.T) {
			m := NewModel()
			m.state = tt.start

			got, _ := sendCommand(m, "/delete")

			assert.Equal(t, tt.want, got.state)
		})
	}
}

func TestChangeDetailsCommandsAreExact(t *testing.T) {
	assert.Equal(t, []string{
		"/new-requirement",
		"/phase",
		"/epic",
		"/types",
		"/edit",
		"/delete",
		"/return",
	}, commandsByState[ChangeDetailsState])
}

func TestChangesListCommandsAreExact(t *testing.T) {
	assert.Equal(t, []string{
		"/new-change",
		"/phase-filter",
		"/epic-filter",
		"/type-filter",
		"/find-filter",
		"/clear-filters",
		"/help",
		"/return",
	}, commandsByState[ChangesListState])
}

func TestReturnAndEscapeTransitions(t *testing.T) {
	returnTests := []struct {
		start State
		want  State
	}{
		{start: ChangesListState, want: MainState},
		{start: ChangeDetailsState, want: ChangesListState},
		{start: RequirementDetailsState, want: ChangeDetailsState},
		{start: EpicsListState, want: MainState},
		{start: EpicDetailsState, want: EpicsListState},
		{start: ProjectsListState, want: MainState},
		{start: ProjectDetailsState, want: ProjectsListState},
		{start: MainHelpState, want: MainState},
		{start: ChangesHelpState, want: ChangesListState},
		{start: EpicsHelpState, want: EpicsListState},
		{start: ProjectsHelpState, want: ProjectsListState},
	}

	for _, tt := range returnTests {
		t.Run("return "+string(tt.start), func(t *testing.T) {
			m := NewModel()
			m.state = tt.start

			got, _ := sendKey(m, tea.KeyEsc)

			assert.Equal(t, tt.want, got.state)
		})
	}

	m := NewModel()
	got, cmd := sendKey(m, tea.KeyEsc)
	assert.Equal(t, DoneState, got.state)
	assert.True(t, got.quitting)
	require.NotNil(t, cmd)

	m = NewModel()
	m.state = ChangeCreateState
	got, _ = sendKey(m, tea.KeyEsc)
	assert.Equal(t, ChangesListState, got.state)
}

func TestSelectorDropdownsLoadAndReturn(t *testing.T) {
	client := &fakeClient{
		projects: []api.Option{{ID: "7", Label: "Project Seven"}},
		phases:   []api.Option{{ID: "backlog", Label: "backlog"}},
		types:    []api.Option{{ID: "feature", Label: "feature"}},
		epics:    []api.Option{{ID: "3", Label: "Epic Three"}},
	}

	m := NewModelWithClient(client)
	got, cmd := sendCommand(m, "/select-project")
	require.Equal(t, SelectProjectDropDown, got.state)
	require.NotNil(t, cmd)
	got = applyMsg(got, cmd())
	got, _ = sendKey(got, tea.KeyEnter)
	assert.Equal(t, MainState, got.state)
	assert.Equal(t, "7", got.currentProject.ID)

	got.state = ChangeDetailsState
	got, cmd = sendCommand(got, "/phase")
	got = applyMsg(got, cmd())
	got, _ = sendKey(got, tea.KeyEnter)
	assert.Equal(t, ChangeDetailsState, got.state)
	assert.Equal(t, 1, client.phaseCalls)

	got, cmd = sendCommand(got, "/types")
	got = applyMsg(got, cmd())
	got, _ = sendKey(got, tea.KeyEnter)
	assert.Equal(t, ChangeDetailsState, got.state)
	assert.Equal(t, 1, client.typeCalls)

	got, cmd = sendCommand(got, "/epic")
	got = applyMsg(got, cmd())
	got, _ = sendKey(got, tea.KeyEnter)
	assert.Equal(t, ChangeDetailsState, got.state)
	assert.Equal(t, 1, client.epicCalls)
	assert.Equal(t, "7", client.projectID)
}

func TestSelectProjectPersistsProjectIDToConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".config", "config.yaml")
	require.NoError(t, saveAppConfig(path, appConfig{BackendURL: defaultBackendURL}))
	client := &fakeClient{
		projects: []api.Option{{ID: "7", Label: "Project Seven"}},
	}
	m := newModelWithConfig(client, appConfig{BackendURL: defaultBackendURL}, path)

	got, cmd := sendCommand(m, "/select-project")
	require.NotNil(t, cmd)
	got = applyMsg(got, cmd())
	got, _ = sendKey(got, tea.KeyEnter)

	assert.Equal(t, MainState, got.state)
	assert.Equal(t, "7", got.currentProject.ID)
	loaded, err := loadAppConfig(path)
	require.NoError(t, err)
	assert.Equal(t, 7, loaded.ProjectID)
}

func TestSelectorFailureAndEscapePreservePreviousState(t *testing.T) {
	client := &fakeClient{err: errors.New("backend unavailable")}
	m := NewModelWithClient(client)
	m.state = ChangeDetailsState

	got, cmd := sendCommand(m, "/phase")
	got = applyMsg(got, cmd())
	assert.Equal(t, SelectPhaseDropDown, got.state)
	assert.NotEmpty(t, got.err)

	got, _ = sendKey(got, tea.KeyEsc)
	assert.Equal(t, ChangeDetailsState, got.state)
}

func TestFilterSelectorsReturnToChangesList(t *testing.T) {
	client := &fakeClient{
		phases: []api.Option{{ID: "done", Label: "done"}},
		epics:  []api.Option{{ID: "epic-1", Label: "Epic One"}},
		types:  []api.Option{{ID: "test", Label: "test"}},
	}
	m := NewModelWithClient(client)
	m.state = ChangesListState
	m.currentProject = api.Option{ID: "project-1", Label: "Project One"}

	got, cmd := sendCommand(m, "/phase-filter")
	require.NotNil(t, cmd)
	assert.Equal(t, ChangesListState, got.state)
	assert.Contains(t, got.View(), "ChangesListScreen - Title: Changes List")
	got = applyMsg(got, cmd())
	phaseDropdown := strings.Split(got.dropdownView(80), "\n")
	require.GreaterOrEqual(t, len(phaseDropdown), 3)
	assert.True(t, strings.HasPrefix(stripANSI(phaseDropdown[1]), "    -done"))
	assert.True(t, strings.HasPrefix(stripANSI(phaseDropdown[len(phaseDropdown)-1]), "    /clear"))
	got, _ = sendKey(got, tea.KeyEnter)
	assert.Equal(t, ChangesListState, got.state)
	assert.Equal(t, "done", got.changesFilters.phase.ID)
	assert.Equal(t, "done", got.changesFilters.phase.Label)

	got, cmd = sendCommand(got, "/epic-filter")
	require.NotNil(t, cmd)
	assert.Equal(t, ChangesListState, got.state)
	assert.Contains(t, got.View(), "ChangesListScreen - Title: Changes List")
	got = applyMsg(got, cmd())
	got, _ = sendKey(got, tea.KeyEnter)
	assert.Equal(t, ChangesListState, got.state)
	assert.Equal(t, "epic-1", got.changesFilters.epic.ID)

	got, cmd = sendCommand(got, "/type-filter")
	require.NotNil(t, cmd)
	assert.Equal(t, ChangesListState, got.state)
	assert.Contains(t, got.View(), "ChangesListScreen - Title: Changes List")
	got = applyMsg(got, cmd())
	got, _ = sendKey(got, tea.KeyEnter)
	assert.Equal(t, ChangesListState, got.state)
	assert.Equal(t, "test", got.changesFilters.typ.ID)

	got, cmd = sendCommand(got, "/phase-filter")
	require.NotNil(t, cmd)
	got = applyMsg(got, cmd())
	got.dropdown.filter = "/clear"
	got, _ = sendKey(got, tea.KeyEnter)
	assert.Equal(t, ChangesListState, got.state)
	assert.Empty(t, got.changesFilters.phase.ID)
	assert.Equal(t, "epic-1", got.changesFilters.epic.ID)
	assert.Equal(t, "test", got.changesFilters.typ.ID)

	got, _ = sendCommand(got, "/find-filter")
	assert.Equal(t, ChangesListState, got.state)

	got, _ = sendCommand(got, "/clear-filters")
	assert.Empty(t, got.changesFilters.epic.ID)
	assert.Empty(t, got.changesFilters.typ.ID)
}

func TestFindInputHighlightsAndEmptyFindErrors(t *testing.T) {
	m := NewModel()
	m.state = MainHelpState

	got, _ := sendCommand(m, "/find")
	assert.Equal(t, FindInputState, got.state)
	assert.Equal(t, MainHelpState, got.previousState)

	got.input.SetValue("phase")
	got, _ = sendKey(got, tea.KeyEnter)
	assert.Equal(t, MainHelpState, got.state)
	assert.Equal(t, "phase", got.helpQuery)

	got, _ = sendCommand(got, "/find")
	got, _ = sendKey(got, tea.KeyEnter)
	assert.Equal(t, MainHelpState, got.state)
	assert.NotEmpty(t, got.err)
}

func TestConfirmationRequiresYesOrCancel(t *testing.T) {
	m := NewModel()
	m.state = ChangeDetailsState

	got, _ := sendCommand(m, "/delete")
	assert.Equal(t, ChangeDeleteConfirmation, got.state)

	got.dropdown.filter = "/no"
	got, _ = sendKey(got, tea.KeyEnter)
	assert.Equal(t, ChangeDeleteConfirmation, got.state)
	assert.NotEmpty(t, got.err)

	got.dropdown.filter = "/cancel"
	got, _ = sendKey(got, tea.KeyEnter)
	assert.Equal(t, ChangeDetailsState, got.state)

	got, _ = sendCommand(got, "/delete")
	got.dropdown.filter = "/yes"
	got, _ = sendKey(got, tea.KeyEnter)
	assert.Equal(t, ChangesListState, got.state)
}

func TestCommandDropdownFiltersAndExecutesSelection(t *testing.T) {
	m := NewModel()

	got, _ := sendRune(m, '/')
	require.Equal(t, MainState, got.state)
	require.Equal(t, dropdownCommand, got.dropdown.kind)
	assert.Contains(t, got.View(), "MainScreen - Title: Main")
	assert.NotContains(t, got.View(), "CommandDropDownScreen - Title: Commands")
	dropdown := got.dropdownView(80)
	lines := strings.Split(dropdown, "\n")
	require.GreaterOrEqual(t, len(lines), 3)
	assert.True(t, strings.HasPrefix(stripANSI(lines[1]), "    /new-change"))
	assert.Equal(t, "15", fmt.Sprint(styles.Default.Selection.GetForeground()))
	assert.Empty(t, strings.TrimSpace(stripANSI(lines[len(lines)-1])))
	got, _ = sendRune(got, 'e')
	got, _ = sendRune(got, 'p')
	got, _ = sendKey(got, tea.KeyEnter)

	assert.Equal(t, EpicsListState, got.state)
}

func TestCommandDropdownPreservesUnderlyingScreenForEveryCommandState(t *testing.T) {
	for state := range commandsByState {
		t.Run(string(state), func(t *testing.T) {
			m := NewModel()
			m.state = state

			got, _ := sendRune(m, '/')

			assert.Equal(t, state, got.state)
			assert.Equal(t, dropdownCommand, got.dropdown.kind)
			assert.Equal(t, CommandDropDownState, got.dropdown.state)
			assert.Contains(t, got.View(), screenTitle(state))
			assert.NotContains(t, got.View(), screenTitle(CommandDropDownState))
		})
	}
}

func TestCreateStatesUseContextSpecificNewCommandVocabulary(t *testing.T) {
	createCommands := map[State]string{
		ChangesListState:        "/new-change",
		ChangeDetailsState:      "/new-requirement",
		RequirementDetailsState: "/new-requirement",
		EpicsListState:          "/new-epic",
		ProjectsListState:       "/new-project",
	}
	for state, want := range createCommands {
		t.Run(string(state), func(t *testing.T) {
			commands := commandsByState[state]
			assert.Contains(t, commands, want)
			assert.NotContains(t, commands, "/new")
			assert.NotContains(t, commands, "/create")
		})
	}
}

func TestUpdateStatesUseEditCommandVocabulary(t *testing.T) {
	updateSources := []State{
		ChangeDetailsState,
		RequirementDetailsState,
		EpicDetailsState,
		ProjectDetailsState,
	}
	for _, state := range updateSources {
		t.Run(string(state), func(t *testing.T) {
			commands := commandsByState[state]
			assert.Contains(t, commands, "/edit")
			assert.NotContains(t, commands, "/update")
		})
	}
}

func TestNoPersistenceAPICallsForNavigationOnlyActions(t *testing.T) {
	client := &fakeClient{
		phases: []api.Option{{ID: "backlog", Label: "backlog"}},
	}
	m := NewModelWithClient(client)
	m.state = ChangeDetailsState

	got, _ := sendCommand(m, "/save")
	got, _ = sendCommand(got, "/delete")
	got.dropdown.filter = "/yes"
	got, _ = sendKey(got, tea.KeyEnter)
	assert.Zero(t, client.listCalls)
	assert.Zero(t, client.phaseCalls)
	assert.Zero(t, client.typeCalls)
	assert.Zero(t, client.epicCalls)

	got.state = ChangesListState
	got, cmd := sendCommand(got, "/phase-filter")
	require.NotNil(t, cmd)
	got = applyMsg(got, cmd())
	got, _ = sendKey(got, tea.KeyEnter)
	assert.Equal(t, 1, client.phaseCalls)
}

func TestEveryDummyScreenTitleRendersExactly(t *testing.T) {
	tests := []struct {
		state State
		title string
	}{
		{MainState, "MainScreen - Title: Main"},
		{ChangesListState, "ChangesListScreen - Title: Changes List"},
		{ChangeDetailsState, "ChangeDetailsScreen - Title: Change Details"},
		{RequirementDetailsState, "RequirementDetailsScreen - Title: Requirement Details"},
		{ChangeCreateState, "ChangeCreateScreen - Title: New Change"},
		{ChangeUpdateState, "ChangeUpdateScreen - Title: Edit Change"},
		{RequirementCreateState, "RequirementCreateScreen - Title: New Requirement"},
		{RequirementUpdateState, "RequirementUpdateScreen - Title: Edit Requirement"},
		{EpicsListState, "EpicsListScreen - Title: Epics List"},
		{EpicDetailsState, "EpicDetailsScreen - Title: Epic Details"},
		{EpicCreateState, "EpicCreateScreen - Title: New Epic"},
		{EpicUpdateState, "EpicUpdateScreen - Title: Edit Epic"},
		{ProjectsListState, "ProjectsListScreen - Title: Projects List"},
		{ProjectDetailsState, "ProjectDetailsScreen - Title: Project Details"},
		{ProjectCreateState, "ProjectCreateScreen - Title: New Project"},
		{ProjectUpdateState, "ProjectUpdateScreen - Title: Edit Project"},
		{MainHelpState, "MainHelpScreen - Title: Main Help"},
		{ChangesHelpState, "ChangesHelpScreen - Title: Changes Help"},
		{EpicsHelpState, "EpicsHelpScreen - Title: Epics Help"},
		{ProjectsHelpState, "ProjectsHelpScreen - Title: Projects Help"},
		{FindInputState, "FindInputScreen - Title: Find"},
		{CommandDropDownState, "CommandDropDownScreen - Title: Commands"},
		{ListSelectionDropDownState, "ListSelectionDropDownScreen - Title: Select Item"},
		{SelectProjectDropDown, "SelectProjectDropDownScreen - Title: Select Project"},
		{SelectPhaseDropDown, "SelectPhaseDropDownScreen - Title: Select Phase"},
		{SelectEpicDropDown, "SelectEpicDropDownScreen - Title: Select Epic"},
		{SelectTypesDropDown, "SelectTypesDropDownScreen - Title: Select Types"},
		{ChangeDeleteConfirmation, "ChangeDeleteConfirmationScreen - Title: Confirm Delete"},
		{RequirementDeleteConfirmation, "RequirementDeleteConfirmationScreen - Title: Confirm Delete"},
		{EpicDeleteConfirmation, "EpicDeleteConfirmationScreen - Title: Confirm Delete"},
		{ProjectDeleteConfirmation, "ProjectDeleteConfirmationScreen - Title: Confirm Delete"},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			m := NewModel()
			m.state = tt.state

			view := m.View()
			assert.Contains(t, view, tt.title)
			assert.NotContains(t, view, "Make a Change")
		})
	}
}

func sendCommand(m Model, command string) (Model, tea.Cmd) {
	m.input.SetValue(command)
	return sendKey(m, tea.KeyEnter)
}

func sendRune(m Model, r rune) (Model, tea.Cmd) {
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	return updated.(Model), cmd
}

func sendKey(m Model, key tea.KeyType) (Model, tea.Cmd) {
	updated, cmd := m.Update(tea.KeyMsg{Type: key})
	return updated.(Model), cmd
}

func applyMsg(m Model, msg tea.Msg) Model {
	updated, _ := m.Update(msg)
	return updated.(Model)
}

func stripANSI(value string) string {
	var b strings.Builder
	inEscape := false
	for _, r := range value {
		if inEscape {
			if r == '[' || (r >= '0' && r <= '?') {
				continue
			}
			if r >= '@' && r <= '~' {
				inEscape = false
			}
			continue
		}
		if r == '\x1b' {
			inEscape = true
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
