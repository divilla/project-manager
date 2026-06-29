package app

import (
	"fmt"
	"strings"

	"mch/internal/changes"
	"mch/internal/epics"
	"mch/internal/help"
	"mch/internal/planning"
	"mch/internal/projects"
	"mch/internal/requirements"
	"mch/internal/styles"
	"mch/internal/ui"

	"github.com/charmbracelet/lipgloss"
)

// View renders the root application shell and active screen.
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
	if m.state == ProjectsListState && !m.hasDropdown() {
		lines = append(lines, "", projects.TableView(m.projectList, width))
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
	lines = append(lines, styles.Default.Footer.Width(width).Render(m.footerText()))
	if m.quitting {
		lines = append(lines, styles.Default.Success.Render("done"))
	}
	return styles.Default.Surface.Width(width).Render(strings.Join(lines, "\n"))
}

func (m Model) inputBand(width int) string {
	width = ui.NormalizeWidth(width)
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

func (m Model) footerText() string {
	if m.status != "" {
		return fmt.Sprintf("/ commands  |  esc safe action  |  status %s", m.status)
	}
	return "/ commands  |  esc safe action"
}

func screenTitle(state State) string {
	titles := map[State]string{
		MainState:                     planning.MainTitle(),
		ChangesListState:              changes.ListTitle(),
		ChangeDetailsState:            changes.DetailTitle(),
		RequirementDetailsState:       requirements.DetailTitle(),
		ChangeCreateState:             "ChangeCreateScreen - Title: New Change",
		ChangeUpdateState:             "ChangeUpdateScreen - Title: Edit Change",
		RequirementCreateState:        requirements.CreateTitle(),
		RequirementUpdateState:        requirements.UpdateTitle(),
		EpicsListState:                epics.ListTitle(),
		EpicDetailsState:              epics.DetailTitle(),
		EpicCreateState:               "EpicCreateScreen - Title: New Epic",
		EpicUpdateState:               "EpicUpdateScreen - Title: Edit Epic",
		ProjectsListState:             projects.ListTitle(),
		ProjectDetailsState:           projects.DetailTitle(),
		ProjectCreateState:            projects.CreateTitle(),
		ProjectUpdateState:            projects.UpdateTitle(),
		MainHelpState:                 help.MainTitle(),
		ChangesHelpState:              help.ChangesTitle(),
		EpicsHelpState:                help.EpicsTitle(),
		ProjectsHelpState:             help.ProjectsTitle(),
		FindInputState:                help.FindInputTitle(),
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

func terminalWidth(width int) int {
	return ui.NormalizeWidth(width)
}
