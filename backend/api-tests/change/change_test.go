package change_test

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
	ID int `json:"id"`
}

type epic struct {
	ID        int   `json:"id"`
	DoneReq   int16 `json:"done_req"`
	TotalReq  int16 `json:"total_req"`
	Completed int16 `json:"completed"`
}

type references struct {
	Phases []referenceOption `json:"phases"`
	Types  []referenceOption `json:"types"`
}

type referenceOption struct {
	Slug string `json:"slug"`
}

type change struct {
	ID          int      `json:"id"`
	Version     int16    `json:"version"`
	ProjectID   int      `json:"project_id"`
	EpicID      *int     `json:"epic_id"`
	ChangePhase string   `json:"change_phase"`
	ChangeTypes []string `json:"change_types"`
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	BodyHTML    string   `json:"body_html"`
	Closed      bool     `json:"closed"`
	DoneReq     int16    `json:"done_req"`
	TotalReq    int16    `json:"total_req"`
	Completed   int16    `json:"completed"`
}

type detail struct {
	Change change `json:"change"`
}

type requirement struct {
	ID int `json:"id"`
}

type requirementMutation struct {
	Requirement *requirement `json:"requirement"`
}

type renderedBodies struct {
	Bodies []struct {
		ID       int    `json:"id"`
		BodyHTML string `json:"body_html"`
	} `json:"bodies"`
}

func TestChangeCRUDAndReferences(t *testing.T) {
	client := shared.NewClient(t)

	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)
	epicID := createEpic(t, client, projectID)

	var refs references
	status := client.Post(t, "/api/v1/change/reference", map[string]any{}, &refs)
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, refs.Phases)
	require.NotEmpty(t, refs.Types)

	title := fmt.Sprintf("api-test-change-%d", time.Now().UnixNano())
	var created change
	status = client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id":   projectID,
		"epic_id":      epicID,
		"title":        title,
		"body":         "Created by change API integration test.",
		"change_phase": "backlog",
		"change_types": []string{"feature"},
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, title, created.Title)
	assert.Equal(t, []string{"feature"}, created.ChangeTypes)
	require.NotNil(t, created.EpicID)
	assert.Equal(t, epicID, *created.EpicID)

	var listed []change
	status = client.Post(t, "/api/v1/change/list", map[string]any{"project_id": projectID}, &listed)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, listed, 1)
	assert.Equal(t, created.ID, listed[0].ID)
	assert.Equal(t, created.Title, listed[0].Title)
	assert.Equal(t, created.EpicID, listed[0].EpicID)

	var fetched detail
	status = client.Post(t, "/api/v1/change/get", map[string]any{"id": created.ID}, &fetched)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, created.ID, fetched.Change.ID)
	assert.Contains(t, fetched.Change.BodyHTML, "<p>Created by change API integration test.</p>")

	var rendered renderedBodies
	status = client.Post(t, "/api/v1/change/rendered-bodies", map[string]any{"ids": []int{created.ID}}, &rendered)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, rendered.Bodies, 1)
	assert.Contains(t, rendered.Bodies[0].BodyHTML, "<p>Created by change API integration test.</p>")

	var updated change
	status = client.Post(t, "/api/v1/change/update", map[string]any{
		"id":           created.ID,
		"title":        title + "-updated",
		"body":         "Updated by change API integration test.",
		"change_types": []string{"fix"},
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, title+"-updated", updated.Title)
	assert.Equal(t, []string{"fix"}, updated.ChangeTypes)
	assert.Equal(t, created.Version+1, updated.Version)

	status = client.Post(t, "/api/v1/change/update-phase", map[string]any{"id": created.ID, "change_phase": "review"}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, "review", updated.ChangePhase)

	status = client.Post(t, "/api/v1/change/update-closed", map[string]any{"id": created.ID, "closed": true}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.True(t, updated.Closed)

	status = client.Post(t, "/api/v1/change/update-epic", map[string]any{"id": created.ID, "epic_id": nil}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Nil(t, updated.EpicID)

	requirementID := createRequirement(t, client, created.ID)

	status = client.Post(t, "/api/v1/change/delete", map[string]any{"id": created.ID}, nil)
	require.Equal(t, http.StatusNoContent, status)

	status = client.Post(t, "/api/v1/change/get", map[string]any{"id": created.ID}, nil)
	assert.Equal(t, http.StatusNotFound, status)

	status = client.Post(t, "/api/v1/requirement/delete", map[string]any{"id": requirementID}, nil)
	assert.Equal(t, http.StatusNotFound, status)
}

func TestChangeCreateRejectsInvalidReferences(t *testing.T) {
	client := shared.NewClient(t)

	status := client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id":   999999999,
		"title":        "orphan change",
		"change_phase": "backlog",
		"change_types": []string{"feature"},
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)

	status = client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id":   projectID,
		"epic_id":      999999999,
		"title":        "missing epic change",
		"change_phase": "backlog",
		"change_types": []string{"feature"},
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestChangeRejectsInvalidInputAndMissingRows(t *testing.T) {
	client := shared.NewClient(t)

	status := client.Post(t, "/api/v1/change/list", map[string]any{}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/change/get", map[string]any{"id": 999999999}, nil)
	assert.Equal(t, http.StatusNotFound, status)

	status = client.Post(t, "/api/v1/change/rendered-bodies", map[string]any{"ids": []int{0}}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/change/update", map[string]any{
		"id":           999999999,
		"title":        "missing change",
		"change_types": []string{"feature"},
	}, nil)
	assert.Equal(t, http.StatusNotFound, status)

	status = client.Post(t, "/api/v1/change/update", map[string]any{
		"id":           999999999,
		"title":        "missing change",
		"change_types": []string{"missing-type"},
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/change/update-epic", map[string]any{"id": 999999999, "epic_id": nil}, nil)
	assert.Equal(t, http.StatusNotFound, status)

	status = client.Post(t, "/api/v1/change/update-phase", map[string]any{
		"id":           999999999,
		"change_phase": "missing-phase",
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/change/update-closed", map[string]any{"id": 999999999, "closed": true}, nil)
	assert.Equal(t, http.StatusNotFound, status)

	status = client.Post(t, "/api/v1/change/delete", map[string]any{}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/change/delete", map[string]any{"id": 999999999}, nil)
	assert.Equal(t, http.StatusNotFound, status)
}

func createProject(t *testing.T, client *shared.Client) int {
	t.Helper()
	var created project
	status := client.Post(t, "/api/v1/project/create", map[string]string{
		"name": fmt.Sprintf("api-test-change-project-%d", time.Now().UnixNano()),
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)
	return created.ID
}

func createEpic(t *testing.T, client *shared.Client, projectID int) int {
	t.Helper()
	var created epic
	status := client.Post(t, "/api/v1/epic/create", map[string]any{
		"project_id": projectID,
		"name":       fmt.Sprintf("api-test-epic-%d", time.Now().UnixNano()),
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)
	return created.ID
}

func createRequirement(t *testing.T, client *shared.Client, changeID int) int {
	t.Helper()

	var created requirementMutation
	status := client.Post(t, "/api/v1/requirement/create", map[string]any{
		"change_id":  changeID,
		"definition": "Change delete removes this requirement.",
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotNil(t, created.Requirement)
	require.NotEmpty(t, created.Requirement.ID)
	return created.Requirement.ID
}
