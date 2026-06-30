package app

import (
	"strconv"
	"strings"

	"mch/internal/dto"
	"mch/internal/navigation"
	"mch/internal/projects"

	tea "github.com/charmbracelet/bubbletea"
)

// Init starts any initial asynchronous command required by the model.
func (m Model) Init() tea.Cmd {
	clear := tea.ClearScreen
	if m.needsProjectSelection() {
		selectProject := func() tea.Msg {
			return startupProjectSelectionMsg{}
		}
		return tea.Batch(clear, selectProject)
	}
	if m.appConfig.ProjectID > 0 {
		return tea.Batch(clear, currentProjectCommand(m.client, m.appConfig.ProjectID))
	}
	return clear
}

// Update applies Bubble Tea messages to the root model.
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
		var options []dto.Option
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
	case projectListLoadedMsg:
		if m.state != ProjectsListState {
			return m, nil
		}
		if msg.err != nil {
			m.projectList = m.projectList.WithError()
			m.err = msg.err.Error()
			return m, nil
		}
		m.projectList = m.projectList.WithRows(msg.projects)
		if len(m.projectList.Rows) == 0 {
			m.status = "no projects"
		}
		return m, nil
	case projectSavedMsg:
		if m.state != msg.source {
			return m, nil
		}
		if msg.err != nil {
			m.err = msg.err.Error()
			m.status = "save failed"
			return m, nil
		}
		m.projectList.Detail = msg.project
		m.state = ProjectDetailsState
		m.status = "save"
		m = m.setPromptValue("")
		return m, nil
	case projectLoadedMsg:
		if m.state != ProjectDetailsState {
			return m, nil
		}
		if currentID, err := projectNumericID(m.projectList.Detail); err == nil && currentID != msg.id {
			return m, nil
		}
		if msg.err != nil {
			m.err = msg.err.Error()
			m.status = "load failed"
			return m, nil
		}
		m.projectList.Detail = msg.project
		m.status = "loaded project"
		return m, nil
	case currentProjectLoadedMsg:
		currentID, err := strconv.Atoi(m.currentProject.ID)
		if err != nil || currentID != msg.id {
			return m, nil
		}
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.currentProject = dto.Option{ID: m.currentProject.ID, Label: strings.TrimSpace(msg.project.Name)}
		return m, nil
	case editorFinishedMsg:
		if m.state != msg.source {
			return m, nil
		}
		if msg.err != nil {
			m.err = msg.err.Error()
			m.status = "editor failed"
			return m, tea.ClearScreen
		}
		next, cmd := m.submitPromptValue(msg.content)
		m = next.(Model)
		if cmd == nil {
			return m, tea.ClearScreen
		}
		return m, tea.Sequence(tea.ClearScreen, cmd)
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

	if isPromptNewlineKey(msg) {
		return m.insertPromptNewline(), nil
	}
	var handled bool
	m, handled = m.handlePendingShiftEnter(msg)
	if handled {
		return m, nil
	}
	if isShiftEnterPrefix(msg) {
		m.pendingAltO = true
		return m, nil
	}

	switch key {
	case "ctrl+c":
		return m.handlePromptCancel()
	case "ctrl+e":
		return m.openPromptEditor(m.state)
	case "esc":
		return m.handleEsc()
	case "up":
		if m.state == ProjectsListState && m.input.Value() == "" {
			m.projectList = m.projectList.MoveSelection(-1)
			return m, nil
		}
	case "down":
		if m.state == ProjectsListState && m.input.Value() == "" {
			m.projectList = m.projectList.MoveSelection(1)
			return m, nil
		}
	case "/":
		if m.input.Value() == "" {
			m.openCommandDropdown()
			return m, nil
		}
	case "enter":
		return m.submitPrompt()
	}

	return m.updatePromptInput(msg)
}

func (m Model) handleFindKey(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isPromptNewlineKey(msg) {
		return m.insertPromptNewline(), nil
	}
	var handled bool
	m, handled = m.handlePendingShiftEnter(msg)
	if handled {
		return m, nil
	}
	if isShiftEnterPrefix(msg) {
		m.pendingAltO = true
		return m, nil
	}
	switch key {
	case "ctrl+c":
		if m.input.Value() != "" {
			m = m.setPromptValue("")
			m.status = "prompt cleared"
			return m, nil
		}
		return m.arrive(m.previousState, "cancel")
	case "ctrl+e":
		return m.openPromptEditor(m.state)
	case "esc":
		m = m.setPromptValue("")
		return m.arrive(m.previousState, "cancel")
	case "enter":
		return m.submitFindValue(m.input.Value())
	}
	return m.updatePromptInput(msg)
}

func isPromptNewlineKey(msg tea.KeyMsg) bool {
	key := msg.String()
	return key == "shift+enter" || (msg.Type == tea.KeyEnter && msg.Alt) || msg.Type == tea.KeyCtrlJ
}

func (m Model) submitPrompt() (tea.Model, tea.Cmd) {
	return m.submitPromptValue(m.input.Value())
}

func (m Model) submitPromptValue(value string) (tea.Model, tea.Cmd) {
	trimmed := strings.TrimSpace(value)
	if commandAllowed(m.state, "/save") {
		if m.state == ProjectCreateState {
			return m.saveProjectCreateValue(value)
		}
		if m.state == ProjectUpdateState {
			return m.saveProjectUpdateValue(value)
		}
		return m.executeCommandFrom(m.state, "/save")
	}
	if m.state == FindInputState {
		return m.submitFindValue(value)
	}
	m = m.setPromptValue("")
	if trimmed == "" {
		return m.handleListSelection()
	}
	if strings.HasPrefix(trimmed, "/") {
		return m.executeCommand(trimmed)
	}
	return m, nil
}

