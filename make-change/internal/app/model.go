package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"mch/internal/styles"
)

type Screen string

const (
	ScreenReady Screen = "ready"
	ScreenDone  Screen = "done"
)

type Model struct {
	input    textinput.Model
	screen   Screen
	width    int
	quitting bool
}

func NewModel() Model {
	input := textinput.New()
	input.Placeholder = "Type a note or press q"
	input.Prompt = "> "
	input.Focus()
	input.CharLimit = 240
	input.Width = 48
	input.PromptStyle = styles.Default.AccentCyan
	input.TextStyle = styles.Default.Foreground
	input.PlaceholderStyle = styles.Default.Muted
	input.Cursor.Style = styles.Default.AccentPurple

	return Model{
		input:  input,
		screen: ScreenReady,
		width:  80,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.input.Width = clampWidth(msg.Width-8, 16, 96)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.screen = ScreenDone
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	width := clampWidth(m.width, 32, 120)
	body := []string{
		styles.Default.Title.Render("mch"),
		styles.Default.Muted.Render("version " + Version),
		"",
		styles.Default.Foreground.Render("Hello World from mch."),
		"",
		m.inputBand(width),
		styles.Default.Footer.Width(width).Render("q quit  |  status ready"),
	}
	if m.quitting {
		body = append(body, styles.Default.Success.Render("done"))
	}
	return styles.Default.Surface.Width(width).Render(strings.Join(body, "\n"))
}

func (m Model) inputBand(width int) string {
	content := m.input.View()
	return styles.Default.InputBand.Width(width).Render(content)
}

func clampWidth(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

var _ tea.Model = Model{}
