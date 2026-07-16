package flow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type storeSave struct {
	changeID int
	artifact Artifact
	content  []byte
}

type fakeStore struct {
	data    map[Artifact][]byte
	loads   []Artifact
	saves   []storeSave
	loadErr error
	saveErr error
	events  *[]string
}

func (s *fakeStore) Load(_ int, artifact Artifact) ([]byte, error) {
	s.loads = append(s.loads, artifact)
	if s.loadErr != nil {
		return nil, s.loadErr
	}
	return append([]byte(nil), s.data[artifact]...), nil
}

func (s *fakeStore) Save(changeID int, artifact Artifact, content []byte) error {
	if s.events != nil {
		*s.events = append(*s.events, "save")
	}
	if s.saveErr != nil {
		return s.saveErr
	}
	s.saves = append(s.saves, storeSave{changeID: changeID, artifact: artifact, content: append([]byte(nil), content...)})
	return nil
}

type fakeOperations struct {
	editorContent      []byte
	editorErr          error
	editorCalls        int
	execRun            func(ExecRequest) error
	execCalls          int
	execContext        context.Context
	execRequest        ExecRequest
	interactiveRun     func(InteractiveRequest) error
	interactiveErr     error
	interactiveCalls   int
	interactiveRequest InteractiveRequest
	previewResult      RenderResult
	previewErr         error
	previewCalls       int
	diffResult         RenderResult
	diffErr            error
	diffCalls          int
	events             *[]string
}

func (o *fakeOperations) Editor(path string, _ string, done func(error) tea.Msg) tea.Cmd {
	o.editorCalls++
	return func() tea.Msg {
		if o.editorErr != nil {
			return done(o.editorErr)
		}
		if o.editorContent != nil {
			if err := os.WriteFile(path, o.editorContent, 0o644); err != nil {
				return done(err)
			}
		}
		return done(nil)
	}
}

func (o *fakeOperations) Exec(ctx context.Context, request ExecRequest, done func(error) tea.Msg) tea.Cmd {
	o.execCalls++
	o.execContext = ctx
	o.execRequest = request
	return func() tea.Msg {
		if o.execRun == nil {
			return done(nil)
		}
		return done(o.execRun(request))
	}
}

func (o *fakeOperations) Interactive(_ context.Context, request InteractiveRequest, done func(error) tea.Msg) tea.Cmd {
	o.interactiveCalls++
	o.interactiveRequest = request
	return func() tea.Msg {
		if o.interactiveErr != nil {
			return done(o.interactiveErr)
		}
		if o.interactiveRun != nil {
			return done(o.interactiveRun(request))
		}
		return done(nil)
	}
}

func (o *fakeOperations) Preview(_ context.Context, _ string, _ string, _ string, done func(RenderResult, error) tea.Msg) tea.Cmd {
	o.previewCalls++
	if o.events != nil {
		*o.events = append(*o.events, "preview")
	}
	return func() tea.Msg { return done(o.previewResult, o.previewErr) }
}

func (o *fakeOperations) Diff(_ context.Context, _ string, _ string, _ string, _ string, done func(RenderResult, error) tea.Msg) tea.Cmd {
	o.diffCalls++
	return func() tea.Msg { return done(o.diffResult, o.diffErr) }
}

