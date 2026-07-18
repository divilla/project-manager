package flow

import (
	"fmt"
	"strings"

	"mch/internal/dto"
)

// ArtifactStore loads and saves plain artifact bytes for an active Change.
type ArtifactStore interface {
	Load(changeID int, artifact Artifact) ([]byte, error)
	Save(changeID int, artifact Artifact, content []byte, provenance SaveProvenance) error
}

// SaveProvenance records who produced a focused artifact save.
type SaveProvenance string

// Supported artifact save provenance values.
const (
	SaveByUser  SaveProvenance = "user"
	SaveByAgent SaveProvenance = "agent"
)

// ChangeAPI is the existing CLI API boundary needed by ChangeArtifactStore.
type ChangeAPI interface {
	GetChange(id int) (dto.Change, error)
	UpdateChangeIdea(id int, idea string, agentEdit bool) (dto.Change, error)
	UpdateChangeSpec(id int, spec string, agentEdit bool) (dto.Change, error)
	UpdateChangePR(id int, pr string, agentEdit bool) (dto.Change, error)
}

// ChangeArtifactStore adapts Idea, Spec, and PR artifacts to focused Change endpoints.
type ChangeArtifactStore struct {
	api ChangeAPI
}

// NewChangeArtifactStore creates a Change API-backed artifact store.
func NewChangeArtifactStore(api ChangeAPI) ChangeArtifactStore {
	return ChangeArtifactStore{api: api}
}

// Load loads one supported artifact from POST /api/v1/change/get through the CLI client.
func (s ChangeArtifactStore) Load(changeID int, artifact Artifact) ([]byte, error) {
	if s.api == nil {
		return nil, fmt.Errorf("change API is required")
	}
	if changeID <= 0 {
		return nil, fmt.Errorf("change ID must be a valid positive number")
	}
	if !supportedChangeArtifact(artifact) {
		return nil, fmt.Errorf("change artifact store does not support artifact %q", artifact)
	}
	change, err := s.api.GetChange(changeID)
	if err != nil {
		return nil, fmt.Errorf("load Change %d: %w", changeID, err)
	}
	switch artifact {
	case ArtifactIdea:
		return []byte(change.Idea), nil
	case ArtifactSpec:
		return []byte(change.Spec), nil
	case ArtifactPR:
		return []byte(change.PR), nil
	}
	return nil, fmt.Errorf("change artifact store does not support artifact %q", artifact)
}

// Save persists through the focused endpoint with explicit agent_edit provenance.
func (s ChangeArtifactStore) Save(changeID int, artifact Artifact, content []byte, provenance SaveProvenance) error {
	if s.api == nil {
		return fmt.Errorf("change API is required")
	}
	if changeID <= 0 {
		return fmt.Errorf("change ID must be a valid positive number")
	}
	if !supportedChangeArtifact(artifact) {
		return fmt.Errorf("change artifact store does not support artifact %q", artifact)
	}
	text := string(content)
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("%s artifact must not be blank", artifact)
	}
	var err error
	agentEdit := false
	switch provenance {
	case SaveByUser:
	case SaveByAgent:
		agentEdit = true
	default:
		return fmt.Errorf("unsupported artifact save provenance %q", provenance)
	}
	switch artifact {
	case ArtifactIdea:
		_, err = s.api.UpdateChangeIdea(changeID, text, agentEdit)
	case ArtifactSpec:
		_, err = s.api.UpdateChangeSpec(changeID, text, agentEdit)
	case ArtifactPR:
		_, err = s.api.UpdateChangePR(changeID, text, agentEdit)
	}
	if err != nil {
		return fmt.Errorf("save %s artifact for Change %d: %w", artifact, changeID, err)
	}
	return nil
}

func supportedChangeArtifact(artifact Artifact) bool {
	return artifact == ArtifactIdea || artifact == ArtifactSpec || artifact == ArtifactPR
}
