package flow

import (
	"errors"
	"testing"

	"mch/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeChangeAPI struct {
	change    dto.Change
	agentEdit bool
	artifact  Artifact
	err       error
}

func (f *fakeChangeAPI) GetChange(int) (dto.Change, error) { return f.change, f.err }
func (f *fakeChangeAPI) UpdateChangeIdea(_ int, value string, agent bool) (dto.Change, error) {
	f.artifact, f.agentEdit, f.change.Idea = ArtifactIdea, agent, value
	return f.change, f.err
}
func (f *fakeChangeAPI) UpdateChangeSpec(_ int, value string, agent bool) (dto.Change, error) {
	f.artifact, f.agentEdit, f.change.Spec = ArtifactSpec, agent, value
	return f.change, f.err
}
func (f *fakeChangeAPI) UpdateChangePR(_ int, value string, agent bool) (dto.Change, error) {
	f.artifact, f.agentEdit, f.change.PR = ArtifactPR, agent, value
	return f.change, f.err
}

func TestChangeArtifactStoreUsesFocusedEndpointAndProvenance(t *testing.T) {
	api := &fakeChangeAPI{change: dto.Change{Idea: "idea", Spec: "spec", PR: "pr"}}
	store := NewChangeArtifactStore(api)
	for _, artifact := range []Artifact{ArtifactIdea, ArtifactSpec, ArtifactPR} {
		_, err := store.Load(1, artifact)
		require.NoError(t, err)
		require.NoError(t, store.Save(1, artifact, []byte("# Title\n\nBody"), SaveByAgent))
		assert.Equal(t, artifact, api.artifact)
		assert.True(t, api.agentEdit)
	}
	require.NoError(t, store.Save(1, ArtifactIdea, []byte("# Title\n\nBody"), SaveByUser))
	assert.False(t, api.agentEdit)
}

func TestChangeArtifactStoreRejectsUnsupportedAndFailures(t *testing.T) {
	store := NewChangeArtifactStore(&fakeChangeAPI{err: errors.New("backend")})
	_, err := store.Load(1, ArtifactImplement)
	require.Error(t, err)
	err = store.Save(1, ArtifactIdea, []byte("x"), SaveProvenance("unknown"))
	require.Error(t, err)
}