func TestEditorStepLifecycleLoadsComparesSavesAndPreviews(t *testing.T) {
	for _, test := range []struct {
		name         string
		artifact     Artifact
		edited       []byte
		expectedSave int
	}{
		{name: "unchanged Idea", artifact: ArtifactIdea, edited: nil, expectedSave: 0},
		{name: "changed Idea", artifact: ArtifactIdea, edited: []byte("changed idea"), expectedSave: 1},
		{name: "changed Spec", artifact: ArtifactSpec, edited: []byte("changed spec"), expectedSave: 1},
		{name: "changed PR", artifact: ArtifactPR, edited: []byte("changed pr"), expectedSave: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			tempDir := t.TempDir()
			events := []string{}
			store := &fakeStore{data: map[Artifact][]byte{test.artifact: []byte("persisted")}, events: &events}
			operations := &fakeOperations{editorContent: test.edited, previewResult: RenderResult{Output: "rendered"}, events: &events}
			model := composeTestModel(t, editorDefinition(test.artifact), "edit", tempDir, store, operations)

			var command tea.Cmd = model.Init()
			require.NotNil(t, command)
			message := command()
			inputPath := filepath.Join(tempDir, string(test.artifact), inputFileName)
			outputPath := filepath.Join(tempDir, string(test.artifact), outputFileName)
			assertFileContent(t, inputPath, "persisted")
			assertFileContent(t, outputPath, "persisted")

			model, command = applyUpdate(t, model, message)
			assert.Equal(t, ScreenID("editor"), model.Screen())
			model = drainCommands(t, model, command, 5)

			assert.NoError(t, model.Error())
			assert.Equal(t, ScreenID("preview"), model.Screen())
			assert.Equal(t, "rendered", model.Rendered())
			assert.Len(t, store.loads, 1)
			assert.Len(t, store.saves, test.expectedSave)
			assertFileContent(t, inputPath, "persisted")
			if test.expectedSave == 1 {
				assert.Equal(t, test.edited, store.saves[0].content)
				assert.Equal(t, []string{"save", "preview"}, events)
			} else {
				assert.Equal(t, []string{"preview"}, events)
			}
		})
	}
}

func TestEditorFailuresAndSaveFailuresEnterErrorWithoutPreview(t *testing.T) {
	tests := []struct {
		name      string
		editorErr error
		saveErr   error
		expected  string
	}{
		{name: "editor failure", editorErr: errors.New("editor unavailable"), expected: "editor unavailable"},
		{name: "save failure", saveErr: errors.New("backend unavailable"), expected: "backend unavailable"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &fakeStore{data: map[Artifact][]byte{ArtifactIdea: []byte("before")}, saveErr: test.saveErr}
			operations := &fakeOperations{editorContent: []byte("after"), editorErr: test.editorErr}
			model := composeTestModel(t, editorDefinition(ArtifactIdea), "edit", t.TempDir(), store, operations)

			model, command := applyUpdate(t, model, model.Init()())
			model = drainCommands(t, model, command, 5)

			require.Error(t, model.Error())
			assert.Contains(t, model.Error().Error(), test.expected)
			assert.Equal(t, ScreenID("error"), model.Screen())
			assert.Equal(t, []CommandID{"/return"}, model.Commands())
			assert.Equal(t, 0, operations.previewCalls)
		})
	}
}

