package app

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"mch/internal/dto"
	"mch/internal/projects"
	"mch/internal/styles"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClient struct {
	projects       []dto.Option
	projectRows    []dto.Project
	createdProject dto.Project
	updatedProject dto.Project
	gotProject     dto.Project
	epics          []dto.Option
	phases         []dto.Option
	types          []dto.Option
	err            error
	createErr      error
	updateErr      error
	getErr         error
	projectID      string
	listCalls      int
	rowListCalls   int
	createCalls    int
	updateCalls    int
	getCalls       int
	phaseCalls     int
	typeCalls      int
	epicCalls      int
	createNames    []string
	updateIDs      []int
	updateNames    []string
	getIDs         []int
}

func (f *fakeClient) ListProjects() ([]dto.Option, error) {
	f.listCalls++
	return f.projects, f.err
}

func (f *fakeClient) ListProjectRows() ([]dto.Project, error) {
	f.rowListCalls++
	return f.projectRows, f.err
}

func (f *fakeClient) GetProject(id int) (dto.Project, error) {
	f.getCalls++
	f.getIDs = append(f.getIDs, id)
	if f.getErr != nil {
		return dto.Project{}, f.getErr
	}
	if f.err != nil {
		return dto.Project{}, f.err
	}
	return f.gotProject, nil
}

func (f *fakeClient) CreateProject(name string) (dto.Project, error) {
	f.createCalls++
	f.createNames = append(f.createNames, name)
	if f.createErr != nil {
		return dto.Project{}, f.createErr
	}
	if f.err != nil {
		return dto.Project{}, f.err
	}
	return f.createdProject, nil
}

func (f *fakeClient) UpdateProject(id int, name string) (dto.Project, error) {
	f.updateCalls++
	f.updateIDs = append(f.updateIDs, id)
	f.updateNames = append(f.updateNames, name)
	if f.updateErr != nil {
		return dto.Project{}, f.updateErr
	}
	if f.err != nil {
		return dto.Project{}, f.err
	}
	return f.updatedProject, nil
}

