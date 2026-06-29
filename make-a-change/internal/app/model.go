package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"mch/internal/api"
	"mch/internal/styles"
)

const defaultBackendURL = "http://localhost:8080"
const noProjectsToSelectError = "No projects to select from. Please create new project and select it on Main Screen."

type State string

const (
	MainState                     State = "MainState"
	ChangesListState              State = "ChangesListState"
	ChangeDetailsState            State = "ChangeDetailsState"
	RequirementDetailsState       State = "RequirementDetailsState"
	ChangeCreateState             State = "ChangeCreateState"
	ChangeUpdateState             State = "ChangeUpdateState"
	RequirementCreateState        State = "RequirementCreateState"
	RequirementUpdateState        State = "RequirementUpdateState"
	EpicsListState                State = "EpicsListState"
	EpicDetailsState              State = "EpicDetailsState"
	EpicCreateState               State = "EpicCreateState"
	EpicUpdateState               State = "EpicUpdateState"
	ProjectsListState             State = "ProjectsListState"
	ProjectDetailsState           State = "ProjectDetailsState"
	ProjectCreateState            State = "ProjectCreateState"
	ProjectUpdateState            State = "ProjectUpdateState"
	MainHelpState                 State = "MainHelpState"
	ChangesHelpState              State = "ChangesHelpState"
	EpicsHelpState                State = "EpicsHelpState"
	ProjectsHelpState             State = "ProjectsHelpState"
	FindInputState                State = "FindInput"
	CommandDropDownState          State = "CommandDropDown"
	ListSelectionDropDownState    State = "ListSelectionDropDown"
	SelectProjectDropDown         State = "SelectProjectDropDown"
	SelectPhaseDropDown           State = "SelectPhaseDropDown"
	SelectEpicDropDown            State = "SelectEpicDropDown"
	SelectTypesDropDown           State = "SelectTypesDropDown"
	ChangeDeleteConfirmation      State = "ChangeDeleteConfirmation"
	RequirementDeleteConfirmation State = "RequirementDeleteConfirmation"
	EpicDeleteConfirmation        State = "EpicDeleteConfirmation"
	ProjectDeleteConfirmation     State = "ProjectDeleteConfirmation"
	DoneState                     State = "DoneState"
)

type dropdownKind string

const (
	dropdownCommand dropdownKind = "command"
	dropdownList    dropdownKind = "list"
	dropdownSelect  dropdownKind = "select"
	dropdownConfirm dropdownKind = "confirm"
)

type selectorSource string

const (
	selectorProjects selectorSource = "projects"
	selectorPhases   selectorSource = "phases"
	selectorEpics    selectorSource = "epics"
	selectorTypes    selectorSource = "types"
)

type filterField string

const (
	filterPhase filterField = "phase"
	filterEpic  filterField = "epic"
	filterType  filterField = "type"
)

type changesFilters struct {
	phase api.Option
	epic  api.Option
	typ   api.Option
}

type dropdownModel struct {
	kind        dropdownKind
	state       State
	previous    State
	onSelect    State
	source      selectorSource
	filterField filterField
	label       string
	options     []api.Option
	filter      string
	highlighted int
	loading     bool
}

type selectorLoadedMsg struct {
	source  selectorSource
	options []api.Option
	err     error
}

type startupProjectSelectionMsg struct{}

type Model struct {
	input          textinput.Model
	state          State
	previousState  State
	width          int
	quitting       bool
	err            string
	status         string
	helpQuery      string
	changesFilters changesFilters
	currentProject api.Option
	client         api.Client
	appConfig      appConfig
	configPath     string
	dropdown       dropdownModel
}

func NewModel() Model {
	configPath := resolveConfigPath(defaultConfigPath)
	cfg, err := loadAppConfig(configPath)
	m := newModelWithConfig(api.NewHTTPClient(cfg.BackendURL), cfg, configPath)
	if err != nil {
		m.err = err.Error()
	}
	return m
}

func NewModelWithClient(client api.Client) Model {
	return newModelWithConfig(client, appConfig{BackendURL: defaultBackendURL}, "")
}