func TestExecExactOutputStopAndInteractiveContinuation(t *testing.T) {
	t.Run("exact final line completes and saves", func(t *testing.T) {
		tempDir := t.TempDir()
		store := &fakeStore{data: map[Artifact][]byte{ArtifactSpec: []byte("before")}}
		operations := &fakeOperations{previewResult: RenderResult{Output: "spec preview"}}
		operations.execRun = func(request ExecRequest) error {
			require.NoError(t, os.WriteFile(filepath.Join(request.Workspace, outputFileName), []byte("after"), 0o644))
			return os.WriteFile(filepath.Join(request.Workspace, agentOutputFileName), []byte("details\nNo blocking issues found.\n"), 0o644)
		}
		model := composeTestModel(t, ConformanceDefinition(), "spec-review", tempDir, store, operations)

		model, command := applyUpdate(t, model, model.Init()())
		assert.Equal(t, []CommandID{"/stop"}, model.Commands())
		model, command = applyUpdate(t, model, command())
		model, command = applyUpdate(t, model, command())
		model, command = applyUpdate(t, model, command())
		model, command = applyUpdate(t, model, command())

		assert.Nil(t, command)
		assert.NoError(t, model.Error())
		assert.Equal(t, ScreenID("spec-preview"), model.Screen())
		assert.Equal(t, "spec preview", model.Rendered())
		require.Len(t, store.saves, 1)
		assert.Equal(t, []byte("after"), store.saves[0].content)
	})

	t.Run("stop cancels and returns without save", func(t *testing.T) {
		store := &fakeStore{data: map[Artifact][]byte{ArtifactSpec: []byte("before")}}
		operations := &fakeOperations{}
		model := composeTestModel(t, ConformanceDefinition(), "spec-review", t.TempDir(), store, operations)
		model, _ = applyUpdate(t, model, model.Init()())
		require.NotNil(t, operations.execContext)

		model, command := applyUpdate(t, model, CommandMsg{ID: "/stop"})

		assert.Nil(t, command)
		assert.True(t, model.Done())
		assert.Equal(t, ScreenID("changes"), model.TerminalScreen())
		assert.Empty(t, store.saves)
		select {
		case <-operations.execContext.Done():
		default:
			t.Fatal("expected active Exec context to be cancelled")
		}
	})

	t.Run("non-matching output enters Interactive and edit reuses baseline", func(t *testing.T) {
		tempDir := t.TempDir()
		store := &fakeStore{data: map[Artifact][]byte{ArtifactSpec: []byte("before")}}
		operations := &fakeOperations{editorContent: []byte("edited"), previewResult: RenderResult{Output: "preview"}}
		operations.execRun = func(request ExecRequest) error {
			return os.WriteFile(filepath.Join(request.Workspace, agentOutputFileName), []byte("questions remain\n"), 0o644)
		}
		model := composeTestModel(t, ConformanceDefinition(), "spec-review", tempDir, store, operations)

		model, command := applyUpdate(t, model, model.Init()())
		model, command = applyUpdate(t, model, command())
		model, command = applyUpdate(t, model, command())
		model, command = applyUpdate(t, model, command())
		assert.Nil(t, command)

		assert.Equal(t, ScreenID("interactive"), model.Screen())
		assert.Equal(t, []CommandID{"/interactive", "/edit", "/cancel"}, model.Commands())
		assert.Contains(t, model.View(), "questions remain")
		assert.Len(t, store.loads, 1)
		model, command = applyUpdate(t, model, CommandMsg{ID: "/edit"})
		model = drainCommands(t, model, command, 5)

		assert.Equal(t, ScreenID("spec-preview"), model.Screen())
		assert.Len(t, store.loads, 1)
		require.Len(t, store.saves, 1)
		assertFileContent(t, filepath.Join(tempDir, string(ArtifactSpec), inputFileName), "before")
	})

	t.Run("execution failure enters Error without save", func(t *testing.T) {
		store := &fakeStore{data: map[Artifact][]byte{ArtifactSpec: []byte("before")}}
		operations := &fakeOperations{execRun: func(ExecRequest) error { return errors.New("exec failed") }}
		model := composeTestModel(t, ConformanceDefinition(), "spec-review", t.TempDir(), store, operations)
		model, command := applyUpdate(t, model, model.Init()())
		model, _ = applyUpdate(t, model, command())

		require.Error(t, model.Error())
		assert.Contains(t, model.Error().Error(), "exec failed")
		assert.Empty(t, store.saves)
		assert.Equal(t, 0, operations.previewCalls)
	})
}

