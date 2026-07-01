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
	ID      int   `json:"id"`
	LastRef int32 `json:"last_ref"`
}

type epic struct {
	ID        int   `json:"id"`
	DoneTC    int16 `json:"done_tc"`
	TotalTC   int16 `json:"total_tc"`
	Completed int16 `json:"completed"`
}

type changeOption struct {
	Slug string `json:"slug"`
}

type change struct {
	ID          int      `json:"id"`
	Version     int16    `json:"version"`
	Ref         int32    `json:"ref"`
	Slug        string   `json:"slug"`
	ProjectID   int      `json:"project_id"`
	EpicID      *int     `json:"epic_id"`
	EpicName    *string  `json:"epic_name"`
	ChangePhase string   `json:"change_phase"`
	ChangeTypes []string `json:"change_types"`
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	HTML        string   `json:"html"`
	PRBody      string   `json:"pr_body"`
	PRHtml      string   `json:"pr_html"`
	PRUrl       string   `json:"pr_url"`
	AgentEdit   bool     `json:"agent_edit"`
	Open        bool     `json:"open"`
	DoneTC      int16    `json:"done_tc"`
	TotalTC     int16    `json:"total_tc"`
	Completed   int16    `json:"completed"`
}

type detail struct {
	Change change `json:"change"`
}

type testCase struct {
	ID int `json:"id"`
}

type testCaseMutation struct {
	TestCase *testCase `json:"test_case"`
}

type renderedBodies struct {
	Bodies []struct {
		ID     int    `json:"id"`
		Html   string `json:"html"`
		PRHtml string `json:"pr_html"`
	} `json:"bodies"`
}

