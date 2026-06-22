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
	ID   string `json:"id"`
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
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	Phase       string `json:"phase"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Complete    int    `json:"complete"`
}

type taskDetail struct {
	Task         task  `json:"task"`
	Requirements []any `json:"requirements"`
}

func TestTaskCRUDAndReferences(t *testing.T) {
	client := shared.NewClient(t)
	db := shared.NewDB(t)

	var refs taskReferences
	status := client.Post(t, "/api/task/reference", map[string]any{}, &refs)
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, refs.Phases)
	require.NotEmpty(t, refs.Types)

	projectID := createProject(t, client)
	defer client.Post(t, "/api/project/delete", map[string]string{"id": projectID}, nil)

	taskName := fmt.Sprintf("api-test-task-%d", time.Now().UnixNano())
	var created task
	status = client.Post(t, "/api/task/create", map[string]any{
		"project_id":  projectID,
		"name":        taskName,
		"description": "Created by task API integration test.",
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, projectID, created.ProjectID)
	assert.Equal(t, taskName, created.Name)
	assert.Equal(t, 0, created.Complete)

	var listed []task
	status = client.Post(t, "/api/task/list", map[string]string{"project_id": projectID}, &listed)
	require.Equal(t, http.StatusOK, status)
	assert.Contains(t, listed, created)

	var detail taskDetail
	status = client.Post(t, "/api/task/get", map[string]string{"id": created.ID}, &detail)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, created.ID, detail.Task.ID)
	assert.Empty(t, detail.Requirements)

	var updated task
	status = client.Post(t, "/api/task/update", map[string]any{
		"id":          created.ID,
		"name":        taskName + "-updated",
		"description": "Updated by task API integration test.",
		"type":        refs.Types[0].Slug,
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, taskName+"-updated", updated.Name)
	assert.Equal(t, refs.Types[0].Slug, updated.Type)
	shared.AssertHistoryNotDeleted(t, db, "task_history", created.ID)

	nextPhase := refs.Phases[0].Slug
	for _, phase := range refs.Phases {
		if phase.Slug != updated.Phase {
			nextPhase = phase.Slug
			break
		}
	}

	var moved task
	status = client.Post(t, "/api/task/phase", map[string]string{"id": created.ID, "phase": nextPhase}, &moved)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, nextPhase, moved.Phase)
	shared.AssertHistoryCountAtLeast(t, db, "task_history", created.ID, false, 2)

	status = client.Post(t, "/api/task/delete", map[string]string{"id": created.ID}, nil)
	require.Equal(t, http.StatusNoContent, status)
	shared.AssertHistoryDeleted(t, db, "task_history", created.ID)

	status = client.Post(t, "/api/task/get", map[string]string{"id": created.ID}, nil)
	assert.Equal(t, http.StatusNotFound, status)
}

func TestTaskCreateRejectsInvalidReferences(t *testing.T) {
	client := shared.NewClient(t)

	status := client.Post(t, "/api/task/create", map[string]any{
		"project_id": "00000000-0000-0000-0000-000000000001",
		"name":       "orphan task",
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	projectID := createProject(t, client)
	defer client.Post(t, "/api/project/delete", map[string]string{"id": projectID}, nil)

	status = client.Post(t, "/api/task/create", map[string]any{
		"project_id": projectID,
		"parent_id":  "00000000-0000-0000-0000-000000000002",
		"name":       "orphan child task",
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestTaskDeleteArchivesAndRemovesChildTaskTree(t *testing.T) {
	client := shared.NewClient(t)
	db := shared.NewDB(t)
	ctx := context.Background()

	projectID := createProject(t, client)
	defer client.Post(t, "/api/project/delete", map[string]string{"id": projectID}, nil)

	var parent task
	status := client.Post(t, "/api/task/create", map[string]any{
		"project_id": projectID,
		"name":       fmt.Sprintf("api-test-parent-task-%d", time.Now().UnixNano()),
	}, &parent)
	require.Equal(t, http.StatusCreated, status)

	var child task
	status = client.Post(t, "/api/task/create", map[string]any{
		"project_id": projectID,
		"parent_id":  parent.ID,
		"name":       fmt.Sprintf("api-test-child-task-%d", time.Now().UnixNano()),
	}, &child)
	require.Equal(t, http.StatusCreated, status)

	var requirementID string
	err := db.QueryRow(ctx, `
		insert into requirement (definition, task_id)
		values ($1, $2)
		returning id
	`, "Task tree delete archives this requirement.", child.ID).Scan(&requirementID)
	require.NoError(t, err)

	status = client.Post(t, "/api/task/delete", map[string]string{"id": parent.ID}, nil)
	require.Equal(t, http.StatusNoContent, status)

	status = client.Post(t, "/api/task/get", map[string]string{"id": parent.ID}, nil)
	assert.Equal(t, http.StatusNotFound, status)
	status = client.Post(t, "/api/task/get", map[string]string{"id": child.ID}, nil)
	assert.Equal(t, http.StatusNotFound, status)

	shared.AssertHistoryDeleted(t, db, "task_history", parent.ID)
	shared.AssertHistoryDeleted(t, db, "task_history", child.ID)
	shared.AssertHistoryDeleted(t, db, "requirement_history", requirementID)
}

func createProject(t *testing.T, client *shared.Client) string {
	t.Helper()

	var created project
	status := client.Post(t, "/api/project/create", map[string]string{
		"name": fmt.Sprintf("api-test-task-project-%d", time.Now().UnixNano()),
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)

	return created.ID
}
