package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppConfigLoadsRepositoryMCHConfigFlowAndHelp(t *testing.T) {
	root := t.TempDir()
	writeMCHFixture(t, root, "backend_url: http://backend.test\n"+"temp_dir: /workspace/custom-mch\n"+"project_id: 7\n")

	cfg, err := loadAppConfig(root)

	require.NoError(t, err)
	assert.Equal(t, root, cfg.RepositoryRoot)
	assert.Equal(t, filepath.Join(root, ".mch", "config.yaml"), cfg.ConfigPath)
	assert.Equal(t, "http://backend.test", cfg.BackendURL)
	assert.Equal(t, "/workspace/custom-mch", cfg.TempDir)
	assert.Equal(t, 7, cfg.ProjectID)
	assert.Equal(t, filepath.Join(root, ".mch", "default"), cfg.FlowDir)
	assert.Equal(t, "default", cfg.Flow.Slug)
	require.Len(t, cfg.Flow.Steps, len(defaultFlowStageSlugs))
	assert.Equal(t, "idea", cfg.Flow.Steps[0].Slug)
	assert.Equal(t, "make idea-entry", cfg.Flow.Steps[0].Entry)
	assert.Equal(t, "make idea-exec", cfg.Flow.Steps[0].Exec)
	assert.Equal(t, "make idea-exit", cfg.Flow.Steps[0].Exit)
	assert.Equal(t, []string{"skip", "prompt", "exec"}, flowOptionSlugs(cfg.FlowHelp.StageModes))
	assert.Equal(t, []string{"queued", "running", "paused", "stopped", "waiting", "completed", "failed"}, flowOptionSlugs(cfg.FlowHelp.TaskStatuses))
	assert.Equal(t, []string{"none", "entry", "prompt", "agent", "exit", "done"}, flowOptionSlugs(cfg.FlowHelp.TaskSteps))
}

func TestAppConfigAllowsMissingAndZeroProjectID(t *testing.T) {
	root := t.TempDir()
	writeMCHFixture(t, root, "backend_url: http://backend.test\n"+"temp_dir: /workspace/custom-mch\n")

	missing, err := loadAppConfig(root)
	require.NoError(t, err)
	assert.Zero(t, missing.ProjectID)

	require.NoError(t, os.WriteFile(filepath.Join(root, ".mch", "config.yaml"), []byte("backend_url: http://backend.test\n"+"temp_dir: /workspace/custom-mch\n"+"project_id: 0\n"), 0o644))
	zero, err := loadAppConfig(root)
	require.NoError(t, err)
	assert.Zero(t, zero.ProjectID)
}

func TestSaveAppConfigPersistsRepositoryProjectIDAndTempDir(t *testing.T) {
	root := t.TempDir()
	writeMCHFixture(t, root, "backend_url: http://backend.test\n"+"temp_dir: /workspace/custom-mch\n"+"project_id: 0\n")
	path := filepath.Join(root, ".mch", "config.yaml")
	cfg, err := loadAppConfig(root)
	require.NoError(t, err)

	cfg.ProjectID = 11
	require.NoError(t, saveAppConfig(path, cfg))

	loaded, err := loadAppConfig(root)
	require.NoError(t, err)
	assert.Equal(t, "http://backend.test", loaded.BackendURL)
	assert.Equal(t, "/workspace/custom-mch", loaded.TempDir)
	assert.Equal(t, 11, loaded.ProjectID)
	body, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(body), "backend_url: http://backend.test")
	assert.Contains(t, string(body), "temp_dir: /workspace/custom-mch")
	assert.Contains(t, string(body), "project_id: 11")
}

func TestResolveGitRepositoryRootFindsRootFromNestedDirectory(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, exec.Command("git", "init", root).Run())
	nested := filepath.Join(root, "one", "two")
	require.NoError(t, os.MkdirAll(nested, 0o755))

	previous, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(nested))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(previous))
	})

	got, err := resolveGitRepositoryRoot(t.Context())

	require.NoError(t, err)
	assert.Equal(t, root, got)
}

func TestAppConfigErrorsWithoutFallbackToLegacyConfig(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "cli", ".config"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "cli", ".config", "config.yaml"), []byte("backend_url: http://legacy.test\nproject_id: 99\n"), 0o644))

	cfg, err := loadAppConfig(root)

	require.Error(t, err)
	assert.Contains(t, err.Error(), ".mch/config.yaml")
	assert.Equal(t, root, cfg.RepositoryRoot)
	assert.Equal(t, filepath.Join(root, ".mch", "config.yaml"), cfg.ConfigPath)
	_, statErr := os.Stat(filepath.Join(root, ".mch", "config.yaml"))
	assert.True(t, os.IsNotExist(statErr))
}

