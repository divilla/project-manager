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
