package flow

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStore struct {
	content    []byte
	loads      int
	saves      int
	provenance SaveProvenance
	saveRun    func() error
}

func (f *fakeStore) Load(int, Artifact) ([]byte, error) {
	f.loads++
	return append([]byte(nil), f.content...), nil
}

func (f *fakeStore) Save(_ int, _ Artifact, content []byte, provenance SaveProvenance) error {
	f.saves++
	f.content = append([]byte(nil), content...)
	f.provenance = provenance
	if f.saveRun != nil {
		return f.saveRun()
	}
	return nil
}

type fakeOperations struct {
	editorRun func(path string) error
	execRun   func(ExecRequest) error
	chatRun   func(ChatRequest) error
	execCalls int
	chatCalls int
}

func (f *fakeOperations) Editor(path, _ string, done func(error) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		if f.editorRun != nil {
			return done(f.editorRun(path))
		}
		return done(nil)
	}
}

func (f *fakeOperations) Exec(_ context.Context, request ExecRequest, done func(error) tea.Msg) tea.Cmd {
	f.execCalls++
	return func() tea.Msg {
		if f.execRun != nil {
			return done(f.execRun(request))
		}
		return done(nil)
	}
}

func (f *fakeOperations) Chat(_ context.Context, request ChatRequest, done func(error) tea.Msg) tea.Cmd {
	f.chatCalls++
	return func() tea.Msg {
		if f.chatRun != nil {
			return done(f.chatRun(request))
		}
		return done(nil)
	}
}

func (*fakeOperations) Preview(_ context.Context, path, _, _ string, done func(RenderResult, error) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		return done(RenderResult{Output: string(content)}, err)
	}
}

func (*fakeOperations) Diff(_ context.Context, input, output, _, _ string, done func(RenderResult, error) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		left, leftErr := os.ReadFile(input)
		right, rightErr := os.ReadFile(output)
		if leftErr != nil {
			return done(RenderResult{}, leftErr)
		}
		if rightErr != nil {
			return done(RenderResult{}, rightErr)
		}
		status := 0
		if string(left) != string(right) {
			status = 1
		}
		return done(RenderResult{Output: string(right), Status: status}, nil)
	}
}

