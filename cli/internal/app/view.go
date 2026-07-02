package app

import (
	"fmt"
	"strings"

	"mch/internal/changes"
	"mch/internal/epics"
	"mch/internal/help"
	"mch/internal/planning"
	"mch/internal/projects"
	"mch/internal/styles"
	"mch/internal/testcases"
	"mch/internal/ui"

	"github.com/charmbracelet/lipgloss"
)

// View renders the root application shell and active screen.
func (m Model) View() string {
	width := terminalWidth(m.width)
	lines := []string{m.headerLine(width)}
	if m.state == ProjectsListState && !m.hasDropdown() {
		lines = append(lines, "")
		lines = append(lines, projects.TableView(m.projectList, width))
	}
	if m.state == ChangesListState && !m.hasDropdown() {
		table := changes.TableView(m.changeList, m.changeFilters(), width, m.changeTableRows())
		lines = append(lines, m.changeFiltersLine(table), table)
	}
	if m.state == ChangeDetailsState {
		details := changes.DetailsView(m.changeList, width, m.changeTableRows())
		if details != "" {
			lines = append(lines, "")
			lines = append(lines, details)
		}
	}
	if m.state == ProjectDetailsState {
		details := projects.DetailsView(m.projectList.Detail, width)
		if details != "" {
			lines = append(lines, "")
			lines = append(lines, details)
		}
	}
	if m.state == FindInputState {
		lines = append(lines, "")
		lines = append(lines, m.inputBand(width))
	} else if m.hasDropdown() {
		lines = append(lines, "")
		lines = append(lines, m.dropdownView(width))
	} else {
		lines = append(lines, "")
		lines = append(lines, m.inputBand(width))
	}
	if m.err != "" {
		lines = append(lines, styles.Default.Error.Render("Error: "+m.err))
	}
	if m.helpQuery != "" {
		lines = append(lines, styles.Default.Success.Render("Highlight: "+m.helpQuery))
	}
	lines = append(lines, "", styles.Default.Footer.Width(width).Render(m.footerText()))
	if m.quitting {
		lines = append(lines, styles.Default.Success.Render("done"))
	}
	return styles.Default.Surface.Width(width).Render(strings.Join(lines, "\n"))
}

func (m Model) headerLine(width int) string {
	left := appTitle()
	right := m.headerRight()
	padding := width - lipgloss.Width(left) - lipgloss.Width(right)
	if padding < 1 {
		padding = 1
	}
	return left + strings.Repeat(" ", padding) + right
}

func (m Model) headerRight() string {
	title := screenTitle(m.state)
	if before, _, ok := strings.Cut(title, " - "); ok {
		title = before
	}
	return styles.Default.Foreground.Render(title)
}

func (m Model) changeFiltersLine(table string) string {
	line := changeFilterLabel("/filter-phase ") + changeFilterValue(m.changeFilters().Phase.Label) +
		"   " + changeFilterLabel("/filter-type ") + changeFilterValue(m.changeFilters().Type.Label) +
		"   " + changeFilterLabel("/filter-epic ") + changeFilterValue(m.changeFilters().Epic.Label) +
		"   " + changeFilterLabel("/filter-find ") + changeFilterValue(m.changeFilters().Find)
	tableWidth := firstLineWidth(table)
	padding := tableWidth - lipgloss.Width(line)
	if padding < 0 {
		padding = 0
	}
	return strings.Repeat(" ", padding) + line
}

func changeFilterLabel(value string) string {
	return styles.Default.Muted.Render(value)
}

func changeFilterValue(value string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Render(value)
}

func firstLineWidth(value string) int {
	first, _, _ := strings.Cut(value, "\n")
	return lipgloss.Width(first)
}

func (m Model) helpText() string {
	if m.hasDropdown() {
		if m.dropdown.kind == dropdownConfirm {
			return "<return> select  |  <esc> or <ctrl+c> cancel"
		}
		return "<return> select  |  <esc> cancel"
	}
	switch m.state {
	case ChangesListState:
		return "<ctrl+n> new change  |  <return> view  |  </> command"
	case ChangeDetailsState:
		return "<ctrl+n> new testcase  |  <return> edit  |  <space> toggle  |  <del> delete  |  </> command"
	case TestCaseCreateState, TestCaseUpdateState:
		return "<return> save  |  <ctrl+c> delete prompt  |  <esc> cancel"
	case ChangeCreateState, ChangeUpdateState, EpicCreateState, EpicUpdateState, ProjectCreateState, ProjectUpdateState:
		return "<return> save  |  <ctrl+c> delete prompt  |  <esc> cancel"
	case FindInputState:
		return "<return> search  |  <ctrl+c> delete prompt  |  <esc> cancel"
	case ProjectsListState, EpicsListState:
		return "<return> view  |  </> command"
	case ProjectDetailsState, EpicDetailsState, TestCaseDetailsState:
		return "<return> edit  |  </> command"
	default:
		return "</> command  |  <esc> cancel"
	}
}

func (m Model) inputBand(width int) string {
	width = ui.NormalizeWidth(width)
	content := m.inputLines(width)
	blank := strings.Repeat(" ", width)
	lines := []string{styles.Default.InputBand.Render(blank)}
	lines = append(lines, content...)
	lines = append(lines, styles.Default.InputBand.Render(blank))
	return strings.Join(lines, "\n")
}

func (m Model) inputLines(width int) []string {
	lines := promptValueLines(m.input.Value())
	padded := make([]string, 0, len(lines))
	for index, value := range lines {
		showCursor := m.input.Focused() && m.input.Value() != "" && index == m.promptCursorRow
		line := m.renderPromptLine(value, showCursor)
		if visible := lipgloss.Width(line); visible < width {
			line += styles.Default.InputBand.Render(strings.Repeat(" ", width-visible))
		}
		padded = append(padded, line)
	}
	return padded
}

