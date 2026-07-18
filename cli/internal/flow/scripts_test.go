package flow

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodexScriptsRequireWorkspaceEnvironment(t *testing.T) {
	scripts := []struct {
		name string
		args []string
	}{
		{name: "codex-exec-restore-session.sh", args: []string{"prompt.md"}},
		{name: "codex-resume-session.sh"},
	}
	for _, script := range scripts {
		t.Run(script.name, func(t *testing.T) {
			path := filepath.Join("..", "..", "..", ".mch", "default", "scripts", script.name)
			for _, test := range []struct {
				name string
				env  []string
				want string
			}{
				{name: "temp", env: []string{"MCH_TEMP_DIR=", "MCH_REF_UUID=", "MCH_ARTIFACT="}, want: "missing MCH_TEMP_DIR"},
				{name: "ref", env: []string{"MCH_TEMP_DIR=" + TmpDir, "MCH_REF_UUID=", "MCH_ARTIFACT=idea"}, want: "missing MCH_REF_UUID"},
				{name: "invalid ref", env: []string{"MCH_TEMP_DIR=" + TmpDir, "MCH_REF_UUID=bad", "MCH_ARTIFACT=idea"}, want: "invalid MCH_REF_UUID"},
				{name: "artifact", env: []string{"MCH_TEMP_DIR=" + TmpDir, "MCH_REF_UUID=" + testRefUUID, "MCH_ARTIFACT="}, want: "missing MCH_ARTIFACT"},
				{name: "unsupported artifact", env: []string{"MCH_TEMP_DIR=" + TmpDir, "MCH_REF_UUID=" + testRefUUID, "MCH_ARTIFACT=review"}, want: "unsupported MCH_ARTIFACT"},
			} {
				t.Run(test.name, func(t *testing.T) {
					command := exec.Command(path, script.args...)
					command.Env = append([]string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME")}, test.env...)
					output, err := command.CombinedOutput()
					require.Error(t, err)
					assert.Contains(t, string(output), test.want)
				})
			}
		})
	}
}

func TestCodexScriptsDoNotCreateMissingFlowTempRoot(t *testing.T) {
	repo := t.TempDir()
	bin := t.TempDir()
	git := filepath.Join(bin, "git")
	require.NoError(t, os.WriteFile(git, []byte("#!/bin/sh\nprintf '%s\\n' '"+repo+"'\n"), 0o755))
	prompt := filepath.Join(repo, "prompt.md")
	require.NoError(t, os.WriteFile(prompt, []byte("prompt"), 0o644))

	for _, script := range []struct {
		name string
		args []string
	}{
		{name: "codex-exec-restore-session.sh", args: []string{prompt}},
		{name: "codex-resume-session.sh"},
	} {
		t.Run(script.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join("..", "..", "..", ".mch", "default", "scripts", script.name))
			require.NoError(t, err)
			command := exec.Command(path, script.args...)
			command.Env = []string{
				"PATH=" + bin + ":" + os.Getenv("PATH"),
				"HOME=" + os.Getenv("HOME"),
				"MCH_TEMP_DIR=" + TmpDir,
				"MCH_REF_UUID=" + testRefUUID,
				"MCH_ARTIFACT=idea",
			}
			output, err := command.CombinedOutput()
			require.Error(t, err)
			assert.Contains(t, string(output), "missing Flow temp root")
			_, statErr := os.Stat(filepath.Join(repo, TmpDir))
			assert.ErrorIs(t, statErr, os.ErrNotExist)
		})
	}
}
