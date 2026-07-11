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

// PhaseColors maps backend phase slugs to optional Lip Gloss color values.
type PhaseColors map[string]string

var defaultPhaseColors = PhaseColors{
	"backlog":    "15",
	"progress":   "10",
	"review":     "11",
	"staging":    "12",
	"production": "13",
	"rejected":   "9",
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
	Label        string
	Text         string
	Selectable   bool
	DividerAfter bool
	TestCaseID   string
	TestCaseText string
	TestCaseDone bool
}

// ParsedSpec stores metadata extracted from spec markdown.
type ParsedSpec struct {
	Title       string
	Spec        string
	ChangeTypes []string
	EpicID      *int
	EpicName    string
}

// ParsedIdea stores the title and full idea text extracted from idea markdown.
type ParsedIdea struct {
	Title string
	Idea  string
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
	m.DetailSelected = firstSelectableDetailSelection(change)
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
	next := nextSelectableDetailSelection(m.Detail, rows, m.DetailSelected, delta)
	m.DetailSelected = next
	if m.DetailSelected < 0 {
		_, textWidth := DetailColumnWidths(m.Detail, width)
		m.DetailOffset = clampLineOffset(m.DetailOffset, detailLineCount(rows, textWidth), detailScrollPageSize(m.Detail, pageSize, width))
		return m
	}
	_, textWidth := DetailColumnWidths(m.Detail, width)
	rowStart := detailRowLineStart(rows, m.DetailSelected, textWidth)
	rowEnd := rowStart + detailRowLineCount(rows[m.DetailSelected], textWidth)
	m.DetailOffset = detailOffsetKeepingRowVisible(m.DetailOffset, rowStart, rowEnd, detailLineCount(rows, textWidth), detailScrollPageSize(m.Detail, pageSize, width))
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
	if !validDetailSelection(m.Detail, rows, m.DetailSelected) {
		m.DetailSelected = firstSelectableDetailSelection(m.Detail)
	}
	_, textWidth := DetailColumnWidths(m.Detail, width)
	m.DetailOffset = clampLineOffset(m.DetailOffset, detailLineCount(rows, textWidth), detailScrollPageSize(m.Detail, pageSize, width))
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
	scrollPageSize := detailScrollPageSize(m.Detail, pageSize, width)
	if abs(delta) >= pageSize {
		if delta < 0 {
			delta = -scrollPageSize
		} else {
			delta = scrollPageSize
		}
	}
	m.DetailOffset = clampLineOffset(m.DetailOffset+delta, detailLineCount(rows, textWidth), scrollPageSize)
	m.DetailSelected = selectableDetailRowAtOffset(rows, m.DetailOffset, textWidth)
	return m
}

// SelectDetailRow returns the currently selected editable detail row.
func (m Model) SelectDetailRow(pageSize int, width int) (Model, DetailRow, bool) {
	m = m.ClampDetailSelection(pageSize, width)
	rows := DetailRows(m.Detail)
	row, ok := detailRowForSelection(m.Detail, rows, m.DetailSelected)
	if !ok {
		return m, DetailRow{}, false
	}
	return m, row, true
}

