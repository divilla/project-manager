package changes

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"mch/internal/dto"
)

// Filters stores active change list filter selections.
type Filters struct {
	Phase dto.Option
	Epic  dto.Option
	Type  dto.Option
	Find  string
}

// Model stores changes list and detail state.
type Model struct {
	Rows           []dto.Change
	Selected       int
	Offset         int
	Detail         dto.Change
	DetailSelected int
	DetailOffset   int
	Loading        bool
}

// DetailRow is one row in the Change details table.
type DetailRow struct {
	Label      string
	Text       string
	Selectable bool
}

// ParsedRequirement stores metadata extracted from requirement markdown.
type ParsedRequirement struct {
	Title       string
	Body        string
	ChangeTypes []string
	EpicID      *int
	EpicName    string
}

// StartLoading returns a changes model in loading state.
func StartLoading() Model {
	return Model{Loading: true}
}

// WithRows returns a changes model populated with loaded rows.
func (m Model) WithRows(rows []dto.Change) Model {
	m.Rows = rows
	m.Selected = 0
	m.Offset = 0
	m.Loading = false
	return m
}

// WithError returns a changes model reset after load failure.
func (m Model) WithError() Model {
	m.Rows = nil
	m.Selected = 0
	m.Offset = 0
	m.Loading = false
	return m
}

// MoveSelection moves the selected change within list bounds.
func (m Model) MoveSelection(delta int, filters Filters, pageSize int) Model {
	m = m.ClampSelection(filters, pageSize)
	visible := FilteredRows(m.Rows, filters)
	if len(visible) == 0 {
		return m
	}
	next := m.Selected + delta
	if next < 0 {
		next = 0
	}
	if next >= len(visible) {
		next = len(visible) - 1
	}
	m.Selected = next
	m.Offset = clampOffset(m.Offset, m.Selected, len(visible), pageSize)
	return m
}

// ClampSelection keeps the selected visible row and scroll offset in bounds.
func (m Model) ClampSelection(filters Filters, pageSize int) Model {
	visible := FilteredRows(m.Rows, filters)
	if len(visible) == 0 {
		m.Selected = 0
		m.Offset = 0
		return m
	}
	if m.Selected < 0 {
		m.Selected = 0
	}
	if m.Selected >= len(visible) {
		m.Selected = len(visible) - 1
	}
	m.Offset = clampOffset(m.Offset, m.Selected, len(visible), pageSize)
	return m
}

// SelectDetail selects the current visible change.
func (m Model) SelectDetail(filters Filters) (Model, dto.Change, bool) {
	visible := FilteredRows(m.Rows, filters)
	if len(visible) == 0 {
		return m, dto.Change{}, false
	}
	m = m.ClampSelection(filters, 1)
	m.Offset = clampOffset(m.Offset, m.Selected, len(visible), 1)
	selected := visible[m.Selected]
	m = m.WithDetail(selected)
	return m, selected, true
}

// WithDetail stores the selected Change and resets detail-table selection.
func (m Model) WithDetail(change dto.Change) Model {
	m.Detail = change
	m.DetailSelected = firstSelectableDetailRow(DetailRows(change))
	m.DetailOffset = 0
	return m
}

// MoveDetailSelection moves within editable detail rows and keeps the row visible.
func (m Model) MoveDetailSelection(delta int, pageSize int, width int) Model {
	rows := DetailRows(m.Detail)
	if len(rows) == 0 {
		m.DetailSelected = 0
		m.DetailOffset = 0
		return m
	}
	m = m.ClampDetailSelection(pageSize, width)
	next := nextSelectableDetailRow(rows, m.DetailSelected, delta)
	if next >= 0 {
		m.DetailSelected = next
	}
	_, textWidth := DetailColumnWidths(m.Detail, width)
	m.DetailOffset = detailRowLineStart(rows, m.DetailSelected, textWidth)
	m.DetailOffset = clampLineOffset(m.DetailOffset, detailLineCount(rows, textWidth), pageSize)
	return m
}