func (f *fakeClient) ListEpics(projectID string) ([]dto.Option, error) {
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

func (f *fakeClient) ListPhases() ([]dto.Option, error) {
	f.phaseCalls++
	return f.phases, f.err
}

func (f *fakeClient) ListTypes() ([]dto.Option, error) {
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

func TestShellChromeRendersTitleAndCurrentProjectInFooter(t *testing.T) {
	m := newModelWithConfig(&fakeClient{}, appConfig{BackendURL: defaultBackendURL, ProjectID: 7}, "")
	m.currentProject = dto.Option{ID: "7", Label: "Project Seven"}
	m.width = 120

	view := stripANSI(m.View())

	assert.Contains(t, view, "Make a Change ver. 0.1")
	assert.NotContains(t, view, "\nversion 0.1")
	assert.NotContains(t, view, "\nProject: ")
	assert.Contains(t, view, "Current Project: #7 Project Seven")
}

func TestStartupTriggersProjectSelectionWhenProjectIDIsUnset(t *testing.T) {
	client := &fakeClient{
		projects: []dto.Option{{ID: "7", Label: "Project Seven"}},
	}
	m := NewModelWithClient(client)

	cmd := m.Init()
	require.NotNil(t, cmd)
	got := applyCommand(m, cmd)
	assert.Equal(t, SelectProjectDropDown, got.state)
	assert.Equal(t, selectorProjects, got.dropdown.source)

	load := selectorCommand(client, got.dropdown.source, got.currentProject.ID)
	got = applyMsg(got, load())

	assert.Equal(t, SelectProjectDropDown, got.state)
	assert.Equal(t, []dto.Option{{ID: "7", Label: "Project Seven"}}, got.dropdown.options)
}

func TestStartupSkipsProjectSelectionWhenProjectIDIsSaved(t *testing.T) {
	client := &fakeClient{gotProject: dto.Project{ID: "7", Name: "Project Seven"}}
	m := newModelWithConfig(client, appConfig{BackendURL: defaultBackendURL, ProjectID: 7}, "")
	m.width = 120

	require.NotNil(t, m.Init())
	got := applyCommand(m, m.Init())
	assert.Equal(t, MainState, m.state)
	assert.Equal(t, MainState, got.state)
	assert.Equal(t, "7", got.currentProject.ID)
	assert.Equal(t, "Project Seven", got.currentProject.Label)
	assert.Contains(t, stripANSI(got.View()), "Current Project: #7 Project Seven")
}

func TestStartupProjectSelectionShowsErrorWhenNoProjectsExist(t *testing.T) {
	client := &fakeClient{}
	m := NewModelWithClient(client)

	cmd := m.Init()
	require.NotNil(t, cmd)
	got := applyCommand(m, cmd)
	load := selectorCommand(client, got.dropdown.source, got.currentProject.ID)
	got = applyMsg(got, load())

	assert.Equal(t, MainState, got.state)
	assert.Empty(t, got.dropdown.kind)
	assert.Equal(t, noProjectsToSelectError, got.err)
}

func TestInputBandUsesCliProtoFullWidthBackground(t *testing.T) {
	m := NewModel()
	m.width = 40
	assert.Equal(t, 1, m.input.Width())

	band := m.inputBand(40)
	lines := strings.Split(band, "\n")
	require.Len(t, lines, 3)
	assert.Contains(t, band, "Type / for commands")
	for i, line := range lines {
		visible := stripANSI(line)
		assert.Falsef(t, strings.TrimSpace(visible) == "" && len(visible) < 40, "blank input band line %d too short: %q", i, visible)
	}
	assert.True(t, strings.HasPrefix(stripANSI(lines[1]), "> Type / for commands"))

	m = m.setPromptValue("typed text")
	typedBand := m.inputBand(40)
	assert.NotContains(t, typedBand, "48;5;0")
	assert.NotContains(t, typedBand, "[40m")
	typedLine := stripANSI(strings.Split(typedBand, "\n")[1])
	assert.True(t, strings.HasPrefix(typedLine, "> typed text"))
	assert.Equal(t, "15", fmt.Sprint(m.input.FocusedStyle.Text.GetForeground()))
	assert.Equal(t, "15", fmt.Sprint(m.input.FocusedStyle.CursorLine.GetForeground()))
	assert.Equal(t, "0", fmt.Sprint(m.input.FocusedStyle.Placeholder.GetForeground()))
	assert.Equal(t, cursor.CursorStatic, m.input.Cursor.Mode())

	wideBand := m.inputBand(180)
	wideLines := strings.Split(wideBand, "\n")
	require.Len(t, wideLines, 3)
	assert.Len(t, stripANSI(wideLines[0]), 180)
	assert.Len(t, stripANSI(wideLines[1]), 180)
	assert.Len(t, stripANSI(wideLines[2]), 180)
}

func TestPromptTextareaGrowsForExplicitNewlines(t *testing.T) {
	m := NewModel()
	m = m.setPromptValue("first line\nsecond line\n")

	band := stripANSI(m.inputBand(40))
	lines := strings.Split(band, "\n")

	require.Len(t, lines, 5)
	assert.True(t, strings.HasPrefix(lines[1], "> first line"))
	assert.True(t, strings.HasPrefix(lines[2], "> second line"))
	assert.True(t, strings.HasPrefix(lines[3], "> "))
}

func TestPromptNewlineKeyAddsBlankPromptLine(t *testing.T) {
	m := NewModel()
	m = m.setPromptValue("first line")

	got, cmd := sendKeyMsg(m, tea.KeyMsg{Type: tea.KeyEnter, Alt: true})

	assert.Nil(t, cmd)
	assert.Equal(t, "first line\n", got.input.Value())
	band := got.inputBand(40)
	assert.NotContains(t, band, "48;5;0")
	assert.NotContains(t, band, "[40m")
	assert.Equal(t, 4, len(strings.Split(stripANSI(band), "\n")))
}

func TestPromptShiftEnterEscapeSequenceAddsNewline(t *testing.T) {
	m := NewModel()
	m = m.setPromptValue("first line")

	got, cmd := sendKeyMsg(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'O'}, Alt: true})
	assert.Nil(t, cmd)
	assert.Equal(t, "first line", got.input.Value())
	assert.True(t, got.pendingAltO)

	got, cmd = sendKeyMsg(got, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'M'}})

	assert.Nil(t, cmd)
	assert.False(t, got.pendingAltO)
	assert.Equal(t, "first line\n", got.input.Value())
	assert.NotContains(t, got.input.Value(), "OM")
	assert.Equal(t, 4, len(strings.Split(stripANSI(got.inputBand(40)), "\n")))
}

func TestPromptInputUsesTerminalWidthForTyping(t *testing.T) {
	m := NewModel()
	m.width = 40

	got, cmd := sendRune(m, 'a')

	assert.Nil(t, cmd)
	assert.Equal(t, "a", got.input.Value())
	assert.Greater(t, got.input.Width(), 1)
}