func TestEditorCanonicalSaveNoOpAndValidationRecovery(t *testing.T) {
	t.Run("changed canonical output saves as user before Preview", func(t *testing.T) {
		store := &fakeStore{content: []byte("# Old\n\nBody")}
		operations := &fakeOperations{editorRun: func(path string) error {
			return os.WriteFile(path, []byte("# New\r\n\r\nTypes: test | feature\r\n\r\nBody"), 0o644)
		}}
		model, cmd := composeIdeaModel(t, IdeaEdit, ChangeDetailsTerminal, store, operations)
		for cmd != nil {
			model, cmd = drive(t, model, cmd)
		}
		assert.Equal(t, IdeaEditPreview, model.Screen())
		assert.Equal(t, SaveByUser, store.provenance)
		assert.Equal(t, "# New\n\nTypes: feature|test\n\nBody", string(store.content))
		assert.Equal(t, []CommandID{"/continue", "/edit", "/cancel"}, model.Commands())
		assert.Equal(t, WorkspaceArtifact, model.FlowContext().WorkspaceScope)
		artifactWorkspace := Workspace{Root: model.FlowContext().Root, ChangeRef: testRefUUID, Artifact: ArtifactIdea}
		artifactInputPath, err := artifactWorkspace.InputPath()
		require.NoError(t, err)
		artifactInput, err := os.ReadFile(artifactInputPath)
		require.NoError(t, err)
		artifactOutputPath, err := artifactWorkspace.OutputPath()
		require.NoError(t, err)
		artifactOutput, err := os.ReadFile(artifactOutputPath)
		require.NoError(t, err)
		assert.Equal(t, "# Old\n\nBody", string(artifactInput))
		assert.Equal(t, "# New\n\nTypes: feature|test\n\nBody", string(artifactOutput))
	})

	t.Run("canonical byte-identical output returns to caller", func(t *testing.T) {
		store := &fakeStore{content: []byte("# Same\n\nTypes: feature|test\n\nBody")}
		operations := &fakeOperations{editorRun: func(path string) error {
			return os.WriteFile(path, []byte("# Same\n\nTypes: test | feature\n\nBody"), 0o644)
		}}
		model, cmd := composeIdeaModel(t, IdeaEdit, ChangeDetailsTerminal, store, operations)
		for cmd != nil {
			model, cmd = drive(t, model, cmd)
		}
		assert.True(t, model.Done())
		assert.Equal(t, ChangeDetailsTerminal, model.TerminalScreen())
		assert.Zero(t, store.saves)
	})

	t.Run("publish failure enters Error after save", func(t *testing.T) {
		store := &fakeStore{content: []byte("# Old\n\nBody")}
		operations := &fakeOperations{editorRun: func(path string) error {
			return os.WriteFile(path, []byte("# New\n\nBody"), 0o644)
		}}
		model, cmd := composeIdeaModel(t, IdeaEdit, ChangeDetailsTerminal, store, operations)
		editorWorkspace := Workspace{
			Root: model.FlowContext().Root, ChangeRef: testRefUUID, Artifact: ArtifactIdea, Scope: WorkspaceEditor,
		}
		store.saveRun = func() error {
			outputPath, err := editorWorkspace.OutputPath()
			if err != nil {
				return err
			}
			return os.Remove(outputPath)
		}
		for index := 0; index < 4; index++ {
			model, cmd = drive(t, model, cmd)
		}
		assert.Equal(t, GenericErrorScreen, model.Screen())
		require.Error(t, model.Error())
		assert.Contains(t, model.Error().Error(), "read Editor output.md")
		assert.Equal(t, 1, store.saves)
	})

	t.Run("invalid output offers fix and cancel", func(t *testing.T) {
		store := &fakeStore{content: []byte("# Old\n\n")}
		operations := &fakeOperations{editorRun: func(path string) error {
			return os.WriteFile(path, []byte("not a title"), 0o644)
		}}
		model, cmd := composeIdeaModel(t, IdeaEdit, ChangeDetailsTerminal, store, operations)
		for index := 0; index < 4; index++ {
			model, cmd = drive(t, model, cmd)
		}
		assert.Equal(t, []CommandID{"/fix", "/cancel"}, model.Commands())
		require.Error(t, model.ValidationError())
		assert.Equal(t, "# Title parsing failed", model.ValidationError().Error())
		updated, _ := model.Update(CommandMsg{ID: "/cancel"})
		model = updated.(Model)
		assert.Equal(t, ChangeDetailsTerminal, model.TerminalScreen())
		assert.Zero(t, store.saves)
	})
}