// ClampDetailSelection keeps the selected detail row and scroll offset in bounds.
func (m Model) ClampDetailSelection(pageSize int, width int) Model {
	rows := DetailRows(m.Detail)
	if len(rows) == 0 {
		m.DetailSelected = 0
		m.DetailOffset = 0
		return m
	}
	if m.DetailSelected < 0 || m.DetailSelected >= len(rows) || !rows[m.DetailSelected].Selectable {
		m.DetailSelected = firstSelectableDetailRow(rows)
	}
	_, textWidth := DetailColumnWidths(m.Detail, width)
	m.DetailOffset = clampLineOffset(m.DetailOffset, detailLineCount(rows, textWidth), pageSize)
	return m
}

// ScrollDetailViewport moves the detail table viewport by rendered lines.
func (m Model) ScrollDetailViewport(delta int, pageSize int, width int) Model {
	rows := DetailRows(m.Detail)
	if len(rows) == 0 {
		m.DetailSelected = 0
		m.DetailOffset = 0
		return m
	}
	_, textWidth := DetailColumnWidths(m.Detail, width)
	m.DetailOffset = clampLineOffset(m.DetailOffset+delta, detailLineCount(rows, textWidth), pageSize)
	m.DetailSelected = selectableDetailRowAtOffset(rows, m.DetailOffset, textWidth)
	return m
}

// SelectDetailRow returns the currently selected editable detail row.
func (m Model) SelectDetailRow(pageSize int, width int) (Model, DetailRow, bool) {
	m = m.ClampDetailSelection(pageSize, width)
	rows := DetailRows(m.Detail)
	if len(rows) == 0 || m.DetailSelected < 0 || m.DetailSelected >= len(rows) || !rows[m.DetailSelected].Selectable {
		return m, DetailRow{}, false
	}
	return m, rows[m.DetailSelected], true
}

// DetailRows returns Change details as label/text table rows.
func DetailRows(change dto.Change) []DetailRow {
	if change.ID == "" && change.Title == "" {
		return nil
	}
	return []DetailRow{
		{Label: "Ref", Text: displayRef(change), Selectable: true},
		{Label: "Slug", Text: change.Slug, Selectable: true},
		{Label: "Phase", Text: change.ChangePhase, Selectable: true},
		{Label: "Epic", Text: epicLabel(change), Selectable: true},
		{Label: "Types", Text: strings.Join(change.ChangeTypes, "|"), Selectable: true},
		{Label: "Title", Text: change.Title, Selectable: true},
		{Label: "Requirement", Text: change.Body, Selectable: true},
		{Label: "Pull Request", Text: change.PRBody, Selectable: true},
		{Label: "PR URL", Text: change.PRUrl, Selectable: true},
		{Label: "Agent Edit", Text: fmt.Sprintf("%t", change.AgentEdit), Selectable: true},
		{Label: "Complete", Text: fmt.Sprintf("%d/%d - %d%%", change.Done, change.Total, change.Completed), Selectable: true},
		{Label: "Open", Text: fmt.Sprintf("%t", change.Open), Selectable: true},
		{Label: "Created", Text: formatListTimestamp(change.Created), Selectable: true},
		{Label: "Modified", Text: formatListTimestamp(change.Modified), Selectable: true},
	}
}

// DetailColumnWidths returns label and text widths for the rendered details table.
func DetailColumnWidths(change dto.Change, width int) (int, int) {
	contentWidth := width - 2
	if width <= 4 {
		contentWidth = 20
	}
	if contentWidth < 20 {
		contentWidth = 20
	}
	labelWidth := detailLabelWidth(DetailRows(change))
	textWidth := contentWidth - labelWidth - 3
	if textWidth < 10 {
		textWidth = 10
		labelWidth = max(1, contentWidth-textWidth-3)
	}
	return labelWidth, textWidth
}

func firstSelectableDetailRow(rows []DetailRow) int {
	for i, row := range rows {
		if row.Selectable {
			return i
		}
	}
	return 0
}

func nextSelectableDetailRow(rows []DetailRow, selected int, delta int) int {
	if len(rows) == 0 || delta == 0 {
		return selected
	}
	step := 1
	if delta < 0 {
		step = -1
	}
	next := selected
	for remaining := abs(delta); remaining > 0; remaining-- {
		candidate := next
		for {
			candidate += step
			if candidate < 0 || candidate >= len(rows) {
				return next
			}
			if rows[candidate].Selectable {
				next = candidate
				break
			}
		}
	}
	return next
}