func TestPromptUpDownMovesVisibleCursorBetweenLines(t *testing.T) {
	m := NewModel()
	m = m.setPromptValue("first\nsecond")

	got, cmd := sendKey(m, tea.KeyUp)

	assert.Nil(t, cmd)
	assert.Equal(t, 0, got.promptCursorRow)
	assert.Equal(t, len("first"), got.promptCursorCol)

	got, cmd = sendKey(got, tea.KeyDown)

	assert.Nil(t, cmd)
	assert.Equal(t, 1, got.promptCursorRow)
	assert.Equal(t, len("first"), got.promptCursorCol)
}

func TestViewAddsBlankLineBetweenPromptAndFooter(t *testing.T) {
	m := NewModelWithClient(&fakeClient{})
	m.width = 40

	lines := strings.Split(stripANSI(m.View()), "\n")
	var promptLine int
	for i, line := range lines {
		if strings.HasPrefix(line, "> Type / for commands") {
			promptLine = i
			break
		}
	}
	require.NotZero(t, promptLine)
	require.Greater(t, len(lines), promptLine+3)
	assert.Empty(t, strings.TrimSpace(lines[promptLine+1]))
	assert.Empty(t, strings.TrimSpace(lines[promptLine+2]))
	assert.Contains(t, lines[promptLine+3], "/ commands")
}

func TestNewProjectUsesNamePlaceholder(t *testing.T) {
	m := NewModelWithClient(&fakeClient{})
	m.state = ProjectsListState

	got, _ := sendCommand(m, "/new-project")

	assert.Equal(t, ProjectCreateState, got.state)
	assert.Equal(t, "Write a Name", got.input.Placeholder)
	assert.Contains(t, stripANSI(got.inputBand(40)), "> Write a Name")

	got, _ = sendCommand(got, "/cancel")
	assert.Equal(t, defaultInputPlaceholder, got.input.Placeholder)
}

func TestProjectFormsExposeEditorCommandFirst(t *testing.T) {
	assert.Equal(t, []string{"/editor", "/save", "/cancel"}, commandsByState[ProjectCreateState])
	assert.Equal(t, []string{"/editor", "/save", "/cancel"}, commandsByState[ProjectUpdateState])
}

func TestProjectEditorSavesResultWithoutReturningToPrompt(t *testing.T) {
	m := NewModelWithClient(&fakeClient{})
	m.state = ProjectCreateState
	m.input.SetValue("Initial Name")

	updated, cmd := m.Update(editorFinishedMsg{source: ProjectCreateState, content: "Edited\nName\n"})
	got := updated.(Model)

	require.NotNil(t, cmd)
	assert.Equal(t, "Initial Name", got.input.Value())
	assert.Equal(t, "saving", got.status)
}

func TestProjectEditorIgnoresStaleResultAndReportsErrors(t *testing.T) {
	m := NewModelWithClient(&fakeClient{})
	m.state = ProjectUpdateState
	m.input.SetValue("Current")

	got := applyMsg(m, editorFinishedMsg{source: ProjectCreateState, content: "Stale"})
	assert.Equal(t, "Current", got.input.Value())

	got = applyMsg(got, editorFinishedMsg{source: ProjectUpdateState, err: errors.New("nano failed")})
	assert.Equal(t, "Current", got.input.Value())
	assert.Equal(t, "nano failed", got.err)
	assert.Equal(t, "editor failed", got.status)
}

func TestProjectEditorUsesEditorEnvWithNanoFallback(t *testing.T) {
	t.Setenv("EDITOR", "")
	fallback := editorCommand("/tmp/project.md")
	assert.Equal(t, "nano", fallback.Args[0])
	assert.Equal(t, "/tmp/project.md", fallback.Args[1])

	t.Setenv("EDITOR", "vim -f")
	fromEnv := editorCommand("/tmp/project.md")
	assert.Equal(t, "sh", fromEnv.Args[0])
	assert.Equal(t, []string{"sh", "-c", "$EDITOR \"$1\"", "mch-editor", "/tmp/project.md"}, fromEnv.Args)
	assert.Contains(t, fromEnv.Env, "EDITOR=vim -f")
}

func TestPromptEnterSavesProjectFormRawMultilineValue(t *testing.T) {
	client := &fakeClient{
		createdProject: dto.Project{ID: "7"},
		gotProject:     dto.Project{ID: "7", Name: "Line 1\nLine 2"},
	}
	m := NewModelWithClient(client)
	m.state = ProjectCreateState
	m.input.SetValue("Line 1\nLine 2")

	updated, cmd := sendKey(m, tea.KeyEnter)
	got := updated

	require.NotNil(t, cmd)
	assert.Equal(t, ProjectCreateState, got.state)
	assert.Equal(t, "saving", got.status)

	got = applyMsg(got, cmd())

	assert.Equal(t, ProjectDetailsState, got.state)
	assert.Equal(t, []string{"Line 1\nLine 2"}, client.createNames)
}