func TestInteractiveSessionEntryCompletionFailureAndCancel(t *testing.T) {
	t.Run("entry does not start session and missing session enters Error", func(t *testing.T) {
		store := &fakeStore{data: map[Artifact][]byte{ArtifactPR: []byte("pr")}}
		operations := &fakeOperations{}
		model := composeTestModel(t, ConformanceDefinition(), "pr-review", t.TempDir(), store, operations)

		model, command := applyUpdate(t, model, model.Init()())
		assert.Equal(t, 0, operations.interactiveCalls)
		require.NotNil(t, command)
		model, _ = applyUpdate(t, model, command())
		model, command = applyUpdate(t, model, CommandMsg{ID: "/interactive"})
		model, _ = applyUpdate(t, model, command())

		require.Error(t, model.Error())
		assert.Contains(t, model.Error().Error(), "session-id")
		assert.Equal(t, 0, operations.interactiveCalls)
		assert.Empty(t, store.saves)
	})

	t.Run("successful session completes through save before Preview", func(t *testing.T) {
		tempDir := t.TempDir()
		store := &fakeStore{data: map[Artifact][]byte{ArtifactPR: []byte("before")}}
		operations := &fakeOperations{previewResult: RenderResult{Output: "preview"}}
		operations.interactiveRun = func(request InteractiveRequest) error {
			return os.WriteFile(filepath.Join(request.Workspace, outputFileName), []byte("after"), 0o644)
		}
		model := composeTestModel(t, ConformanceDefinition(), "pr-review", tempDir, store, operations)
		model, optionalOutput := applyUpdate(t, model, model.Init()())
		model, _ = applyUpdate(t, model, optionalOutput())
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, string(ArtifactPR), sessionFileName), []byte("session-123\n"), 0o644))

		model, command := applyUpdate(t, model, CommandMsg{ID: "/interactive"})
		model, command = applyUpdate(t, model, command())
		assert.Equal(t, "session-123", operations.interactiveRequest.SessionID)
		model, command = applyUpdate(t, model, command())
		model, command = applyUpdate(t, model, command())
		model, command = applyUpdate(t, model, command())

		assert.Nil(t, command)
		assert.NoError(t, model.Error())
		assert.Equal(t, ScreenID("pr-preview"), model.Screen())
		require.Len(t, store.saves, 1)
		assert.Equal(t, []byte("after"), store.saves[0].content)
	})

	t.Run("cancel returns to origin without save", func(t *testing.T) {
		tempDir := t.TempDir()
		store := &fakeStore{data: map[Artifact][]byte{ArtifactPR: []byte("before")}}
		model := composeTestModel(t, ConformanceDefinition(), "pr-review", tempDir, store, &fakeOperations{})
		model, _ = applyUpdate(t, model, model.Init()())
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, string(ArtifactPR), outputFileName), []byte("local change"), 0o644))

		model, command := applyUpdate(t, model, CommandMsg{ID: "/cancel"})

		assert.Nil(t, command)
		assert.True(t, model.Done())
		assert.Equal(t, ScreenID("changes"), model.TerminalScreen())
		assert.Empty(t, store.saves)
	})

	t.Run("session failure enters Error without save", func(t *testing.T) {
		tempDir := t.TempDir()
		store := &fakeStore{data: map[Artifact][]byte{ArtifactPR: []byte("before")}}
		operations := &fakeOperations{interactiveErr: errors.New("resume failed")}
		model := composeTestModel(t, ConformanceDefinition(), "pr-review", tempDir, store, operations)
		model, optionalOutput := applyUpdate(t, model, model.Init()())
		model, _ = applyUpdate(t, model, optionalOutput())
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, string(ArtifactPR), sessionFileName), []byte("session"), 0o644))
		model, command := applyUpdate(t, model, CommandMsg{ID: "/interactive"})
		model, command = applyUpdate(t, model, command())
		model, _ = applyUpdate(t, model, command())

		require.Error(t, model.Error())
		assert.Contains(t, model.Error().Error(), "resume failed")
		assert.Empty(t, store.saves)
		assert.Equal(t, 0, operations.previewCalls)
	})
}