func (m Model) submitFindValue(value string) (tea.Model, tea.Cmd) {
	query := strings.TrimSpace(value)
	m = m.setPromptValue("")
	if query == "" {
		m.state = m.previousState
		m.err = "find text is required"
		return m, nil
	}
	m.helpQuery = query
	return m.arrive(m.previousState, "highlight "+query)
}

func (m Model) handlePromptCancel() (tea.Model, tea.Cmd) {
	if m.input.Value() != "" {
		m = m.setPromptValue("")
		m.status = "prompt cleared"
		return m, nil
	}
	switch {
	case commandAllowed(m.state, "/cancel"):
		return m.executeCommandFrom(m.state, "/cancel")
	case commandAllowed(m.state, "/return"):
		return m.executeCommandFrom(m.state, "/return")
	case commandAllowed(m.state, "/quit"):
		return m.executeCommandFrom(m.state, "/quit")
	default:
		return m.handleEsc()
	}
}

func (m Model) handleEsc() (tea.Model, tea.Cmd) {
	switch m.state {
	case MainState:
		m.state = DoneState
		m.quitting = true
		return m, tea.Quit
	case ChangeCreateState, ChangeUpdateState, TestCaseCreateState, TestCaseUpdateState,
		EpicCreateState, EpicUpdateState, ProjectCreateState, ProjectUpdateState:
		return m.arrive(navigation.CancelTarget(m.state), "cancel")
	default:
		if target, ok := navigation.ReturnTargets()[m.state]; ok {
			return m.arrive(target, "return")
		}
		m.err = "cannot cancel from this state"
		return m, nil
	}
}

func (m Model) handleListSelection() (tea.Model, tea.Cmd) {
	switch m.state {
	case ChangesListState:
		m.openDropdown(ListSelectionDropDownState, dropdownList, m.state, ChangeDetailsState, "Changes", []dto.Option{{ID: "change-1", Label: "Example Change"}}, false)
	case ChangeDetailsState:
		m.openDropdown(ListSelectionDropDownState, dropdownList, m.state, TestCaseDetailsState, "Test Cases", []dto.Option{{ID: "test-case-1", Label: "Example Test Case"}}, false)
	case EpicsListState:
		m.openDropdown(ListSelectionDropDownState, dropdownList, m.state, EpicDetailsState, "Epics", []dto.Option{{ID: "epic-1", Label: "Example Epic"}}, false)
	case ProjectsListState:
		next, selected, ok := m.projectList.SelectDetail()
		m.projectList = next
		if !ok {
			m.err = projects.NoSelectableError
			return m, nil
		}
		m.state = ProjectDetailsState
		m.status = "selected " + projects.DisplayName(selected)
		id, err := projectNumericID(selected)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		return m, projectGetCommand(m.client, id)
	default:
		m.err = "nothing selectable in current state"
	}
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
		return m.arrive(ProjectsListState, string(ProjectsListState))
	case "/select-project":
		return m.beginSelector(SelectProjectDropDown)
	case "/help":
		m.state = helpStateFor(source)
	case "/find":
		m.previousState = source
		m.state = FindInputState
		m = m.setPromptValue("")
	case "/find-filter":
		m.status = "find filter"
	case "/clear-filters":
		m.changesFilters = changesFilters{}
		m.status = "filters cleared"
	case "/return":
		return m.arrive(navigation.ReturnTargets()[source], "return")
	case "/new-change", "/new-test-case", "/new-epic", "/new-project":
		m.state = navigation.CreateTarget(source)
		if m.state == ProjectCreateState {
			m = m.setPromptValue("")
			m.input.Placeholder = "Write a Name"
		} else {
			m.input.Placeholder = defaultInputPlaceholder
		}
	case "/edit":
		m.state = navigation.UpdateTarget(source)
		if m.state == ProjectUpdateState {
			m = m.setPromptValue(m.projectList.Detail.Name)
		}
		m.input.Placeholder = defaultInputPlaceholder
	case "/save":
		if source == ProjectCreateState {
			return m.saveProjectCreate()
		}
		if source == ProjectUpdateState {
			return m.saveProjectUpdate()
		}
		m.state = navigation.SaveTarget(source)
		m.status = "save"
	case "/editor":
		return m.openPromptEditor(source)
	case "/cancel":
		return m.arrive(navigation.CancelTarget(source), "cancel")
	case "/delete":
		m.openConfirmation(navigation.DeleteConfirmationState(source), source, navigation.DeleteReturnState(source))
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

func (m Model) arrive(state State, status string) (tea.Model, tea.Cmd) {
	m.state = state
	m.status = status
	if state != ProjectCreateState {
		m.input.Placeholder = defaultInputPlaceholder
	}
	switch state {
	case ProjectsListState:
		m.projectList = projects.StartLoading()
		return m, projectListCommand(m.client)
	case ProjectDetailsState:
		id, err := projectNumericID(m.projectList.Detail)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.status = "loading project"
		return m, projectGetCommand(m.client, id)
	default:
		return m, nil
	}
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
