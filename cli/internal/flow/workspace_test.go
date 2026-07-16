package flow

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceResolvesEverySupportedArtifactBelowConfiguredTempDir(t *testing.T) {
	tempDir := t.TempDir()
	for _, artifact := range []Artifact{
		ArtifactIdea,
		ArtifactSpec,
		ArtifactPR,
		ArtifactImplement,
		ArtifactReview,
		ArtifactFinalize,
	} {
		t.Run(string(artifact), func(t *testing.T) {
			workspace := Workspace{TempDir: tempDir, Artifact: artifact}
			directory, err := workspace.Directory()
			require.NoError(t, err)
			assert.Equal(t, filepath.Join(tempDir, string(artifact)), directory)
			session, err := workspace.SessionPath()
			require.NoError(t, err)
			assert.Equal(t, filepath.Join(directory, sessionFileName), session)
			input, err := workspace.InputPath()
			require.NoError(t, err)
			assert.Equal(t, filepath.Join(directory, inputFileName), input)
			output, err := workspace.OutputPath()
			require.NoError(t, err)
			assert.Equal(t, filepath.Join(directory, outputFileName), output)
			agentOutput, err := workspace.AgentOutputPath()
			require.NoError(t, err)
			assert.Equal(t, filepath.Join(directory, agentOutputFileName), agentOutput)
		})
	}
}

func TestWorkspaceRejectsMissingTempDirAndUnsupportedArtifact(t *testing.T) {
	_, err := (Workspace{Artifact: ArtifactIdea}).Directory()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "temp_dir")

	_, err = (Workspace{TempDir: t.TempDir(), Artifact: "unknown"}).Directory()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported artifact")
}

func TestFinalOutputLineUsesExactLastTerminalLine(t *testing.T) {
	assert.Equal(t, "Done.", finalOutputLine("progress\nDone.\n"))
	assert.Equal(t, " Done. ", finalOutputLine("progress\r\n Done. \r\n"))
	assert.Equal(t, "", finalOutputLine(""))
}
