package flow

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	sessionFileName     = "session-id"
	inputFileName       = "input.md"
	outputFileName      = "output.md"
	agentOutputFileName = "agent-output.md"
)

// Workspace resolves artifact-scoped runtime resources beneath a configured temp_dir.
type Workspace struct {
	TempDir  string
	Artifact Artifact
}

// Directory returns the artifact workspace directory.
func (w Workspace) Directory() (string, error) {
	if strings.TrimSpace(w.TempDir) == "" {
		return "", fmt.Errorf("flow context temp_dir is required")
	}
	if !supportedArtifact(w.Artifact) {
		return "", fmt.Errorf("unsupported artifact %q", w.Artifact)
	}
	return filepath.Join(w.TempDir, string(w.Artifact)), nil
}

// SessionPath returns the session-id resource path.
func (w Workspace) SessionPath() (string, error) {
	return w.resourcePath(sessionFileName)
}

// InputPath returns the immutable Step baseline path.
func (w Workspace) InputPath() (string, error) {
	return w.resourcePath(inputFileName)
}

// OutputPath returns the mutable artifact output path.
func (w Workspace) OutputPath() (string, error) {
	return w.resourcePath(outputFileName)
}

// AgentOutputPath returns the external execution output path.
func (w Workspace) AgentOutputPath() (string, error) {
	return w.resourcePath(agentOutputFileName)
}

func (w Workspace) resourcePath(name string) (string, error) {
	directory, err := w.Directory()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, name), nil
}

func (w Workspace) replaceBaseline(content []byte) error {
	directory, err := w.Directory()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create artifact workspace %q: %w", directory, err)
	}
	input, err := w.InputPath()
	if err != nil {
		return err
	}
	output, err := w.OutputPath()
	if err != nil {
		return err
	}
	if err := os.WriteFile(input, content, 0o644); err != nil {
		return fmt.Errorf("write input.md: %w", err)
	}
	if err := os.WriteFile(output, content, 0o644); err != nil {
		return fmt.Errorf("write output.md: %w", err)
	}
	return nil
}

func (w Workspace) compare(expectedBaseline []byte) ([]byte, bool, error) {
	input, err := w.InputPath()
	if err != nil {
		return nil, false, err
	}
	output, err := w.OutputPath()
	if err != nil {
		return nil, false, err
	}
	baseline, err := os.ReadFile(input)
	if err != nil {
		return nil, false, fmt.Errorf("read input.md: %w", err)
	}
	if !bytes.Equal(baseline, expectedBaseline) {
		return nil, false, fmt.Errorf("input.md changed during Step")
	}
	current, err := os.ReadFile(output)
	if err != nil {
		return nil, false, fmt.Errorf("read output.md: %w", err)
	}
	return current, bytes.Equal(baseline, current), nil
}

func (w Workspace) readSession() (string, error) {
	path, err := w.SessionPath()
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read session-id: %w", err)
	}
	session := strings.TrimSpace(string(content))
	if session == "" {
		return "", fmt.Errorf("session-id is empty")
	}
	return session, nil
}

func (w Workspace) readAgentOutput(required bool) (string, error) {
	path, err := w.AgentOutputPath()
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		if !required && os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read agent-output.md: %w", err)
	}
	return string(content), nil
}