func newModelWithConfig(client api.Client, cfg appConfig, configPath string) Model {
	input := textinput.New()
	input.Placeholder = "Type / for commands"
	input.Prompt = "> "
	input.Focus()
	input.CharLimit = 240
	input.Width = 0
	input.PromptStyle = styles.Default.InputBand.Copy().Foreground(lipgloss.Color("183"))
	input.TextStyle = styles.Default.InputBand.Copy().Foreground(lipgloss.Color("15"))
	input.PlaceholderStyle = styles.Default.InputBand.Copy().Foreground(lipgloss.Color("0"))
	input.Cursor.Style = styles.Default.InputBand.Copy().Foreground(lipgloss.Color("15"))
	input.Cursor.TextStyle = input.TextStyle

	currentProject := api.Option{}
	if cfg.ProjectID > 0 {
		currentProject = api.Option{
			ID:    strconv.Itoa(cfg.ProjectID),
			Label: fmt.Sprintf("Project #%d", cfg.ProjectID),
		}
	}

	return Model{
		input:          input,
		state:          MainState,
		width:          80,
		currentProject: currentProject,
		client:         client,
		appConfig:      cfg,
		configPath:     configPath,
		status:         "MainState",
	}
}

func (m Model) Init() tea.Cmd {
	if m.needsProjectSelection() {
		return func() tea.Msg {
			return startupProjectSelectionMsg{}
		}
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case startupProjectSelectionMsg:
		if !m.needsProjectSelection() {
			return m, nil
		}
		return m.beginSelector(SelectProjectDropDown)
	case selectorLoadedMsg:
		if m.dropdown.source != msg.source {
			return m, nil
		}
		m.dropdown.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			m.dropdown.options = nil
			return m, nil
		}
		var options []api.Option
		if m.dropdown.filterField != "" {
			options = filterOptions(msg.options)
		} else {
			options = msg.options
		}
		m.dropdown.options = options
		m.dropdown.highlighted = 0
		if len(options) == 0 {
			if m.dropdown.source == selectorProjects && m.needsProjectSelection() {
				m.state = MainState
				m.dropdown = dropdownModel{}
				m.err = noProjectsToSelectError
				return m, nil
			}
			m.err = "no options available"
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	m.err = ""

	if m.isDropdownState() {
		return m.handleDropdownKey(key, msg)
	}
	if m.state == FindInputState {
		return m.handleFindKey(key, msg)
	}

	switch key {
	case "ctrl+c":
		m.state = DoneState
		m.quitting = true
		return m, tea.Quit
	case "esc":
		return m.handleEsc()
	case "/":
		m.openCommandDropdown()
		return m, nil
	case "enter":
		text := strings.TrimSpace(m.input.Value())
		m.input.SetValue("")
		if text == "" {
			return m.handleListSelection()
		}
		if strings.HasPrefix(text, "/") {
			return m.executeCommand(text)
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) handleDropdownKey(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		return m.cancelDropdown()
	case "up":
		m.moveHighlight(-1)
		return m, nil
	case "down":
		m.moveHighlight(1)
		return m, nil
	case "backspace":
		if len(m.dropdown.filter) > 0 {
			m.dropdown.filter = m.dropdown.filter[:len(m.dropdown.filter)-1]
			m.dropdown.highlighted = 0
		}
		return m, nil
	case "enter":
		if m.dropdown.loading {
			return m, nil
		}
		return m.confirmDropdown()
	}
	if len(msg.Runes) > 0 {
		m.dropdown.filter += string(msg.Runes)
		m.dropdown.highlighted = 0
	}
	return m, nil
}

func (m Model) handleFindKey(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.state = m.previousState
		m.input.SetValue("")
		m.status = "cancel"
		return m, nil
	case "enter":
		query := strings.TrimSpace(m.input.Value())
		m.input.SetValue("")
		m.state = m.previousState
		if query == "" {
			m.err = "find text is required"
			return m, nil
		}
		m.helpQuery = query
		m.status = "highlight " + query
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) handleEsc() (tea.Model, tea.Cmd) {
	switch m.state {
	case MainState:
		m.state = DoneState
		m.quitting = true
		return m, tea.Quit
	case ChangeCreateState, ChangeUpdateState, RequirementCreateState, RequirementUpdateState,
		EpicCreateState, EpicUpdateState, ProjectCreateState, ProjectUpdateState:
		m.state = m.cancelTarget(m.state)
		m.status = "cancel"
		return m, nil
	default:
		if target, ok := returnTargets[m.state]; ok {
			m.state = target
			m.status = "return"
			return m, nil
		}
		m.err = "cannot cancel from this state"
		return m, nil
	}
}

func (m Model) handleListSelection() (tea.Model, tea.Cmd) {
	switch m.state {
	case ChangesListState:
		m.openDropdown(ListSelectionDropDownState, dropdownList, m.state, ChangeDetailsState, "Changes", []api.Option{{ID: "change-1", Label: "Example Change"}}, false)
	case ChangeDetailsState:
		m.openDropdown(ListSelectionDropDownState, dropdownList, m.state, RequirementDetailsState, "Requirements", []api.Option{{ID: "requirement-1", Label: "Example Requirement"}}, false)
	case EpicsListState:
		m.openDropdown(ListSelectionDropDownState, dropdownList, m.state, EpicDetailsState, "Epics", []api.Option{{ID: "epic-1", Label: "Example Epic"}}, false)
	case ProjectsListState:
		m.openDropdown(ListSelectionDropDownState, dropdownList, m.state, ProjectDetailsState, "Projects", []api.Option{{ID: "project-1", Label: "Example Project"}}, false)
	default:
		m.err = "nothing selectable in current state"
	}
	return m, nil
}

func (m *Model) openCommandDropdown() {
	options := commandOptions(m.state)
	m.previousState = m.state
	m.dropdown = dropdownModel{
		kind:     dropdownCommand,
		state:    CommandDropDownState,
		previous: m.state,
		onSelect: m.state,
		label:    "Commands",
		options:  options,
	}
	m.status = string(CommandDropDownState)
}

func (m *Model) openDropdown(state State, kind dropdownKind, previous State, onSelect State, label string, options []api.Option, loading bool) {
	m.previousState = previous
	m.state = state
	m.dropdown = dropdownModel{
		kind:     kind,
		state:    state,
		previous: previous,
		onSelect: onSelect,
		label:    label,
		options:  options,
		loading:  loading,
	}
	m.status = string(state)
}

func (m *Model) openSelectorDropdown(state State, previous State, onSelect State, label string, source selectorSource) {
	m.previousState = previous
	m.state = state
	m.dropdown = dropdownModel{
		kind:     dropdownSelect,
		state:    state,
		previous: previous,
		onSelect: onSelect,
		source:   source,
		label:    label,
		loading:  true,
	}
	m.status = label
}

func (m *Model) openFilterDropdown(label string, source selectorSource, field filterField) {
	m.previousState = ChangesListState
	m.state = ChangesListState
	m.dropdown = dropdownModel{
		kind:        dropdownSelect,
		previous:    ChangesListState,
		onSelect:    ChangesListState,
		source:      source,
		filterField: field,
		label:       label,
		loading:     true,
	}
	m.status = label
}

func (m Model) cancelDropdown() (tea.Model, tea.Cmd) {
	m.state = m.dropdown.previous
	m.status = "cancel"
	m.dropdown = dropdownModel{}
	return m, nil
}

func (m Model) confirmDropdown() (tea.Model, tea.Cmd) {
	if m.dropdown.kind == dropdownConfirm {
		selected := m.selectedOption()
		if selected.Label == "" {
			m.err = "confirmation requires /yes or /cancel"
			return m, nil
		}
		switch selected.ID {
		case "/yes":
			m.state = m.dropdown.onSelect
			m.status = "confirmed"
			m.dropdown = dropdownModel{}
			return m, nil
		case "/cancel":
			return m.cancelDropdown()
		default:
			m.err = "confirmation requires /yes or /cancel"
			return m, nil
		}
	}

	if m.dropdown.kind == dropdownCommand {
		selected := m.selectedOption()
		if selected.ID == "" {
			m.err = "unknown command"
			return m, nil
		}
		return m.executeCommandFrom(m.dropdown.previous, selected.ID)
	}

	selected := m.selectedOption()
	if selected.Label == "" {
		m.err = "no matching option"
		return m, nil
	}
	if m.dropdown.filterField != "" {
		if selected.ID == "/clear" {
			m.clearChangesFilter(m.dropdown.filterField)
			m.state = m.dropdown.onSelect
			m.status = "cleared " + string(m.dropdown.filterField) + " filter"
			m.dropdown = dropdownModel{}
			return m, nil
		}
		m.setChangesFilter(m.dropdown.filterField, selected)
	}
	if m.state == SelectProjectDropDown {
		m.currentProject = selected
		if err := m.saveCurrentProject(selected); err != nil {
			m.err = err.Error()
		}
	}
	m.state = m.dropdown.onSelect
	m.status = "selected " + selected.Label
	m.dropdown = dropdownModel{}
	return m, nil
}

func (m Model) executeCommand(command string) (tea.Model, tea.Cmd) {
	return m.executeCommandFrom(m.state, command)
}

func (m Model) executeCommandFrom(source State, command string) (tea.Model, tea.Cmd) {
	m.state = source
	m.dropdown = dropdownModel{}
	if command != "/quit" && !commandAllowed(source, command) {
		m.err = "unknown command: " + command
		return m, nil
	}

	switch command {
	case "/quit":
		if source != MainState {
			m.err = "/quit is only available from MainState"
			return m, nil
		}
		m.state = DoneState
		m.quitting = true
		return m, tea.Quit
	case "/changes":
		m.state = ChangesListState
	case "/epics":
		m.state = EpicsListState
	case "/projects":
		m.state = ProjectsListState
	case "/select-project":
		return m.beginSelector(SelectProjectDropDown)
	case "/help":
		m.state = helpStateFor(source)
	case "/find":
		m.previousState = source
		m.state = FindInputState
		m.input.SetValue("")
	case "/find-filter":
		m.status = "find filter"
	case "/clear-filters":
		m.changesFilters = changesFilters{}
		m.status = "filters cleared"
	case "/return":
		m.state = returnTargets[source]
		m.status = "return"
	case "/new-change", "/new-requirement", "/new-epic", "/new-project":
		m.state = createStateFor(source)
	case "/edit":
		m.state = updateStateFor(source)
	case "/save":
		m.state = saveTarget(source)
		m.status = "save"
	case "/cancel":
		m.state = m.cancelTarget(source)
		m.status = "cancel"
	case "/delete":
		m.openConfirmation(deleteConfirmationState(source), source, deleteReturnState(source))
	case "/phase":
		return m.beginSelector(SelectPhaseDropDown)
	case "/epic":
		return m.beginSelector(SelectEpicDropDown)
	case "/types":
		return m.beginSelector(SelectTypesDropDown)
	case "/phase-filter":
		return m.beginFilter("Phase Filter", selectorPhases, filterPhase)
	case "/epic-filter":
		return m.beginFilter("Epic Filter", selectorEpics, filterEpic)
	case "/type-filter":
		return m.beginFilter("Type Filter", selectorTypes, filterType)
	}
	if m.state == "" {
		m.state = source
		m.err = "unknown command: " + command
	}
	return m, nil
}

func (m Model) beginSelector(state State) (tea.Model, tea.Cmd) {
	previous := m.state
	if previous == CommandDropDownState {
		previous = m.dropdown.previous
	}
	onSelect := previous
	if state == SelectProjectDropDown {
		onSelect = MainState
	}
	source := selectorSourceForState(state)
	m.openSelectorDropdown(state, previous, onSelect, string(state), source)
	return m, selectorCommand(m.client, source, m.currentProject.ID)
}

func (m Model) beginFilter(label string, source selectorSource, field filterField) (tea.Model, tea.Cmd) {
	m.openFilterDropdown(label, source, field)
	return m, selectorCommand(m.client, source, m.currentProject.ID)
}

func (m Model) needsProjectSelection() bool {
	return m.currentProject.ID == "" && m.appConfig.ProjectID <= 0
}

func filterOptions(options []api.Option) []api.Option {
	filtered := make([]api.Option, 0, len(options)+1)
	filtered = append(filtered, options...)
	filtered = append(filtered, api.Option{ID: "/clear", Label: "/clear"})
	return filtered
}

func (m *Model) setChangesFilter(field filterField, option api.Option) {
	switch field {
	case filterPhase:
		m.changesFilters.phase = option
	case filterEpic:
		m.changesFilters.epic = option
	case filterType:
		m.changesFilters.typ = option
	}
}

func (m *Model) clearChangesFilter(field filterField) {
	switch field {
	case filterPhase:
		m.changesFilters.phase = api.Option{}
	case filterEpic:
		m.changesFilters.epic = api.Option{}
	case filterType:
		m.changesFilters.typ = api.Option{}
	}
}

func (m *Model) saveCurrentProject(project api.Option) error {
	if m.configPath == "" {
		return nil
	}
	projectID, err := strconv.Atoi(project.ID)
	if err != nil {
		return fmt.Errorf("failed to save project_id: current project ID must be numeric")
	}
	m.appConfig.ProjectID = projectID
	if m.appConfig.BackendURL == "" {
		m.appConfig.BackendURL = defaultBackendURL
	}
	if err := saveAppConfig(m.configPath, m.appConfig); err != nil {
		return fmt.Errorf("failed to save project_id: %w", err)
	}
	return nil
}

func selectorSourceForState(state State) selectorSource {
	switch state {
	case SelectProjectDropDown:
		return selectorProjects
	case SelectPhaseDropDown:
		return selectorPhases
	case SelectEpicDropDown:
		return selectorEpics
	case SelectTypesDropDown:
		return selectorTypes
	default:
		return ""
	}
}

func selectorCommand(client api.Client, source selectorSource, projectID string) tea.Cmd {
	return func() tea.Msg {
		var (
			options []api.Option
			err     error
		)
		switch source {
		case selectorProjects:
			options, err = client.ListProjects()
		case selectorEpics:
			options, err = client.ListEpics(projectID)
		case selectorPhases:
			options, err = client.ListPhases()
		case selectorTypes:
			options, err = client.ListTypes()
		}
		return selectorLoadedMsg{source: source, options: options, err: err}
	}
}

func (m *Model) openConfirmation(state, previous, onYes State) {
	m.openDropdown(state, dropdownConfirm, previous, onYes, "Confirm", []api.Option{
		{ID: "/yes", Label: "/yes"},
		{ID: "/cancel", Label: "/cancel"},
	}, false)
}

func (m Model) View() string {
	width := terminalWidth(m.width)
	lines := []string{
		styles.Default.Title.Render("mch"),
		styles.Default.Muted.Render("version " + Version),
		"",
		styles.Default.Foreground.Render(screenTitle(m.state)),
	}
	if m.currentProject.Label != "" {
		lines = append(lines, styles.Default.Muted.Render("Project: "+m.currentProject.Label))
	}
	if m.state == FindInputState {
		lines = append(lines, "", m.inputBand(width))
	} else if m.hasDropdown() {
		lines = append(lines, "", m.dropdownView(width))
	} else {
		lines = append(lines, "", m.inputBand(width))
	}
	if m.err != "" {
		lines = append(lines, styles.Default.Error.Render("Error: "+m.err))
	}
	if m.helpQuery != "" {
		lines = append(lines, styles.Default.Success.Render("Highlight: "+m.helpQuery))
	}
	lines = append(lines, styles.Default.Footer.Width(width).Render(m.footerText()))
	if m.quitting {
		lines = append(lines, styles.Default.Success.Render("done"))
	}
	return styles.Default.Surface.Width(width).Render(strings.Join(lines, "\n"))
}

func (m Model) inputBand(width int) string {
	width = terminalWidth(width)
	content := m.inputLine(width)
	blank := strings.Repeat(" ", width)
	return strings.Join([]string{
		styles.Default.InputBand.Render(blank),
		styles.Default.InputBand.Render(content),
		styles.Default.InputBand.Render(blank),
	}, "\n")
}

func (m Model) inputLine(width int) string {
	content := m.input.View()
	if visible := lipgloss.Width(content); visible < width {
		content += styles.Default.InputBand.Render(strings.Repeat(" ", width-visible))
	}
	return content
}

func (m Model) dropdownView(width int) string {
	if m.dropdown.loading {
		return styles.Default.InputBand.Width(width).Render(m.dropdown.label + ": loading")
	}
	options := m.filteredOptions()
	if len(options) == 0 {
		return styles.Default.InputBand.Width(width).Render(m.dropdown.label + ": no options")
	}
	lines := []string{m.dropdown.label + " " + m.dropdown.filter}
	for i, option := range options {
		line := m.dropdownLine(option.Label)
		if i == m.dropdown.highlighted {
			line = styles.Default.Selection.Render(line)
		}
		lines = append(lines, line)
	}
	rendered := styles.Default.InputBand.Width(width).Render(strings.Join(lines, "\n"))
	if m.dropdown.kind == dropdownCommand {
		rendered += "\n" + styles.Default.Background.Width(width).Render("")
	}
	return rendered
}

func (m Model) dropdownLine(label string) string {
	if m.dropdown.filterField != "" && label != "/clear" {
		label = "-" + strings.TrimPrefix(label, "-")
	}
	return "    " + label
}

func (m Model) footerText() string {
	if m.status != "" {
		return fmt.Sprintf("/ commands  |  esc safe action  |  status %s", m.status)
	}
	return "/ commands  |  esc safe action"
}

func (m Model) isDropdownState() bool {
	if m.hasDropdown() {
		return true
	}
	switch m.state {
	case CommandDropDownState, ListSelectionDropDownState, SelectProjectDropDown, SelectPhaseDropDown,
		SelectEpicDropDown, SelectTypesDropDown, ChangeDeleteConfirmation, RequirementDeleteConfirmation,
		EpicDeleteConfirmation, ProjectDeleteConfirmation:
		return true
	default:
		return false
	}
}

func (m Model) hasDropdown() bool {
	return m.dropdown.kind != ""
}

func (m Model) selectedOption() api.Option {
	options := m.filteredOptions()
	if len(options) == 0 {
		return api.Option{}
	}
	if m.dropdown.highlighted >= len(options) {
		m.dropdown.highlighted = len(options) - 1
	}
	return options[m.dropdown.highlighted]
}

func (m *Model) moveHighlight(delta int) {
	options := m.filteredOptions()
	if len(options) == 0 {
		m.dropdown.highlighted = 0
		return
	}
	m.dropdown.highlighted = (m.dropdown.highlighted + delta + len(options)) % len(options)
}

func (m Model) filteredOptions() []api.Option {
	filter := strings.ToLower(strings.TrimSpace(m.dropdown.filter))
	if filter == "" {
		return m.dropdown.options
	}
	var options []api.Option
	for _, option := range m.dropdown.options {
		label := strings.ToLower(option.Label)
		id := strings.ToLower(option.ID)
		if !strings.HasPrefix(filter, "/") {
			label = strings.TrimPrefix(label, "/")
			id = strings.TrimPrefix(id, "/")
		}
		if strings.Contains(label, filter) || strings.Contains(id, filter) {
			options = append(options, option)
		}
	}
	return options
}

func screenTitle(state State) string {
	titles := map[State]string{
		MainState:                     "MainScreen - Title: Main",
		ChangesListState:              "ChangesListScreen - Title: Changes List",
		ChangeDetailsState:            "ChangeDetailsScreen - Title: Change Details",
		RequirementDetailsState:       "RequirementDetailsScreen - Title: Requirement Details",
		ChangeCreateState:             "ChangeCreateScreen - Title: New Change",
		ChangeUpdateState:             "ChangeUpdateScreen - Title: Edit Change",
		RequirementCreateState:        "RequirementCreateScreen - Title: New Requirement",
		RequirementUpdateState:        "RequirementUpdateScreen - Title: Edit Requirement",
		EpicsListState:                "EpicsListScreen - Title: Epics List",
		EpicDetailsState:              "EpicDetailsScreen - Title: Epic Details",
		EpicCreateState:               "EpicCreateScreen - Title: New Epic",
		EpicUpdateState:               "EpicUpdateScreen - Title: Edit Epic",
		ProjectsListState:             "ProjectsListScreen - Title: Projects List",
		ProjectDetailsState:           "ProjectDetailsScreen - Title: Project Details",
		ProjectCreateState:            "ProjectCreateScreen - Title: New Project",
		ProjectUpdateState:            "ProjectUpdateScreen - Title: Edit Project",
		MainHelpState:                 "MainHelpScreen - Title: Main Help",
		ChangesHelpState:              "ChangesHelpScreen - Title: Changes Help",
		EpicsHelpState:                "EpicsHelpScreen - Title: Epics Help",
		ProjectsHelpState:             "ProjectsHelpScreen - Title: Projects Help",
		FindInputState:                "FindInputScreen - Title: Find",
		CommandDropDownState:          "CommandDropDownScreen - Title: Commands",
		ListSelectionDropDownState:    "ListSelectionDropDownScreen - Title: Select Item",
		SelectProjectDropDown:         "SelectProjectDropDownScreen - Title: Select Project",
		SelectPhaseDropDown:           "SelectPhaseDropDownScreen - Title: Select Phase",
		SelectEpicDropDown:            "SelectEpicDropDownScreen - Title: Select Epic",
		SelectTypesDropDown:           "SelectTypesDropDownScreen - Title: Select Types",
		ChangeDeleteConfirmation:      "ChangeDeleteConfirmationScreen - Title: Confirm Delete",
		RequirementDeleteConfirmation: "RequirementDeleteConfirmationScreen - Title: Confirm Delete",
		EpicDeleteConfirmation:        "EpicDeleteConfirmationScreen - Title: Confirm Delete",
		ProjectDeleteConfirmation:     "ProjectDeleteConfirmationScreen - Title: Confirm Delete",
		DoneState:                     "DoneScreen - Title: Done",
	}
	if title, ok := titles[state]; ok {
		return title
	}
	return "UnknownScreen - Title: Unknown"
}

var returnTargets = map[State]State{
	ChangesListState:        MainState,
	ChangeDetailsState:      ChangesListState,
	RequirementDetailsState: ChangeDetailsState,
	EpicsListState:          MainState,
	EpicDetailsState:        EpicsListState,
	ProjectsListState:       MainState,
	ProjectDetailsState:     ProjectsListState,
	MainHelpState:           MainState,
	ChangesHelpState:        ChangesListState,
	EpicsHelpState:          EpicsListState,
	ProjectsHelpState:       ProjectsListState,
}

var commandsByState = map[State][]string{
	MainState:               {"/new-change", "/changes", "/epics", "/projects", "/select-project", "/help", "/quit"},
	ChangesListState:        {"/new-change", "/phase-filter", "/epic-filter", "/type-filter", "/find-filter", "/clear-filters", "/help", "/return"},
	ChangeDetailsState:      {"/new-requirement", "/phase", "/epic", "/types", "/edit", "/delete", "/return"},
	RequirementDetailsState: {"/new-requirement", "/edit", "/delete", "/save", "/cancel", "/return"},
	ChangeCreateState:       {"/save", "/cancel"},
	ChangeUpdateState:       {"/save", "/cancel"},
	RequirementCreateState:  {"/save", "/cancel"},
	RequirementUpdateState:  {"/save", "/cancel"},
	EpicsListState:          {"/new-epic", "/help", "/find", "/return"},
	EpicDetailsState:        {"/edit", "/delete", "/help", "/find", "/return"},
	EpicCreateState:         {"/save", "/cancel"},
	EpicUpdateState:         {"/save", "/cancel"},
	ProjectsListState:       {"/new-project", "/help", "/find", "/return"},
	ProjectDetailsState:     {"/edit", "/delete", "/help", "/find", "/return"},
	ProjectCreateState:      {"/save", "/cancel"},
	ProjectUpdateState:      {"/save", "/cancel"},
	MainHelpState:           {"/find", "/return"},
	ChangesHelpState:        {"/find", "/return"},
	EpicsHelpState:          {"/find", "/return"},
	ProjectsHelpState:       {"/find", "/return"},
}

func commandOptions(state State) []api.Option {
	commands := commandsByState[state]
	options := make([]api.Option, 0, len(commands))
	for _, command := range commands {
		options = append(options, api.Option{ID: command, Label: command})
	}
	return options
}

func commandAllowed(state State, command string) bool {
	for _, allowed := range commandsByState[state] {
		if allowed == command {
			return true
		}
	}
	return false
}

func helpStateFor(state State) State {
	switch state {
	case MainState:
		return MainHelpState
	case ChangesListState, ChangeDetailsState, RequirementDetailsState:
		return ChangesHelpState
	case EpicsListState, EpicDetailsState:
		return EpicsHelpState
	case ProjectsListState, ProjectDetailsState:
		return ProjectsHelpState
	default:
		return state
	}
}

func createStateFor(state State) State {
	switch state {
	case MainState:
		return ChangeCreateState
	case ChangesListState:
		return ChangeCreateState
	case ChangeDetailsState, RequirementDetailsState:
		return RequirementCreateState
	case EpicsListState:
		return EpicCreateState
	case ProjectsListState:
		return ProjectCreateState
	default:
		return state
	}
}

func updateStateFor(state State) State {
	switch state {
	case ChangeDetailsState:
		return ChangeUpdateState
	case RequirementDetailsState:
		return RequirementUpdateState
	case EpicDetailsState:
		return EpicUpdateState
	case ProjectDetailsState:
		return ProjectUpdateState
	default:
		return state
	}
}

func saveTarget(state State) State {
	switch state {
	case ChangeCreateState, ChangeUpdateState:
		return ChangeDetailsState
	case RequirementCreateState, RequirementUpdateState:
		return RequirementDetailsState
	case EpicCreateState, EpicUpdateState:
		return EpicDetailsState
	case ProjectCreateState, ProjectUpdateState:
		return ProjectDetailsState
	default:
		return state
	}
}

func (m Model) cancelTarget(state State) State {
	switch state {
	case ChangeCreateState:
		return ChangesListState
	case ChangeUpdateState:
		return ChangeDetailsState
	case RequirementCreateState, RequirementUpdateState:
		return RequirementDetailsState
	case EpicCreateState:
		return EpicsListState
	case EpicUpdateState:
		return EpicDetailsState
	case ProjectCreateState:
		return ProjectsListState
	case ProjectUpdateState:
		return ProjectDetailsState
	default:
		return state
	}
}

func deleteConfirmationState(state State) State {
	switch state {
	case ChangeDetailsState:
		return ChangeDeleteConfirmation
	case RequirementDetailsState:
		return RequirementDeleteConfirmation
	case EpicDetailsState:
		return EpicDeleteConfirmation
	case ProjectDetailsState:
		return ProjectDeleteConfirmation
	default:
		return state
	}
}

func deleteReturnState(state State) State {
	switch state {
	case ChangeDetailsState:
		return ChangesListState
	case RequirementDetailsState:
		return ChangeDetailsState
	case EpicDetailsState:
		return EpicsListState
	case ProjectDetailsState:
		return ProjectsListState
	default:
		return state
	}
}

func terminalWidth(width int) int {
	if width < 20 {
		return 100
	}
	return width
}

var _ tea.Model = Model{}
