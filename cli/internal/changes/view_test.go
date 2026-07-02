package changes

import (
	"regexp"
	"strings"
	"testing"

	"mch/internal/dto"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestDetailsViewSeparatesBodyAndTestCases(t *testing.T) {
	model := Model{}.WithDetail(dto.Change{
		ID:    "12",
		Ref:   "3",
		Title: "Backend Change",
		Body:  "Body text",
		TestCases: []dto.TestCase{
			{ID: "31", Scenario: "first scenario", Done: true},
		},
	})

	view := stripANSI(DetailsView(model, 120, 20))

	bodyIndex := strings.Index(view, "Body │ Body text")
	dividerIndex := strings.Index(view[bodyIndex:], "───────────┼")
	testCaseIndex := strings.Index(view, "✅ │ first scenario (#31)")
	require.NotEqual(t, -1, bodyIndex)
	require.NotEqual(t, -1, dividerIndex)
	require.NotEqual(t, -1, testCaseIndex)
	assert.Less(t, bodyIndex, bodyIndex+dividerIndex)
	assert.Less(t, bodyIndex+dividerIndex, testCaseIndex)
}

func TestDetailsViewEmojiRowsDoNotOverflowSelectionWidth(t *testing.T) {
	change := dto.Change{
		ID:      "12",
		Ref:     "3",
		Title:   "Backend Change",
		Body:    "Body text",
		Open:    true,
		Created: "2026-06-29T08:15:00Z",
		TestCases: []dto.TestCase{
			{ID: "31", Scenario: "first scenario", Done: true},
			{ID: "32", Scenario: "second scenario", Done: false},
		},
	}

	for _, selected := range []int{7, 13} {
		view := DetailsView(Model{Detail: change, DetailSelected: selected}, 120, 20)
		for _, line := range strings.Split(view, "\n") {
			assert.LessOrEqual(t, lipgloss.Width(line), 120)
		}
	}
}

func stripANSI(value string) string {
	return ansiPattern.ReplaceAllString(value, "")
}
