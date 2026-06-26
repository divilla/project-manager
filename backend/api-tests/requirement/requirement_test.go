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
	ID int `json:"id"`
}

type epic struct {
	ID        int   `json:"id"`
	DoneReq   int16 `json:"done_req"`
	TotalReq  int16 `json:"total_req"`
	Completed int16 `json:"completed"`
}

type change struct {
	ID        int   `json:"id"`
	Version   int16 `json:"version"`
	DoneReq   int16 `json:"done_req"`
	TotalReq  int16 `json:"total_req"`
	Completed int16 `json:"completed"`
}

type requirement struct {
	ID         int    `json:"id"`
	Version    int16  `json:"version"`
	ChangeID   int    `json:"change_id"`
	Definition string `json:"definition"`
	Done       bool   `json:"done"`
}

type mutation struct {
	Requirement  *requirement  `json:"requirement"`
	Change       change        `json:"change"`
	Requirements []requirement `json:"requirements"`
}

func TestRequirementCRUDRecalculatesChangeAndEpicCompleteness(t *testing.T) {
	client := shared.NewClient(t)
	db := shared.NewDB(t)

	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)
	epicID := createEpic(t, client, projectID)
	changeID := createChange(t, client, projectID, &epicID)

	var listed []requirement
	status := client.Post(t, "/api/v1/requirement/list", map[string]any{"change_id": changeID}, &listed)
	require.Equal(t, http.StatusOK, status)
	assert.Empty(t, listed)

	first := createRequirement(t, client, changeID, "Add requirement create endpoint test.")
	second := createRequirement(t, client, changeID, "Add requirement update endpoint test.")
	assert.Equal(t, changeID, first.ChangeID)
	assert.False(t, first.Done)
	assert.False(t, second.Done)

	status = client.Post(t, "/api/v1/requirement/list", map[string]any{"change_id": changeID}, &listed)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, listed, 2)

	var updated mutation
	status = client.Post(t, "/api/v1/requirement/update-done", map[string]any{"id": first.ID, "done": true}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.Requirement)
	assert.True(t, updated.Requirement.Done)
	assert.Equal(t, int16(50), updated.Change.Completed)
	require.Len(t, updated.Requirements, 2)
	shared.AssertHistoryCount(t, db, "requirement_history", first.ID, false, 0)
	assert.Equal(t, first.Version, updated.Requirement.Version)
	first = *updated.Requirement
	assertEpicCompleteness(t, client, epicID, 1, 2, 50)

	status = client.Post(t, "/api/v1/requirement/update-done", map[string]any{"id": second.ID, "done": true}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, int16(100), updated.Change.Completed)
	second = *updated.Requirement

	status = client.Post(t, "/api/v1/requirement/update", map[string]any{
		"id":         second.ID,
		"definition": "Updated requirement definition.",
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.Requirement)
	assert.Equal(t, "Updated requirement definition.", updated.Requirement.Definition)
	second = *updated.Requirement

	status = client.Post(t, "/api/v1/requirement/update-done", map[string]any{"id": second.ID, "done": false}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.Requirement)
	assert.False(t, updated.Requirement.Done)
	assert.Equal(t, int16(50), updated.Change.Completed)
	shared.AssertHistoryCountAtLeast(t, db, "requirement_history", second.ID, false, 1)

	newChangeID := createChange(t, client, projectID, &epicID)
	previousVersion := updated.Requirement.Version
	status = client.Post(t, "/api/v1/requirement/update-change", map[string]any{
		"id": second.ID, "change_id": newChangeID,
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.Requirement)
	assert.Equal(t, newChangeID, updated.Requirement.ChangeID)
	assert.Equal(t, previousVersion, updated.Requirement.Version)
	assert.Equal(t, newChangeID, updated.Change.ID)

	var deleted mutation
	status = client.Post(t, "/api/v1/requirement/delete", map[string]any{"id": first.ID}, &deleted)
	require.Equal(t, http.StatusOK, status)
	assert.Nil(t, deleted.Requirement)
	assert.Equal(t, int16(0), deleted.Change.Completed)
	assert.Empty(t, deleted.Requirements)
	shared.AssertHistoryDeleted(t, db, "requirement_history", first.ID)
}

func TestRequirementRejectsInvalidInputAndMissingRows(t *testing.T) {
	client := shared.NewClient(t)

	status := client.Post(t, "/api/v1/requirement/create", map[string]any{
		"change_id":  -1,
		"definition": "orphan requirement",
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/requirement/create", map[string]any{
		"change_id":  -1,
		"definition": "   ",
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/requirement/update", map[string]any{
		"id":         999999999,
		"definition": "missing requirement",
	}, nil)
	assert.Equal(t, http.StatusNotFound, status)
}

func TestRequirementDeleteLastItemReturnsZeroCompleted(t *testing.T) {
	client := shared.NewClient(t)
	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)
	changeID := createChange(t, client, projectID, nil)
	requirement := createRequirement(t, client, changeID, "Delete the final requirement transactionally.")

	var deleted mutation
	status := client.Post(t, "/api/v1/requirement/delete", map[string]any{"id": requirement.ID}, &deleted)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, int16(0), deleted.Change.Completed)
	assert.Equal(t, int16(0), deleted.Change.DoneReq)
	assert.Equal(t, int16(0), deleted.Change.TotalReq)
	assert.Empty(t, deleted.Requirements)
}

func TestRequirementDoneUpdateDoesNotIncrementVersion(t *testing.T) {
	client := shared.NewClient(t)
	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)
	changeID := createChange(t, client, projectID, nil)
	requirement := createRequirement(t, client, changeID, "Done changes preserve the version.")

	var updated mutation
	status := client.Post(t, "/api/v1/requirement/update-done", map[string]any{"id": requirement.ID, "done": true}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.Requirement)
	assert.Equal(t, requirement.Version, updated.Requirement.Version)
	assert.True(t, updated.Requirement.Done)
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

func createEpic(t *testing.T, client *shared.Client, projectID int) int {
	t.Helper()
	var created epic
	status := client.Post(t, "/api/v1/epic/create", map[string]any{
		"project_id": projectID,
		"name":       fmt.Sprintf("api-test-requirement-epic-%d", time.Now().UnixNano()),
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)
	return created.ID
}

func createChange(t *testing.T, client *shared.Client, projectID int, epicID *int) int {
	t.Helper()
	var created change
	status := client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id":   projectID,
		"epic_id":      epicID,
		"title":        fmt.Sprintf("api-test-requirement-change-%d", time.Now().UnixNano()),
		"change_phase": "backlog",
		"change_types": []string{"feature"},
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)
	assert.Equal(t, int16(0), created.Completed)
	return created.ID
}

func createRequirement(t *testing.T, client *shared.Client, changeID int, definition string) requirement {
	t.Helper()
	var created mutation
	status := client.Post(t, "/api/v1/requirement/create", map[string]any{
		"change_id":  changeID,
		"definition": definition,
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotNil(t, created.Requirement)
	require.NotEmpty(t, created.Requirement.ID)
	assert.Equal(t, int16(0), created.Change.Completed)
	return *created.Requirement
}

func assertEpicCompleteness(t *testing.T, client *shared.Client, id int, doneReq, totalReq, completed int16) {
	t.Helper()
	var fetched epic
	status := client.Post(t, "/api/v1/epic/get", map[string]any{"id": id}, &fetched)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, doneReq, fetched.DoneReq)
	assert.Equal(t, totalReq, fetched.TotalReq)
	assert.Equal(t, completed, fetched.Completed)
}
