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
	Rows     []dto.Change
	Selected int
	Offset   int
	Detail   dto.Change
	Loading  bool
}

// ParsedRequirement stores metadata extracted from requirement markdown.
type ParsedRequirement struct {
	Title           string
	RequirementBody string
	ChangeTypes     []string
	EpicID          *int
	EpicName        string
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
	m.Detail = selected
	return m, selected, true
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

// ParseRequirementBody extracts backend fields while preserving the full body.
func ParseRequirementBody(body string, validTypes, epics []dto.Option) (ParsedRequirement, error) {
	parsed, err := ParseRequirementBodyStructure(body)
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

// ParseRequirementBodyStructure extracts locally validated metadata before reference lookups.
func ParseRequirementBodyStructure(body string) (ParsedRequirement, error) {
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
		Title:           title,
		RequirementBody: normalized,
		ChangeTypes:     types,
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
	body := strings.TrimSpace(change.RequirementBody)
	if body != "" && hasRequirementMetadata(body) {
		return requirementMarkdownWithBackendEpic(change.RequirementBody, change.EpicName)
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
		lines = append(lines, change.RequirementBody)
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
		change.RequirementBody,
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