func TestPromptCtrlCClearsBeforeCancelingOrQuitting(t *testing.T) {
	m := NewModelWithClient(&fakeClient{})
	m.state = ProjectCreateState
	m.input.SetValue("draft")

	got, cmd := sendKey(m, tea.KeyCtrlC)
	assert.Nil(t, cmd)
	assert.Equal(t, ProjectCreateState, got.state)
	assert.Empty(t, got.input.Value())
	assert.Equal(t, "prompt cleared", got.status)

	got, cmd = sendKey(got, tea.KeyCtrlC)
	assert.NotNil(t, cmd)
	assert.Equal(t, ProjectsListState, got.state)

	got, cmd = sendKey(NewModelWithClient(&fakeClient{}), tea.KeyCtrlC)
	assert.NotNil(t, cmd)
	assert.Equal(t, DoneState, got.state)
	assert.True(t, got.quitting)
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

func TestProjectsCommandReloadsAndRendersSelectableTable(t *testing.T) {
	client := &fakeClient{
		projectRows: []dto.Project{
			{
				ID:          "7",
				Name:        "Project Seven",
				ChangeCount: 3,
				Created:     "2026-06-29T08:15:00Z",
				Modified:    "2026-06-29T10:45:00Z",
			},
			{
				ID:          "8",
				Name:        "Project Eight",
				ChangeCount: 0,
				Created:     "bad timestamp",
				Modified:    "",
			},
		},
	}
	m := NewModelWithClient(client)

	got, cmd := sendCommand(m, "/projects")
	require.Equal(t, ProjectsListState, got.state)
	require.NotNil(t, cmd)
	assert.True(t, got.projectList.Loading)

	got = applyMsg(got, cmd())

	assert.Equal(t, 1, client.rowListCalls)
	assert.False(t, got.projectList.Loading)
	assert.Equal(t, 0, got.projectList.Selected)
	view := stripANSI(got.View())
	assert.Contains(t, view, "ProjectsListScreen - Title: Projects List")
	assert.Contains(t, view, "id")
	assert.Contains(t, view, "Name")
	assert.Contains(t, view, "Changes")
	assert.Contains(t, view, "Created")
	assert.Contains(t, view, "Modified")
	assert.Contains(t, view, "     7  Project Seven")
	assert.Contains(t, view, "Project Seven")
	assert.Contains(t, view, "3")
	assert.Contains(t, view, "2026-06-29")
	assert.Contains(t, view, "not a date")

	got, _ = sendCommand(got, "/return")
	got, cmd = sendCommand(got, "/projects")
	require.NotNil(t, cmd)
	got = applyMsg(got, cmd())
	assert.Equal(t, 2, client.rowListCalls)
	assert.Equal(t, ProjectsListState, got.state)
}

func TestProjectsTableUsesDynamicNameWidthAndTrimsVeryLongNames(t *testing.T) {
	longName := "This is a project with a real name that is long enough to resize the name column"
	tooLongName := longName + " and has additional words on the right that must be removed"
	m := NewModelWithClient(&fakeClient{})
	m.state = ProjectsListState
	m.projectList.Rows = []dto.Project{
		{ID: "1", Name: "demo1", ChangeCount: 2, Created: "2026-06-23T04:51:00Z", Modified: "2026-06-23T04:51:00Z"},
		{ID: "350", Name: longName, ChangeCount: 0, Created: "2026-06-29T15:57:00Z", Modified: "2026-06-29T15:57:00Z"},
		{ID: "351", Name: tooLongName, ChangeCount: 1, Created: "2026-06-29T15:58:00Z", Modified: "2026-06-29T15:58:00Z"},
	}

	rendered := stripANSI(projects.TableView(m.projectList, 160))
	lines := strings.Split(rendered, "\n")
	require.Len(t, lines, 4)

	createdColumn := strings.Index(lines[0], "Created")
	require.NotEqual(t, -1, createdColumn)
	assert.Equal(t, createdColumn, strings.Index(lines[1], "2026-"))
	assert.Equal(t, createdColumn, strings.Index(lines[2], "2026-"))
	assert.Equal(t, createdColumn, strings.Index(lines[3], "2026-"))
	assert.Contains(t, lines[2], longName)
	assert.NotContains(t, lines[3], "must be removed")
	trimmedName := projects.ProjectTableName(tooLongName)
	assert.True(t, strings.HasSuffix(trimmedName, "..."))
	assert.Less(t, len([]rune(trimmedName)), 78)
	assert.Contains(t, lines[3], trimmedName)
}

func TestProjectsTableSelectionIsBounded(t *testing.T) {
	m := NewModelWithClient(&fakeClient{})
	m.state = ProjectsListState
	m.projectList.Rows = []dto.Project{
		{ID: "1", Name: "One"},
		{ID: "2", Name: "Two"},
	}

	got, _ := sendKey(m, tea.KeyUp)
	assert.Equal(t, 0, got.projectList.Selected)

	got, _ = sendKey(got, tea.KeyDown)
	assert.Equal(t, 1, got.projectList.Selected)

	got, _ = sendKey(got, tea.KeyDown)
	assert.Equal(t, 1, got.projectList.Selected)

	got, _ = sendKey(got, tea.KeyUp)
	assert.Equal(t, 0, got.projectList.Selected)
}

func TestProjectsEnterOpensDetailsWithoutMutatingCurrentProject(t *testing.T) {
	current := dto.Option{ID: "99", Label: "Current Project"}
	client := &fakeClient{
		gotProject: dto.Project{ID: "8", Name: "Fresh Project Eight", ChangeCount: 5, Created: "2026-06-30T08:15:00Z", Modified: "2026-06-30T11:45:00Z"},
	}
	m := NewModelWithClient(client)
	m.state = ProjectsListState
	m.currentProject = current
	m.projectList.Rows = []dto.Project{
		{ID: "7", Name: "Project Seven", ChangeCount: 3, Created: "2026-06-29T08:15:00Z", Modified: "2026-06-29T10:45:00Z"},
		{ID: "8", Name: "Project Eight", ChangeCount: 4, Created: "2026-06-30T08:15:00Z", Modified: "2026-06-30T10:45:00Z"},
	}
	m.projectList.Selected = 1

	got, cmd := sendKey(m, tea.KeyEnter)

	assert.Equal(t, ProjectDetailsState, got.state)
	assert.Equal(t, current, got.currentProject)
	assert.Equal(t, dto.Project{ID: "8", Name: "Project Eight", ChangeCount: 4, Created: "2026-06-30T08:15:00Z", Modified: "2026-06-30T10:45:00Z"}, got.projectList.Detail)
	require.NotNil(t, cmd)
	got = applyMsg(got, cmd())
	assert.Equal(t, []int{8}, client.getIDs)
	assert.Equal(t, client.gotProject, got.projectList.Detail)
	view := stripANSI(got.View())
	assert.Contains(t, view, "ProjectDetailsScreen - Title: Project Details")
	assert.Contains(t, view, "         #ID: 8")
	assert.Contains(t, view, "        Name: Fresh Project Eight")
	assert.Contains(t, view, "Changes: 5")
}

func TestProjectDetailsRenderRequiredLabelsAndTimestampFallback(t *testing.T) {
	m := NewModelWithClient(&fakeClient{})
	m.state = ProjectDetailsState
	m.width = 32
	m.projectList.Detail = dto.Project{
		ID:          "7",
		Name:        "Project Seven",
		ChangeCount: 3,
		Created:     "2026-06-29T13:04:59.999Z",
		Modified:    "malformed",
	}

	rawDetails := projects.DetailsView(m.projectList.Detail, 32)
	view := stripANSI(m.View())
	whiteValue := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	pinkValue := lipgloss.NewStyle().Foreground(lipgloss.Color("218"))
	timestampValue := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	createdValue := projects.FormatTimestamp("2026-06-29T13:04:59.999Z")

	assert.Contains(t, view, "         #ID: 7")
	assert.Contains(t, view, "        Name: Project Seven")
	assert.Contains(t, view, "     Changes: 3")
	assert.Contains(t, view, "     Created: "+createdValue)
	assert.Contains(t, view, "    Modified: not a date")
	assert.Contains(t, rawDetails, pinkValue.Render("7"))
	assert.Contains(t, rawDetails, whiteValue.Render("3"))
	assert.Contains(t, rawDetails, timestampValue.Render(createdValue))
	assert.Contains(t, rawDetails, timestampValue.Render("not a date"))
	assert.Contains(t, rawDetails, styles.Default.AccentCyan.Render("Project Seven"))
	for _, line := range strings.Split(stripANSI(rawDetails), "\n") {
		assert.LessOrEqual(t, len(line), 32)
	}
}

func TestProjectDetailsWrapsNameAtEightyCharactersWithoutBreakingWords(t *testing.T) {
	name := "This project name is deliberately long and should wrap onto the next line without breaking any words in half"
	m := NewModelWithClient(&fakeClient{})
	m.state = ProjectDetailsState
	m.width = 120
	m.projectList.Detail = dto.Project{ID: "7", Name: name}

	view := stripANSI(m.View())

	assert.Contains(t, view, "        Name: This project name is deliberately long and should wrap onto the next line")
	assert.Contains(t, view, "\n              without breaking any words in half")
	assert.NotContains(t, view, "witho\n")
}

func TestProjectDetailsPreservesExplicitNameNewlines(t *testing.T) {
	m := NewModelWithClient(&fakeClient{})
	m.state = ProjectDetailsState
	m.width = 120
	m.projectList.Detail = dto.Project{ID: "7", Name: "First line\nSecond line"}

	view := stripANSI(m.View())

	assert.Contains(t, view, "        Name: First line")
	assert.Contains(t, view, "\n              Second line")
}

func TestProjectPagesReloadOnArrival(t *testing.T) {
	client := &fakeClient{
		projectRows: []dto.Project{{ID: "7", Name: "Reloaded List Project"}},
		gotProject:  dto.Project{ID: "7", Name: "Reloaded Detail Project"},
	}

	m := NewModelWithClient(client)
	m.state = ProjectDetailsState
	m.projectList.Detail = dto.Project{ID: "7", Name: "Stale Detail Project"}
	got, cmd := sendCommand(m, "/return")
	require.NotNil(t, cmd)
	assert.Equal(t, ProjectsListState, got.state)
	assert.True(t, got.projectList.Loading)
	got = applyMsg(got, cmd())
	assert.Equal(t, 1, client.rowListCalls)
	assert.Equal(t, []dto.Project{{ID: "7", Name: "Reloaded List Project"}}, got.projectList.Rows)

	got.state = ProjectUpdateState
	got.projectList.Detail = dto.Project{ID: "7", Name: "Stale Detail Project"}
	got, cmd = sendCommand(got, "/cancel")
	require.NotNil(t, cmd)
	assert.Equal(t, ProjectDetailsState, got.state)
	got = applyMsg(got, cmd())
	assert.Equal(t, []int{7}, client.getIDs)
	assert.Equal(t, client.gotProject, got.projectList.Detail)
}

func TestProjectsEnterWithNoSelectableRowErrors(t *testing.T) {
	m := NewModelWithClient(&fakeClient{})
	m.state = ProjectsListState

	got, _ := sendKey(m, tea.KeyEnter)

	assert.Equal(t, ProjectsListState, got.state)
	assert.NotEmpty(t, got.err)
}

func TestProjectsLoadFailureAndEmptyListAreDeterministic(t *testing.T) {
	failing := &fakeClient{err: errors.New("backend unavailable")}
	m := NewModelWithClient(failing)

	got, cmd := sendCommand(m, "/projects")
	require.NotNil(t, cmd)
	got = applyMsg(got, cmd())

	assert.Equal(t, ProjectsListState, got.state)
	assert.False(t, got.projectList.Loading)
	assert.Equal(t, "backend unavailable", got.err)
	assert.Contains(t, stripANSI(got.View()), "No projects.")

	empty := NewModelWithClient(&fakeClient{})
	got, cmd = sendCommand(empty, "/projects")
	require.NotNil(t, cmd)
	got = applyMsg(got, cmd())

	assert.Equal(t, ProjectsListState, got.state)
	assert.Contains(t, stripANSI(got.View()), "No projects.")
}

func TestProjectCreateSavePersistsFetchesDetailsAndDoesNotMutateConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".config", "config.yaml")
	require.NoError(t, saveAppConfig(path, appConfig{BackendURL: defaultBackendURL, ProjectID: 99}))
	client := &fakeClient{
		createdProject: dto.Project{ID: "7"},
		gotProject: dto.Project{
			ID:          "7",
			Name:        "New Project",
			ChangeCount: 0,
			Created:     "2026-06-29T11:04:59Z",
			Modified:    "2026-06-29T11:04:59Z",
		},
	}
	m := newModelWithConfig(client, appConfig{BackendURL: defaultBackendURL, ProjectID: 99}, path)
	m.state = ProjectCreateState
	m.input.SetValue("  New\nProject  ")

	updated, cmd := m.executeCommandFrom(ProjectCreateState, "/save")
	got := updated.(Model)
	require.NotNil(t, cmd)
	assert.Equal(t, ProjectCreateState, got.state)
	assert.Equal(t, "saving", got.status)

	got = applyMsg(got, cmd())

	assert.Equal(t, ProjectDetailsState, got.state)
	assert.Equal(t, []string{"  New\nProject  "}, client.createNames)
	assert.Equal(t, []int{7}, client.getIDs)
	assert.Equal(t, client.gotProject, got.projectList.Detail)
	assert.Equal(t, "99", got.currentProject.ID)
	loaded, err := loadAppConfig(path)
	require.NoError(t, err)
	assert.Equal(t, 99, loaded.ProjectID)
	view := stripANSI(got.View())
	assert.Contains(t, view, "Name: New Project")
	assert.Contains(t, view, "Changes: 0")
}

