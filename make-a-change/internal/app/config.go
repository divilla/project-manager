package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	goconfig "github.com/ridgelines/go-config"
	"gopkg.in/yaml.v3"
)

const defaultConfigPath = ".config/config.yaml"

type appConfig struct {
	BackendURL string
	ProjectID  int
}

type configFile struct {
	BackendURL string `yaml:"backend_url"`
	ProjectID  int    `yaml:"project_id"`
}

func loadAppConfig(path string) (appConfig, error) {
	if err := ensureAppConfig(path); err != nil {
		return appConfig{BackendURL: defaultBackendURL}, err
	}

	cfg := goconfig.NewConfig([]goconfig.Provider{goconfig.NewYAMLFile(path)})
	backendURL, err := cfg.StringOr("backend_url", defaultBackendURL)
	if err != nil {
		return appConfig{BackendURL: defaultBackendURL}, err
	}
	backendURL = strings.TrimSpace(backendURL)
	if backendURL == "" {
		return appConfig{BackendURL: defaultBackendURL}, fmt.Errorf("backend_url is required")
	}
	projectID, err := cfg.IntOr("project_id", 0)
	if err != nil {
		return appConfig{BackendURL: backendURL}, err
	}
	return appConfig{BackendURL: backendURL, ProjectID: projectID}, nil
}

func resolveConfigPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if _, err := os.Stat(path); err == nil {
		return path
	}
	dir, err := os.Getwd()
	if err != nil {
		return path
	}
	for {
		candidate := filepath.Join(dir, path)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return path
		}
		dir = parent
	}
}

func ensureAppConfig(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return saveAppConfig(path, appConfig{BackendURL: defaultBackendURL})
}

func saveAppConfig(path string, cfg appConfig) error {
	backendURL := strings.TrimSpace(cfg.BackendURL)
	if backendURL == "" {
		backendURL = defaultBackendURL
	}
	body, err := yaml.Marshal(configFile{
		BackendURL: backendURL,
		ProjectID:  cfg.ProjectID,
	})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}
