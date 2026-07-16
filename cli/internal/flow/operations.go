package flow

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ExecRequest describes a configured non-interactive external operation.
type ExecRequest struct {
	Script    string
	Prompt    string
	FlowDir   string
	Workspace string
	ChangeID  int
	ChangeRef string
	Artifact  Artifact
}

// InteractiveRequest describes a configured terminal session handoff.
type InteractiveRequest struct {
	Script    string
	FlowDir   string
	Workspace string
	SessionID string
	ChangeID  int
	ChangeRef string
	Artifact  Artifact
}

// RenderResult contains rendered terminal text and the source command status.
type RenderResult struct {
	Output string
	Status int
}

// Operations provides fakeable external editor, process, session, and rendering boundaries.
type Operations interface {
	Editor(path string, workingDirectory string, done func(error) tea.Msg) tea.Cmd
	Exec(ctx context.Context, request ExecRequest, done func(error) tea.Msg) tea.Cmd
	Interactive(ctx context.Context, request InteractiveRequest, done func(error) tea.Msg) tea.Cmd
	Preview(ctx context.Context, path string, workingDirectory string, theme string, done func(RenderResult, error) tea.Msg) tea.Cmd
	Diff(ctx context.Context, inputPath string, outputPath string, workingDirectory string, theme string, done func(RenderResult, error) tea.Msg) tea.Cmd
}

// ProcessOperations executes the real external commands used by composed Flows.
type ProcessOperations struct{}

// NewProcessOperations creates process-backed external operation boundaries.
func NewProcessOperations() ProcessOperations {
	return ProcessOperations{}
}

// Editor opens the configured editor with Bubble Tea terminal handoff.
func (ProcessOperations) Editor(path string, workingDirectory string, done func(error) tea.Msg) tea.Cmd {
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	var command *exec.Cmd
	if editor == "" {
		command = exec.Command("nano", path)
	} else {
		command = exec.Command("sh", "-c", "$EDITOR \"$1\"", "mch-flow-editor", path)
		command.Env = append(os.Environ(), "EDITOR="+editor)
	}
	command.Dir = workingDirectory
	return tea.ExecProcess(command, done)
}

// Exec runs the configured script without blocking the Bubble Tea update loop.
func (ProcessOperations) Exec(ctx context.Context, request ExecRequest, done func(error) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		script, err := resolveDefinitionPath(request.FlowDir, request.Script)
		if err != nil {
			return done(err)
		}
		prompt, err := resolveDefinitionPath(request.FlowDir, request.Prompt)
		if err != nil {
			return done(err)
		}
		command := exec.CommandContext(ctx, script, prompt)
		configureCommand(command, request.FlowDir, request.Workspace, request.ChangeID, request.ChangeRef, request.Artifact)
		configureExecProcessGroup(command)
		output, err := command.CombinedOutput()
		if err != nil {
			return done(commandError("exec", err, output))
		}
		return done(nil)
	}
}

// Interactive hands terminal control to the configured session-resume script.
func (ProcessOperations) Interactive(ctx context.Context, request InteractiveRequest, done func(error) tea.Msg) tea.Cmd {
	script, err := resolveDefinitionPath(request.FlowDir, request.Script)
	if err != nil {
		return func() tea.Msg { return done(err) }
	}
	command := exec.CommandContext(ctx, script)
	configureCommand(command, request.FlowDir, request.Workspace, request.ChangeID, request.ChangeRef, request.Artifact)
	command.Env = append(command.Env, "MCH_SESSION_ID="+request.SessionID)
	return tea.ExecProcess(command, done)
}

// Preview renders an artifact with bat.
func (ProcessOperations) Preview(ctx context.Context, path string, workingDirectory string, theme string, done func(RenderResult, error) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		command := exec.CommandContext(ctx, "bat", "-pp", "--theme", renderTheme(theme), path)
		command.Dir = workingDirectory
		output, err := command.CombinedOutput()
		if err != nil {
			return done(RenderResult{}, commandError("preview", err, output))
		}
		return done(RenderResult{Output: string(output), Status: 0}, nil)
	}
}

// Diff captures Git's status before rendering its output with bat.
func (ProcessOperations) Diff(ctx context.Context, inputPath string, outputPath string, workingDirectory string, theme string, done func(RenderResult, error) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		gitCommand := exec.CommandContext(ctx, "git", "--no-pager", "diff", "--no-index", "--no-ext-diff", "--color=never", "--", inputPath, outputPath)
		gitCommand.Dir = workingDirectory
		var diff bytes.Buffer
		var gitErrorOutput bytes.Buffer
		gitCommand.Stdout = &diff
		gitCommand.Stderr = &gitErrorOutput
		gitErr := gitCommand.Run()
		status := commandStatus(gitErr)
		if status > 1 {
			return done(RenderResult{Status: status}, commandError("git diff", gitErr, gitErrorOutput.Bytes()))
		}
		batCommand := exec.CommandContext(ctx, "bat", "-pp", "--theme", renderTheme(theme), "--language", "diff")
		batCommand.Dir = workingDirectory
		batCommand.Stdin = bytes.NewReader(diff.Bytes())
		rendered, batErr := batCommand.CombinedOutput()
		if batErr != nil {
			return done(RenderResult{Status: status}, commandError("diff preview", batErr, rendered))
		}
		return done(RenderResult{Output: string(rendered), Status: status}, nil)
	}
}

func renderTheme(theme string) string {
	if strings.TrimSpace(theme) == "" {
		return "Coldark-Dark"
	}
	return theme
}

func resolveDefinitionPath(flowDirectory string, path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("definition path is required")
	}
	if filepath.IsAbs(path) {
		return path, nil
	}
	if strings.TrimSpace(flowDirectory) == "" {
		return "", fmt.Errorf("active Flow directory is required for relative path %q", path)
	}
	return filepath.Join(flowDirectory, path), nil
}

func configureCommand(command *exec.Cmd, flowDirectory string, workspace string, changeID int, changeRef string, artifact Artifact) {
	command.Dir = flowDirectory
	command.Env = append(os.Environ(),
		"MCH_CHANGE_ID="+strconv.Itoa(changeID),
		"MCH_REF_UUID="+changeRef,
		"MCH_ARTIFACT="+string(artifact),
		"MCH_TEMP_DIR="+workspace,
	)
}

func commandStatus(err error) int {
	if err == nil {
		return 0
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		status := exitError.ExitCode()
		if status >= 0 {
			return status
		}
	}
	return 2
}

func commandError(operation string, err error, output []byte) error {
	message := strings.TrimSpace(string(output))
	if message == "" {
		return fmt.Errorf("%s failed: %w", operation, err)
	}
	return fmt.Errorf("%s failed: %w: %s", operation, err, message)
}
