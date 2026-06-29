package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppConfigCreatesDefaultsAndPersistsProjectID(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".config", "config.yaml")

	cfg, err := loadAppConfig(path)

	require.NoError(t, err)
	assert.Equal(t, defaultBackendURL, cfg.BackendURL)
	assert.Zero(t, cfg.ProjectID)
	require.FileExists(t, path)
	body, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(body), "project_id: 0")

	cfg.ProjectID = 7
	require.NoError(t, saveAppConfig(path, cfg))

	loaded, err := loadAppConfig(path)
	require.NoError(t, err)
	assert.Equal(t, defaultBackendURL, loaded.BackendURL)
	assert.Equal(t, 7, loaded.ProjectID)

	body, err = os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(body), "backend_url: http://localhost:8080")
	assert.Contains(t, string(body), "project_id: 7")
}

func TestResolveConfigPathFindsParentConfig(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module mch\n"), 0o644))
	path := filepath.Join(root, ".config", "config.yaml")
	require.NoError(t, saveAppConfig(path, appConfig{BackendURL: defaultBackendURL}))
	nested := filepath.Join(root, "internal", "app")
	require.NoError(t, os.MkdirAll(nested, 0o755))

	previous, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(nested))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(previous))
	})

	assert.Equal(t, path, resolveConfigPath(defaultConfigPath))
}

func TestResolveConfigPathUsesModuleRootWhenConfigIsMissing(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module mch\n"), 0o644))
	nested := filepath.Join(root, "internal", "app")
	require.NoError(t, os.MkdirAll(nested, 0o755))

	previous, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(nested))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(previous))
	})

	assert.Equal(t, filepath.Join(root, ".config", "config.yaml"), resolveConfigPath(defaultConfigPath))
}