func TestPreviewNavigationTogglingAndRenderingErrors(t *testing.T) {
	t.Run("Step destination fresh-loads and terminal destination does not", func(t *testing.T) {
		definition := twoEditorStepDefinition()
		store := &fakeStore{data: map[Artifact][]byte{ArtifactIdea: []byte("idea"), ArtifactSpec: []byte("spec")}}
		operations := &fakeOperations{previewResult: RenderResult{Output: "preview"}}
		model := composeTestModel(t, definition, "idea", t.TempDir(), store, operations)
		model = completeUnchangedEditor(t, model)

		model, command := applyUpdate(t, model, CommandMsg{ID: "/anything"})
		require.NotNil(t, command)
		model, command = applyUpdate(t, model, command())
		assert.Equal(t, []Artifact{ArtifactIdea, ArtifactSpec}, store.loads)
		assert.Equal(t, ScreenID("editor"), model.Screen())
		assertFileContent(t, filepath.Join(model.FlowContext().TempDir, string(ArtifactSpec), inputFileName), "spec")
		assert.NotNil(t, command)

		model = completeUnchangedEditorFromCommand(t, model, command)
		loadsBefore := len(store.loads)
		model, command = applyUpdate(t, model, CommandMsg{ID: "/done"})
		assert.Nil(t, command)
		assert.True(t, model.Done())
		assert.Equal(t, ScreenID("changes"), model.TerminalScreen())
		assert.Len(t, store.loads, loadsBefore)
	})

	t.Run("both horizontal arrows toggle Preview and Diff", func(t *testing.T) {
		store := &fakeStore{data: map[Artifact][]byte{ArtifactIdea: []byte("idea")}}
		operations := &fakeOperations{
			previewResult: RenderResult{Output: "preview"},
			diffResult:    RenderResult{Output: "diff", Status: 1},
		}
		model := completeUnchangedEditor(t, composeTestModel(t, editorDefinition(ArtifactIdea), "edit", t.TempDir(), store, operations))

		model, command := applyUpdate(t, model, tea.KeyMsg{Type: tea.KeyLeft})
		model, _ = applyUpdate(t, model, command())
		assert.Equal(t, PreviewDiff, model.Mode())
		assert.Equal(t, "diff", model.Rendered())
		model, command = applyUpdate(t, model, tea.KeyMsg{Type: tea.KeyRight})
		model, _ = applyUpdate(t, model, command())
		assert.Equal(t, PreviewArtifact, model.Mode())
		assert.Equal(t, "preview", model.Rendered())
	})

	t.Run("Git status zero is a successful identical Diff", func(t *testing.T) {
		store := &fakeStore{data: map[Artifact][]byte{ArtifactIdea: []byte("idea")}}
		operations := &fakeOperations{
			previewResult: RenderResult{Output: "preview"},
			diffResult:    RenderResult{Output: "identical", Status: 0},
		}
		model := completeUnchangedEditor(t, composeTestModel(t, editorDefinition(ArtifactIdea), "edit", t.TempDir(), store, operations))
		model, command := applyUpdate(t, model, tea.KeyMsg{Type: tea.KeyLeft})
		model, _ = applyUpdate(t, model, command())

		assert.NoError(t, model.Error())
		assert.Equal(t, PreviewDiff, model.Mode())
		assert.Equal(t, "identical", model.Rendered())
	})

	t.Run("Git status greater than one enters Error", func(t *testing.T) {
		store := &fakeStore{data: map[Artifact][]byte{ArtifactIdea: []byte("idea")}}
		operations := &fakeOperations{previewResult: RenderResult{Output: "preview"}, diffResult: RenderResult{Status: 2}}
		model := completeUnchangedEditor(t, composeTestModel(t, editorDefinition(ArtifactIdea), "edit", t.TempDir(), store, operations))

		model, command := applyUpdate(t, model, tea.KeyMsg{Type: tea.KeyRight})
		model, _ = applyUpdate(t, model, command())

		require.Error(t, model.Error())
		assert.Contains(t, model.Error().Error(), "status 2")
		assert.Equal(t, ScreenID("error"), model.Screen())
	})

	t.Run("renderer failure enters Error", func(t *testing.T) {
		store := &fakeStore{data: map[Artifact][]byte{ArtifactIdea: []byte("idea")}}
		operations := &fakeOperations{previewErr: errors.New("bat failed")}
		model := composeTestModel(t, editorDefinition(ArtifactIdea), "edit", t.TempDir(), store, operations)
		model, command := applyUpdate(t, model, model.Init()())
		model = drainCommands(t, model, command, 5)

		require.Error(t, model.Error())
		assert.Contains(t, model.Error().Error(), "bat failed")
		assert.Equal(t, ScreenID("error"), model.Screen())
	})

	t.Run("unsupported Preview artifact enters Error", func(t *testing.T) {
		store := &fakeStore{data: map[Artifact][]byte{ArtifactImplement: []byte("implementation")}}
		operations := &fakeOperations{}
		model := composeTestModel(t, editorDefinition(ArtifactImplement), "edit", t.TempDir(), store, operations)
		model, command := applyUpdate(t, model, model.Init()())
		model = drainCommands(t, model, command, 5)

		require.Error(t, model.Error())
		assert.Contains(t, model.Error().Error(), "does not support artifact")
		assert.Equal(t, 0, operations.previewCalls)
	})
}

