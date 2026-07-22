package agent

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Workspace stores paths for the temporary agent planning files.
type Workspace struct {
	Dir       string
	RootDir   string
	RefUUID   string
	Stage     string
	Operation WriteOperation
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
	if w.RootDir == "" {
		return filepath.Join(w.Dir, IdeaFileName)
	}
	return w.OutputPath()
}

// InputPath returns the baseline Idea artifact path.
func (w Workspace) InputPath() string { return filepath.Join(w.Dir, InputFileName) }

// OutputPath returns the editable and rewritten Idea artifact path.
func (w Workspace) OutputPath() string { return filepath.Join(w.Dir, OutputFileName) }

// GeneratedPath returns the generated Change spec path.
func (w Workspace) GeneratedPath() string {
	return filepath.Join(w.Dir, GeneratedFileName)
}

// AgentOutputPath returns the Codex final text output path.
func (w Workspace) AgentOutputPath() string {
	return filepath.Join(w.Dir, CodexOutputName)
}

// LogPath returns the Codex JSON event log path.
func (w Workspace) LogPath() string {
	return filepath.Join(w.Dir, CodexRunLogName)
}

// SessionPath returns the stage-local Codex session file path.
func (w Workspace) SessionPath() string { return filepath.Join(w.Dir, SessionFileName) }

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

// InitializeChange creates a new blank input/output stage and refuses file reuse.
func (w Workspace) InitializeChange() error {
	if strings.TrimSpace(w.RootDir) == "" || strings.TrimSpace(w.RefUUID) == "" {
		return fmt.Errorf("change workspace identity is required")
	}
	if err := os.MkdirAll(w.Dir, 0o755); err != nil {
		_ = w.RemoveChange()
		return err
	}
	for _, path := range []string{w.InputPath(), w.OutputPath()} {
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			_ = w.RemoveChange()
			return err
		}
		if err := file.Close(); err != nil {
			_ = w.RemoveChange()
			return err
		}
	}
	return nil
}

// EqualIdeaFiles reports whether the current editor pass changed output.
func (w Workspace) EqualIdeaFiles() (bool, error) {
	input, err := os.ReadFile(w.InputPath())
	if err != nil {
		return false, err
	}
	output, err := os.ReadFile(w.OutputPath())
	if err != nil {
		return false, err
	}
	return bytes.Equal(input, output), nil
}

// PromoteOutput makes the current output the next immutable input baseline.
func (w Workspace) PromoteOutput() error {
	output, err := os.ReadFile(w.OutputPath())
	if err != nil {
		return err
	}
	return os.WriteFile(w.InputPath(), output, 0o644)
}

// PrepareArtifact refreshes input/output from backend artifact text without
// reinitializing the Change workspace or touching its session files.
func (w Workspace) PrepareArtifact(content string) error {
	if strings.TrimSpace(w.RootDir) == "" || strings.TrimSpace(w.RefUUID) == "" {
		return fmt.Errorf("change workspace identity is required")
	}
	if err := w.Ensure(); err != nil {
		return err
	}
	if err := os.WriteFile(w.OutputPath(), []byte(content), 0o644); err != nil {
		return err
	}
	return w.PromoteOutput()
}

// ReadSessionID reads the existing stage-local Codex session when present.
func (w Workspace) ReadSessionID() (string, error) {
	content, err := os.ReadFile(w.SessionPath())
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

// WriteSessionID preserves the Codex session for later artifact-write operations.
func (w Workspace) WriteSessionID(sessionID string) error {
	if strings.TrimSpace(w.RootDir) == "" {
		return nil
	}
	if err := w.Ensure(); err != nil {
		return err
	}
	return os.WriteFile(w.SessionPath(), []byte(strings.TrimSpace(sessionID)+"\n"), 0o644)
}

// RemoveChange removes only this generated UUID workspace.
func (w Workspace) RemoveChange() error {
	root := strings.TrimSpace(w.RootDir)
	if root == "" {
		return fmt.Errorf("change workspace root is required")
	}
	return os.RemoveAll(root)
}

// WriteIdea replaces the idea file contents.
func (w Workspace) WriteIdea(content string) error {
	if err := w.Ensure(); err != nil {
		return err
	}
	return os.WriteFile(w.IdeaPath(), []byte(content), 0o644)
}

// RemoveIdea removes the idea file when it exists.
func (w Workspace) RemoveIdea() error {
	if err := os.Remove(w.IdeaPath()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ReadIdea reads the current idea file contents.
func (w Workspace) ReadIdea() (string, error) {
	content, err := os.ReadFile(w.IdeaPath())
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// ReadGenerated reads the generated Change spec contents.
func (w Workspace) ReadGenerated() (string, error) {
	content, err := os.ReadFile(w.GeneratedPath())
	if err != nil {
		return "", err
	}
	return string(content), nil
}
