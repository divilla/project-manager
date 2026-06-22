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
	ID   string `json:"id"`
	Name string `json:"name"`
}

type task struct {
	ID       string `json:"id"`
	Complete int    `json:"complete"`
	Name     string `json:"name"`
}

type requirement struct {
	ID         string `json:"id"`
	TaskID     string `json:"task_id"`
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
	defer client.Post(t, "/api/project/delete", map[string]string{"id": projectID}, nil)
	taskID := createTask(t, client, projectID)

	var listed []requirement
	status := client.Post(t, "/api/requirement/list", map[string]string{"task_id": taskID}, &listed)
	require.Equal(t, http.StatusOK, status)
	assert.Empty(t, listed)

	first := createRequirement(t, client, taskID, "Add requirement create endpoint test.")
	second := createRequirement(t, client, taskID, "Add requirement update endpoint test.")
	assert.Equal(t, taskID, first.TaskID)
	assert.False(t, first.Done)
	assert.False(t, second.Done)

	status = client.Post(t, "/api/requirement/list", map[string]string{"task_id": taskID}, &listed)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, listed, 2)

	var updated requirementMutation
	status = client.Post(t, "/api/requirement/update", map[string]any{
		"id":         first.ID,
		"definition": first.Definition,
		"done":       true,
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.Requirement)
	assert.True(t, updated.Requirement.Done)
	assert.Equal(t, 50, updated.Task.Complete)
	require.Len(t, updated.Requirements, 2)
	shared.AssertHistoryNotDeleted(t, db, "requirement_history", first.ID)
	shared.AssertHistoryNotDeleted(t, db, "task_history", taskID)

	status = client.Post(t, "/api/requirement/update", map[string]any{
		"id":         second.ID,
		"definition": second.Definition,
		"done":       true,
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, 100, updated.Task.Complete)

	status = client.Post(t, "/api/requirement/update", map[string]any{
		"id":         second.ID,
		"definition": "Text-only update preserves completion state.",
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.Requirement)
	assert.True(t, updated.Requirement.Done)
	assert.Equal(t, 100, updated.Task.Complete)

	status = client.Post(t, "/api/requirement/update", map[string]any{
		"id":         second.ID,
		"definition": "Add requirement toggle endpoint test.",
		"done":       false,
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.Requirement)
	assert.Equal(t, "Add requirement toggle endpoint test.", updated.Requirement.Definition)
	assert.False(t, updated.Requirement.Done)
	assert.Equal(t, 50, updated.Task.Complete)
	shared.AssertHistoryCountAtLeast(t, db, "requirement_history", second.ID, false, 2)

	var deleted requirementMutation
	status = client.Post(t, "/api/requirement/delete", map[string]string{"id": first.ID}, &deleted)
	require.Equal(t, http.StatusOK, status)
	assert.Nil(t, deleted.Requirement)
	assert.Equal(t, 0, deleted.Task.Complete)
	require.Len(t, deleted.Requirements, 1)
	shared.AssertHistoryDeleted(t, db, "requirement_history", first.ID)
	shared.AssertHistoryCountAtLeast(t, db, "task_history", taskID, false, 3)
}

func TestRequirementRejectsInvalidInputAndMissingRows(t *testing.T) {
	client := shared.NewClient(t)

	status := client.Post(t, "/api/requirement/create", map[string]string{
		"task_id":    "00000000-0000-0000-0000-000000000001",
		"definition": "orphan requirement",
	}, nil)
	assert.Equal(t, http.StatusNotFound, status)

	status = client.Post(t, "/api/requirement/create", map[string]string{
		"task_id":    "00000000-0000-0000-0000-000000000001",
		"definition": "   ",
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/requirement/update", map[string]any{
		"id":         "00000000-0000-0000-0000-000000000002",
		"definition": "missing requirement",
		"done":       true,
	}, nil)
	assert.Equal(t, http.StatusNotFound, status)
}

func createProject(t *testing.T, client *shared.Client) string {
	t.Helper()

	var created project
	status := client.Post(t, "/api/project/create", map[string]string{
		"name": fmt.Sprintf("api-test-requirement-project-%d", time.Now().UnixNano()),
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)

	return created.ID
}

func createTask(t *testing.T, client *shared.Client, projectID string) string {
	t.Helper()

	var created task
	status := client.Post(t, "/api/task/create", map[string]string{
		"project_id": projectID,
		"name":       fmt.Sprintf("api-test-requirement-task-%d", time.Now().UnixNano()),
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, 0, created.Complete)

	return created.ID
}

func createRequirement(t *testing.T, client *shared.Client, taskID, definition string) requirement {
	t.Helper()

	var created requirementMutation
	status := client.Post(t, "/api/requirement/create", map[string]string{
		"task_id":    taskID,
		"definition": definition,
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotNil(t, created.Requirement)
	require.NotEmpty(t, created.Requirement.ID)
	assert.Equal(t, 0, created.Task.Complete)

	return *created.Requirement
}