func TestExecTraversesOrderedChatAndSynthesizesPreviewChat(t *testing.T) {
	t.Run("matching output skips Chat", func(t *testing.T) {
		store := &fakeStore{content: []byte("# Idea\n\nBody")}
		operations := &fakeOperations{execRun: func(request ExecRequest) error {
			require.Equal(t, TmpDir, commandTempDirForTest(request))
			return os.WriteFile(filepath.Join(request.Workspace, agentOutputFileName), []byte("Done.\n"), 0o644)
		}}
		model, cmd := composeIdeaModel(t, IdeaRewriteExec, ChangesListTerminal, store, operations)
		for index := 0; index < 5; index++ {
			model, cmd = drive(t, model, cmd)
		}
		assert.Equal(t, IdeaRewritePreview, model.Screen())
		assert.Zero(t, operations.chatCalls)
		assert.Equal(t, []CommandID{"/continue", "/chat", "/edit", "/cancel"}, model.Commands())
		updated, chatCommand := model.Update(CommandMsg{ID: "/chat"})
		model = updated.(Model)
		assert.Equal(t, IdeaRewriteChat, model.Screen())
		assert.Empty(t, model.Commands())
		assert.NotContains(t, model.View(), "/chat")
		updated, prematureCommand := model.Update(CommandMsg{ID: "/chat"})
		model = updated.(Model)
		assert.Nil(t, prematureCommand)
		assert.Zero(t, operations.chatCalls)
		model, _ = drive(t, model, chatCommand)
		assert.Equal(t, IdeaRewriteChat, model.Screen())
		assert.Equal(t, []CommandID{"/chat", "/edit", "/cancel"}, model.Commands())
		assert.Equal(t, IdeaRewriteExec, model.FlowContext().Step)
		assert.Equal(t, 1, model.FlowContext().TaskIndex)
		assert.Equal(t, 2, store.loads)
	})

	t.Run("unexpected output advances to Chat and Chat completion reaches Preview", func(t *testing.T) {
		store := &fakeStore{content: []byte("# Idea\n\nBody")}
		operations := &fakeOperations{
			execRun: func(request ExecRequest) error {
				require.NoError(t, os.WriteFile(filepath.Join(request.Workspace, agentOutputFileName), []byte("Needs work\n"), 0o644))
				return os.WriteFile(filepath.Join(request.Workspace, sessionFileName), []byte(testRefUUID), 0o644)
			},
			chatRun: func(request ChatRequest) error {
				return os.WriteFile(filepath.Join(request.Workspace, outputFileName), []byte("# Idea\n\nImproved"), 0o644)
			},
		}
		model, cmd := composeIdeaModel(t, IdeaRewriteExec, ChangesListTerminal, store, operations)
		for index := 0; index < 4; index++ {
			model, cmd = drive(t, model, cmd)
		}
		assert.Equal(t, IdeaRewriteChat, model.Screen())
		assert.Empty(t, model.Commands())
		assert.NotContains(t, model.View(), "/chat")
		model, _ = drive(t, model, cmd)
		assert.Equal(t, []CommandID{"/chat", "/edit", "/cancel"}, model.Commands())
		assert.Contains(t, model.View(), "Needs work")
		updated, cmd := model.Update(CommandMsg{ID: "/chat"})
		model = updated.(Model)
		for index := 0; index < 4; index++ {
			model, cmd = drive(t, model, cmd)
		}
		assert.Equal(t, IdeaRewritePreview, model.Screen())
		assert.Equal(t, SaveByAgent, store.provenance)
		assert.Equal(t, "# Idea\n\nImproved", string(store.content))
	})

	t.Run("empty readable output advances to Chat", func(t *testing.T) {
		store := &fakeStore{content: []byte("# Idea\n\nBody")}
		operations := &fakeOperations{execRun: func(request ExecRequest) error {
			return os.WriteFile(filepath.Join(request.Workspace, agentOutputFileName), nil, 0o644)
		}}
		model, cmd := composeIdeaModel(t, IdeaRewriteExec, ChangesListTerminal, store, operations)
		for index := 0; index < 5; index++ {
			model, cmd = drive(t, model, cmd)
		}
		assert.Equal(t, IdeaRewriteChat, model.Screen())
	})

	t.Run("byte-identical Edit preserves the exact Preview workspace", func(t *testing.T) {
		store := &fakeStore{content: []byte("# Idea\n\nOriginal")}
		operations := &fakeOperations{execRun: func(request ExecRequest) error {
			require.NoError(t, os.WriteFile(filepath.Join(request.Workspace, outputFileName), []byte("# Idea\n\nRewritten"), 0o644))
			return os.WriteFile(filepath.Join(request.Workspace, agentOutputFileName), []byte("Done.\n"), 0o644)
		}}
		model, cmd := composeIdeaModel(t, IdeaRewriteExec, ChangesListTerminal, store, operations)
		for index := 0; index < 5; index++ {
			model, cmd = drive(t, model, cmd)
		}
		require.Equal(t, IdeaRewritePreview, model.Screen())
		artifactWorkspace := Workspace{Root: model.FlowContext().Root, ChangeRef: testRefUUID, Artifact: ArtifactIdea}

		updated, cmd := model.Update(CommandMsg{ID: "/edit"})
		model = updated.(Model)
		for cmd != nil {
			model, cmd = drive(t, model, cmd)
		}

		assert.Equal(t, IdeaRewritePreview, model.Screen())
		assert.Equal(t, WorkspaceArtifact, model.FlowContext().WorkspaceScope)
		assert.Equal(t, 1, store.saves)
		inputPath, err := artifactWorkspace.InputPath()
		require.NoError(t, err)
		input, err := os.ReadFile(inputPath)
		require.NoError(t, err)
		outputPath, err := artifactWorkspace.OutputPath()
		require.NoError(t, err)
		output, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Equal(t, "# Idea\n\nOriginal", string(input))
		assert.Equal(t, "# Idea\n\nRewritten", string(output))

		updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyLeft})
		model = updated.(Model)
		model, _ = drive(t, model, cmd)
		assert.Equal(t, PreviewDiff, model.Mode())
	})

	t.Run("byte-identical Edit restores the exact Chat caller", func(t *testing.T) {
		store := &fakeStore{content: []byte("# Idea\n\nBody")}
		operations := &fakeOperations{execRun: func(request ExecRequest) error {
			return os.WriteFile(filepath.Join(request.Workspace, agentOutputFileName), []byte("Needs work"), 0o644)
		}}
		model, cmd := composeIdeaModel(t, IdeaRewriteExec, ChangesListTerminal, store, operations)
		for index := 0; index < 5; index++ {
			model, cmd = drive(t, model, cmd)
		}
		updated, cmd := model.Update(CommandMsg{ID: "/edit"})
		model = updated.(Model)
		for cmd != nil {
			model, cmd = drive(t, model, cmd)
		}
		assert.Equal(t, IdeaRewriteChat, model.Screen())
		assert.Equal(t, IdeaRewriteExec, model.FlowContext().Step)
		assert.Equal(t, TaskID("IdeaRewriteChatTask"), model.FlowContext().Task)
		assert.Equal(t, 1, model.FlowContext().TaskIndex)
		assert.Contains(t, model.View(), "Needs work")
		assert.Zero(t, store.saves)
	})
}

