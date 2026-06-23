package task_test

import (
	"aipm/api-tests/shared"
	"context"
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

type referenceOption struct {
	Slug     string `json:"slug"`
	Priority int    `json:"priority"`
}

type taskReferences struct {
	Phases []referenceOption `json:"phases"`
	Types  []referenceOption `json:"types"`
}

type task struct {
	ID          int    `json:"id"`
	Version     int16  `json:"version"`
	ProjectID   int    `json:"project_id"`
	TaskPhase   string `json:"task_phase"`
	TaskType    string `json:"task_type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Difficulty  int16  `json:"difficulty"`
	Priority    int16  `json:"priority"`
	ParentID    *int   `json:"parent_id"`
	DoneReq     int16  `json:"done_req"`
	TotalReq    int16  `json:"total_req"`
	Completed   int16  `json:"completed"`
}

type taskDetail struct {
	Task         task  `json:"task"`
	Requirements []any `json:"requirements"`
}

type renderedDescriptionsResponse struct {
	Descriptions []struct {
		ID              int    `json:"id"`
		DescriptionHTML string `json:"description_html"`
	} `json:"descriptions"`
}

func TestTaskCRUDAndReferences(t *testing.T) {
	client := shared.NewClient(t)
	db := shared.NewDB(t)

	var refs taskReferences
	status := client.Post(t, "/api/v1/task/reference", map[string]any{}, &refs)
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, refs.Phases)
	require.NotEmpty(t, refs.Types)

	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)

	taskName := fmt.Sprintf("api-test-task-%d", time.Now().UnixNano())
	var created task
	status = client.Post(t, "/api/v1/task/create", map[string]any{
		"project_id":  projectID,
		"name":        taskName,
		"description": "Created by task API integration test.",
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, projectID, created.ProjectID)
	assert.Equal(t, taskName, created.Name)
	assert.Equal(t, int16(0), created.Completed)

	var listed []task
	status = client.Post(t, "/api/v1/task/list", map[string]any{"project_id": projectID}, &listed)
	require.Equal(t, http.StatusOK, status)
	assert.Contains(t, listed, created)

	var detail taskDetail
	status = client.Post(t, "/api/v1/task/get", map[string]any{"id": created.ID}, &detail)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, created.ID, detail.Task.ID)
	assert.Empty(t, detail.Requirements)

	var rendered renderedDescriptionsResponse
	status = client.Post(t, "/api/v1/task/rendered-descriptions", map[string]any{
		"ids": []int{created.ID},
	}, &rendered)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, rendered.Descriptions, 1)
	assert.Equal(t, created.ID, rendered.Descriptions[0].ID)
	assert.Contains(t, rendered.Descriptions[0].DescriptionHTML, "<p>Created by task API integration test.</p>")

	var updated task
	status = client.Post(t, "/api/v1/task/update", map[string]any{
		"id":          created.ID,
		"name":        taskName + "-updated",
		"description": "Updated by task API integration test.",
		"task_type":   refs.Types[0].Slug,
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, taskName+"-updated", updated.Name)
	assert.Equal(t, refs.Types[0].Slug, updated.TaskType)
	assert.Equal(t, created.Version+1, updated.Version)
	shared.AssertHistoryNotDeleted(t, db, "task_history", created.ID)

	nextPhase := refs.Phases[0].Slug
	for _, phase := range refs.Phases {
		if phase.Slug != updated.TaskPhase {
			nextPhase = phase.Slug
			break
		}
	}

	var moved task
	status = client.Post(t, "/api/v1/task/update-phase", map[string]any{"id": created.ID, "task_phase": nextPhase}, &moved)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, nextPhase, moved.TaskPhase)
	assert.Equal(t, updated.Version, moved.Version)
	shared.AssertHistoryCount(t, db, "task_history", created.ID, false, 1)

	status = client.Post(t, "/api/v1/task/delete", map[string]any{"id": created.ID}, nil)
	require.Equal(t, http.StatusNoContent, status)
	shared.AssertHistoryDeleted(t, db, "task_history", created.ID)

	status = client.Post(t, "/api/v1/task/get", map[string]any{"id": created.ID}, nil)
	assert.Equal(t, http.StatusNotFound, status)
}

func TestTaskCreateRejectsInvalidReferences(t *testing.T) {
	client := shared.NewClient(t)

	status := client.Post(t, "/api/v1/task/create", map[string]any{
		"project_id": 999999999,
		"name":       "orphan task",
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)

	status = client.Post(t, "/api/v1/task/create", map[string]any{
		"project_id": projectID,
		"parent_id":  999999999,
		"name":       "orphan child task",
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestTaskDeleteArchivesAndRemovesChildTaskTree(t *testing.T) {
	client := shared.NewClient(t)
	db := shared.NewDB(t)
	ctx := context.Background()

	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)

	var parent task
	status := client.Post(t, "/api/v1/task/create", map[string]any{
		"project_id": projectID,
		"name":       fmt.Sprintf("api-test-parent-task-%d", time.Now().UnixNano()),
	}, &parent)
	require.Equal(t, http.StatusCreated, status)

	var child task
	status = client.Post(t, "/api/v1/task/create", map[string]any{
		"project_id": projectID,
		"parent_id":  parent.ID,
		"name":       fmt.Sprintf("api-test-child-task-%d", time.Now().UnixNano()),
	}, &child)
	require.Equal(t, http.StatusCreated, status)

	var requirementID int
	err := db.QueryRow(ctx, `
		insert into requirement (definition, task_id)
		values ($1, $2)
		returning id
	`, "Task tree delete archives this requirement.", child.ID).Scan(&requirementID)
	require.NoError(t, err)

	status = client.Post(t, "/api/v1/task/delete", map[string]any{"id": parent.ID}, nil)
	require.Equal(t, http.StatusNoContent, status)

	status = client.Post(t, "/api/v1/task/get", map[string]any{"id": parent.ID}, nil)
	assert.Equal(t, http.StatusNotFound, status)
	status = client.Post(t, "/api/v1/task/get", map[string]any{"id": child.ID}, nil)
	assert.Equal(t, http.StatusNotFound, status)

	shared.AssertHistoryDeleted(t, db, "task_history", parent.ID)
	shared.AssertHistoryDeleted(t, db, "task_history", child.ID)
	shared.AssertHistoryDeleted(t, db, "requirement_history", requirementID)
}

func createProject(t *testing.T, client *shared.Client) int {
	t.Helper()

	var created project
	status := client.Post(t, "/api/v1/project/create", map[string]string{
		"name": fmt.Sprintf("api-test-task-project-%d", time.Now().UnixNano()),
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)

	return created.ID
}
