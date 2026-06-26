package project_test

import (
	"aipm/api-tests/shared"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
	db := shared.NewDB(t)
	ctx := context.Background()

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

	var requirementID int
	err := db.QueryRow(ctx, `
		insert into requirement (definition, change_id)
		values ($1, $2)
		returning id
	`, "Project delete keeps this requirement.", createdChange.ID).Scan(&requirementID)
	require.NoError(t, err)

	status = client.Post(t, "/api/v1/project/delete", map[string]any{"id": created.ID}, nil)
	require.Equal(t, http.StatusConflict, status)

	assertRowExists(t, db, "project", created.ID)
	assertRowExists(t, db, "change", createdChange.ID)
	assertRowExists(t, db, "requirement", requirementID)
	shared.AssertHistoryCount(t, db, "change_history", createdChange.ID, true, 0)
	shared.AssertHistoryCount(t, db, "requirement_history", requirementID, true, 0)
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

func assertRowExists(t *testing.T, db *pgxpool.Pool, table string, id int) {
	t.Helper()

	var exists bool
	err := db.QueryRow(context.Background(), "select exists(select 1 from "+table+" where id = $1)", id).Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists)
}
