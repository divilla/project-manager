package app

import (
	"bytes"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRunVersionPrintsVersion(t *testing.T) {
	var out bytes.Buffer

	if err := Run([]string{"--version"}, &out); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "mch") {
		t.Fatalf("version output missing executable name: %q", got)
	}
	if !strings.Contains(got, Version) {
		t.Fatalf("version output missing version %q: %q", Version, got)
	}
}

func TestNewModelStartupState(t *testing.T) {
	m := NewModel()

	if m.screen != ScreenReady {
		t.Fatalf("screen = %q, want %q", m.screen, ScreenReady)
	}
	if !m.input.Focused() {
		t.Fatal("input should start focused")
	}
}

func TestQuitKeyTransitionsToDone(t *testing.T) {
	m := NewModel()

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	got := updated.(Model)

	if got.screen != ScreenDone {
		t.Fatalf("screen = %q, want %q", got.screen, ScreenDone)
	}
	if !got.quitting {
		t.Fatal("model should mark quit state")
	}
	if cmd == nil {
		t.Fatal("quit key should return a command")
	}
}

func TestViewContainsBaselineCopy(t *testing.T) {
	view := NewModel().View()

	for _, want := range []string{"mch", Version, "Hello World", "q quit"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "Make a Change") {
		t.Fatalf("view should not show formal name:\n%s", view)
	}
}
