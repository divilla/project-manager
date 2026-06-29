package app

import (
	"strings"

	"mch/internal/dto"
	"mch/internal/styles"

	tea "github.com/charmbracelet/bubbletea"
)

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

func (m *Model) openDropdown(state State, kind dropdownKind, previous State, onSelect State, label string, options []dto.Option, loading bool) {
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

func (m *Model) openConfirmation(state, previous, onYes State) {
	m.openDropdown(state, dropdownConfirm, previous, onYes, "Confirm", []dto.Option{
		{ID: "/yes", Label: "/yes"},
		{ID: "/cancel", Label: "/cancel"},
	}, false)
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

func (m Model) selectedOption() dto.Option {
	options := m.filteredOptions()
	if len(options) == 0 {
		return dto.Option{}
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

func (m Model) filteredOptions() []dto.Option {
	filter := strings.ToLower(strings.TrimSpace(m.dropdown.filter))
	if filter == "" {
		return m.dropdown.options
	}
	var options []dto.Option
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