func TestProjectCreateValidationDoesNotCallBackend(t *testing.T) {
	client := &fakeClient{}
	m := NewModelWithClient(client)
	m.state = ProjectCreateState
	m.input.SetValue("   ")

	got, cmd := sendKey(m, tea.KeyEnter)

	assert.Nil(t, cmd)
	assert.Equal(t, ProjectCreateState, got.state)
	assert.Contains(t, got.err, "project name is required")
	assert.Zero(t, client.createCalls)
	assert.Zero(t, client.getCalls)
}

func TestProjectUpdateSavePersistsFetchesDetailsAndDoesNotMutateConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".config", "config.yaml")
	require.NoError(t, saveAppConfig(path, appConfig{BackendURL: defaultBackendURL, ProjectID: 99}))
	client := &fakeClient{
		updatedProject: dto.Project{ID: "7"},
		gotProject: dto.Project{
			ID:          "7",
			Name:        "Renamed Project",
			ChangeCount: 2,
			Created:     "2026-06-29T08:15:00Z",
			Modified:    "2026-06-29T13:04:59Z",
		},
	}
	m := newModelWithConfig(client, appConfig{BackendURL: defaultBackendURL, ProjectID: 99}, path)
	m.state = ProjectDetailsState
	m.projectList.Detail = dto.Project{ID: "7", Name: "Old Project", ChangeCount: 2}

	got, _ := sendCommand(m, "/edit")
	assert.Equal(t, ProjectUpdateState, got.state)
	assert.Equal(t, "Old Project", got.input.Value())
	got.input.SetValue("  Renamed\nProject  ")

	updated, cmd := got.executeCommandFrom(ProjectUpdateState, "/save")
	got = updated.(Model)
	require.NotNil(t, cmd)
	got = applyMsg(got, cmd())

	assert.Equal(t, ProjectDetailsState, got.state)
	assert.Equal(t, []int{7}, client.updateIDs)
	assert.Equal(t, []string{"  Renamed\nProject  "}, client.updateNames)
	assert.Equal(t, []int{7}, client.getIDs)
	assert.Equal(t, client.gotProject, got.projectList.Detail)
	assert.Equal(t, "99", got.currentProject.ID)
	loaded, err := loadAppConfig(path)
	require.NoError(t, err)
	assert.Equal(t, 99, loaded.ProjectID)
}

