package integration_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"mch/internal/app"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIProgramExistingSpecUsesUUIDIdeaWorkspaceAndSpecWrite(t *testing.T) {
	const refUUID = "0198a86f-9b8a-7d89-ae5b-6f25b528b04c"
	backend := newArtifactProgramBackend(t, refUUID)
	repo := t.TempDir()
	require.NoError(t, exec.Command("git", "init", repo).Run())
	flowDir := filepath.Join(repo, ".mch", "default")
	require.NoError(t, os.MkdirAll(filepath.Join(flowDir, "prompts"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".mch", "config.yaml"), []byte("backend_url: "+backend.URL+"\nproject_id: 7\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(flowDir, "flow.yaml"), []byte("version: 1\nslug: default\nhelp: help.yaml\nmakefile: Makefile\nsteps:\n  - slug: spec-write\n    stage: idea\n    mode: exec\n    prompt: prompts/spec-write.md\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(flowDir, "help.yaml"), []byte("version: 1\nstage_modes: []\ntask_statuses: []\ntask_steps: []\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(flowDir, "prompts", "spec-write.md"), []byte("write spec"), 0o644))

	workspace := filepath.Join(repo, ".mch", "tmp", refUUID, "idea")
	require.NoError(t, os.MkdirAll(workspace, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "session-id"), []byte("22222222-2222-2222-2222-222222222222\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(workspace, "events.jsonl"), []byte("preserve"), 0o644))

	binDir := filepath.Join(t.TempDir(), "bin")
	require.NoError(t, os.MkdirAll(binDir, 0o755))
	editorPath := filepath.Join(binDir, "editor")
	codexPath := filepath.Join(binDir, "codex")
	invocationPath := filepath.Join(t.TempDir(), "codex-invocation")
	require.NoError(t, os.WriteFile(editorPath, []byte("#!/bin/sh\nprintf 'Types: feature\\n\\nUser-edited spec\\n' > \"$1\"\n"), 0o755))
	require.NoError(t, os.WriteFile(codexPath, []byte("#!/bin/sh\noutput=''\nprevious=''\nfor argument in \"$@\"; do\n  if [ \"$previous\" = '-o' ]; then output=\"$argument\"; fi\n  previous=\"$argument\"\ndone\nworkspace=\"$PWD/$MCH_TEMP_DIR/$MCH_REF_UUID/$MCH_STAGE\"\nprintf '%s|%s|%s\\n' \"$MCH_STAGE\" \"$MCH_REF_UUID\" \"$*\" > \"$MCH_TEST_INVOCATION\"\nif [ \"$(cat \"$workspace/input.md\")\" != 'Types: feature\n\nUser-edited spec' ]; then\n  printf 'stale artifact input' >&2\n  exit 1\nfi\nprintf 'Types: feature\\n\\nAgent-written spec\\n' > \"$workspace/output.md\"\nprintf 'Done.' > \"$output\"\nprintf '{\"type\":\"thread.started\",\"thread_id\":\"33333333-3333-3333-3333-333333333333\"}\\n'\n"), 0o755))

	previous, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { require.NoError(t, os.Chdir(previous)) })
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("EDITOR", editorPath)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("MCH_TEST_INVOCATION", invocationPath)

	inputReader, inputWriter, err := os.Pipe()
	require.NoError(t, err)
	outputPath := filepath.Join(t.TempDir(), "program-output")
	output, err := os.Create(outputPath)
	require.NoError(t, err)
	done := make(chan error, 1)
	go func() { done <- app.RunWithIO(nil, inputReader, output) }()

	waitForProgramOutput(t, outputPath, "MainScreen", done)
	_, err = io.WriteString(inputWriter, "/changes\r")
	require.NoError(t, err)
	waitForProgramOutput(t, outputPath, "ChangesListScreen", done)
	_, err = inputWriter.Write([]byte{'\r'})
	require.NoError(t, err)
	waitForProgramOutput(t, outputPath, "ChangeDetailsScreen", done)
	editOffset := programOutputSize(t, outputPath)
	_, err = io.WriteString(inputWriter, "/edit-spec\r")
	require.NoError(t, err)
	backend.waitForAgentSpec(t, outputPath)
	waitForProgramOutputAfter(t, outputPath, "status save", editOffset, done)

	assert.Equal(t, []artifactProgramUpdate{
		{Text: "Types: feature\n\nUser-edited spec\n", AgentEdit: false},
		{Text: "Types: feature\n\nAgent-written spec\n", AgentEdit: true},
	}, backend.updatesSnapshot())
	assert.Equal(t, "Types: feature\n\nAgent-written spec\n", readFile(t, filepath.Join(workspace, "output.md")))
	assert.Equal(t, "preserve", readFile(t, filepath.Join(workspace, "events.jsonl")))
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", strings.TrimSpace(readFile(t, filepath.Join(workspace, "session-id"))))
	invocation := readFile(t, invocationPath)
	assert.Contains(t, invocation, "idea|"+refUUID+"|")
	assert.Contains(t, invocation, "spec-write.md")
	assert.Contains(t, invocation, "22222222-2222-2222-2222-222222222222")

	returnOffset := programOutputSize(t, outputPath)
	_, err = io.WriteString(inputWriter, "/return\r")
	require.NoError(t, err)
	waitForProgramOutputAfter(t, outputPath, "ChangesListScreen", returnOffset, done)
	mainOffset := programOutputSize(t, outputPath)
	_, err = io.WriteString(inputWriter, "/return\r")
	require.NoError(t, err)
	waitForProgramOutputAfter(t, outputPath, "MainScreen", mainOffset, done)
	_, err = inputWriter.Write([]byte{3})
	require.NoError(t, err)
	require.NoError(t, inputWriter.Close())
	select {
	case err := <-done:
		require.NoError(t, err, readFile(t, outputPath))
	case <-time.After(5 * time.Second):
		t.Fatal("CLI program did not exit")
	}
	require.NoError(t, output.Close())
}

func TestCLIProgramExistingSpecKeepsPersistedArtifactAfterFollowUpFailure(t *testing.T) {
	const refUUID = "0198a86f-9b8a-7d89-ae5b-6f25b528b04c"
	tests := []struct {
		name             string
		failTypeUpdateAt int
		wantSpec         string
		wantUpdates      []artifactProgramUpdate
	}{
		{
			name:             "user save",
			failTypeUpdateAt: 1,
			wantSpec:         "User-edited spec",
			wantUpdates: []artifactProgramUpdate{
				{Text: "Types: feature\n\nUser-edited spec\n", AgentEdit: false},
			},
		},
		{
			name:             "agent save",
			failTypeUpdateAt: 2,
			wantSpec:         "Agent-written spec",
			wantUpdates: []artifactProgramUpdate{
				{Text: "Types: feature\n\nUser-edited spec\n", AgentEdit: false},
				{Text: "Types: feature\n\nAgent-written spec\n", AgentEdit: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := newArtifactProgramBackend(t, refUUID)
			backend.failTypeUpdateAt = tt.failTypeUpdateAt
			repo := t.TempDir()
			require.NoError(t, exec.Command("git", "init", repo).Run())
			flowDir := filepath.Join(repo, ".mch", "default")
			require.NoError(t, os.MkdirAll(filepath.Join(flowDir, "prompts"), 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(repo, ".mch", "config.yaml"), []byte("backend_url: "+backend.URL+"\nproject_id: 7\n"), 0o644))
			require.NoError(t, os.WriteFile(filepath.Join(flowDir, "flow.yaml"), []byte("version: 1\nslug: default\nhelp: help.yaml\nmakefile: Makefile\nsteps:\n  - slug: spec-write\n    stage: idea\n    mode: exec\n    prompt: prompts/spec-write.md\n"), 0o644))
			require.NoError(t, os.WriteFile(filepath.Join(flowDir, "help.yaml"), []byte("version: 1\nstage_modes: []\ntask_statuses: []\ntask_steps: []\n"), 0o644))
			require.NoError(t, os.WriteFile(filepath.Join(flowDir, "prompts", "spec-write.md"), []byte("write spec"), 0o644))

			workspace := filepath.Join(repo, ".mch", "tmp", refUUID, "idea")
			require.NoError(t, os.MkdirAll(workspace, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(workspace, "session-id"), []byte("22222222-2222-2222-2222-222222222222\n"), 0o644))

			binDir := filepath.Join(t.TempDir(), "bin")
			require.NoError(t, os.MkdirAll(binDir, 0o755))
			editorPath := filepath.Join(binDir, "editor")
			codexPath := filepath.Join(binDir, "codex")
			require.NoError(t, os.WriteFile(editorPath, []byte("#!/bin/sh\nprintf 'Types: feature\\n\\nUser-edited spec\\n' > \"$1\"\n"), 0o755))
			require.NoError(t, os.WriteFile(codexPath, []byte("#!/bin/sh\noutput=''\nprevious=''\nfor argument in \"$@\"; do\n  if [ \"$previous\" = '-o' ]; then output=\"$argument\"; fi\n  previous=\"$argument\"\ndone\nworkspace=\"$PWD/$MCH_TEMP_DIR/$MCH_REF_UUID/$MCH_STAGE\"\nprintf 'Types: feature\\n\\nAgent-written spec\\n' > \"$workspace/output.md\"\nprintf 'Done.' > \"$output\"\nprintf '{\"type\":\"thread.started\",\"thread_id\":\"33333333-3333-3333-3333-333333333333\"}\\n'\n"), 0o755))

			previous, err := os.Getwd()
			require.NoError(t, err)
			require.NoError(t, os.Chdir(repo))
			t.Cleanup(func() { require.NoError(t, os.Chdir(previous)) })
			t.Setenv("TERM", "xterm-256color")
			t.Setenv("EDITOR", editorPath)
			t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

			inputReader, inputWriter, err := os.Pipe()
			require.NoError(t, err)
			outputPath := filepath.Join(t.TempDir(), "program-output")
			output, err := os.Create(outputPath)
			require.NoError(t, err)
			done := make(chan error, 1)
			go func() { done <- app.RunWithIO(nil, inputReader, output) }()

			waitForProgramOutput(t, outputPath, "MainScreen", done)
			_, err = io.WriteString(inputWriter, "/changes\r")
			require.NoError(t, err)
			waitForProgramOutput(t, outputPath, "ChangesListScreen", done)
			_, err = inputWriter.Write([]byte{'\r'})
			require.NoError(t, err)
			waitForProgramOutput(t, outputPath, "ChangeDetailsScreen", done)
			editOffset := programOutputSize(t, outputPath)
			_, err = io.WriteString(inputWriter, "/edit-spec\r")
			require.NoError(t, err)
			waitForProgramOutputAfter(t, outputPath, "save failed", editOffset, done)
			detailOffset := programOutputSize(t, outputPath)
			_, err = io.WriteString(inputWriter, strings.Repeat("\x1b[B", 9))
			require.NoError(t, err)
			waitForProgramOutputAfter(t, outputPath, tt.wantSpec, detailOffset, done)

			assert.Equal(t, tt.wantUpdates, backend.updatesSnapshot())
			assert.Equal(t, tt.failTypeUpdateAt, backend.typeUpdateCount())

			returnOffset := programOutputSize(t, outputPath)
			_, err = inputWriter.Write([]byte{3})
			require.NoError(t, err)
			waitForProgramOutputAfter(t, outputPath, "ChangesListScreen", returnOffset, done)
			mainOffset := programOutputSize(t, outputPath)
			_, err = inputWriter.Write([]byte{3})
			require.NoError(t, err)
			waitForProgramOutputAfter(t, outputPath, "MainScreen", mainOffset, done)
			_, err = inputWriter.Write([]byte{3})
			require.NoError(t, err)
			require.NoError(t, inputWriter.Close())
			select {
			case err := <-done:
				require.NoError(t, err, readFile(t, outputPath))
			case <-time.After(5 * time.Second):
				t.Fatal("CLI program did not exit")
			}
			require.NoError(t, output.Close())
		})
	}
}

type artifactProgramUpdate struct {
	Text      string
	AgentEdit bool
}

type artifactProgramBackend struct {
	*httptest.Server
	mu               sync.Mutex
	refUUID          string
	spec             string
	updates          []artifactProgramUpdate
	typeUpdates      int
	failTypeUpdateAt int
}

func newArtifactProgramBackend(t *testing.T, refUUID string) *artifactProgramBackend {
	t.Helper()
	backend := &artifactProgramBackend{refUUID: refUUID, spec: "Original spec"}
	backend.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		switch r.URL.Path {
		case "/api/v1/options/change-phases-list":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"slug": "backlog"}})
		case "/api/v1/options/change-types-list":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"slug": "feature"}})
		case "/api/v1/project/get":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 7, "name": "Program Project"})
		case "/api/v1/change/list":
			_ = json.NewEncoder(w).Encode(map[string]any{"changes": []map[string]any{{"id": 12, "ref_uuid": refUUID, "ref": 3, "title": "Program Change", "spec": backend.currentSpec()}}})
		case "/api/v1/change/get":
			_ = json.NewEncoder(w).Encode(map[string]any{"change": map[string]any{"id": 12, "project_id": 7, "ref_uuid": refUUID, "ref": 3, "title": "Program Change", "idea": "# Program Change\n\nIdea", "spec": backend.currentSpec()}, "test_cases": []any{}})
		case "/api/v1/change/update-spec":
			spec, _ := payload["spec"].(string)
			agentEdit, _ := payload["agent_edit"].(bool)
			backend.recordUpdate(spec, agentEdit)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 12, "ref_uuid": refUUID, "spec": spec, "agent_edit": agentEdit})
		case "/api/v1/change/update-change-types":
			backend.mu.Lock()
			backend.typeUpdates++
			fail := backend.failTypeUpdateAt == backend.typeUpdates
			backend.mu.Unlock()
			if fail {
				http.Error(w, "type update failed", http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 12, "change_types": payload["change_types"]})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(backend.Close)
	return backend
}

