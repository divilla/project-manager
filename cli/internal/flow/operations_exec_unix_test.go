//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package flow

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type execProcessFinishedMsg struct {
	err error
}

func TestProcessOperationsExecCancellationTerminatesChildProcesses(t *testing.T) {
	flowDir := t.TempDir()
	workspace := t.TempDir()
	promptPath := filepath.Join(flowDir, "prompt.md")
	scriptPath := filepath.Join(flowDir, "exec.sh")
	require.NoError(t, os.WriteFile(promptPath, []byte("prompt"), 0o644))
	require.NoError(t, os.WriteFile(scriptPath, []byte(`#!/bin/sh
set -eu
(
	while :; do
		printf x >> "${MCH_TEMP_DIR}/heartbeat"
		sleep 0.05
	done
) >/dev/null 2>&1 &
printf '%s\n' "$!" > "${MCH_TEMP_DIR}/child-pid"
wait "$!"
`), 0o755))

	ctx, cancel := context.WithCancel(context.Background())
	command := NewProcessOperations().Exec(ctx, ExecRequest{
		Script:    filepath.Base(scriptPath),
		Prompt:    filepath.Base(promptPath),
		FlowDir:   flowDir,
		Workspace: workspace,
		ChangeID:  42,
		ChangeRef: "ref-42",
		Artifact:  ArtifactSpec,
	}, func(err error) tea.Msg {
		return execProcessFinishedMsg{err: err}
	})
	result := make(chan tea.Msg, 1)
	go func() {
		result <- command()
	}()

	childPIDPath := filepath.Join(workspace, "child-pid")
	heartbeatPath := filepath.Join(workspace, "heartbeat")
	require.Eventually(t, func() bool {
		_, childErr := os.Stat(childPIDPath)
		_, heartbeatErr := os.Stat(heartbeatPath)
		return childErr == nil && heartbeatErr == nil
	}, 2*time.Second, 10*time.Millisecond)
	childPIDBytes, err := os.ReadFile(childPIDPath)
	require.NoError(t, err)
	childPID, err := strconv.Atoi(strings.TrimSpace(string(childPIDBytes)))
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = syscall.Kill(childPID, syscall.SIGKILL)
	})

	cancel()
	var message tea.Msg
	require.Eventually(t, func() bool {
		select {
		case message = <-result:
			return true
		default:
			return false
		}
	}, 2*time.Second, 10*time.Millisecond)
	finished, ok := message.(execProcessFinishedMsg)
	require.True(t, ok)
	require.Error(t, finished.err)

	before, err := os.Stat(heartbeatPath)
	require.NoError(t, err)
	time.Sleep(200 * time.Millisecond)
	after, err := os.Stat(heartbeatPath)
	require.NoError(t, err)
	assert.Equal(t, before.Size(), after.Size(), "child process continued after Exec cancellation")
}
