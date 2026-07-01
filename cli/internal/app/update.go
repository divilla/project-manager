package app

import (
	"strconv"
	"strings"

	"mch/internal/changes"
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
		if m.dropdown.editField == detailEditEpic {
			options = append(options, dto.Option{ID: "@none", Label: "@none"})
		}
		m.dropdown.options = options
		m.dropdown.highlighted = m.dropdownCurrentValueIndex(options)
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
	case changeListLoadedMsg:
		if m.state != ChangesListState {
			return m, nil
		}
		if msg.err != nil {
			m.changeList = m.changeList.WithError()
			m.err = msg.err.Error()
			return m, nil
		}
		m.changeList = m.changeList.WithRows(msg.changes)
		if len(m.changeList.Rows) == 0 {
			m.status = "no changes"
		}
		return m, nil
	case changeLoadedMsg:
		if m.state != ChangeDetailsState {
			return m, nil
		}
		if currentID, err := changeNumericID(m.changeList.Detail); err == nil && currentID != msg.id {
			return m, nil
		}
		if msg.err != nil {
			m.err = msg.err.Error()
			m.status = "load failed"
			return m, nil
		}
		m.changeList = m.changeList.WithDetail(msg.change)
		m.status = "loaded change"
		return m, nil
	case changeSavedMsg:
		if m.state != msg.source {
			return m, nil
		}
		if msg.err != nil {
			m.err = msg.err.Error()
			m.status = "save failed"
			return m, nil
		}
		detailSelected := m.changeList.DetailSelected
		detailOffset := m.changeList.DetailOffset
		preserveDetailSelection := msg.source == ChangeDetailsState || m.detailEditField != ""
		m.changeList = m.changeList.WithDetail(msg.change)
		if preserveDetailSelection {
			m.changeList.DetailSelected = detailSelected
			m.changeList.DetailOffset = detailOffset
			m.changeList = m.changeList.ClampDetailSelection(m.changeTableRows(), terminalWidth(m.width))
		}
		m.state = ChangeDetailsState
		m.status = "save"
		if msg.reloadErr != nil {
			m.err = msg.reloadErr.Error()
			m.status = "load failed"
		}
		m.detailEditField = ""
		m = m.setPromptValue("")
		return m, nil
	case changeDeletedMsg:
		if msg.err != nil {
			m.state = ChangeDetailsState
			m.err = msg.err.Error()
			m.status = "delete failed"
			return m, nil
		}
		return m.arrive(msg.target, "deleted")
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
		if msg.source == ChangeCreateState || msg.source == ChangeUpdateState ||
			(msg.source == ChangeDetailsState && m.detailEditField != "") {
			m = m.setPromptValue(msg.content)
		}
		next, cmd := m.submitPromptValue(msg.content)
		m = next.(Model)
		if cmd == nil {
			return m, tea.ClearScreen
		}
		if msg.source == ChangeDetailsState && m.detailEditField != "" {
			return m, tea.Batch(tea.ClearScreen, cmd)
		}
		return m, tea.Sequence(tea.ClearScreen, cmd)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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

	if updated, cmd, ok := m.handleListNavigationKey(key, msg); ok {
		return updated, cmd
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
		if m.state == ChangeDetailsState && m.detailEditField != "" {
			return m.handlePromptCancel()
		}
		return m.handleEsc()
	case "up":
	case "down":
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

func (m Model) handleListNavigationKey(key string, msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	if m.input.Value() != "" {
		return m, nil, false
	}
	switch m.state {
	case ChangesListState:
		switch {
		case key == "up":
			m.changeList = m.changeList.MoveSelection(-1, m.changeFilters(), m.changeTableRows())
			return m, nil, true
		case key == "down":
			m.changeList = m.changeList.MoveSelection(1, m.changeFilters(), m.changeTableRows())
			return m, nil, true
		case key == "pgup":
			m.changeList = m.changeList.MoveSelection(-m.changeTableRows(), m.changeFilters(), m.changeTableRows())
			return m, nil, true
		case key == "pgdown":
			m.changeList = m.changeList.MoveSelection(m.changeTableRows(), m.changeFilters(), m.changeTableRows())
			return m, nil, true
		case key == "enter" || msg.Type == tea.KeyCtrlJ:
			updated, cmd := m.handleListSelection()
			return updated.(Model), cmd, true
		}
	case ChangeDetailsState:
		switch {
		case key == "up":
			m.changeList = m.changeList.MoveDetailSelection(-1, m.changeTableRows(), terminalWidth(m.width))
			return m, nil, true
		case key == "down":
			m.changeList = m.changeList.MoveDetailSelection(1, m.changeTableRows(), terminalWidth(m.width))
			return m, nil, true
		case key == "pgup":
			m.changeList = m.changeList.ScrollDetailViewport(-m.changeTableRows(), m.changeTableRows(), terminalWidth(m.width))
			return m, nil, true
		case key == "pgdown":
			m.changeList = m.changeList.ScrollDetailViewport(m.changeTableRows(), m.changeTableRows(), terminalWidth(m.width))
			return m, nil, true
		case key == "enter" || msg.Type == tea.KeyCtrlJ:
			updated, cmd := m.handleListSelection()
			return updated.(Model), cmd, true
		}
	case ProjectsListState:
		switch {
		case key == "up":
			m.projectList = m.projectList.MoveSelection(-1)
			return m, nil, true
		case key == "down":
			m.projectList = m.projectList.MoveSelection(1)
			return m, nil, true
		case key == "enter" || msg.Type == tea.KeyCtrlJ:
			updated, cmd := m.handleListSelection()
			return updated.(Model), cmd, true
		}
	}
	return m, nil, false
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
	if m.detailEditField == detailEditTitle && strings.HasPrefix(trimmed, "/") {
		return m.executeCommand(trimmed)
	}
	if m.detailEditField != "" && (m.state == ChangeDetailsState || m.state == ChangeUpdateState) {
		return m.saveChangeDetailTextValue(value)
	}
	if commandAllowed(m.state, "/save") {
		if m.state == ChangeCreateState {
			return m.saveChangeCreateValue(value)
		}
		if m.state == ChangeUpdateState {
			return m.saveChangeUpdateValue(value)
		}
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
	if m.previousState == ChangesListState {
		m.changesFilters.find = query
		m.clampChangeListSelection()
		m.state = ChangesListState
		m.input.Placeholder = defaultInputPlaceholder
		m.status = "find filter"
		return m, nil
	}
	m.helpQuery = query
	return m.arrive(m.previousState, "highlight "+query)
}

func (m Model) handlePromptCancel() (tea.Model, tea.Cmd) {
	if m.input.Value() != "" || m.detailEditField != "" {
		m.detailEditField = ""
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
	case ChangeUpdateState:
		if m.detailEditField != "" {
			m.detailEditField = ""
			m = m.setPromptValue("")
			return m.arrive(ChangeDetailsState, "cancel")
		}
		return m.arrive(navigation.CancelTarget(m.state), "cancel")
	case ChangeCreateState, TestCaseCreateState, TestCaseUpdateState,
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
		next, selected, ok := m.changeList.SelectDetail(m.changeFilters())
		m.changeList = next
		if !ok {
			m.err = "no changes selectable"
			return m, nil
		}
		m.state = ChangeDetailsState
		m.status = "selected " + selected.Title
		id, err := changeNumericID(selected)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		return m, changeGetCommand(m.client, id)
	case ChangeDetailsState:
		next, row, ok := m.changeList.SelectDetailRow(m.changeTableRows(), terminalWidth(m.width))
		m.changeList = next
		if !ok {
			m.err = "no change details selectable"
			return m, nil
		}
		switch row.Label {
		case "Phase":
			return m.beginDetailFieldSelector(detailEditPhase)
		case "Epic":
			return m.beginDetailFieldSelector(detailEditEpic)
		case "Types":
			return m.beginDetailFieldSelector(detailEditTypes)
		case "Title":
			return m.beginDetailTitleEdit()
		case "Requirement":
			return m.beginDetailTextEditor(detailEditRequirement)
		case "Pull Request":
			return m.beginDetailTextEditor(detailEditPullRequest)
		}
		m.status = "selected " + row.Label
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
		return m.arrive(ChangesListState, string(ChangesListState))
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
		m.previousState = ChangesListState
		m.state = FindInputState
		m = m.setPromptValue("")
		m.input.Placeholder = "Find changes"
	case "/clear-filters":
		m.changesFilters = changesFilters{}
		m.clampChangeListSelection()
		m.status = "filters cleared"
	case "/return":
		return m.arrive(navigation.ReturnTargets()[source], "return")
	case "/new-change", "/new-test-case", "/new-epic", "/new-project":
		m.state = navigation.CreateTarget(source)
		if m.state == ChangeCreateState {
			m = m.setPromptValue("")
			m.input.Placeholder = defaultInputPlaceholder
			return m.openPromptEditor(ChangeCreateState)
		}
		if m.state == ProjectCreateState {
			m = m.setPromptValue("")
			m.input.Placeholder = "Write a Name"
		} else {
			m.input.Placeholder = defaultInputPlaceholder
		}
	case "/edit":
		m.state = navigation.UpdateTarget(source)
		if m.state == ChangeUpdateState {
			m = m.setPromptValue(changes.RequirementMarkdown(m.changeList.Detail))
			m.input.Placeholder = defaultInputPlaceholder
			return m.openPromptEditor(ChangeUpdateState)
		}
		if m.state == ProjectUpdateState {
			m = m.setPromptValue(m.projectList.Detail.Name)
		}
		m.input.Placeholder = defaultInputPlaceholder
	case "/save":
		if source == ChangeCreateState {
			return m.saveChangeCreate()
		}
		if source == ChangeUpdateState {
			return m.saveChangeUpdate()
		}
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
		if source == ChangeUpdateState && m.detailEditField != "" {
			m.detailEditField = ""
			m = m.setPromptValue("")
		}
		return m.arrive(navigation.CancelTarget(source), "cancel")
	case "/delete":
		m.openConfirmation(navigation.DeleteConfirmationState(source), source, navigation.DeleteReturnState(source))
	case "/phase":
		if source == ChangeDetailsState {
			return m.beginDetailFieldSelector(detailEditPhase)
		}
		return m.beginSelector(SelectPhaseDropDown)
	case "/epic":
		if source == ChangeDetailsState {
			return m.beginDetailFieldSelector(detailEditEpic)
		}
		return m.beginSelector(SelectEpicDropDown)
	case "/types":
		if source == ChangeDetailsState {
			return m.beginDetailFieldSelector(detailEditTypes)
		}
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
	m.applyPromptLimit()
	if state != ProjectCreateState {
		m.input.Placeholder = defaultInputPlaceholder
	}
	switch state {
	case ChangesListState:
		m.changeList = changes.StartLoading()
		return m, changeListCommand(m.client, m.currentProject.ID)
	case ChangeDetailsState:
		id, err := changeNumericID(m.changeList.Detail)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.status = "loading change"
		return m, changeGetCommand(m.client, id)
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

func (m Model) beginDetailTitleEdit() (tea.Model, tea.Cmd) {
	m.previousState = ChangeDetailsState
	m.state = ChangeUpdateState
	m.detailEditField = detailEditTitle
	m.input.Placeholder = "Write a Title"
	m = m.setPromptValue(m.changeList.Detail.Title)
	m.status = "editing title"
	return m, nil
}

func (m Model) beginDetailTextEditor(field detailEditField) (tea.Model, tea.Cmd) {
	m.detailEditField = field
	switch field {
	case detailEditRequirement:
		m = m.setPromptValue(m.changeList.Detail.RequirementBody)
	case detailEditPullRequest:
		m = m.setPromptValue(m.changeList.Detail.PullRequestBody)
	default:
		m.err = "unsupported editable detail text field"
		return m, nil
	}
	return m.openPromptEditor(ChangeDetailsState)
}

func (m Model) beginDetailFieldSelector(field detailEditField) (tea.Model, tea.Cmd) {
	m.detailEditField = ""
	switch field {
	case detailEditPhase:
		m.openSelectorDropdown(SelectPhaseDropDown, ChangeDetailsState, ChangeDetailsState, "Phase", selectorPhases)
	case detailEditEpic:
		m.openSelectorDropdown(SelectEpicDropDown, ChangeDetailsState, ChangeDetailsState, "Epic", selectorEpics)
	case detailEditTypes:
		m.openSelectorDropdown(SelectTypesDropDown, ChangeDetailsState, ChangeDetailsState, "Types", selectorTypes)
	default:
		m.err = "unsupported editable detail field"
		return m, nil
	}
	m.dropdown.editField = field
	return m, selectorCommand(m.client, m.dropdown.source, m.currentProject.ID)
}

func (m Model) saveChangeDetailTextValue(value string) (tea.Model, tea.Cmd) {
	field := m.detailEditField
	m.status = "saving " + string(field)
	return m, changeDetailTextUpdateCommand(m.client, m.state, m.changeList.Detail, field, value)
}

func (m Model) needsProjectSelection() bool {
	return m.currentProject.ID == "" && m.appConfig.ProjectID <= 0
}