func detailLineCount(rows []DetailRow, textWidth int) int {
	total := 0
	for _, row := range rows {
		total += detailRowLineCount(row, textWidth)
		if detailDividerAfter(row) {
			total++
		}
	}
	return total
}

func detailRowLineStart(rows []DetailRow, rowIndex int, textWidth int) int {
	start := 0
	for i, row := range rows {
		if i == rowIndex {
			return start
		}
		start += detailRowLineCount(row, textWidth)
		if detailDividerAfter(row) {
			start++
		}
	}
	return start
}

func detailRowLineCount(row DetailRow, textWidth int) int {
	return len(detailRowTextLines(row, textWidth))
}

func selectableDetailRowAtOffset(rows []DetailRow, offset int, textWidth int) int {
	line := 0
	for i, row := range rows {
		count := detailRowLineCount(row, textWidth)
		if row.Selectable && line+count > offset {
			return i
		}
		line += count
		if detailDividerAfter(row) {
			line++
		}
	}
	for i := len(rows) - 1; i >= 0; i-- {
		if rows[i].Selectable {
			return i
		}
	}
	return 0
}

func detailRowTextLines(row DetailRow, textWidth int) []string {
	value := strings.TrimSpace(row.Text)
	if value == "" {
		value = "-"
	}
	parts := strings.Split(normalizeNewlines(value), "\n")
	textLines := make([]string, 0, len(parts))
	for _, part := range parts {
		wrapped := wrapWords(part, textWidth)
		if len(wrapped) == 0 {
			wrapped = []string{""}
		}
		textLines = append(textLines, wrapped...)
	}
	if detailRowShouldTruncate(row) && len(textLines) > 15 {
		textLines = append(append([]string(nil), textLines[:15]...), "...")
	}
	return textLines
}

func detailRowShouldTruncate(row DetailRow) bool {
	return row.Label == "Requirement" || row.Label == "Pull Request"
}

func detailDividerAfter(row DetailRow) bool {
	switch row.Label {
	case "Types", "Title", "Requirement", "Pull Request":
		return true
	default:
		return false
	}
}

func clampOffset(offset, selected, total, pageSize int) int {
	if total <= 0 {
		return 0
	}
	if pageSize < 1 {
		pageSize = 1
	}
	if selected < 0 {
		selected = 0
	}
	if selected >= total {
		selected = total - 1
	}
	if offset > selected {
		offset = selected
	}
	if selected >= offset+pageSize {
		offset = selected - pageSize + 1
	}
	maxOffset := total - pageSize
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		return 0
	}
	return offset
}