func TestProjectUpdateValidationDoesNotCallBackend(t *testing.T) {
	tests := []dto.Project{
		{},
		{ID: "0", Name: "Zero"},
		{ID: "-1", Name: "Negative"},
		{ID: "not-a-number", Name: "Bad"},
	}

	for _, project := range tests {
		t.Run(project.ID, func(t *testing.T) {
			client := &fakeClient{}
			m := NewModelWithClient(client)
			m.state = ProjectUpdateState
			m.projectList.Detail = project
			m.input.SetValue("Renamed")

			updated, cmd := m.executeCommandFrom(ProjectUpdateState, "/save")
			got := updated.(Model)

			assert.Nil(t, cmd)
			assert.Equal(t, ProjectUpdateState, got.state)
			assert.Contains(t, got.err, "project ID must be a valid positive number")
			assert.Zero(t, client.updateCalls)
			assert.Zero(t, client.getCalls)
		})
	}
}

func TestProjectSaveBackendFailurePreservesRecoverableFormState(t *testing.T) {
	client := &fakeClient{createErr: errors.New("invalid project payload")}
	m := NewModelWithClient(client)
	m.state = ProjectCreateState
	m.input.SetValue("New Project")

	updated, cmd := m.executeCommandFrom(ProjectCreateState, "/save")
	got := updated.(Model)
	require.NotNil(t, cmd)
	got = applyMsg(got, cmd())

	assert.Equal(t, ProjectCreateState, got.state)
	assert.Equal(t, "New Project", got.input.Value())
	assert.Equal(t, "invalid project payload", got.err)
	assert.Equal(t, 1, client.createCalls)
	assert.Zero(t, client.getCalls)

	client = &fakeClient{
		updatedProject: dto.Project{ID: "7"},
		getErr:         errors.New("project not found"),
	}
	m = NewModelWithClient(client)
	m.state = ProjectUpdateState
	m.projectList.Detail = dto.Project{ID: "7", Name: "Old Project"}
	m.input.SetValue("Renamed Project")

	updated, cmd = m.executeCommandFrom(ProjectUpdateState, "/save")
	got = updated.(Model)
	require.NotNil(t, cmd)
	got = applyMsg(got, cmd())

	assert.Equal(t, ProjectUpdateState, got.state)
	assert.Equal(t, "Renamed Project", got.input.Value())
	assert.Equal(t, "project not found", got.err)
	assert.Equal(t, 1, client.updateCalls)
	assert.Equal(t, 1, client.getCalls)
}

