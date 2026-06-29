package app

import (
	"strings"

	"mch/internal/dto"
	"mch/internal/navigation"
	"mch/internal/projects"

	tea "github.com/charmbracelet/bubbletea"
)

// Init starts any initial asynchronous command required by the model.
func (m Model) Init() tea.Cmd {
	if m.needsProjectSelection() {
		return func() tea.Msg {
			return startupProjectSelectionMsg{}
		}
	}
	return nil
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
	case "up":
		if m.state == ProjectsListState {
			m.projectList = m.projectList.MoveSelection(-1)
			return m, nil
		}
	case "down":
		if m.state == ProjectsListState {
			m.projectList = m.projectList.MoveSelection(1)
			return m, nil
		}
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
		m.state = navigation.CancelTarget(m.state)
		m.status = "cancel"
		return m, nil
	default:
		if target, ok := navigation.ReturnTargets()[m.state]; ok {
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
		m.openDropdown(ListSelectionDropDownState, dropdownList, m.state, ChangeDetailsState, "Changes", []dto.Option{{ID: "change-1", Label: "Example Change"}}, false)
	case ChangeDetailsState:
		m.openDropdown(ListSelectionDropDownState, dropdownList, m.state, RequirementDetailsState, "Requirements", []dto.Option{{ID: "requirement-1", Label: "Example Requirement"}}, false)
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
		m.state = ProjectsListState
		m.projectList = projects.StartLoading()
		m.status = string(ProjectsListState)
		return m, projectListCommand(m.client)
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
		m.state = navigation.ReturnTargets()[source]
		m.status = "return"
	case "/new-change", "/new-requirement", "/new-epic", "/new-project":
		m.state = navigation.CreateTarget(source)
	case "/edit":
		m.state = navigation.UpdateTarget(source)
	case "/save":
		m.state = navigation.SaveTarget(source)
		m.status = "save"
	case "/cancel":
		m.state = navigation.CancelTarget(source)
		m.status = "cancel"
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
