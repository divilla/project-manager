package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Workspace stores paths for the temporary agent planning files.
type Workspace struct {
	Dir string
}

// Ensure creates the workspace directory, replacing a regular file at that path.
func (w Workspace) Ensure() error {
	if strings.TrimSpace(w.Dir) == "" {
		return fmt.Errorf("agent workspace directory is required")
	}
	info, err := os.Stat(w.Dir)
	if err == nil {
		if info.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s exists and is not a directory", w.Dir)
		}
		if err := os.Remove(w.Dir); err != nil {
			return err
		}
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.MkdirAll(w.Dir, 0o755)
}

// IdeaPath returns the Markdown idea file path.
func (w Workspace) IdeaPath() string {
	return filepath.Join(w.Dir, IdeaFileName)
}

// GeneratedPath returns the generated Change file path.
func (w Workspace) GeneratedPath() string {
	return filepath.Join(w.Dir, GeneratedFileName)
}

// OutputPath returns the Codex final text output path.
func (w Workspace) OutputPath() string {
	return filepath.Join(w.Dir, CodexOutputName)
}

// LogPath returns the Codex JSON event log path.
func (w Workspace) LogPath() string {
	return filepath.Join(w.Dir, CodexRunLogName)
}

// IdeaExists reports whether the idea file already exists.
func (w Workspace) IdeaExists() (bool, error) {
	_, err := os.Stat(w.IdeaPath())
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// ResetIdea creates or replaces the idea file with an empty Markdown file.
func (w Workspace) ResetIdea() error {
	if err := w.Ensure(); err != nil {
		return err
	}
	return os.WriteFile(w.IdeaPath(), []byte{}, 0o644)
}

// ReadIdea reads the current idea file contents.
func (w Workspace) ReadIdea() (string, error) {
	content, err := os.ReadFile(w.IdeaPath())
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// ReadGenerated reads the generated Change file contents.
func (w Workspace) ReadGenerated() (string, error) {
	content, err := os.ReadFile(w.GeneratedPath())
	if err != nil {
		return "", err
	}
	return string(content), nil
}
