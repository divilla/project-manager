package requirement_test

import (
	"aipm/api-tests/shared"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type project struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type task struct {
	ID        int    `json:"id"`
	Version   int16  `json:"version"`
	DoneReq   int16  `json:"done_req"`
	TotalReq  int16  `json:"total_req"`
	Completed int16  `json:"completed"`
	Name      string `json:"name"`
}

type requirement struct {
	ID         int    `json:"id"`
	Version    int16  `json:"version"`
	TaskID     int    `json:"task_id"`
	Definition string `json:"definition"`
	Done       bool   `json:"done"`
}

type requirementMutation struct {
	Requirement  *requirement  `json:"requirement"`
	Task         task          `json:"task"`
	Requirements []requirement `json:"requirements"`
}

func TestRequirementCRUDRecalculatesTaskCompleteness(t *testing.T) {
	client := shared.NewClient(t)
	db := shared.NewDB(t)

	projectID := createProject(t, client)
	defer client.Post(t, "/api/v1/project/delete", map[string]any{"id": projectID}, nil)
	taskID := createTask(t, client, projectID)

	var listed []requirement
	status := client.Post(t, "/api/v1/requirement/list", map[string]any{"task_id": taskID}, &listed)
	require.Equal(t, http.StatusOK, status)
	assert.Empty(t, listed)

	first := createRequirement(t, client, taskID, "Add requirement create endpoint test.")
	second := createRequirement(t, client, taskID, "Add requirement update endpoint test.")
	assert.Equal(t, taskID, first.TaskID)
	assert.False(t, first.Done)
	assert.False(t, second.Done)

	status = client.Post(t, "/api/v1/requirement/list", map[string]any{"task_id": taskID}, &listed)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, listed, 2)

	var updated requirementMutation
	status = client.Post(t, "/api/v1/requirement/update-done", map[string]any{
		"id": first.ID, "done": true,
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.Requirement)
	assert.True(t, updated.Requirement.Done)
	assert.Equal(t, int16(50), updated.Task.Completed)
	require.Len(t, updated.Requirements, 2)
	shared.AssertHistoryCount(t, db, "requirement_history", first.ID, false, 0)
	assert.Equal(t, first.Version, updated.Requirement.Version)
	first = *updated.Requirement

	status = client.Post(t, "/api/v1/requirement/update-done", map[string]any{
		"id": second.ID, "done": true,
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, int16(100), updated.Task.Completed)
	second = *updated.Requirement

	status = client.Post(t, "/api/v1/requirement/update", map[string]any{
		"id":         second.ID,
		"definition": "Text-only update preserves completion state.",
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.Requirement)
	assert.True(t, updated.Requirement.Done)
	assert.Equal(t, int16(100), updated.Task.Completed)
	second = *updated.Requirement

	status = client.Post(t, "/api/v1/requirement/update", map[string]any{
		"id":         second.ID,
		"definition": "Add requirement toggle endpoint test.",
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.Requirement)
	assert.Equal(t, "Add requirement toggle endpoint test.", updated.Requirement.Definition)
	second = *updated.Requirement

	status = client.Post(t, "/api/v1/requirement/update-done", map[string]any{
		"id": second.ID, "done": false,
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.Requirement)
	assert.False(t, updated.Requirement.Done)
	assert.Equal(t, int16(50), updated.Task.Completed)
	shared.AssertHistoryCountAtLeast(t, db, "requirement_history", second.ID, false, 2)

	newTaskID := createTask(t, client, projectID)
	previousVersion := updated.Requirement.Version
	status = client.Post(t, "/api/v1/requirement/update-task", map[string]any{
		"id": second.ID, "task_id": newTaskID,
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.Requirement)
	assert.Equal(t, newTaskID, updated.Requirement.TaskID)
	assert.Equal(t, previousVersion, updated.Requirement.Version)
	assert.Equal(t, newTaskID, updated.Task.ID)

	var deleted requirementMutation
	status = client.Post(t, "/api/v1/requirement/delete", map[string]any{"id": first.ID}, &deleted)
	require.Equal(t, http.StatusOK, status)
	assert.Nil(t, deleted.Requirement)
	assert.Equal(t, int16(0), deleted.Task.Completed)
	assert.Empty(t, deleted.Requirements)
	shared.AssertHistoryDeleted(t, db, "requirement_history", first.ID)
}

func TestRequirementCountersPropagateToAncestors(t *testing.T) {
	client := shared.NewClient(t)
	projectID := createProject(t, client)
	defer client.Post(t, "/api/v1/project/delete", map[string]any{"id": projectID}, nil)

	var parent task
	status := client.Post(t, "/api/v1/task/create", map[string]any{
		"project_id": projectID,
		"name":       fmt.Sprintf("api-test-aggregate-parent-%d", time.Now().UnixNano()),
	}, &parent)
	require.Equal(t, http.StatusCreated, status)

	var child task
	status = client.Post(t, "/api/v1/task/create", map[string]any{
		"project_id": projectID,
		"parent_id":  parent.ID,
		"name":       fmt.Sprintf("api-test-aggregate-child-%d", time.Now().UnixNano()),
	}, &child)
	require.Equal(t, http.StatusCreated, status)

	requirement := createRequirement(t, client, child.ID, "Verify counters reach the parent task.")
	var parentDetail struct {
		Task task `json:"task"`
	}
	status = client.Post(t, "/api/v1/task/get", map[string]any{"id": parent.ID}, &parentDetail)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, int16(0), parentDetail.Task.DoneReq)
	assert.Equal(t, int16(1), parentDetail.Task.TotalReq)

	var mutation requirementMutation
	status = client.Post(t, "/api/v1/requirement/update-done", map[string]any{
		"id": requirement.ID, "done": true,
	}, &mutation)
	require.Equal(t, http.StatusOK, status)

	status = client.Post(t, "/api/v1/task/get", map[string]any{"id": parent.ID}, &parentDetail)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, int16(1), parentDetail.Task.DoneReq)
	assert.Equal(t, int16(1), parentDetail.Task.TotalReq)
	assert.Equal(t, int16(100), parentDetail.Task.Completed)
}

func TestRequirementRejectsInvalidInputAndMissingRows(t *testing.T) {
	client := shared.NewClient(t)

	status := client.Post(t, "/api/v1/requirement/create", map[string]any{
		"task_id":    -1,
		"definition": "orphan requirement",
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/requirement/create", map[string]any{
		"task_id":    -1,
		"definition": "   ",
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/requirement/update", map[string]any{
		"id":         999999999,
		"definition": "missing requirement",
	}, nil)
	assert.Equal(t, http.StatusNotFound, status)
}

func createProject(t *testing.T, client *shared.Client) int {
	t.Helper()

	var created project
	status := client.Post(t, "/api/v1/project/create", map[string]string{
		"name": fmt.Sprintf("api-test-requirement-project-%d", time.Now().UnixNano()),
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)

	return created.ID
}

func createTask(t *testing.T, client *shared.Client, projectID int) int {
	t.Helper()

	var created task
	status := client.Post(t, "/api/v1/task/create", map[string]any{
		"project_id": projectID,
		"name":       fmt.Sprintf("api-test-requirement-task-%d", time.Now().UnixNano()),
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, int16(0), created.Completed)

	return created.ID
}

func createRequirement(t *testing.T, client *shared.Client, taskID int, definition string) requirement {
	t.Helper()

	var created requirementMutation
	status := client.Post(t, "/api/v1/requirement/create", map[string]any{
		"task_id":    taskID,
		"definition": definition,
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotNil(t, created.Requirement)
	require.NotEmpty(t, created.Requirement.ID)
	assert.Equal(t, int16(0), created.Task.Completed)

	return *created.Requirement
}