func TestChangeCRUDAndOptions(t *testing.T) {
	client := shared.NewClient(t)

	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)
	epicID := createEpic(t, client, projectID)

	var phases []changeOption
	status := client.Post(t, "/api/v1/options/change-phases-list", map[string]any{}, &phases)
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, phases)
	var types []changeOption
	status = client.Post(t, "/api/v1/options/change-types-list", map[string]any{}, &types)
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, types)

	title := fmt.Sprintf("api-test-change-%d", time.Now().UnixNano())
	var created change
	status = client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id":   projectID,
		"epic_id":      epicID,
		"title":        title,
		"body":         "Created by change API integration test.",
		"change_types": []string{"feature"},
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)
	require.NotZero(t, created.Ref)
	require.NotEmpty(t, created.Slug)
	assert.Equal(t, title, created.Title)
	assert.Equal(t, "backlog", created.ChangePhase)
	assert.Equal(t, "Created by change API integration test.", created.Body)
	assert.Empty(t, created.PRBody)
	assert.Empty(t, created.PRUrl)
	assert.False(t, created.AgentEdit)
	assert.True(t, created.Open)
	assert.Equal(t, []string{"feature"}, created.ChangeTypes)
	require.NotNil(t, created.EpicID)
	assert.Equal(t, epicID, *created.EpicID)

	var listed []change
	status = client.Post(t, "/api/v1/change/list", map[string]any{"project_id": projectID}, &listed)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, listed, 1)
	assert.Equal(t, created.ID, listed[0].ID)
	assert.Equal(t, created.Ref, listed[0].Ref)
	assert.Equal(t, created.Slug, listed[0].Slug)
	assert.Equal(t, created.Title, listed[0].Title)
	assert.Equal(t, created.EpicID, listed[0].EpicID)
	require.NotNil(t, listed[0].EpicName)

	var listedFields []map[string]any
	status = client.Post(t, "/api/v1/change/list", map[string]any{"project_id": projectID}, &listedFields)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, listedFields, 1)
	assert.NotContains(t, listedFields[0], "body")
	assert.NotContains(t, listedFields[0], "html")
	assert.NotContains(t, listedFields[0], "pr_body")
	assert.NotContains(t, listedFields[0], "pr_url")
	assert.NotContains(t, listedFields[0], "version")
	assert.NotContains(t, listedFields[0], "created")

	var fetched detail
	status = client.Post(t, "/api/v1/change/get", map[string]any{"id": created.ID}, &fetched)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, created.ID, fetched.Change.ID)
	assert.Equal(t, created.Ref, fetched.Change.Ref)
	assert.Equal(t, created.Slug, fetched.Change.Slug)
	assert.Contains(t, fetched.Change.HTML, "<p>Created by change API integration test.</p>")

	var rendered renderedBodies
	status = client.Post(t, "/api/v1/change/rendered-bodies", map[string]any{"ids": []int{created.ID}}, &rendered)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, rendered.Bodies, 1)
	assert.Contains(t, rendered.Bodies[0].Html, "<p>Created by change API integration test.</p>")

	var updated change
	status = client.Post(t, "/api/v1/change/update-title", map[string]any{"id": created.ID, "title": title + "-title"}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, title+"-title", updated.Title)
	assert.Equal(t, created.Ref, updated.Ref)
	assert.Equal(t, created.Slug, updated.Slug)

	status = client.Post(t, "/api/v1/change/update-body", map[string]any{
		"id":   created.ID,
		"body": "Focused body update.",
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, "Focused body update.", updated.Body)

	status = client.Post(t, "/api/v1/change/update-pr-body", map[string]any{
		"id":      created.ID,
		"pr_body": "Focused pull request body update.",
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, "Focused pull request body update.", updated.PRBody)

	status = client.Post(t, "/api/v1/change/update-pr-url", map[string]any{
		"id":     created.ID,
		"pr_url": "https://example.test/project-manager/pull/1",
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, "https://example.test/project-manager/pull/1", updated.PRUrl)

	status = client.Post(t, "/api/v1/change/update-pr-url", map[string]any{
		"id":     created.ID,
		"pr_url": "javascript:alert(1)",
	}, nil)
	require.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/change/get", map[string]any{"id": created.ID}, &fetched)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, "https://example.test/project-manager/pull/1", fetched.Change.PRUrl)

	status = client.Post(t, "/api/v1/change/update-agent-edit", map[string]any{
		"id":         created.ID,
		"agent_edit": true,
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.True(t, updated.AgentEdit)

	status = client.Post(t, "/api/v1/change/update-change-types", map[string]any{
		"id":           created.ID,
		"change_types": []string{"docs"},
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, []string{"docs"}, updated.ChangeTypes)

	status = client.Post(t, "/api/v1/change/update-phase", map[string]any{"id": created.ID, "change_phase": "review"}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, "review", updated.ChangePhase)

	status = client.Post(t, "/api/v1/change/update-open", map[string]any{"id": created.ID, "open": false}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.False(t, updated.Open)

	status = client.Post(t, "/api/v1/change/update-epic", map[string]any{"id": created.ID, "epic_id": nil}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Nil(t, updated.EpicID)

	testCaseID := createTestCase(t, client, created.ID)

	status = client.Post(t, "/api/v1/change/delete", map[string]any{"id": created.ID}, nil)
	require.Equal(t, http.StatusNoContent, status)

	status = client.Post(t, "/api/v1/change/get", map[string]any{"id": created.ID}, nil)
	assert.Equal(t, http.StatusNotFound, status)

	status = client.Post(t, "/api/v1/test-case/delete", map[string]any{"id": testCaseID}, nil)
	assert.Equal(t, http.StatusNotFound, status)
}

func TestChangeListOrdersByModifiedDescending(t *testing.T) {
	client := shared.NewClient(t)

	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)

	var older change
	status := client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id":   projectID,
		"title":        fmt.Sprintf("api-test-older-change-%d", time.Now().UnixNano()),
		"change_types": []string{"feature"},
	}, &older)
	require.Equal(t, http.StatusCreated, status)

	time.Sleep(10 * time.Millisecond)

	var newer change
	status = client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id":   projectID,
		"title":        fmt.Sprintf("api-test-newer-change-%d", time.Now().UnixNano()),
		"change_types": []string{"feature"},
	}, &newer)
	require.Equal(t, http.StatusCreated, status)

	var listed []change
	status = client.Post(t, "/api/v1/change/list", map[string]any{"project_id": projectID}, &listed)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, listed, 2)
	assert.Equal(t, newer.ID, listed[0].ID)
	assert.Equal(t, older.ID, listed[1].ID)

	time.Sleep(10 * time.Millisecond)

	var updated change
	status = client.Post(t, "/api/v1/change/update-title", map[string]any{
		"id":    older.ID,
		"title": older.Title + "-updated",
	}, &updated)
	require.Equal(t, http.StatusOK, status)

	status = client.Post(t, "/api/v1/change/list", map[string]any{"project_id": projectID}, &listed)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, listed, 2)
	assert.Equal(t, older.ID, listed[0].ID)
	assert.Equal(t, newer.ID, listed[1].ID)
}

