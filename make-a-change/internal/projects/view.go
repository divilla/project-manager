package projects

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"mch/internal/dto"
	"mch/internal/styles"
	"mch/internal/ui"
)

// ListTitle returns the projects list screen title.
func ListTitle() string {
	return "ProjectsListScreen - Title: Projects List"
}

// DetailTitle returns the project details screen title.
func DetailTitle() string {
	return "ProjectDetailsScreen - Title: Project Details"
}

// CreateTitle returns the new project screen title.
func CreateTitle() string {
	return "ProjectCreateScreen - Title: New Project"
}

// UpdateTitle returns the edit project screen title.
func UpdateTitle() string {
	return "ProjectUpdateScreen - Title: Edit Project"
}

// TableView renders the selectable projects table.
func TableView(m Model, width int) string {
	if m.Loading {
		return styles.Default.InputBand.Width(width).Render("Projects: loading")
	}
	if len(m.Rows) == 0 {
		return styles.Default.Muted.Render("No projects.")
	}
	rows := []string{
		tableLine("id", "Name", "Changes", "Created", "Modified"),
	}
	for i, project := range m.Rows {
		line := tableLine(
			FormatID(project.ID),
			project.Name,
			strconv.Itoa(project.ChangeCount),
			FormatTimestamp(project.Created),
			FormatTimestamp(project.Modified),
		)
		if i == m.Selected {
			line = styles.Default.Selection.Render(line)
		}
		rows = append(rows, line)
	}
	return ui.TruncateBlock(strings.Join(rows, "\n"), width)
}

// DetailsView renders selected project details.
func DetailsView(project dto.Project, width int) string {
	if project.ID == "" && project.Name == "" {
		return ""
	}
	lines := []string{
		"Project: " + DisplayName(project),
		"Change Count: " + strconv.Itoa(project.ChangeCount),
		"Created: " + FormatTimestamp(project.Created),
		"Modified: " + FormatTimestamp(project.Modified),
	}
	return styles.Default.Muted.Render(ui.TruncateBlock(strings.Join(lines, "\n"), width))
}

func tableLine(id, name, changeCount, created, modified string) string {
	return fmt.Sprintf("%6s  %-24s  %7s  %-16s  %-16s", id, name, changeCount, created, modified)
}

// FormatID renders a project ID without a leading number prefix.
func FormatID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return "?"
	}
	return strings.TrimPrefix(id, "#")
}

// FormatTimestamp renders a project timestamp in local display format.
func FormatTimestamp(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "Invalid"
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.Local().Format("2006-01-02 15:04")
		}
	}
	return "Invalid"
}

// DisplayName returns a compact project label for status and details.
func DisplayName(project dto.Project) string {
	name := strings.TrimSpace(project.Name)
	id := FormatID(project.ID)
	if name == "" {
		return id
	}
	return id + " " + name
}