func TestTaskSpecificMissingResourcesEnterErrorBeforeExternalOperation(t *testing.T) {
	t.Run("Editor requires output.md", func(t *testing.T) {
		tempDir := t.TempDir()
		store := &fakeStore{data: map[Artifact][]byte{ArtifactIdea: []byte("idea")}}
		operations := &fakeOperations{}
		model := composeTestModel(t, editorDefinition(ArtifactIdea), "edit", tempDir, store, operations)
		model, command := applyUpdate(t, model, model.Init()())
		require.NoError(t, os.Remove(filepath.Join(tempDir, string(ArtifactIdea), outputFileName)))
		model, _ = applyUpdate(t, model, command())

		require.Error(t, model.Error())
		assert.Contains(t, model.Error().Error(), "output.md")
		assert.Equal(t, 0, operations.editorCalls)
	})

	t.Run("Exec requires agent-output.md", func(t *testing.T) {
		store := &fakeStore{data: map[Artifact][]byte{ArtifactSpec: []byte("spec")}}
		operations := &fakeOperations{}
		model := composeTestModel(t, ConformanceDefinition(), "spec-review", t.TempDir(), store, operations)
		model, command := applyUpdate(t, model, model.Init()())
		model, command = applyUpdate(t, model, command())
		model, _ = applyUpdate(t, model, command())

		require.Error(t, model.Error())
		assert.Contains(t, model.Error().Error(), "agent-output.md")
		assert.Empty(t, store.saves)
	})

	t.Run("Preview requires input.md and output.md", func(t *testing.T) {
		tempDir := t.TempDir()
		store := &fakeStore{data: map[Artifact][]byte{ArtifactIdea: []byte("idea")}}
		operations := &fakeOperations{}
		model := composeTestModel(t, editorDefinition(ArtifactIdea), "edit", tempDir, store, operations)
		model, command := applyUpdate(t, model, model.Init()())
		model, command = applyUpdate(t, model, command())
		model, command = applyUpdate(t, model, command())
		model, command = applyUpdate(t, model, command())
		require.NoError(t, os.Remove(filepath.Join(tempDir, string(ArtifactIdea), inputFileName)))
		model, _ = applyUpdate(t, model, command())

		require.Error(t, model.Error())
		assert.Contains(t, model.Error().Error(), "input.md")
		assert.Equal(t, 0, operations.previewCalls)
	})
}

func TestChangedInputBaselineFailsStepWithoutPersistence(t *testing.T) {
	tempDir := t.TempDir()
	store := &fakeStore{data: map[Artifact][]byte{ArtifactIdea: []byte("baseline")}}
	operations := &fakeOperations{editorContent: []byte("output")}
	model := composeTestModel(t, editorDefinition(ArtifactIdea), "edit", tempDir, store, operations)
	model, command := applyUpdate(t, model, model.Init()())
	model, command = applyUpdate(t, model, command())
	editorFinished := command()
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, string(ArtifactIdea), inputFileName), []byte("tampered"), 0o644))
	model, command = applyUpdate(t, model, editorFinished)
	model, _ = applyUpdate(t, model, command())

	require.Error(t, model.Error())
	assert.Contains(t, model.Error().Error(), "input.md changed")
	assert.Empty(t, store.saves)
	assert.Equal(t, 0, operations.previewCalls)
}

func TestWorkspaceAndLoadFailuresEnterConcreteError(t *testing.T) {
	t.Run("missing temp_dir prevents load", func(t *testing.T) {
		store := &fakeStore{data: map[Artifact][]byte{ArtifactIdea: []byte("idea")}}
		model := composeTestModel(t, editorDefinition(ArtifactIdea), "edit", "", store, &fakeOperations{})
		model, _ = applyUpdate(t, model, model.Init()())

		require.Error(t, model.Error())
		assert.Contains(t, model.Error().Error(), "temp_dir")
		assert.Empty(t, store.loads)
	})

	t.Run("load failure", func(t *testing.T) {
		store := &fakeStore{loadErr: errors.New("backend down")}
		model := composeTestModel(t, editorDefinition(ArtifactIdea), "edit", t.TempDir(), store, &fakeOperations{})
		model, _ = applyUpdate(t, model, model.Init()())

		require.Error(t, model.Error())
		assert.Contains(t, model.Error().Error(), "backend down")
	})

	t.Run("artifact directory creation failure", func(t *testing.T) {
		tempDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, string(ArtifactIdea)), []byte("not a directory"), 0o644))
		store := &fakeStore{data: map[Artifact][]byte{ArtifactIdea: []byte("idea")}}
		model := composeTestModel(t, editorDefinition(ArtifactIdea), "edit", tempDir, store, &fakeOperations{})
		model, _ = applyUpdate(t, model, model.Init()())

		require.Error(t, model.Error())
		assert.Contains(t, model.Error().Error(), "artifact workspace")
	})
}

