package changes

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"mch/internal/dto"
	"mch/internal/styles"
	"mch/internal/ui"

	"github.com/charmbracelet/lipgloss"
)

// ListTitle returns the changes list screen title.
func ListTitle() string {
	return "ChangesListScreen - Title: Changes List"
}

// DetailTitle returns the change details screen title.
func DetailTitle() string {
	return "ChangeDetailsScreen - Title: Change Details"
}

// TableView renders the selectable changes list.
func TableView(m Model, filters Filters, width int, pageSize int) string {
	width = ui.NormalizeWidth(width)
	if m.Loading {
		return styles.Default.InputBand.Width(width).Render("Changes: loading")
	}
	rows := FilteredRows(m.Rows, filters)
	if len(rows) == 0 {
		if len(m.Rows) == 0 {
			return styles.Default.Muted.Render("No changes.")
		}
		return styles.Default.Muted.Render("No changes match filters.")
	}
	if pageSize < 1 {
		pageSize = 1
	}
	selected := m.ClampSelection(filters, pageSize).Selected
	offset := clampOffset(m.Offset, selected, len(rows), pageSize)
	end := offset + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	terminalTableWidth := innerTableWidth(width)
	typesWidth, epicWidth, titleWidth := changeTableColumnWidths(terminalTableWidth)
	tableWidth := changeTableContentWidth(typesWidth, epicWidth, titleWidth)
	lines := []string{styles.Default.Muted.Render(changeTableLine("#Ref", "Phase", "Types", "Epic", "Title", "Don", "Tot", "%", "Modified", typesWidth, epicWidth, titleWidth))}
	for i, change := range rows[offset:end] {
		rowIndex := offset + i
		line := changeTableRowLine(
			displayRef(change),
			change.ChangePhase,
			strings.Join(change.ChangeTypes, "|"),
			epicLabel(change),
			change.Title,
			strconv.Itoa(change.Done),
			strconv.Itoa(change.Total),
			strconv.Itoa(change.Completed),
			formatListTimestamp(change.Modified),
			typesWidth,
			epicWidth,
			titleWidth,
			rowIndex == selected,
		)
		lines = append(lines, line)
	}
	for len(lines) < pageSize+1 {
		lines = append(lines, "")
	}
	lines = append(lines, styles.Default.Foreground.Render(fmt.Sprintf("Rows %d-%d of %d", offset+1, end, len(rows))))
	content := ui.TruncateBlock(strings.Join(lines, "\n"), tableWidth)
	return boxedTable(content, tableWidth)
}

func changeTableLine(ref, phase, types, epic, title, done, total, completed, modified string, typesWidth, epicWidth, titleWidth int) string {
	return fmt.Sprintf(
		"%6s %-10s %-*s %-*s %-*s %3s %3s %3s %-16s",
		tableText(ref, 6),
		tableText(phase, 10),
		typesWidth,
		tableText(types, typesWidth),
		epicWidth,
		tableText(epic, epicWidth),
		titleWidth,
		tableText(title, titleWidth),
		tableText(done, 3),
		tableText(total, 3),
		tableText(completed, 3),
		tableText(modified, 16),
	)
}

func changeTableRowLine(ref, phase, types, epic, title, done, total, completed, modified string, typesWidth, epicWidth, titleWidth int, selected bool) string {
	prefix := fmt.Sprintf("%6s ", tableText(ref, 6))
	phaseValue := fmt.Sprintf("%-10s", tableText(phase, 10))
	beforeTitle := fmt.Sprintf(
		" %-*s %-*s ",
		typesWidth,
		tableText(types, typesWidth),
		epicWidth,
		tableText(epic, epicWidth),
	)
	titleValue := fmt.Sprintf("%-*s", titleWidth, tableText(title, titleWidth))
	beforeCompleted := fmt.Sprintf(
		" %3s %3s ",
		tableText(done, 3),
		tableText(total, 3),
	)
	completedValue := fmt.Sprintf("%3s", tableText(completed, 3))
	afterCompleted := fmt.Sprintf(" %-16s", tableText(modified, 16))

	if selected {
		base := styles.Default.Selection
		return base.Render(prefix) +
			phaseStyle(phase).Background(lipgloss.Color("60")).Render(phaseValue) +
			base.Render(beforeTitle) +
			base.Foreground(lipgloss.Color("15")).Render(titleValue) +
			base.Render(beforeCompleted) +
			base.Foreground(lipgloss.Color("14")).Render(completedValue) +
			base.Render(afterCompleted)
	}

	return styles.Default.Muted.Render(prefix) +
		phaseStyle(phase).Render(phaseValue) +
		styles.Default.Muted.Render(beforeTitle) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Render(titleValue) +
		styles.Default.Muted.Render(beforeCompleted) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Render(completedValue) +
		styles.Default.Muted.Render(afterCompleted)
}

func phaseStyle(phase string) lipgloss.Style {
	switch strings.ToLower(strings.TrimSpace(phase)) {
	case "backlog":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	case "progress":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	case "review":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
	case "staging":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	case "production", "done":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	case "rejected":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	}
}

