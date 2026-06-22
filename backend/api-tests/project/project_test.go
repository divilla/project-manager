package project_test

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

func TestProjectCRUD(t *testing.T) {
	client := shared.NewClient(t)
	name := fmt.Sprintf("api-test-project-%d", time.Now().UnixNano())
	updatedName := name + "-updated"

	var created project
	status := client.Post(t, "/api/project/create", map[string]string{"name": name}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, name, created.Name)

	defer client.Post(t, "/api/project/delete", map[string]string{"id": created.ID}, nil)

	var listed []project
	status = client.Post(t, "/api/project/list", map[string]int{"limit": 200}, &listed)
	require.Equal(t, http.StatusOK, status)
	assert.Contains(t, listed, created)

	var fetched project
	status = client.Post(t, "/api/project/get", map[string]string{"id": created.ID}, &fetched)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, created, fetched)

	var updated project
	status = client.Post(t, "/api/project/update", map[string]string{"id": created.ID, "name": updatedName}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, updatedName, updated.Name)

	status = client.Post(t, "/api/project/delete", map[string]string{"id": created.ID}, nil)
	require.Equal(t, http.StatusNoContent, status)

	status = client.Post(t, "/api/project/get", map[string]string{"id": created.ID}, nil)
	assert.Equal(t, http.StatusNotFound, status)
}

func TestProjectDeleteArchivesChildTasksAndRequirements(t *testing.T) {
	client := shared.NewClient(t)
	db := shared.NewDB(t)
	ctx := context.Background()

	var created project
	status := client.Post(t, "/api/project/create", map[string]string{
		"name": fmt.Sprintf("api-test-project-cascade-%d", time.Now().UnixNano()),
	}, &created)
	require.Equal(t, http.StatusCreated, status)

	var createdTask task
	status = client.Post(t, "/api/task/create", map[string]any{
		"project_id": created.ID,
		"name":       fmt.Sprintf("api-test-project-delete-task-%d", time.Now().UnixNano()),
	}, &createdTask)
	require.Equal(t, http.StatusCreated, status)

	var requirementID string
	err := db.QueryRow(ctx, `
		insert into requirement (definition, task_id)
		values ($1, $2)
		returning id
	`, "Project delete archives this requirement.", createdTask.ID).Scan(&requirementID)
	require.NoError(t, err)

	status = client.Post(t, "/api/project/delete", map[string]string{"id": created.ID}, nil)
	require.Equal(t, http.StatusNoContent, status)

	shared.AssertHistoryDeleted(t, db, "task_history", createdTask.ID)
	shared.AssertHistoryDeleted(t, db, "requirement_history", requirementID)
}

type task struct {
	ID string `json:"id"`
}