func TestErrorReturnUsesRecordedOrigin(t *testing.T) {
	store := &fakeStore{loadErr: errors.New("failed")}
	model := composeTestModel(t, editorDefinition(ArtifactIdea), "edit", t.TempDir(), store, &fakeOperations{})
	model, _ = applyUpdate(t, model, model.Init()())

	model, command := applyUpdate(t, model, CommandMsg{ID: "/return"})

	assert.Nil(t, command)
	assert.True(t, model.Done())
	assert.Equal(t, ScreenID("changes"), model.TerminalScreen())
}

func composeTestModel(t *testing.T, definition Definition, step StepID, tempDir string, store ArtifactStore, operations Operations) Model {
	t.Helper()
	model := Compose(Composition{
		Definition: definition,
		Context: Context{
			TempDir:   tempDir,
			FlowDir:   "/flow",
			ChangeID:  42,
			ChangeRef: "ref-42",
			Origin:    "changes",
			Step:      step,
		},
		TerminalScreens: []ScreenID{"changes"},
		Store:           store,
		Operations:      operations,
	})
	require.NoError(t, model.Error())
	return model
}

func applyUpdate(t *testing.T, model Model, message tea.Msg) (Model, tea.Cmd) {
	t.Helper()
	next, command := model.Update(message)
	updated, ok := next.(Model)
	require.True(t, ok)
	return updated, command
}

func completeUnchangedEditor(t *testing.T, model Model) Model {
	t.Helper()
	model, command := applyUpdate(t, model, model.Init()())
	return completeUnchangedEditorFromCommand(t, model, command)
}

func completeUnchangedEditorFromCommand(t *testing.T, model Model, command tea.Cmd) Model {
	t.Helper()
	return drainCommands(t, model, command, 5)
}

func drainCommands(t *testing.T, model Model, command tea.Cmd, limit int) Model {
	t.Helper()
	for count := 0; command != nil && count < limit; count++ {
		model, command = applyUpdate(t, model, command())
	}
	assert.Nil(t, command, "command chain exceeded limit")
	return model
}

func assertFileContent(t *testing.T, path string, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, expected, string(content))
}

func editorDefinition(artifact Artifact) Definition {
	return Definition{
		ID: "editor-definition",
		Steps: []StepDefinition{{
			ID: "edit",
			Tasks: []TaskDefinition{{
				ID:       "edit-artifact",
				Type:     TaskEditor,
				Artifact: artifact,
				Screen:   "editor",
				Preview:  "preview",
				Error:    "error",
			}},
		}},
		Screens: []ScreenDefinition{
			{ID: "editor", Type: ScreenEditor},
			{ID: "preview", Type: ScreenPreview, Commands: []CommandDefinition{{ID: "/done", Destination: ScreenDestination("changes")}}},
			{ID: "error", Type: ScreenError},
		},
	}
}

func twoEditorStepDefinition() Definition {
	definition := editorDefinition(ArtifactIdea)
	definition.ID = "two-editor-steps"
	definition.Steps[0].ID = "idea"
	definition.Steps[0].Tasks[0].ID = "edit-idea"
	definition.Steps = append(definition.Steps, StepDefinition{
		ID: "spec",
		Tasks: []TaskDefinition{{
			ID:       "edit-spec",
			Type:     TaskEditor,
			Artifact: ArtifactSpec,
			Screen:   "editor",
			Preview:  "preview",
			Error:    "error",
		}},
	})
	definition.Screens[1].Commands = []CommandDefinition{
		{ID: "/anything", Destination: StepDestination("spec")},
		{ID: "/done", Destination: ScreenDestination("changes")},
	}
	return definition
}