func (b *artifactProgramBackend) currentSpec() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.spec
}

func (b *artifactProgramBackend) recordUpdate(spec string, agentEdit bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.spec = spec
	b.updates = append(b.updates, artifactProgramUpdate{Text: spec, AgentEdit: agentEdit})
}

func (b *artifactProgramBackend) updatesSnapshot() []artifactProgramUpdate {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]artifactProgramUpdate(nil), b.updates...)
}

func (b *artifactProgramBackend) typeUpdateCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.typeUpdates
}

func (b *artifactProgramBackend) waitForAgentSpec(t *testing.T, outputPath string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		updates := b.updatesSnapshot()
		if len(updates) == 2 && updates[1].AgentEdit {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	output, _ := os.ReadFile(outputPath)
	t.Fatalf("timed out waiting for user and agent Spec updates: %#v\n%s", b.updatesSnapshot(), output)
}

func waitForProgramOutput(t *testing.T, path, marker string, done <-chan error) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		content, _ := os.ReadFile(path)
		if strings.Contains(string(content), marker) {
			return
		}
		select {
		case err := <-done:
			t.Fatalf("CLI exited before %q: %v\n%s", marker, err, content)
		default:
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %q in %s", marker, path)
}

func waitForProgramOutputAfter(t *testing.T, path, marker string, offset int64, done <-chan error) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		content, _ := os.ReadFile(path)
		if offset < int64(len(content)) && strings.Contains(string(content[offset:]), marker) {
			return
		}
		select {
		case err := <-done:
			t.Fatalf("CLI exited before new %q: %v\n%s", marker, err, content)
		default:
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for new %q in %s", marker, path)
}

func programOutputSize(t *testing.T, path string) int64 {
	t.Helper()
	info, err := os.Stat(path)
	require.NoError(t, err)
	return info.Size()
}
