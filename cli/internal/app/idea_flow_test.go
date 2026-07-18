package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"mch/internal/dto"
	"mch/internal/flow"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeFlowOperations struct {
	editorRun func(string) error
}

func (f fakeFlowOperations) Editor(path, _ string, done func(error) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		if f.editorRun != nil {
			return done(f.editorRun(path))
		}
		return done(nil)
	}
}

func (fakeFlowOperations) Exec(_ context.Context, _ flow.ExecRequest, done func(error) tea.Msg) tea.Cmd {
	return func() tea.Msg { return done(nil) }
}

func (fakeFlowOperations) Chat(_ context.Context, _ flow.ChatRequest, done func(error) tea.Msg) tea.Cmd {
	return func() tea.Msg { return done(nil) }
}

func (fakeFlowOperations) Preview(_ context.Context, path, _, _ string, done func(flow.RenderResult, error) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		return done(flow.RenderResult{Output: string(content)}, err)
	}
}

func (fakeFlowOperations) Diff(_ context.Context, _, output, _, _ string, done func(flow.RenderResult, error) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(output)
		return done(flow.RenderResult{Output: string(content), Status: 1}, err)
	}
}

func TestIdeaCreateCreatesAnIsolatedFreshWorkspace(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, flow.TmpDir), 0o755))
	m := newModelWithConfig(&fakeClient{}, testAppConfig(appConfig{RepositoryRoot: root, FlowDir: filepath.Join(root, ".mch", "default")}))
	m.state = ChangesListState
	m.currentProject = dto.Option{ID: "7"}

	updated, command := m.beginIdeaCreate()
	m = updated.(Model)
	require.NotNil(t, command)
	assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`, m.ideaCreateAttempt.uuid)
	assert.Equal(t, filepath.Join(root, flow.TmpDir, m.ideaCreateAttempt.uuid, "new-idea.md"), m.ideaCreateAttempt.path)
	content, err := os.ReadFile(m.ideaCreateAttempt.path)
	require.NoError(t, err)
	assert.Empty(t, content)
	assert.Equal(t, CreateIdeaState, m.state)
}

func TestIdeaCreateCanonicalizesBeforeConfirmation(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, flow.TmpDir), 0o755))
	attemptUUID, path, err := flow.CreateIdeaWorkspace(root)
	require.NoError(t, err)
	client := &fakeClient{types: []dto.Option{{ID: "feature"}, {ID: "test"}}}
	message := validateIdeaCreateCommand(context.Background(), attemptUUID, client, "7", path, []byte("# Idea\r\n\r\nTypes: test | feature\r\n\r\nBody"))().(ideaCreateValidatedMsg)
	require.NoError(t, message.err)
	m := newModelWithConfig(client, testAppConfig(appConfig{RepositoryRoot: root}))
	m.state = IdeaProcessingState
	m.ideaCreateAttempt = ideaCreateAttempt{uuid: attemptUUID, path: path}

	updated, command := m.Update(message)
	m = updated.(Model)
	require.Nil(t, command)
	assert.Equal(t, "Create Change?", m.dropdown.label)
	assert.Equal(t, "Idea", m.ideaCreateTitle)
	assert.Equal(t, "# Idea\n\nTypes: feature|test\n\nBody", string(m.ideaCreateBytes))
}

func TestIdeaCreateValidationUsesCancelableProcessingState(t *testing.T) {
	t.Run("shows animated Processing state with only cancel", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, flow.TmpDir), 0o755))
		attemptUUID, path, err := flow.CreateIdeaWorkspace(root)
		require.NoError(t, err)
		m := newModelWithConfig(&fakeClient{}, testAppConfig(appConfig{RepositoryRoot: root}))
		m.state = CreateIdeaState
		m.ideaCreateAttempt = ideaCreateAttempt{uuid: attemptUUID, path: path}

		updated, command := m.handleIdeaCreateEditorFinished(editorFinishedMsg{content: "# Idea\n\nBody"})
		m = updated.(Model)

		require.NotNil(t, command)
		assert.Equal(t, IdeaProcessingState, m.state)
		assert.Equal(t, []string{"/cancel"}, commandsByState[m.state])
		assert.Contains(t, stripANSI(m.View()), "| Processing...")

		updated, tickCommand := m.Update(m.spinner.Tick())
		m = updated.(Model)
		require.NotNil(t, tickCommand)
		assert.Contains(t, stripANSI(m.View()), "/ Processing...")
	})

	t.Run("ignores a result after cancellation", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, flow.TmpDir), 0o755))
		attemptUUID, path, err := flow.CreateIdeaWorkspace(root)
		require.NoError(t, err)
		m := newModelWithConfig(&fakeClient{}, testAppConfig(appConfig{RepositoryRoot: root}))
		m.state = IdeaProcessingState
		canceled := false
		m.ideaCreateAttempt = ideaCreateAttempt{uuid: attemptUUID, path: path, cancel: func() { canceled = true }}

		updated, _ := m.executeCommandFrom(IdeaProcessingState, "/cancel")
		m = updated.(Model)
		updated, command := m.Update(ideaCreateValidatedMsg{attemptUUID: attemptUUID, content: []byte("# Stale\n\nBody"), title: "Stale"})
		m = updated.(Model)

		require.Nil(t, command)
		assert.True(t, canceled)
		assert.Empty(t, m.ideaCreateAttempt.uuid)
		assert.Equal(t, ChangesListState, m.state)
		assert.Empty(t, m.ideaCreateBytes)
		assert.Empty(t, m.ideaCreateTitle)
		assert.Empty(t, m.dropdown.kind)
		_, err = os.Stat(path)
		assert.ErrorIs(t, err, os.ErrNotExist)
		_, err = os.Stat(filepath.Dir(path))
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("ignores a result from a superseded validation", func(t *testing.T) {
		m := newModelWithConfig(&fakeClient{}, testAppConfig(appConfig{}))
		m.state = IdeaProcessingState
		m.ideaCreateAttempt = ideaCreateAttempt{uuid: "22222222-2222-4222-8222-222222222222"}

		updated, command := m.Update(ideaCreateValidatedMsg{attemptUUID: "11111111-1111-4111-8111-111111111111", content: []byte("# Stale\n\nBody"), title: "Stale"})
		m = updated.(Model)

		require.Nil(t, command)
		assert.Equal(t, IdeaProcessingState, m.state)
		assert.Empty(t, m.ideaCreateBytes)
		assert.Empty(t, m.ideaCreateTitle)
	})

	t.Run("a superseded validation can only rewrite its own workspace", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, flow.TmpDir), 0o755))
		oldUUID, oldPath, err := flow.CreateIdeaWorkspace(root)
		require.NoError(t, err)
		currentUUID, currentPath, err := flow.CreateIdeaWorkspace(root)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(currentPath, []byte("# Current\n\nCurrent body"), 0o644))
		m := newModelWithConfig(&fakeClient{}, testAppConfig(appConfig{RepositoryRoot: root}))
		m.state = IdeaProcessingState
		m.ideaCreateAttempt = ideaCreateAttempt{uuid: currentUUID, path: currentPath}

		message := validateIdeaCreateCommand(context.Background(), oldUUID, m.client, "7", oldPath, []byte("# Old\n\nOld body"))().(ideaCreateValidatedMsg)
		updated, command := m.Update(message)
		m = updated.(Model)

		require.Nil(t, command)
		assert.Equal(t, currentUUID, m.ideaCreateAttempt.uuid)
		current, err := os.ReadFile(currentPath)
		require.NoError(t, err)
		assert.Equal(t, "# Current\n\nCurrent body", string(current))
		old, err := os.ReadFile(oldPath)
		require.NoError(t, err)
		assert.Equal(t, "# Old\n\nOld body", string(old))
	})
}

func TestSuccessfulIdeaCreateStartsRewriteWithListOrigin(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, flow.TmpDir), 0o755))
	change := dto.Change{ID: "12", RefUUID: "11111111-2222-4333-8444-555555555555", Idea: "# Idea\n\nBody"}
	client := &fakeClient{createdChange: change, gotChange: change}
	m := newModelWithConfig(client, testAppConfig(appConfig{RepositoryRoot: root, FlowDir: filepath.Join(root, ".mch", "default")}))
	m.currentProject = dto.Option{ID: "7"}
	m.state = CreateIdeaState
	attemptUUID, path, err := flow.CreateIdeaWorkspace(root)
	require.NoError(t, err)
	m.ideaCreateAttempt = ideaCreateAttempt{uuid: attemptUUID, path: path}

	message := createChangeForIdeaFlowCommand(client, attemptUUID, 7, "Idea", []byte(change.Idea))().(changeCreatedForIdeaFlowMsg)
	require.NoError(t, message.err)
	updated, command := m.Update(message)
	m = updated.(Model)
	require.NotNil(t, command)
	assert.True(t, m.ideaFlowActive)
	assert.Equal(t, State(flow.IdeaRewriteExec), m.state)
	assert.Equal(t, flow.ChangesListTerminal, m.ideaFlow.FlowContext().Origin)
	assert.Equal(t, flow.IdeaRewriteExec, m.ideaFlow.FlowContext().Step)
	assert.Equal(t, []dto.ChangeCreateInput{{ProjectID: 7, Title: "Idea", Idea: change.Idea}}, client.changeCreateInputs)
	_, err = os.Stat(path)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestIdeaCreateAttemptCorrelationRejectsAStaleCreateReply(t *testing.T) {
	m := newModelWithConfig(&fakeClient{}, testAppConfig(appConfig{}))
	m.state = CreateIdeaState
	m.ideaCreateAttempt = ideaCreateAttempt{uuid: "22222222-2222-4222-8222-222222222222"}

	updated, command := m.Update(changeCreatedForIdeaFlowMsg{
		attemptUUID: "11111111-1111-4111-8111-111111111111",
		change:      dto.Change{ID: "12", RefUUID: "33333333-3333-4333-8333-333333333333"},
	})
	m = updated.(Model)

	require.Nil(t, command)
	assert.False(t, m.ideaFlowActive)
	assert.Equal(t, "22222222-2222-4222-8222-222222222222", m.ideaCreateAttempt.uuid)
}

func TestIdeaCreateConfirmationNoCleansItsAttemptWorkspace(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, flow.TmpDir), 0o755))
	attemptUUID, path, err := flow.CreateIdeaWorkspace(root)
	require.NoError(t, err)
	m := newModelWithConfig(&fakeClient{}, testAppConfig(appConfig{RepositoryRoot: root}))
	m.state = CreateIdeaState
	m.ideaCreateAttempt = ideaCreateAttempt{uuid: attemptUUID, path: path}
	updated, _ := m.handleIdeaCreateValidated(ideaCreateValidatedMsg{attemptUUID: attemptUUID, content: []byte("# Idea\n\nBody"), title: "Idea"})
	m = updated.(Model)
	m.dropdown.filter = "/no"

	updated, command := m.confirmDropdown()
	m = updated.(Model)

	require.NotNil(t, command)
	assert.Equal(t, ChangesListState, m.state)
	assert.Empty(t, m.ideaCreateAttempt.uuid)
	_, err = os.Stat(path)
	assert.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(filepath.Dir(path))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestIdeaCreateOperationalFailuresUseFlowErrorAndReturnToOrigin(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, flow.TmpDir), 0o755))
	attemptUUID, path, err := flow.CreateIdeaWorkspace(root)
	require.NoError(t, err)
	m := newModelWithConfig(&fakeClient{}, testAppConfig(appConfig{RepositoryRoot: root}))
	m.state = CreateIdeaState
	m.ideaCreateAttempt = ideaCreateAttempt{uuid: attemptUUID, path: path}

	updated, command := m.handleIdeaCreateValidated(ideaCreateValidatedMsg{attemptUUID: attemptUUID, err: errors.New("option API unavailable")})
	m = updated.(Model)
	require.Nil(t, command)
	assert.Equal(t, FlowErrorState, m.state)
	assert.Equal(t, ChangesListState, m.flowErrorOrigin)
	assert.Equal(t, []string{"/return"}, commandsByState[m.state])
	assert.Contains(t, m.View(), "option API unavailable")
	_, err = os.Stat(path)
	assert.ErrorIs(t, err, os.ErrNotExist)

	updated, command = m.executeCommandFrom(FlowErrorState, "/return")
	m = updated.(Model)
	require.NotNil(t, command)
	assert.Equal(t, ChangesListState, m.state)
	assert.Empty(t, m.err)
}

func TestIdeaCreateBackendFailureUsesFlowError(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, flow.TmpDir), 0o755))
	attemptUUID, path, err := flow.CreateIdeaWorkspace(root)
	require.NoError(t, err)
	m := newModelWithConfig(&fakeClient{}, testAppConfig(appConfig{RepositoryRoot: root}))
	m.state = CreateIdeaState
	m.ideaCreateAttempt = ideaCreateAttempt{uuid: attemptUUID, path: path}

	updated, command := m.Update(changeCreatedForIdeaFlowMsg{attemptUUID: attemptUUID, err: errors.New("create unavailable")})
	m = updated.(Model)
	require.Nil(t, command)
	assert.Equal(t, FlowErrorState, m.state)
	assert.Equal(t, ChangesListState, m.flowErrorOrigin)
	assert.Contains(t, m.err, "create unavailable")
	_, err = os.Stat(path)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestIdeaEditUsesGenericFlowAndUserSaveProvenance(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, flow.TmpDir), 0o755))
	client := &fakeClient{
		types:     []dto.Option{{ID: "feature"}},
		gotChange: dto.Change{ID: "12", RefUUID: "11111111-2222-4333-8444-555555555555", Idea: "# Old\n\nBody"},
	}
	m := newModelWithConfig(client, testAppConfig(appConfig{RepositoryRoot: root, FlowDir: filepath.Join(root, ".mch", "default")}))
	m.currentProject = dto.Option{ID: "7"}
	m.changeList = m.changeList.WithDetail(client.gotChange)
	m.flowOperations = fakeFlowOperations{editorRun: func(path string) error {
		return os.WriteFile(path, []byte("# New\n\nTypes: feature\n\nBody"), 0o644)
	}}

	updated, command := m.beginIdeaEdit()
	m = updated.(Model)
	for index := 0; index < 4; index++ {
		require.NotNil(t, command)
		updated, command = m.Update(command())
		m = updated.(Model)
	}
	assert.True(t, m.ideaFlowActive)
	assert.Equal(t, State(flow.IdeaEditPreview), m.state)
	assert.Equal(t, []string{"# New\n\nTypes: feature\n\nBody"}, client.changeIdeaUpdates)
	assert.Equal(t, []bool{false}, client.changeIdeaAgentEdits)
}

func TestSpecSaveCanonicalizesArtifactWithoutMutatingChangeFields(t *testing.T) {
	client := &fakeClient{
		types:     []dto.Option{{ID: "feature"}, {ID: "test"}},
		epics:     []dto.Option{{ID: "11", Label: "Canonical Epic"}},
		gotChange: dto.Change{ID: "12", Title: "Change Title", ChangeTypes: []string{"ci"}, EpicID: "9"},
	}
	original := client.gotChange
	message := changeUpdateCommand(client, 12, "7", original, "# Artifact Title\n\nEpic: stale #0011\n\nTypes: test | feature\n\nBody", nil)().(changeSavedMsg)
	require.NoError(t, message.err)
	assert.Equal(t, []string{"# Artifact Title\n\nTypes: feature|test\n\nEpic: Canonical Epic #11\n\nBody"}, client.changeSpecUpdates)
	assert.Empty(t, client.changeTitleUpdates)
	assert.Empty(t, client.changeTypesUpdates)
	assert.Empty(t, client.changeEpicUpdates)
}
