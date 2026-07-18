package flow

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testRefUUID = "11111111-2222-4333-8444-555555555555"

func TestWorkspaceUsesRepositoryTmpDirAndChangeArtifactScope(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, TmpDir), 0o755))
	workspace := Workspace{Root: root, ChangeRef: testRefUUID, Artifact: ArtifactIdea}
	directory, err := workspace.Directory()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, TmpDir, testRefUUID, "idea"), directory)
	editorDirectory, err := (Workspace{Root: root, ChangeRef: testRefUUID, Artifact: ArtifactIdea, Scope: WorkspaceEditor}).Directory()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, TmpDir, testRefUUID, "idea", "editor"), editorDirectory)
	attemptUUID, path, err := CreateIdeaWorkspace(root)
	require.NoError(t, err)
	assert.Regexp(t, uuidPattern, attemptUUID)
	assert.Equal(t, filepath.Join(root, TmpDir, attemptUUID, "new-idea.md"), path)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Empty(t, content)
}

func TestCreateIdeaWorkspaceRetriesAnAtomicUUIDDirectoryCollision(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, TmpDir), 0o755))
	collisionUUID := "00000000-0000-4000-8000-000000000000"
	require.NoError(t, os.Mkdir(filepath.Join(root, TmpDir, collisionUUID), 0o755))
	random := bytes.NewReader(append(make([]byte, 16), bytes.Repeat([]byte{1}, 16)...))

	attemptUUID, path, err := createIdeaWorkspace(root, random)
	require.NoError(t, err)
	assert.Equal(t, "01010101-0101-4101-8101-010101010101", attemptUUID)
	assert.Equal(t, filepath.Join(root, TmpDir, attemptUUID, "new-idea.md"), path)
}

func TestCleanupIdeaWorkspaceRemovesFileThenOnlyAnEmptyDirectory(t *testing.T) {
	t.Run("empty workspace", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, TmpDir), 0o755))
		attemptUUID, path, err := CreateIdeaWorkspace(root)
		require.NoError(t, err)

		require.NoError(t, CleanupIdeaWorkspace(root, attemptUUID))
		_, err = os.Stat(path)
		assert.ErrorIs(t, err, os.ErrNotExist)
		_, err = os.Stat(filepath.Dir(path))
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("captured non-empty workspace", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, TmpDir), 0o755))
		attemptUUID, path, err := CreateIdeaWorkspace(root)
		require.NoError(t, err)
		captured := filepath.Join(filepath.Dir(path), "input.md")
		require.NoError(t, os.WriteFile(captured, []byte("captured"), 0o644))

		require.NoError(t, CleanupIdeaWorkspace(root, attemptUUID))
		_, err = os.Stat(path)
		assert.ErrorIs(t, err, os.ErrNotExist)
		content, err := os.ReadFile(captured)
		require.NoError(t, err)
		assert.Equal(t, "captured", string(content))
	})
}

func TestWorkspaceNeverCreatesMissingTempRoot(t *testing.T) {
	root := t.TempDir()
	workspace := Workspace{Root: root, ChangeRef: testRefUUID, Artifact: ArtifactIdea}

	err := workspace.replaceBaseline([]byte("# Idea\n\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
	_, statErr := os.Stat(filepath.Join(root, TmpDir))
	assert.ErrorIs(t, statErr, os.ErrNotExist)

	_, _, err = CreateIdeaWorkspace(root)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestWorkspaceRejectsInvalidContext(t *testing.T) {
	_, err := (Workspace{Root: t.TempDir(), ChangeRef: "bad", Artifact: ArtifactIdea}).Directory()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ref_uuid")

	_, err = (Workspace{Root: t.TempDir(), ChangeRef: testRefUUID, Artifact: ArtifactIdea, Scope: "unknown"}).Directory()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace scope")
}
