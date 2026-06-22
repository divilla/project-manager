package requirement_test

import (
	"aipm/api-tests/shared"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequirementDeleteLastItemReturnsZeroCompletedFromView(t *testing.T) {
	client := shared.NewClient(t)
	projectID := createProject(t, client)
	defer client.Post(t, "/api/v1/project/delete", map[string]any{"id": projectID}, nil)
	taskID := createTask(t, client, projectID)
	requirement := createRequirement(t, client, taskID, "Delete the final requirement transactionally.")

	var deleted requirementMutation
	status := client.Post(t, "/api/v1/requirement/delete", map[string]any{
		"id": requirement.ID,
	}, &deleted)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, int16(0), deleted.Task.Completed)
	assert.Equal(t, int16(0), deleted.Task.DoneReq)
	assert.Equal(t, int16(0), deleted.Task.TotalReq)
	assert.Empty(t, deleted.Requirements)
}

func TestRequirementDoneUpdateDoesNotIncrementVersion(t *testing.T) {
	client := shared.NewClient(t)
	projectID := createProject(t, client)
	defer client.Post(t, "/api/v1/project/delete", map[string]any{"id": projectID}, nil)
	taskID := createTask(t, client, projectID)
	requirement := createRequirement(t, client, taskID, "Done changes preserve the version.")

	var updated requirementMutation
	status := client.Post(t, "/api/v1/requirement/update-done", map[string]any{
		"id": requirement.ID, "done": true,
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.Requirement)
	assert.Equal(t, requirement.Version, updated.Requirement.Version)
	assert.True(t, updated.Requirement.Done)
}