func TestAppConfigErrorsOnMalformedFlowAndEmptyHelpSlugs(t *testing.T) {
	root := t.TempDir()
	writeMCHFixture(t, root, "backend_url: http://backend.test\n"+"temp_dir: /workspace/custom-mch\n")
	require.NoError(t, os.WriteFile(filepath.Join(root, ".mch", "default", "flow.yaml"), []byte("version: 1\nslug: default\nname: Default\nhelp: help.yaml\nmakefile: Makefile\nsteps:\n  - slug: unknown\n    mode: exec\n"), 0o644))

	_, err := loadAppConfig(root)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "flow steps must contain the default ordered stages")

	writeMCHFixture(t, root, "backend_url: http://backend.test\n"+"temp_dir: /workspace/custom-mch\n")
	require.NoError(t, os.WriteFile(filepath.Join(root, ".mch", "default", "help.yaml"), []byte("stage_modes:\n  - slug: ''\n"), 0o644))

	_, err = loadAppConfig(root)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "stage_modes option slug is required")
}

func TestAppConfigErrorsOnDuplicateFlowStepSlugAndInvalidMode(t *testing.T) {
	root := t.TempDir()
	writeMCHFixture(t, root, "backend_url: http://backend.test\n"+"temp_dir: /workspace/custom-mch\n")
	flow := testFlowYAML()
	flow = strings.Replace(flow, "  - slug: spec\n", "  - slug: idea\n", 1)
	require.NoError(t, os.WriteFile(filepath.Join(root, ".mch", "default", "flow.yaml"), []byte(flow), 0o644))

	_, err := loadAppConfig(root)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicates slug \"idea\"")

	writeMCHFixture(t, root, "backend_url: http://backend.test\n"+"temp_dir: /workspace/custom-mch\n")
	flow = strings.Replace(testFlowYAML(), "    mode: prompt\n", "    mode: custom\n", 1)
	require.NoError(t, os.WriteFile(filepath.Join(root, ".mch", "default", "flow.yaml"), []byte(flow), 0o644))

	_, err = loadAppConfig(root)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported Flow step mode \"custom\"")

	writeMCHFixture(t, root, "backend_url: http://backend.test\n"+"temp_dir: /workspace/custom-mch\n")
	flow = strings.Replace(testFlowYAML(), "    mode: prompt\n", "    mode: \n", 1)
	require.NoError(t, os.WriteFile(filepath.Join(root, ".mch", "default", "flow.yaml"), []byte(flow), 0o644))

	_, err = loadAppConfig(root)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "flow step 1 mode is required")
}

func TestAppConfigAllowsCustomAndMissingFlowHelpOptions(t *testing.T) {
	root := t.TempDir()
	writeMCHFixture(t, root, "backend_url: http://backend.test\n"+"temp_dir: /workspace/custom-mch\n")
	require.NoError(t, os.WriteFile(filepath.Join(root, ".mch", "default", "help.yaml"), []byte("stage_modes:\n  - slug: custom-mode\n    help: custom mode\n"), 0o644))

	cfg, err := loadAppConfig(root)

	require.NoError(t, err)
	assert.Equal(t, []string{"custom-mode"}, flowOptionSlugs(cfg.FlowHelp.StageModes))
	assert.Empty(t, cfg.FlowHelp.TaskStatuses)
	assert.Empty(t, cfg.FlowHelp.TaskSteps)
}

