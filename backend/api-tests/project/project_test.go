package project_test

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
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
	ChangeCount int       `json:"change_count"`
}

func TestProjectCRUD(t *testing.T) {
	client := shared.NewClient(t)
	name := fmt.Sprintf("api-test-project-%d", time.Now().UnixNano())
	updatedName := name + "-updated"

	var created project
	status := client.Post(t, "/api/v1/project/create", map[string]string{"name": name}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, name, created.Name)
	assert.False(t, created.Created.IsZero())
	assert.False(t, created.Modified.IsZero())
	assert.Equal(t, 0, created.ChangeCount)

	defer shared.CleanupProject(t, client, created.ID)

	var listed []project
	status = client.Post(t, "/api/v1/project/list", map[string]int{"limit": 200}, &listed)
	require.Equal(t, http.StatusOK, status)
	assert.Contains(t, listed, created)

	var fetched project
	status = client.Post(t, "/api/v1/project/get", map[string]any{"id": created.ID}, &fetched)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, created, fetched)

	var updated project
	status = client.Post(t, "/api/v1/project/update", map[string]any{"id": created.ID, "name": updatedName}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, updatedName, updated.Name)
	assert.False(t, updated.Modified.Before(updated.Created))

	status = client.Post(t, "/api/v1/project/delete", map[string]any{"id": created.ID}, nil)
	require.Equal(t, http.StatusNoContent, status)

	status = client.Post(t, "/api/v1/project/get", map[string]any{"id": created.ID}, nil)
	assert.Equal(t, http.StatusNotFound, status)
}

func TestProjectDeleteRejectsProjectsWithChanges(t *testing.T) {
	client := shared.NewClient(t)

	var created project
	status := client.Post(t, "/api/v1/project/create", map[string]string{
		"name": fmt.Sprintf("api-test-project-cascade-%d", time.Now().UnixNano()),
	}, &created)
	require.Equal(t, http.StatusCreated, status)

	var createdChange change
	status = client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id":   created.ID,
		"title":        fmt.Sprintf("api-test-project-delete-change-%d", time.Now().UnixNano()),
		"change_phase": "backlog",
		"change_types": []string{"feature"},
	}, &createdChange)
	require.Equal(t, http.StatusCreated, status)

	t.Cleanup(func() {
		shared.CleanupProject(t, client, created.ID)
	})

	createdRequirement := createRequirement(t, client, createdChange.ID)

	status = client.Post(t, "/api/v1/project/delete", map[string]any{"id": created.ID}, nil)
	require.Equal(t, http.StatusConflict, status)

	var fetched project
	status = client.Post(t, "/api/v1/project/get", map[string]any{"id": created.ID}, &fetched)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, created.ID, fetched.ID)

	var fetchedChange changeDetail
	status = client.Post(t, "/api/v1/change/get", map[string]any{"id": createdChange.ID}, &fetchedChange)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, createdChange.ID, fetchedChange.Change.ID)

	var requirements []requirement
	status = client.Post(t, "/api/v1/requirement/list", map[string]any{"change_id": createdChange.ID}, &requirements)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, requirements, 1)
	assert.Equal(t, createdRequirement.ID, requirements[0].ID)
}

func TestProjectRejectsInvalidInputAndMissingRows(t *testing.T) {
	client := shared.NewClient(t)

	status := client.Post(t, "/api/v1/project/create", map[string]any{"name": "   "}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/project/get", map[string]any{"id": 999999999}, nil)
	assert.Equal(t, http.StatusNotFound, status)

	status = client.Post(t, "/api/v1/project/update", map[string]any{
		"id":   999999999,
		"name": "missing project",
	}, nil)
	assert.Equal(t, http.StatusNotFound, status)

	status = client.Post(t, "/api/v1/project/update", map[string]any{
		"id":   999999999,
		"name": "   ",
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/project/delete", map[string]any{"id": 999999999}, nil)
	assert.Equal(t, http.StatusNotFound, status)

	status = client.Post(t, "/api/v1/project/delete", map[string]any{}, nil)
	assert.Equal(t, http.StatusBadRequest, status)
}

type change struct {
	ID int `json:"id"`
}

type changeDetail struct {
	Change change `json:"change"`
}

type requirement struct {
	ID         int    `json:"id"`
	ChangeID   int    `json:"change_id"`
	Definition string `json:"definition"`
}

type requirementMutation struct {
	Requirement *requirement `json:"requirement"`
}

func createRequirement(t *testing.T, client *shared.Client, changeID int) requirement {
	t.Helper()

	var created requirementMutation
	status := client.Post(t, "/api/v1/requirement/create", map[string]any{
		"change_id":  changeID,
		"definition": "Project delete keeps this requirement.",
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotNil(t, created.Requirement)
	require.NotEmpty(t, created.Requirement.ID)
	assert.Equal(t, changeID, created.Requirement.ChangeID)
	return *created.Requirement
}