func changeTableColumnWidths(width int) (int, int, int) {
	const targetTypesWidth = 30
	const maxEpicWidth = 20
	const maxTitleWidth = 80
	const minTypesWidth = 5
	const minEpicWidth = 4
	const minTitleWidth = 5

	available := width - changeTableFixedWidth
	if available >= targetTypesWidth+maxEpicWidth+maxTitleWidth {
		return targetTypesWidth, maxEpicWidth, maxTitleWidth
	}
	if available <= minTypesWidth+minEpicWidth+minTitleWidth {
		return shrinkColumns(available, minTypesWidth, minEpicWidth, minTitleWidth)
	}

	typesWidth := targetTypesWidth
	epicWidth := maxEpicWidth
	titleWidth := available - typesWidth - epicWidth

	if titleWidth < minTitleWidth {
		deficit := minTitleWidth - titleWidth
		epicReduction := min(deficit, epicWidth-minEpicWidth)
		epicWidth -= epicReduction
		deficit -= epicReduction
		typesWidth -= min(deficit, typesWidth-minTypesWidth)
		titleWidth = available - typesWidth - epicWidth
	}
	if epicWidth < minEpicWidth {
		deficit := minEpicWidth - epicWidth
		epicWidth = minEpicWidth
		titleWidth -= deficit
	}
	if titleWidth > maxTitleWidth {
		extra := titleWidth - maxTitleWidth
		titleWidth = maxTitleWidth
		typesWidth += extra
	}
	return typesWidth, epicWidth, titleWidth
}

const changeTableFixedWidth = 6 + 1 + 10 + 1 + 1 + 1 + 1 + 3 + 1 + 3 + 1 + 3 + 1 + 16

func changeTableContentWidth(typesWidth, epicWidth, titleWidth int) int {
	return changeTableFixedWidth + typesWidth + epicWidth + titleWidth
}

func shrinkColumns(available, typesWidth, epicWidth, titleWidth int) (int, int, int) {
	if available <= 0 {
		return 1, 1, 1
	}
	for typesWidth+epicWidth+titleWidth > available {
		switch {
		case titleWidth > 1:
			titleWidth--
		case typesWidth > 1:
			typesWidth--
		case epicWidth > 1:
			epicWidth--
		default:
			return typesWidth, epicWidth, titleWidth
		}
	}
	return typesWidth, epicWidth, titleWidth
}

// DetailsView renders selected change details.
func DetailsView(change dto.Change, width int) string {
	if change.ID == "" && change.Title == "" {
		return ""
	}
	lines := []string{
		detailLine("Ref", displayRef(change)),
		detailLine("Slug", change.Slug),
		detailLine("Title", change.Title),
		detailLine("Phase", change.ChangePhase),
		detailLine("Types", strings.Join(change.ChangeTypes, "|")),
		detailLine("Epic", epicLabel(change)),
		detailLine("Closed", fmt.Sprintf("%t", change.Closed)),
	}
	if strings.TrimSpace(change.RequirementBody) != "" {
		lines = append(lines, detailWrappedLines("Body", change.RequirementBody, 88)...)
	}
	return ui.TruncateBlock(strings.Join(lines, "\n"), width)
}

func tableText(value string, limit int) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func boxedTable(content string, width int) string {
	if width < 1 {
		width = 1
	}
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(width).
		Render(content)
}

func innerTableWidth(width int) int {
	width = ui.NormalizeWidth(width)
	if width <= 4 {
		return 20
	}
	return width - 2
}

func detailLine(label, value string) string {
	if strings.TrimSpace(value) == "" {
		value = "-"
	}
	return styles.Default.Muted.Render(fmt.Sprintf("    %8s: ", label)) + styles.Default.Foreground.Render(value)
}

func detailWrappedLines(label, value string, limit int) []string {
	value = strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\n"), "\r", "\n")
	parts := strings.Split(value, "\n")
	lines := make([]string, 0, len(parts))
	prefix := styles.Default.Muted.Render(fmt.Sprintf("    %8s: ", label))
	indent := styles.Default.Muted.Render(strings.Repeat(" ", 14))
	for i, part := range parts {
		wrapped := wrapWords(part, limit)
		if len(wrapped) == 0 {
			wrapped = []string{""}
		}
		for j, line := range wrapped {
			switch {
			case i == 0 && j == 0:
				lines = append(lines, prefix+styles.Default.Foreground.Render(line))
			default:
				lines = append(lines, indent+styles.Default.Foreground.Render(line))
			}
		}
	}
	return lines
}

func wrapWords(value string, limit int) []string {
	words := strings.Fields(strings.TrimSpace(value))
	if len(words) == 0 {
		return nil
	}
	lines := make([]string, 0, len(words))
	current := words[0]
	for _, word := range words[1:] {
		if len([]rune(current))+1+len([]rune(word)) > limit {
			lines = append(lines, current)
			current = word
			continue
		}
		current += " " + word
	}
	lines = append(lines, current)
	return lines
}

func displayRef(change dto.Change) string {
	ref := strings.TrimPrefix(strings.TrimSpace(change.Ref), "#")
	if ref != "" {
		if value, err := strconv.Atoi(ref); err == nil {
			return fmt.Sprintf("%06d", value)
		}
		return ref
	}
	if strings.TrimSpace(change.ID) != "" {
		return "id:" + strings.TrimSpace(change.ID)
	}
	return "?"
}

func epicLabel(change dto.Change) string {
	if strings.TrimSpace(change.EpicName) != "" {
		return strings.TrimSpace(change.EpicName)
	}
	if strings.TrimSpace(change.EpicID) != "" {
		return "#" + strings.TrimPrefix(strings.TrimSpace(change.EpicID), "#")
	}
	return ""
}

func formatListTimestamp(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "not a date"
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
			return parsed.Format("2006-01-02 15.04")
		}
	}
	return "not a date"
}