func TestPendingOperationsBlockCommandsAndIgnoreStaleReplies(t *testing.T) {
	t.Run("Chat save blocks overlapping commands", func(t *testing.T) {
		store := &fakeStore{content: []byte("# Idea\n\nBody")}
		operations := &fakeOperations{
			execRun: func(request ExecRequest) error {
				require.NoError(t, os.WriteFile(filepath.Join(request.Workspace, agentOutputFileName), []byte("Needs work\n"), 0o644))
				return os.WriteFile(filepath.Join(request.Workspace, sessionFileName), []byte(testRefUUID), 0o644)
			},
			chatRun: func(request ChatRequest) error {
				return os.WriteFile(filepath.Join(request.Workspace, outputFileName), []byte("# Idea\n\nImproved"), 0o644)
			},
		}
		model, command := composeIdeaModel(t, IdeaRewriteExec, ChangesListTerminal, store, operations)
		for index := 0; index < 5; index++ {
			model, command = drive(t, model, command)
		}

		updated, command := model.Update(CommandMsg{ID: "/chat"})
		model = updated.(Model)
		model, chatCommand := drive(t, model, command)
		assert.Empty(t, model.Commands())
		require.NotNil(t, chatCommand)

		chatResult := chatCommand()
		updated, saveCommand := model.Update(chatResult)
		model = updated.(Model)
		require.NotNil(t, saveCommand)
		assert.Empty(t, model.Commands())
		assert.Zero(t, store.saves)
		assert.Equal(t, 1, operations.chatCalls)

		for _, commandID := range []CommandID{"/chat", "/edit", "/cancel"} {
			updated, overlappingCommand := model.Update(CommandMsg{ID: commandID})
			model = updated.(Model)
			assert.Nil(t, overlappingCommand)
		}
		assert.Equal(t, IdeaRewriteChat, model.Screen())
		assert.Equal(t, 1, operations.chatCalls)

		updated, staleCommand := model.Update(chatResult)
		model = updated.(Model)
		assert.Nil(t, staleCommand)
		assert.Zero(t, store.saves)

		saveResult := saveCommand()
		updated, renderCommand := model.Update(saveResult)
		model = updated.(Model)
		assert.Equal(t, IdeaRewritePreview, model.Screen())
		assert.Empty(t, model.Commands())
		assert.Equal(t, 1, store.saves)

		updated, staleCommand = model.Update(saveResult)
		model = updated.(Model)
		assert.Nil(t, staleCommand)
		assert.Equal(t, 1, store.saves)

		model, _ = drive(t, model, renderCommand)
		assert.Equal(t, []CommandID{"/continue", "/chat", "/edit", "/cancel"}, model.Commands())
	})

	t.Run("Preview destination load blocks commands and stale load replies", func(t *testing.T) {
		store := &fakeStore{content: []byte("# Idea\n\nBody")}
		operations := &fakeOperations{execRun: func(request ExecRequest) error {
			return os.WriteFile(filepath.Join(request.Workspace, agentOutputFileName), []byte("Done.\n"), 0o644)
		}}
		model, command := composeIdeaModel(t, IdeaRewriteExec, ChangesListTerminal, store, operations)
		for command != nil {
			model, command = drive(t, model, command)
		}

		updated, loadCommand := model.Update(CommandMsg{ID: "/continue"})
		model = updated.(Model)
		require.NotNil(t, loadCommand)
		assert.Empty(t, model.Commands())

		updated, overlappingCommand := model.Update(CommandMsg{ID: "/edit"})
		model = updated.(Model)
		assert.Nil(t, overlappingCommand)
		assert.Equal(t, IdeaRewritePreview, model.Screen())

		loaded := loadCommand()
		updated, execCommand := model.Update(loaded)
		model = updated.(Model)
		require.NotNil(t, execCommand)
		assert.Equal(t, IdeaReviewExec, model.FlowContext().Step)
		assert.Equal(t, 2, operations.execCalls)

		updated, staleCommand := model.Update(loaded)
		model = updated.(Model)
		assert.Nil(t, staleCommand)
		assert.Equal(t, IdeaReviewExec, model.FlowContext().Step)
		assert.Equal(t, 2, operations.execCalls)
	})
}