// DetailRows returns Change details as label/text table rows.
func DetailRows(change dto.Change) []DetailRow {
	if change.ID == "" && change.Title == "" {
		return nil
	}
	rows := []DetailRow{
		{Label: "Ref", Text: displayRef(change), Selectable: true},
		{Label: "Slug", Text: change.Slug, Selectable: true},
		{Label: "Phase", Text: change.ChangePhase, Selectable: true},
		{Label: "Epic", Text: epicLabel(change), Selectable: true},
		{Label: "Types", Text: strings.Join(change.ChangeTypes, "|"), Selectable: true, DividerAfter: true},
		{Label: "Title", Text: change.Title, Selectable: true, DividerAfter: true},
		{Label: "Idea", Text: change.Idea, Selectable: true, DividerAfter: true},
		{Label: "Spec", Text: change.Spec, Selectable: true, DividerAfter: true},
	}
	for i, testCase := range change.TestCases {
		rows = append(rows, DetailRow{
			Label:        testCaseDoneIcon(testCase.Done),
			Text:         fmt.Sprintf("%s (#%s)", testCase.Scenario, testCase.ID),
			Selectable:   true,
			DividerAfter: i == len(change.TestCases)-1,
			TestCaseID:   testCase.ID,
			TestCaseText: testCase.Scenario,
			TestCaseDone: testCase.Done,
		})
	}
	rows = append(rows,
		DetailRow{Label: "PR", Text: change.PR, Selectable: true, DividerAfter: true},
		DetailRow{Label: "PR URL", Text: change.PRUrl, Selectable: true},
		DetailRow{Label: "Agent Edit", Text: agentEditIcon(change.AgentEdit), Selectable: true},
		DetailRow{Label: "Complete", Text: fmt.Sprintf("%d/%d - %d%%", change.Done, change.Total, change.Completed), Selectable: true},
		DetailRow{Label: "Open", Text: testCaseDoneIcon(change.Open), Selectable: true},
		DetailRow{Label: "Created", Text: formatListTimestamp(change.Created), Selectable: true},
		DetailRow{Label: "Modified", Text: formatListTimestamp(change.Modified), Selectable: true},
	)
	return rows
}

func fixedDetailRows(change dto.Change) []DetailRow {
	return []DetailRow{
		{Label: "ID", Text: change.ID, Selectable: true},
		{Label: "Ref UUID", Text: change.RefUUID, Selectable: true},
	}
}

// DetailCopyValue returns the value copied for a selected detail row.
func DetailCopyValue(row DetailRow) string {
	return row.Text
}

func firstSelectableDetailSelection(change dto.Change) int {
	fixedRows := fixedDetailRows(change)
	for i, row := range fixedRows {
		if row.Selectable {
			return i - len(fixedRows)
		}
	}
	return firstSelectableDetailRow(DetailRows(change))
}

func validDetailSelection(change dto.Change, rows []DetailRow, selected int) bool {
	_, ok := detailRowForSelection(change, rows, selected)
	return ok
}

func detailRowForSelection(change dto.Change, rows []DetailRow, selected int) (DetailRow, bool) {
	if selected < 0 {
		fixedRows := fixedDetailRows(change)
		index := selected + len(fixedRows)
		if index < 0 || index >= len(fixedRows) || !fixedRows[index].Selectable {
			return DetailRow{}, false
		}
		return fixedRows[index], true
	}
	if selected >= len(rows) || !rows[selected].Selectable {
		return DetailRow{}, false
	}
	return rows[selected], true
}

func selectableDetailSelections(change dto.Change, rows []DetailRow) []int {
	fixedRows := fixedDetailRows(change)
	selections := make([]int, 0, len(fixedRows)+len(rows))
	for i, row := range fixedRows {
		if row.Selectable {
			selections = append(selections, i-len(fixedRows))
		}
	}
	for i, row := range rows {
		if row.Selectable {
			selections = append(selections, i)
		}
	}
	return selections
}

func nextSelectableDetailSelection(change dto.Change, rows []DetailRow, selected int, delta int) int {
	selections := selectableDetailSelections(change, rows)
	if len(selections) == 0 || delta == 0 {
		return selected
	}
	position := -1
	for i, value := range selections {
		if value == selected {
			position = i
			break
		}
	}
	if position < 0 {
		return selections[0]
	}
	next := position + delta
	if next < 0 {
		next = 0
	}
	if next >= len(selections) {
		next = len(selections) - 1
	}
	return selections[next]
}

func detailScrollPageSize(change dto.Change, pageSize int, width int) int {
	if pageSize < 1 {
		pageSize = 1
	}
	_, textWidth := DetailColumnWidths(change, width)
	for _, row := range fixedDetailRows(change) {
		pageSize -= detailRowLineCount(row, textWidth)
	}
	if pageSize < 1 {
		return 1
	}
	return pageSize
}

