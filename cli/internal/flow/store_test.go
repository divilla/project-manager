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
	getIDs    []int
	ideaCalls []artifactAPICall
	specCalls []artifactAPICall
	prCalls   []artifactAPICall
	getErr    error
	updateErr error
}

type artifactAPICall struct {
	id        int
	content   string
	agentEdit bool
}

func (a *fakeChangeAPI) GetChange(id int) (dto.Change, error) {
	a.getIDs = append(a.getIDs, id)
	return a.change, a.getErr
}

func (a *fakeChangeAPI) UpdateChangeIdea(id int, content string, agentEdit bool) (dto.Change, error) {
	a.ideaCalls = append(a.ideaCalls, artifactAPICall{id: id, content: content, agentEdit: agentEdit})
	return a.change, a.updateErr
}

func (a *fakeChangeAPI) UpdateChangeSpec(id int, content string, agentEdit bool) (dto.Change, error) {
	a.specCalls = append(a.specCalls, artifactAPICall{id: id, content: content, agentEdit: agentEdit})
	return a.change, a.updateErr
}

func (a *fakeChangeAPI) UpdateChangePR(id int, content string, agentEdit bool) (dto.Change, error) {
	a.prCalls = append(a.prCalls, artifactAPICall{id: id, content: content, agentEdit: agentEdit})
	return a.change, a.updateErr
}

func TestChangeArtifactStoreLoadsAndSavesFocusedArtifacts(t *testing.T) {
	for _, test := range []struct {
		name     string
		artifact Artifact
		loaded   string
		calls    func(*fakeChangeAPI) []artifactAPICall
	}{
		{name: "Idea", artifact: ArtifactIdea, loaded: "idea", calls: func(api *fakeChangeAPI) []artifactAPICall { return api.ideaCalls }},
		{name: "Spec", artifact: ArtifactSpec, loaded: "spec", calls: func(api *fakeChangeAPI) []artifactAPICall { return api.specCalls }},
		{name: "PR", artifact: ArtifactPR, loaded: "pr", calls: func(api *fakeChangeAPI) []artifactAPICall { return api.prCalls }},
	} {
		t.Run(test.name, func(t *testing.T) {
			api := &fakeChangeAPI{change: dto.Change{Idea: "idea", Spec: "spec", PR: "pr"}}
			store := NewChangeArtifactStore(api)

			loaded, err := store.Load(17, test.artifact)
			require.NoError(t, err)
			assert.Equal(t, test.loaded, string(loaded))
			require.NoError(t, store.Save(17, test.artifact, []byte("changed")))

			assert.Equal(t, []int{17}, api.getIDs)
			require.Len(t, test.calls(api), 1)
			assert.Equal(t, artifactAPICall{id: 17, content: "changed", agentEdit: false}, test.calls(api)[0])
			assert.Len(t, api.ideaCalls, boolToInt(test.artifact == ArtifactIdea))
			assert.Len(t, api.specCalls, boolToInt(test.artifact == ArtifactSpec))
			assert.Len(t, api.prCalls, boolToInt(test.artifact == ArtifactPR))
		})
	}
}

func TestChangeArtifactStoreRejectsUnsupportedBlankAndFailedOperations(t *testing.T) {
	api := &fakeChangeAPI{change: dto.Change{Idea: "idea"}}
	store := NewChangeArtifactStore(api)

	_, err := store.Load(1, ArtifactImplement)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not support")
	assert.Empty(t, api.getIDs)
	err = store.Save(1, ArtifactFinalize, []byte("content"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not support")
	err = store.Save(1, ArtifactIdea, []byte(" \n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not be blank")
	assert.Empty(t, api.ideaCalls)

	api.getErr = errors.New("get failed")
	_, err = store.Load(1, ArtifactIdea)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get failed")
	api.getErr = nil
	api.updateErr = errors.New("save failed")
	err = store.Save(1, ArtifactIdea, []byte("changed"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save failed")
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