func clampLineOffset(offset, total, pageSize int) int {
	if total <= 0 {
		return 0
	}
	if pageSize < 1 {
		pageSize = 1
	}
	maxOffset := total - pageSize
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset > maxOffset {
		return maxOffset
	}
	if offset < 0 {
		return 0
	}
	return offset
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// FilteredRows returns changes matching active filters.
func FilteredRows(rows []dto.Change, filters Filters) []dto.Change {
	filtered := make([]dto.Change, 0, len(rows))
	find := strings.ToLower(strings.TrimSpace(filters.Find))
	for _, change := range rows {
		if filters.Phase.ID != "" && change.ChangePhase != filters.Phase.ID && change.ChangePhase != filters.Phase.Label {
			continue
		}
		if filters.Type.ID != "" && !hasChangeType(change, filters.Type.ID, filters.Type.Label) {
			continue
		}
		if filters.Epic.ID != "" && change.EpicID != filters.Epic.ID && change.EpicName != filters.Epic.Label {
			continue
		}
		if find != "" && !matchesFind(change, find) {
			continue
		}
		filtered = append(filtered, change)
	}
	return filtered
}

// ParseBody extracts backend fields while preserving the full body.
func ParseBody(body string, validTypes, epics []dto.Option) (ParsedRequirement, error) {
	parsed, err := ParseBodyStructure(body)
	if err != nil {
		return ParsedRequirement{}, err
	}
	validTypeSet := optionSet(validTypes)
	for _, typ := range parsed.ChangeTypes {
		if _, ok := validTypeSet[typ]; !ok {
			return ParsedRequirement{}, fmt.Errorf("invalid change type: %s", typ)
		}
	}
	if parsed.EpicName == "" {
		return parsed, nil
	}
	epicID, ok := resolveEpic(parsed.EpicName, epics)
	if !ok {
		return ParsedRequirement{}, fmt.Errorf("unknown epic: %s", parsed.EpicName)
	}
	parsed.EpicID = &epicID
	return parsed, nil
}

// ParseBodyStructure extracts locally validated metadata before reference lookups.
func ParseBodyStructure(body string) (ParsedRequirement, error) {
	normalized := strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(normalized, "\n")
	firstIndex := firstNonBlankLine(lines, 0)
	if firstIndex < 0 || !strings.HasPrefix(strings.TrimSpace(lines[firstIndex]), "# ") || strings.HasPrefix(strings.TrimSpace(lines[firstIndex]), "## ") {
		return ParsedRequirement{}, fmt.Errorf("requirement title is required")
	}
	title := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[firstIndex]), "# "))
	if title == "" {
		return ParsedRequirement{}, fmt.Errorf("requirement title is required")
	}

	typeIndex := firstNonBlankLine(lines, firstIndex+1)
	if typeIndex < 0 {
		return ParsedRequirement{}, fmt.Errorf("types line is required")
	}
	typeLine := strings.TrimSpace(lines[typeIndex])
	if !strings.HasPrefix(typeLine, "Types: ") {
		return ParsedRequirement{}, fmt.Errorf("types line is required")
	}
	typeValue := strings.TrimPrefix(typeLine, "Types: ")
	if strings.TrimSpace(typeValue) == "" || strings.Contains(typeValue, " ") {
		return ParsedRequirement{}, fmt.Errorf("types line must contain backend type slugs joined by |")
	}
	types := strings.Split(typeValue, "|")
	for _, typ := range types {
		if typ == "" {
			return ParsedRequirement{}, fmt.Errorf("types line must contain backend type slugs joined by |")
		}
	}

	parsed := ParsedRequirement{
		Title:       title,
		Body:        normalized,
		ChangeTypes: types,
	}
	epicIndex := firstNonBlankLine(lines, typeIndex+1)
	if epicIndex < 0 {
		return parsed, nil
	}
	epicLine := strings.TrimSpace(lines[epicIndex])
	if !strings.HasPrefix(epicLine, "Epic:") {
		return parsed, nil
	}
	epicName := strings.TrimSpace(strings.TrimPrefix(epicLine, "Epic:"))
	if epicName == "" {
		return parsed, nil
	}
	parsed.EpicName = epicName
	return parsed, nil
}

// RequirementEpicName returns the non-blank Epic metadata value when present.
func RequirementEpicName(body string) string {
	normalized := strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(normalized, "\n")
	firstIndex := firstNonBlankLine(lines, 0)
	if firstIndex < 0 {
		return ""
	}
	typeIndex := firstNonBlankLine(lines, firstIndex+1)
	if typeIndex < 0 {
		return ""
	}
	epicIndex := firstNonBlankLine(lines, typeIndex+1)
	if epicIndex < 0 {
		return ""
	}
	epicLine := strings.TrimSpace(lines[epicIndex])
	if !strings.HasPrefix(epicLine, "Epic:") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(epicLine, "Epic:"))
}