func (m Model) renderPromptLine(value string, showCursor bool) string {
	prompt := styles.Default.InputBand.Foreground(lipgloss.Color("183")).Render("> ")
	if m.input.Value() == "" {
		placeholder := styles.Default.InputBand.Foreground(lipgloss.Color("0")).Render(m.input.Placeholder)
		return prompt + placeholder
	}
	if showCursor {
		runes := []rune(value)
		col := m.promptCursorCol
		if col < 0 {
			col = 0
		}
		if col > len(runes) {
			col = len(runes)
		}
		before := styles.Default.InputBand.Foreground(lipgloss.Color("15")).Render(string(runes[:col]))
		after := styles.Default.InputBand.Foreground(lipgloss.Color("15")).Render(string(runes[col:]))
		return prompt + before + promptCursor() + after
	}
	return prompt + styles.Default.InputBand.Foreground(lipgloss.Color("15")).Render(value)
}

func promptCursor() string {
	return styles.Default.InputBand.
		Background(lipgloss.Color("15")).
		Foreground(lipgloss.Color("0")).
		Render(" ")
}

func promptValueLines(value string) []string {
	if value == "" {
		return []string{""}
	}
	return strings.Split(value, "\n")
}

func (m Model) footerText() string {
	currentProject := "Current Project: " + m.currentProjectFooter()
	if m.status != "" {
		return fmt.Sprintf("%s  |  status %s  |  %s  |  %s", m.helpText(), m.status, currentProject, footerColorStrip())
	}
	return m.helpText() + "  |  " + currentProject + "  |  " + footerColorStrip()
}

func footerColorStrip() string {
	cells := make([]string, 0, 17)
	for color := 0; color <= 16; color++ {
		label := fmt.Sprintf("%d", color)
		foreground := lipgloss.Color("15")
		switch color {
		case 7, 10, 11, 12, 14, 15, 16:
			foreground = lipgloss.Color("0")
		}
		cells = append(cells, lipgloss.NewStyle().
			Background(lipgloss.Color(label)).
			Foreground(foreground).
			Render(label))
	}
	return strings.Join(cells, " ")
}

func appTitle() string {
	return styles.Default.Title.Render("Make a change") + styles.Default.Muted.Render(" v"+Version)
}

func (m Model) currentProjectFooter() string {
	id := strings.TrimSpace(m.currentProject.ID)
	label := strings.TrimSpace(m.currentProject.Label)
	if id == "" {
		return "none"
	}
	if label == "" || label == id || label == "Project #"+id {
		return "#" + id
	}
	return "#" + id + " " + label
}

func screenTitle(state State) string {
	titles := map[State]string{
		MainState:                  planning.MainTitle(),
		ChangesListState:           changes.ListTitle(),
		ChangeDetailsState:         changes.DetailTitle(),
		TestCaseDetailsState:       testcases.DetailTitle(),
		ChangeCreateState:          "ChangeCreateScreen - Title: New Change",
		ChangeUpdateState:          "ChangeUpdateScreen - Title: Edit Change",
		TestCaseCreateState:        testcases.CreateTitle(),
		TestCaseUpdateState:        testcases.UpdateTitle(),
		EpicsListState:             epics.ListTitle(),
		EpicDetailsState:           epics.DetailTitle(),
		EpicCreateState:            "EpicCreateScreen - Title: New Epic",
		EpicUpdateState:            "EpicUpdateScreen - Title: Edit Epic",
		ProjectsListState:          projects.ListTitle(),
		ProjectDetailsState:        projects.DetailTitle(),
		ProjectCreateState:         projects.CreateTitle(),
		ProjectUpdateState:         projects.UpdateTitle(),
		MainHelpState:              help.MainTitle(),
		ChangesHelpState:           help.ChangesTitle(),
		EpicsHelpState:             help.EpicsTitle(),
		ProjectsHelpState:          help.ProjectsTitle(),
		FindInputState:             help.FindInputTitle(),
		CommandDropDownState:       "CommandDropDownScreen - Title: Commands",
		ListSelectionDropDownState: "ListSelectionDropDownScreen - Title: Select Item",
		SelectProjectDropDown:      "SelectProjectDropDownScreen - Title: Select Project",
		SelectPhaseDropDown:        "SelectChangePhasesDropDownScreen - Title: Select Change Phases",
		SelectEpicDropDown:         "SelectEpicDropDownScreen - Title: Select Epic",
		SelectTypesDropDown:        "SelectChangeTypesDropDownScreen - Title: Select Change Types",
		ChangeDeleteConfirmation:   "ChangeDeleteConfirmationScreen - Title: Are you sure?",
		TestCaseDeleteConfirmation: "TestCaseDeleteConfirmationScreen - Title: Are you sure?",
		EpicDeleteConfirmation:     "EpicDeleteConfirmationScreen - Title: Are you sure?",
		ProjectDeleteConfirmation:  "ProjectDeleteConfirmationScreen - Title: Are you sure?",
		DoneState:                  "DoneScreen - Title: Done",
	}
	if title, ok := titles[state]; ok {
		return title
	}
	return "UnknownScreen - Title: Unknown"
}

func terminalWidth(width int) int {
	return ui.NormalizeWidth(width)
}

func (m Model) changeTableRows() int {
	const reservedRows = 12
	available := m.height - reservedRows
	if available < 3 {
		return 3
	}
	return available
}
