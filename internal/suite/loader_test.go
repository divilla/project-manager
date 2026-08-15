package suite

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestLoaderFindsYAMLFilesInWorkDir(t *testing.T) {
	workDir := t.TempDir()
	yamlPath := filepath.Join(workDir, "root.yaml")
	ymlPath := filepath.Join(workDir, "settings.yml")
	writeTestFile(t, yamlPath)
	writeTestFile(t, ymlPath)

	got, err := NewLoader(workDir).Files()
	if err != nil {
		t.Fatalf("Files() error = %v", err)
	}
	want := []string{yamlPath, ymlPath}
	sort.Strings(got)
	sort.Strings(want)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Files() = %v, want %v", got, want)
	}
}

func TestLoaderFindsYAMLFilesRecursively(t *testing.T) {
	workDir := t.TempDir()
	firstPath := filepath.Join(workDir, "one", "defaults.yaml")
	secondPath := filepath.Join(workDir, "one", "two", "settings.yml")
	writeTestFile(t, firstPath)
	writeTestFile(t, secondPath)

	got, err := NewLoader(workDir).Files()
	if err != nil {
		t.Fatalf("Files() error = %v", err)
	}
	want := []string{firstPath, secondPath}
	sort.Strings(got)
	sort.Strings(want)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Files() = %v, want %v", got, want)
	}
}

func TestLoaderExcludesNonYAMLEntries(t *testing.T) {
	workDir := t.TempDir()
	want := filepath.Join(workDir, "root.yaml")
	writeTestFile(t, want)
	writeTestFile(t, filepath.Join(workDir, "root.json"))
	writeTestFile(t, filepath.Join(workDir, "root.YAML"))
	if err := os.Mkdir(filepath.Join(workDir, "directory.yml"), 0o755); err != nil {
		t.Fatalf("create directory: %v", err)
	}

	got, err := NewLoader(workDir).Files()
	if err != nil {
		t.Fatalf("Files() error = %v", err)
	}
	if !reflect.DeepEqual(got, []string{want}) {
		t.Fatalf("Files() = %v, want [%s]", got, want)
	}
}

func TestLoaderHandlesWorkDirWithoutYAMLFiles(t *testing.T) {
	workDir := t.TempDir()

	got, err := NewLoader(workDir).Files()
	if err != nil {
		t.Fatalf("Files() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Files() = %v, want no paths", got)
	}
}

func TestLoaderHandlesInvalidWorkDir(t *testing.T) {
	parentDir := t.TempDir()
	nonDirectory := filepath.Join(parentDir, "root.yaml")
	writeTestFile(t, nonDirectory)

	for _, workDir := range []string{
		filepath.Join(parentDir, "missing"),
		nonDirectory,
	} {
		got, err := NewLoader(workDir).Files()
		if err == nil {
			t.Errorf("Files() for %q error = nil, want an error", workDir)
		}
		if len(got) != 0 {
			t.Errorf("Files() for %q = %v, want no paths", workDir, got)
		}
	}
}

func TestLoaderPreservesPublicAPI(t *testing.T) {
	var constructor func(string) *Loader = NewLoader
	var filesMethod func(*Loader) ([]string, error) = (*Loader).Files

	loader := constructor(t.TempDir())
	_, _ = filesMethod(loader)
}

func writeTestFile(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent directory: %v", err)
	}
	if err := os.WriteFile(path, []byte("test: true\n"), 0o600); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}