// RequirementMarkdown returns editable requirement markdown for a change.
func RequirementMarkdown(change dto.Change) string {
	body := strings.TrimSpace(change.Body)
	if body != "" && hasRequirementMetadata(body) {
		return requirementMarkdownWithBackendEpic(change.Body, change.EpicName)
	}
	var lines []string
	if strings.TrimSpace(change.Title) != "" {
		lines = append(lines, "# "+strings.TrimSpace(change.Title), "")
	}
	if len(change.ChangeTypes) > 0 {
		lines = append(lines, "Types: "+strings.Join(change.ChangeTypes, "|"), "")
	}
	if strings.TrimSpace(change.EpicName) != "" {
		lines = append(lines, "Epic: "+strings.TrimSpace(change.EpicName), "")
	}
	if body != "" {
		lines = append(lines, change.Body)
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

func requirementMarkdownWithBackendEpic(body, epicName string) string {
	epicName = strings.TrimSpace(epicName)
	if epicName == "" || hasRequirementEpicLine(body) {
		return body
	}
	normalized := strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(normalized, "\n")
	firstIndex := firstNonBlankLine(lines, 0)
	typeIndex := firstNonBlankLine(lines, firstIndex+1)
	epicIndex := firstNonBlankLine(lines, typeIndex+1)
	if epicIndex < 0 {
		epicIndex = len(lines)
	}
	insert := []string{}
	if epicIndex > 0 && strings.TrimSpace(lines[epicIndex-1]) != "" {
		insert = append(insert, "")
	}
	insert = append(insert, "Epic: "+epicName)
	if epicIndex < len(lines) && strings.TrimSpace(lines[epicIndex]) != "" {
		insert = append(insert, "")
	}
	lines = append(lines[:epicIndex], append(insert, lines[epicIndex:]...)...)
	return strings.Join(lines, "\n")
}

// SameTypes reports whether two type slices contain the same values in order.
func SameTypes(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func firstNonBlankLine(lines []string, start int) int {
	for i := start; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			return i
		}
	}
	return -1
}

func hasRequirementMetadata(body string) bool {
	normalized := strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(normalized, "\n")
	firstIndex := firstNonBlankLine(lines, 0)
	if firstIndex < 0 {
		return false
	}
	titleLine := strings.TrimSpace(lines[firstIndex])
	if !strings.HasPrefix(titleLine, "# ") || strings.HasPrefix(titleLine, "## ") || strings.TrimSpace(strings.TrimPrefix(titleLine, "# ")) == "" {
		return false
	}
	typeIndex := firstNonBlankLine(lines, firstIndex+1)
	if typeIndex < 0 {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(lines[typeIndex]), "Types: ")
}

func hasRequirementEpicLine(body string) bool {
	normalized := strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(normalized, "\n")
	firstIndex := firstNonBlankLine(lines, 0)
	if firstIndex < 0 {
		return false
	}
	typeIndex := firstNonBlankLine(lines, firstIndex+1)
	if typeIndex < 0 {
		return false
	}
	epicIndex := firstNonBlankLine(lines, typeIndex+1)
	if epicIndex < 0 {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(lines[epicIndex]), "Epic:")
}

// RequirementHasEpicLine reports whether the editable requirement metadata includes an Epic line.
func RequirementHasEpicLine(body string) bool {
	return hasRequirementEpicLine(body)
}

func optionSet(options []dto.Option) map[string]struct{} {
	values := make(map[string]struct{}, len(options)*2)
	for _, option := range options {
		if option.ID != "" {
			values[option.ID] = struct{}{}
		}
		if option.Label != "" {
			values[option.Label] = struct{}{}
		}
	}
	return values
}

func resolveEpic(name string, epics []dto.Option) (int, bool) {
	for _, epic := range epics {
		if strings.TrimSpace(epic.Label) != name {
			continue
		}
		id, err := strconv.Atoi(strings.TrimSpace(epic.ID))
		if err != nil || id <= 0 {
			return 0, false
		}
		return id, true
	}
	return 0, false
}

func hasChangeType(change dto.Change, values ...string) bool {
	for _, changeType := range change.ChangeTypes {
		for _, value := range values {
			if value != "" && changeType == value {
				return true
			}
		}
	}
	return false
}

func matchesFind(change dto.Change, query string) bool {
	values := []string{
		change.ID,
		change.Ref,
		displayRef(change),
		change.Slug,
		change.Title,
		change.ChangePhase,
		change.EpicID,
		change.EpicName,
		change.Body,
	}
	values = append(values, change.ChangeTypes...)
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), query) {
			return true
		}
	}
	return false
}

// SortedTypeOptions returns deterministic type options for tests and rendering.
func SortedTypeOptions(options []dto.Option) []dto.Option {
	sorted := append([]dto.Option(nil), options...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Label < sorted[j].Label
	})
	return sorted
}
