package suite

import (
	"testing"

	"apih/internal/models"
)

func TestDocumentKindsUseRootDefaultsAndStepsVocabulary(t *testing.T) {
	want := map[models.DocumentKind]string{
		models.KindRoot:     "root",
		models.KindDefaults: "defaults",
		models.KindSteps:    "steps",
	}
	if len(want) != 3 {
		t.Fatal("document kind contract must contain exactly three kinds")
	}
	for kind, value := range want {
		if string(kind) != value {
			t.Errorf("document kind = %q, want %q", kind, value)
		}
	}
}

func TestSuiteFileUsesDefaultsLifecycle(t *testing.T) {
	declared := models.Defaults{BaseURL: "https://example.test"}
	file := models.File{
		Kind:            models.KindDefaults,
		ParentDefaults:  &declared,
		RuntimeDefaults: models.RuntimeDefaults{BaseURL: declared.BaseURL},
	}

	if file.Kind != models.KindDefaults {
		t.Fatalf("file kind = %q, want %q", file.Kind, models.KindDefaults)
	}
	if file.ParentDefaults != &declared {
		t.Fatal("file did not retain parent defaults")
	}
	if file.RuntimeDefaults.BaseURL != declared.BaseURL {
		t.Fatalf("runtime base URL = %q, want %q", file.RuntimeDefaults.BaseURL, declared.BaseURL)
	}
}
