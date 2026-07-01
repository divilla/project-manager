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
	lines := []string{
		appTitle(),
		"",
		styles.Default.Foreground.Render(screenTitle(m.state)),
	}
	if m.state == ProjectsListState && !m.hasDropdown() {
		lines = append(lines, "", projects.TableView(m.projectList, width))
	}
	if m.state == ChangesListState && !m.hasDropdown() {
		lines = append(lines, "", changes.TableView(m.changeList, m.changeFilters(), width, m.changeTableRows()))
	}
	if m.state == ChangeDetailsState {
		details := changes.DetailsView(m.changeList.Detail, width)
		if details != "" {
			lines = append(lines, "", details)
		}
	}
	if m.state == ProjectDetailsState {
		details := projects.DetailsView(m.projectList.Detail, width)
		if details != "" {
			lines = append(lines, "", details)
		}
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
	lines = append(lines, "", styles.Default.Footer.Width(width).Render(m.footerText()))
	if m.quitting {
		lines = append(lines, styles.Default.Success.Render("done"))
	}
	return styles.Default.Surface.Width(width).Render(strings.Join(lines, "\n"))
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
		return fmt.Sprintf("/ commands  |  esc safe action  |  status %s  |  %s  |  %s", m.status, currentProject, footerColorStrip())
	}
	return "/ commands  |  esc safe action  |  " + currentProject + "  |  " + footerColorStrip()
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
	return styles.Default.Title.Render("Make a Change") + styles.Default.Muted.Render(" ver. "+Version)
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
		SelectPhaseDropDown:        "SelectPhaseDropDownScreen - Title: Select Phase",
		SelectEpicDropDown:         "SelectEpicDropDownScreen - Title: Select Epic",
		SelectTypesDropDown:        "SelectTypesDropDownScreen - Title: Select Types",
		ChangeDeleteConfirmation:   "ChangeDeleteConfirmationScreen - Title: Confirm Delete",
		TestCaseDeleteConfirmation: "TestCaseDeleteConfirmationScreen - Title: Confirm Delete",
		EpicDeleteConfirmation:     "EpicDeleteConfirmationScreen - Title: Confirm Delete",
		ProjectDeleteConfirmation:  "ProjectDeleteConfirmationScreen - Title: Confirm Delete",
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