func agentEditIcon(value bool) string {
	if value {
		return "\u2714"
	}
	return "\u2718"
}

func testCaseDoneIcon(value bool) string {
	if value {
		return "\u2705"
	}
	return "\u274c"
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
	return row.Label == "Idea" || row.Label == "Spec" || row.Label == "PR"
}

func detailDividerAfter(row DetailRow) bool {
	return row.DividerAfter
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

func detailOffsetKeepingRowVisible(offset, rowStart, rowEnd, total, pageSize int) int {
	offset = clampLineOffset(offset, total, pageSize)
	if pageSize < 1 {
		pageSize = 1
	}
	if rowEnd <= rowStart {
		rowEnd = rowStart + 1
	}
	if rowStart < offset {
		return clampLineOffset(rowStart, total, pageSize)
	}
	if rowEnd > offset+pageSize {
		return clampLineOffset(rowEnd-pageSize, total, pageSize)
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

// ParseIdeaStructure extracts the Change title and idea text.
func ParseIdeaStructure(idea string) (ParsedIdea, error) {
	normalized := strings.ReplaceAll(strings.ReplaceAll(idea, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(normalized, "\n")
	firstIndex := firstNonBlankLine(lines, 0)
	if firstIndex < 0 || !strings.HasPrefix(strings.TrimSpace(lines[firstIndex]), "# ") || strings.HasPrefix(strings.TrimSpace(lines[firstIndex]), "## ") {
		return ParsedIdea{}, fmt.Errorf("idea title is required")
	}
	title := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[firstIndex]), "# "))
	if title == "" {
		return ParsedIdea{}, fmt.Errorf("idea title is required")
	}
	if strings.TrimSpace(normalized) == "" {
		return ParsedIdea{}, fmt.Errorf("idea is required")
	}
	return ParsedIdea{Title: title, Idea: normalized}, nil
}

// ParseSpec extracts backend fields while preserving the full spec.
func ParseSpec(spec string, validTypes, epics []dto.Option) (ParsedSpec, error) {
	parsed, err := ParseSpecStructure(spec)
	if err != nil {
		return ParsedSpec{}, err
	}
	validTypeSet := optionSet(validTypes)
	for _, typ := range parsed.ChangeTypes {
		if _, ok := validTypeSet[typ]; !ok {
			return ParsedSpec{}, fmt.Errorf("invalid change type: %s", typ)
		}
	}
	if parsed.EpicName == "" {
		return parsed, nil
	}
	epicID, ok := resolveEpic(parsed.EpicName, epics)
	if !ok {
		return ParsedSpec{}, fmt.Errorf("unknown epic: %s", parsed.EpicName)
	}
	parsed.EpicID = &epicID
	return parsed, nil
}

// ParseSpecStructure extracts locally validated metadata before reference lookups.
func ParseSpecStructure(spec string) (ParsedSpec, error) {
	normalized := strings.ReplaceAll(strings.ReplaceAll(spec, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(normalized, "\n")
	firstIndex := firstNonBlankLine(lines, 0)
	if firstIndex < 0 || !strings.HasPrefix(strings.TrimSpace(lines[firstIndex]), "# ") || strings.HasPrefix(strings.TrimSpace(lines[firstIndex]), "## ") {
		return ParsedSpec{}, fmt.Errorf("spec title is required")
	}
	title := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[firstIndex]), "# "))
	if title == "" {
		return ParsedSpec{}, fmt.Errorf("spec title is required")
	}

	types := []string{}
	metadataEndIndex := firstIndex
	typeIndex := firstNonBlankLine(lines, firstIndex+1)
	if typeIndex >= 0 {
		typeLine := strings.TrimSpace(lines[typeIndex])
		if strings.HasPrefix(typeLine, "Types:") {
			metadataEndIndex = typeIndex
			typeValue := strings.TrimPrefix(typeLine, "Types:")
			if strings.TrimSpace(typeValue) != "" {
				if !strings.HasPrefix(typeValue, " ") || strings.Contains(strings.TrimPrefix(typeValue, " "), " ") {
					return ParsedSpec{}, fmt.Errorf("types line must contain backend type slugs joined by |")
				}
				types = strings.Split(strings.TrimPrefix(typeValue, " "), "|")
				for _, typ := range types {
					if typ == "" {
						return ParsedSpec{}, fmt.Errorf("types line must contain backend type slugs joined by |")
					}
				}
			}
		}
	}

	parsed := ParsedSpec{
		Title:       title,
		Spec:        normalized,
		ChangeTypes: types,
	}
	epicIndex := firstNonBlankLine(lines, metadataEndIndex+1)
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

// SpecEpicName returns the non-blank Epic metadata value when present.
func SpecEpicName(spec string) string {
	normalized := strings.ReplaceAll(strings.ReplaceAll(spec, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(normalized, "\n")
	firstIndex := firstNonBlankLine(lines, 0)
	if firstIndex < 0 {
		return ""
	}
	epicIndex := firstMetadataLineAfterTypes(lines, firstIndex)
	if epicIndex < 0 {
		return ""
	}
	epicLine := strings.TrimSpace(lines[epicIndex])
	if !strings.HasPrefix(epicLine, "Epic:") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(epicLine, "Epic:"))
}

// SpecMarkdown returns editable spec markdown for a change.
func SpecMarkdown(change dto.Change) string {
	spec := strings.TrimSpace(change.Spec)
	if spec != "" && hasSpecMetadata(spec) {
		return specMarkdownWithBackendEpic(change.Spec, change.EpicName)
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
	if spec != "" {
		lines = append(lines, change.Spec)
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

func specMarkdownWithBackendEpic(spec, epicName string) string {
	epicName = strings.TrimSpace(epicName)
	if epicName == "" || hasSpecEpicLine(spec) {
		return spec
	}
	normalized := strings.ReplaceAll(strings.ReplaceAll(spec, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(normalized, "\n")
	firstIndex := firstNonBlankLine(lines, 0)
	epicIndex := firstMetadataLineAfterTypes(lines, firstIndex)
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

func hasSpecMetadata(spec string) bool {
	normalized := strings.ReplaceAll(strings.ReplaceAll(spec, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(normalized, "\n")
	firstIndex := firstNonBlankLine(lines, 0)
	if firstIndex < 0 {
		return false
	}
	titleLine := strings.TrimSpace(lines[firstIndex])
	if !strings.HasPrefix(titleLine, "# ") || strings.HasPrefix(titleLine, "## ") || strings.TrimSpace(strings.TrimPrefix(titleLine, "# ")) == "" {
		return false
	}
	return true
}

func hasSpecEpicLine(spec string) bool {
	normalized := strings.ReplaceAll(strings.ReplaceAll(spec, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(normalized, "\n")
	firstIndex := firstNonBlankLine(lines, 0)
	if firstIndex < 0 {
		return false
	}
	epicIndex := firstMetadataLineAfterTypes(lines, firstIndex)
	if epicIndex < 0 {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(lines[epicIndex]), "Epic:")
}

func firstMetadataLineAfterTypes(lines []string, titleIndex int) int {
	nextIndex := firstNonBlankLine(lines, titleIndex+1)
	if nextIndex < 0 {
		return -1
	}
	if strings.HasPrefix(strings.TrimSpace(lines[nextIndex]), "Types:") {
		return firstNonBlankLine(lines, nextIndex+1)
	}
	return nextIndex
}

// SpecHasEpicLine reports whether the editable spec metadata includes an Epic line.
func SpecHasEpicLine(spec string) bool {
	return hasSpecEpicLine(spec)
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
		change.RefUUID,
		change.Ref,
		displayRef(change),
		change.Slug,
		change.Title,
		change.ChangePhase,
		change.EpicID,
		change.EpicName,
		change.Idea,
		change.Spec,
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