func TestChangeBooleanUpdatesRequireExplicitFields(t *testing.T) {
	client := shared.NewClient(t)

	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)

	var created change
	status := client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id":   projectID,
		"title":        fmt.Sprintf("api-test-boolean-change-%d", time.Now().UnixNano()),
		"change_types": []string{"feature"},
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	assert.False(t, created.AgentEdit)
	assert.True(t, created.Open)

	status = client.Post(t, "/api/v1/change/update-open", map[string]any{"id": created.ID}, nil)
	require.Equal(t, http.StatusBadRequest, status)
	status = client.Post(t, "/api/v1/change/update-open", map[string]any{"id": created.ID, "closed": true}, nil)
	require.Equal(t, http.StatusBadRequest, status)

	var fetched detail
	status = client.Post(t, "/api/v1/change/get", map[string]any{"id": created.ID}, &fetched)
	require.Equal(t, http.StatusOK, status)
	assert.True(t, fetched.Change.Open)

	var updated change
	status = client.Post(t, "/api/v1/change/update-agent-edit", map[string]any{
		"id":         created.ID,
		"agent_edit": true,
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.True(t, updated.AgentEdit)

	status = client.Post(t, "/api/v1/change/update-agent-edit", map[string]any{"id": created.ID}, nil)
	require.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/change/get", map[string]any{"id": created.ID}, &fetched)
	require.Equal(t, http.StatusOK, status)
	assert.True(t, fetched.Change.AgentEdit)
}

func TestChangeCreateRejectsInvalidReferences(t *testing.T) {
	client := shared.NewClient(t)

	status := client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id":   999999999,
		"title":        "orphan change",
		"change_types": []string{"feature"},
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)

	status = client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id":   projectID,
		"epic_id":      999999999,
		"title":        "missing epic change",
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

	status = client.Post(t, "/api/v1/change/update-title", map[string]any{
		"id":    999999999,
		"title": "missing change",
	}, nil)
	assert.Equal(t, http.StatusNotFound, status)

	status = client.Post(t, "/api/v1/change/update-change-types", map[string]any{
		"id":           999999999,
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

	status = client.Post(t, "/api/v1/change/update-open", map[string]any{"id": 999999999, "open": false}, nil)
	assert.Equal(t, http.StatusNotFound, status)

	status = client.Post(t, "/api/v1/change/update-pr-url", map[string]any{"id": 999999999, "pr_url": "https://example.test"}, nil)
	assert.Equal(t, http.StatusNotFound, status)

	status = client.Post(t, "/api/v1/change/update-agent-edit", map[string]any{"id": 999999999, "agent_edit": true}, nil)
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

func createTestCase(t *testing.T, client *shared.Client, changeID int) int {
	t.Helper()

	var created testCaseMutation
	status := client.Post(t, "/api/v1/test-case/create", map[string]any{
		"change_id": changeID,
		"scenario":  "Change delete removes this test case.",
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotNil(t, created.TestCase)
	require.NotEmpty(t, created.TestCase.ID)
	return created.TestCase.ID
}