func TestProjectCancelDoesNotCallPersistence(t *testing.T) {
	client := &fakeClient{}
	m := NewModelWithClient(client)
	m.state = ProjectCreateState
	m.input.SetValue("New Project")

	got, _ := sendCommand(m, "/cancel")

	assert.Equal(t, ProjectsListState, got.state)
	assert.Zero(t, client.createCalls)
	assert.Zero(t, client.updateCalls)

	m = NewModelWithClient(client)
	m.state = ProjectUpdateState
	m.projectList.Detail = dto.Project{ID: "7", Name: "Old Project"}
	m.input.SetValue("Renamed Project")

	got, _ = sendKey(m, tea.KeyEsc)

	assert.Equal(t, ProjectDetailsState, got.state)
	assert.Zero(t, client.createCalls)
	assert.Zero(t, client.updateCalls)
}

func TestProjectsTableNarrowWidthDoesNotOverflow(t *testing.T) {
	m := NewModelWithClient(&fakeClient{})
	m.state = ProjectsListState
	m.width = 24
	m.projectList.Rows = []dto.Project{{
		ID:          "777777",
		Name:        "Very Long Project Name That Must Be Truncated",
		ChangeCount: 123,
		Created:     "2026-06-29T08:15:00Z",
		Modified:    "2026-06-29T10:45:00Z",
	}}

	for _, line := range strings.Split(stripANSI(projects.TableView(m.projectList, 24)), "\n") {
		assert.LessOrEqual(t, len(line), 24)
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
		projects: []dto.Option{{ID: "7", Label: "Project Seven"}},
		phases:   []dto.Option{{ID: "backlog", Label: "backlog"}},
		types:    []dto.Option{{ID: "feature", Label: "feature"}},
		epics:    []dto.Option{{ID: "3", Label: "Epic Three"}},
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
		projects: []dto.Option{{ID: "7", Label: "Project Seven"}},
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
		phases: []dto.Option{{ID: "done", Label: "done"}},
		epics:  []dto.Option{{ID: "epic-1", Label: "Epic One"}},
		types:  []dto.Option{{ID: "test", Label: "test"}},
	}
	m := NewModelWithClient(client)
	m.state = ChangesListState
	m.currentProject = dto.Option{ID: "project-1", Label: "Project One"}

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

func TestProjectsCommandMenuPreservesListTitle(t *testing.T) {
	m := NewModelWithClient(&fakeClient{})
	m.state = ProjectsListState
	m.projectList.Rows = []dto.Project{{ID: "7", Name: "Project Seven"}}

	got, _ := sendRune(m, '/')

	assert.Equal(t, ProjectsListState, got.state)
	assert.Equal(t, dropdownCommand, got.dropdown.kind)
	view := stripANSI(got.View())
	assert.Contains(t, view, "ProjectsListScreen - Title: Projects List")
	assert.Contains(t, view, "/new-project")
	assert.Contains(t, view, "/help")
	assert.Contains(t, view, "/find")
	assert.Contains(t, view, "/return")
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
		phases: []dto.Option{{ID: "backlog", Label: "backlog"}},
	}
	m := NewModelWithClient(client)
	m.state = ChangeDetailsState

	got, _ := sendCommand(m, "/save")
	got, _ = sendCommand(got, "/delete")
	got.dropdown.filter = "/yes"
	got, _ = sendKey(got, tea.KeyEnter)
	assert.Zero(t, client.listCalls)
	assert.Zero(t, client.rowListCalls)
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
			assert.Contains(t, view, "Make a Change ver. 0.1")
		})
	}
}

func sendCommand(m Model, command string) (Model, tea.Cmd) {
	updated, cmd := m.executeCommand(command)
	return updated.(Model), cmd
}

func sendRune(m Model, r rune) (Model, tea.Cmd) {
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	return updated.(Model), cmd
}

func sendKey(m Model, key tea.KeyType) (Model, tea.Cmd) {
	return sendKeyMsg(m, tea.KeyMsg{Type: key})
}

func sendKeyMsg(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	updated, cmd := m.Update(msg)
	return updated.(Model), cmd
}

func applyMsg(m Model, msg tea.Msg) Model {
	updated, _ := m.Update(msg)
	return updated.(Model)
}

func applyCommand(m Model, cmd tea.Cmd) Model {
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, next := range batch {
			if next == nil {
				continue
			}
			m = applyMsg(m, next())
		}
		return m
	}
	return applyMsg(m, msg)
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