func testAppConfig(overrides appConfig) appConfig {
	cfg := appConfig{
		RepositoryRoot: "/repo",
		ConfigPath:     "/repo/.mch/config.yaml",
		BackendURL:     defaultBackendURL,
		TempDir:        "/workspace/mch",
		FlowDir:        "/repo/.mch/default",
		Flow: flowConfig{
			Version:        1,
			Slug:           "default",
			Name:           "Default Change Automation",
			Description:    "Default test Flow.",
			Help:           "help.yaml",
			Makefile:       "Makefile",
			Steps:          []flowStep{{Slug: "idea", Help: "capture idea", Mode: "prompt", Prompt: "prompts/change-idea.md", Entry: "make idea-entry", Exec: "make idea-exec", Exit: "make idea-exit"}},
			UtilityPrompts: map[string]string{"change-idea-tmp": "prompts/change-idea-tmp.md"},
		},
		FlowHelp: flowHelpConfig{
			Version:      1,
			StageModes:   []flowOption{{Slug: "skip", Help: "skip"}, {Slug: "prompt", Help: "prompt"}, {Slug: "exec", Help: "exec"}},
			TaskStatuses: []flowOption{{Slug: "queued", Help: "queued"}, {Slug: "running", Help: "running"}, {Slug: "paused", Help: "paused"}, {Slug: "stopped", Help: "stopped"}, {Slug: "waiting", Help: "waiting"}, {Slug: "completed", Help: "completed"}, {Slug: "failed", Help: "failed"}},
			TaskSteps:    []flowOption{{Slug: "none", Help: "none"}, {Slug: "entry", Help: "entry"}, {Slug: "prompt", Help: "prompt"}, {Slug: "agent", Help: "agent"}, {Slug: "exit", Help: "exit"}, {Slug: "done", Help: "done"}},
		},
	}
	if overrides.RepositoryRoot != "" {
		cfg.RepositoryRoot = overrides.RepositoryRoot
	}
	if overrides.ConfigPath != "" {
		cfg.ConfigPath = overrides.ConfigPath
	}
	if overrides.BackendURL != "" {
		cfg.BackendURL = overrides.BackendURL
	}
	if overrides.TempDir != "" {
		cfg.TempDir = overrides.TempDir
	}
	if overrides.ProjectID != 0 {
		cfg.ProjectID = overrides.ProjectID
	}
	if overrides.FlowDir != "" {
		cfg.FlowDir = overrides.FlowDir
	}
	if overrides.Flow.Slug != "" {
		cfg.Flow = overrides.Flow
	}
	if len(overrides.FlowHelp.StageModes) > 0 {
		cfg.FlowHelp = overrides.FlowHelp
	}
	return cfg
}

func writeMCHFixture(t *testing.T, root string, config string) {
	t.Helper()
	flowDir := filepath.Join(root, ".mch", "default")
	require.NoError(t, os.MkdirAll(filepath.Join(flowDir, "prompts"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".mch", "config.yaml"), []byte(config), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(flowDir, "flow.yaml"), []byte(testFlowYAML()), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(flowDir, "help.yaml"), []byte(testHelpYAML()), 0o644))
}

func testFlowYAML() string {
	var b strings.Builder
	b.WriteString(`version: 1
slug: default
name: Default Change Automation
description: Test Flow.
help: help.yaml
makefile: Makefile
steps:
`)
	for _, slug := range defaultFlowStageSlugs {
		mode := "exec"
		if slug == "idea" || slug == "polish" {
			mode = "prompt"
		}
		fmt.Fprintf(&b, "  - slug: %s\n", slug)
		fmt.Fprintf(&b, "    help: %s help\n", slug)
		fmt.Fprintf(&b, "    mode: %s\n", mode)
		fmt.Fprintf(&b, "    prompt: prompts/%s.md\n", slug)
		fmt.Fprintf(&b, "    entry: make %s-entry\n", slug)
		fmt.Fprintf(&b, "    exec: make %s-exec\n", slug)
		fmt.Fprintf(&b, "    exit: make %s-exit\n", slug)
	}
	b.WriteString(`utility_prompts:
  change-idea-tmp: prompts/change-idea-tmp.md
`)
	return b.String()
}

func testHelpYAML() string {
	return `version: 1
stage_modes:
  - slug: skip
    help: stage will not execute
  - slug: prompt
    help: stage will run an interactive session
  - slug: exec
    help: stage will run an automated agent
task_statuses:
  - slug: queued
    help: task is waiting to start
  - slug: running
    help: task is actively executing
  - slug: paused
    help: task is temporarily paused
  - slug: stopped
    help: task was manually stopped
  - slug: waiting
    help: task is waiting for input
  - slug: completed
    help: task finished successfully
  - slug: failed
    help: task finished with an error
task_steps:
  - slug: none
    help: task has not started yet
  - slug: entry
    help: entry script is executing
  - slug: prompt
    help: prompt is being prepared or shown
  - slug: agent
    help: automated agent is executing
  - slug: exit
    help: exit script is executing
  - slug: done
    help: task has finished
`
}

func flowOptionSlugs(options []flowOption) []string {
	slugs := make([]string, 0, len(options))
	for _, option := range options {
		slugs = append(slugs, option.Slug)
	}
	return slugs
}