func TestSameStepChatEditorDestinationRemainsRepresentable(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, TmpDir), 0o755))
	store := &fakeStore{content: []byte("# PR\n\nBody")}
	operations := &fakeOperations{}
	model := Compose(Composition{
		Definition: ConformanceDefinition(),
		Context: Context{
			Root: root, FlowDir: root, ChangeID: 1, ChangeRef: testRefUUID,
			Origin: ChangeDetailsTerminal, Step: "pr-chat",
		},
		TerminalScreens: ideaTerminals,
		Store:           store,
		Options:         &fakeDocumentOptions{},
		Operations:      operations,
	})
	require.NoError(t, model.Error())
	model, command := drive(t, model, model.Init())
	model, _ = drive(t, model, command)
	assert.Equal(t, ScreenID("pr-chat"), model.Screen())

	updated, command := model.Update(CommandMsg{ID: "/edit"})
	model = updated.(Model)
	for command != nil {
		model, command = drive(t, model, command)
	}
	assert.Equal(t, ScreenID("pr-chat"), model.Screen())
	assert.Equal(t, StepID("pr-chat"), model.FlowContext().Step)
	assert.Zero(t, store.saves)
}

func composeIdeaModel(t *testing.T, step StepID, caller ScreenID, store ArtifactStore, operations Operations) (Model, tea.Cmd) {
	t.Helper()
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, TmpDir), 0o755))
	model := Compose(Composition{
		Definition:      IdeaDefinition(),
		Context:         Context{Root: root, FlowDir: root, ChangeID: 1, ChangeRef: testRefUUID, Origin: ChangesListTerminal, Step: step, EditorCaller: caller},
		TerminalScreens: ideaTerminals,
		Store:           store,
		Options:         &fakeDocumentOptions{types: []TypeOption{{Slug: "feature"}, {Slug: "test"}}, epics: []EpicOption{{ID: 1, Title: "Epic"}}},
		Operations:      operations,
	})
	require.NoError(t, model.Error())
	return model, model.Init()
}

func drive(t *testing.T, model Model, cmd tea.Cmd) (Model, tea.Cmd) {
	t.Helper()
	require.NotNil(t, cmd)
	updated, next := model.Update(cmd())
	return updated.(Model), next
}

func commandTempDirForTest(ExecRequest) string { return TmpDir }
