package change_test

import (
	"aipm/api-tests/shared"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
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
	Slug  string `json:"slug"`
	Color string `json:"color"`
}

type change struct {
	ID          int      `json:"id"`
	Version     int16    `json:"version"`
	Ref         *int32   `json:"ref"`
	Slug        *string  `json:"slug"`
	ProjectID   int      `json:"project_id"`
	EpicID      *int     `json:"epic_id"`
	EpicName    *string  `json:"epic_name"`
	ChangePhase string   `json:"change_phase"`
	ChangeTypes []string `json:"change_types"`
	Title       string   `json:"title"`
	Idea        string   `json:"idea"`
	Spec        *string  `json:"spec"`
	SpecHTML    *string  `json:"spec_html"`
	PRBody      *string  `json:"pr_body"`
	PRHtml      *string  `json:"pr_html"`
	PRUrl       *string  `json:"pr_url"`
	AgentEdit   bool     `json:"agent_edit"`
	Open        bool     `json:"open"`
	DoneTC      int16    `json:"done_tc"`
	TotalTC     int16    `json:"total_tc"`
	Completed   int16    `json:"completed"`
}

type detail struct {
	Change    change     `json:"change"`
	TestCases []testCase `json:"test_cases"`
}

type testCase struct {
	ID int `json:"id"`
}

type testCaseMutation struct {
	TestCase *testCase `json:"test_case"`
}

type renderedArtifacts struct {
	Artifacts []struct {
		ID       int    `json:"id"`
		SpecHTML string `json:"spec_html"`
		PRHtml   string `json:"pr_html"`
	} `json:"artifacts"`
}

type referenceResult struct {
	status int
	change change
	err    error
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
	require.NotEmpty(t, phases[0].Color)
	var types []changeOption
	status = client.Post(t, "/api/v1/options/change-types-list", map[string]any{}, &types)
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, types)

	title := fmt.Sprintf("api-test-change-%d", time.Now().UnixNano())
	idea := "Created by change API integration test."
	var created change
	status = client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id": projectID,
		"title":      title,
		"idea":       idea,
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotEmpty(t, created.ID)
	assert.Nil(t, created.Ref)
	assert.Nil(t, created.Slug)
	assert.Equal(t, title, created.Title)
	assert.Equal(t, idea, created.Idea)
	assert.Equal(t, "backlog", created.ChangePhase)
	assert.Nil(t, created.Spec)
	assert.Nil(t, created.PRBody)
	assert.Nil(t, created.PRUrl)
	assert.False(t, created.AgentEdit)
	assert.True(t, created.Open)
	assert.Empty(t, created.ChangeTypes)
	assert.Nil(t, created.EpicID)

	var listed []change
	status = client.Post(t, "/api/v1/change/list", map[string]any{"project_id": projectID}, &listed)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, listed, 1)
	assert.Equal(t, created.ID, listed[0].ID)
	assert.Equal(t, created.Ref, listed[0].Ref)
	assert.Equal(t, created.Slug, listed[0].Slug)
	assert.Equal(t, created.Title, listed[0].Title)
	assert.Nil(t, listed[0].EpicID)
	assert.Nil(t, listed[0].EpicName)

	var listedFields []map[string]any
	status = client.Post(t, "/api/v1/change/list", map[string]any{"project_id": projectID}, &listedFields)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, listedFields, 1)
	assert.NotContains(t, listedFields[0], "idea")
	assert.NotContains(t, listedFields[0], "spec")
	assert.NotContains(t, listedFields[0], "spec_html")
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
	assert.Equal(t, idea, fetched.Change.Idea)
	assert.Nil(t, fetched.Change.SpecHTML)

	var referenced change
	status = client.Post(t, "/api/v1/change/reference", map[string]any{"id": created.ID}, &referenced)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, referenced.Ref)
	require.NotNil(t, referenced.Slug)
	assert.Equal(t, idea, referenced.Idea)

	firstRef := *referenced.Ref
	firstSlug := *referenced.Slug
	var projectAfterFirstReference project
	status = client.Post(t, "/api/v1/project/get", map[string]any{"id": projectID}, &projectAfterFirstReference)
	require.Equal(t, http.StatusOK, status)

	status = client.Post(t, "/api/v1/change/reference", map[string]any{"id": created.ID}, &referenced)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, referenced.Ref)
	require.NotNil(t, referenced.Slug)
	assert.Equal(t, firstRef, *referenced.Ref)
	assert.Equal(t, firstSlug, *referenced.Slug)
	var projectAfterSecondReference project
	status = client.Post(t, "/api/v1/project/get", map[string]any{"id": projectID}, &projectAfterSecondReference)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, projectAfterFirstReference.LastRef, projectAfterSecondReference.LastRef)

	var rendered renderedArtifacts
	status = client.Post(t, "/api/v1/change/rendered-artifacts", map[string]any{"ids": []int{created.ID}}, &rendered)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, rendered.Artifacts, 1)
	assert.Empty(t, rendered.Artifacts[0].SpecHTML)

	var updated change
	punctuationTitle := "perf(json): pooled-buffer JSON deserialize"
	status = client.Post(t, "/api/v1/change/update-title", map[string]any{"id": created.ID, "title": punctuationTitle}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, punctuationTitle, updated.Title)
	require.NotNil(t, updated.Ref)
	require.NotNil(t, updated.Slug)
	assert.Equal(t, firstRef, *updated.Ref)
	assert.Equal(t, firstSlug, *updated.Slug)

	status = client.Post(t, "/api/v1/change/reference", map[string]any{"id": created.ID}, &referenced)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, referenced.Ref)
	require.NotNil(t, referenced.Slug)
	assert.Equal(t, firstRef, *referenced.Ref)
	assert.Equal(t, fmt.Sprintf("%03d-perf-json-pooled-buffer-json-deserialize", firstRef), *referenced.Slug)
	var projectAfterSlugRefresh project
	status = client.Post(t, "/api/v1/project/get", map[string]any{"id": projectID}, &projectAfterSlugRefresh)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, projectAfterSecondReference.LastRef, projectAfterSlugRefresh.LastRef)

	status = client.Post(t, "/api/v1/change/update-idea", map[string]any{
		"id":   created.ID,
		"idea": "Focused idea update.",
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, "Focused idea update.", updated.Idea)

	status = client.Post(t, "/api/v1/change/update-idea-agent-edit", map[string]any{
		"id":   created.ID,
		"idea": "Agent rewritten idea.",
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, "Agent rewritten idea.", updated.Idea)
	assert.True(t, updated.AgentEdit)

	status = client.Post(t, "/api/v1/change/update-spec", map[string]any{
		"id":   created.ID,
		"spec": "Focused spec update.",
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.Spec)
	assert.Equal(t, "Focused spec update.", *updated.Spec)
	require.NotNil(t, updated.SpecHTML)
	assert.Contains(t, *updated.SpecHTML, "<p>Focused spec update.</p>")

	status = client.Post(t, "/api/v1/change/update-spec", map[string]any{
		"id":   created.ID,
		"spec": nil,
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Nil(t, updated.Spec)

	status = client.Post(t, "/api/v1/change/update-pr-body", map[string]any{
		"id":      created.ID,
		"pr_body": "Focused pull request body update.",
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.PRBody)
	assert.Equal(t, "Focused pull request body update.", *updated.PRBody)

	status = client.Post(t, "/api/v1/change/update-pr-url", map[string]any{
		"id":     created.ID,
		"pr_url": "https://example.test/project-manager/pull/1",
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.PRUrl)
	assert.Equal(t, "https://example.test/project-manager/pull/1", *updated.PRUrl)

	status = client.Post(t, "/api/v1/change/update-pr-url", map[string]any{
		"id":     created.ID,
		"pr_url": "javascript:alert(1)",
	}, nil)
	require.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/change/get", map[string]any{"id": created.ID}, &fetched)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, fetched.Change.PRUrl)
	assert.Equal(t, "https://example.test/project-manager/pull/1", *fetched.Change.PRUrl)

	status = client.Post(t, "/api/v1/change/update-agent-edit", map[string]any{
		"id":         created.ID,
		"agent_edit": true,
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.True(t, updated.AgentEdit)

	status = client.Post(t, "/api/v1/change/update-change-types", map[string]any{
		"id":           created.ID,
		"change_types": []string{},
	}, &updated)
	require.Equal(t, http.StatusOK, status)
	assert.Empty(t, updated.ChangeTypes)

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

	status = client.Post(t, "/api/v1/change/update-epic", map[string]any{"id": created.ID, "epic_id": epicID}, &updated)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, updated.EpicID)
	assert.Equal(t, epicID, *updated.EpicID)

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

func TestChangeReferenceConcurrentRequestsPreserveSingleRef(t *testing.T) {
	client := shared.NewClient(t)

	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)

	var created change
	status := client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id": projectID,
		"title":      "Concurrent reference",
		"idea":       "# Concurrent reference\n\nAssign one reference.",
	}, &created)
	require.Equal(t, http.StatusCreated, status)

	var projectBefore project
	status = client.Post(t, "/api/v1/project/get", map[string]any{"id": projectID}, &projectBefore)
	require.Equal(t, http.StatusOK, status)

	const requestCount = 12
	start := make(chan struct{})
	results := make(chan referenceResult, requestCount)
	var wg sync.WaitGroup
	for range requestCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			status, referenced, err := postChangeReference(client.BaseURL(), created.ID)
			results <- referenceResult{status: status, change: referenced, err: err}
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	var firstRef int32
	var firstSlug string
	for result := range results {
		require.NoError(t, result.err)
		require.Equal(t, http.StatusOK, result.status)
		require.NotNil(t, result.change.Ref)
		require.NotNil(t, result.change.Slug)
		if firstRef == 0 {
			firstRef = *result.change.Ref
			firstSlug = *result.change.Slug
		}
		assert.Equal(t, firstRef, *result.change.Ref)
		assert.Equal(t, firstSlug, *result.change.Slug)
	}

	var projectAfter project
	status = client.Post(t, "/api/v1/project/get", map[string]any{"id": projectID}, &projectAfter)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, projectBefore.LastRef+1, projectAfter.LastRef)
}

func postChangeReference(baseURL string, changeID int) (int, change, error) {
	payload, err := json.Marshal(map[string]any{"id": changeID})
	if err != nil {
		return 0, change{}, err
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/change/reference", bytes.NewReader(payload))
	if err != nil {
		return 0, change{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return 0, change{}, err
	}
	defer res.Body.Close()

	var referenced change
	if res.StatusCode != http.StatusOK {
		return res.StatusCode, referenced, nil
	}
	if err := json.NewDecoder(res.Body).Decode(&referenced); err != nil {
		return res.StatusCode, change{}, err
	}
	return res.StatusCode, referenced, nil
}

func TestChangeListOrdersByModifiedDescending(t *testing.T) {
	client := shared.NewClient(t)

	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)

	var older change
	status := client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id": projectID,
		"title":      fmt.Sprintf("api-test-older-change-%d", time.Now().UnixNano()),
		"idea":       "Older idea",
	}, &older)
	require.Equal(t, http.StatusCreated, status)

	time.Sleep(10 * time.Millisecond)

	var newer change
	status = client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id": projectID,
		"title":      fmt.Sprintf("api-test-newer-change-%d", time.Now().UnixNano()),
		"idea":       "Newer idea",
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

func TestChangeGetReturnsTestCasesOrderedByID(t *testing.T) {
	client := shared.NewClient(t)

	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)

	var created change
	status := client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id": projectID,
		"title":      fmt.Sprintf("api-test-testcase-order-change-%d", time.Now().UnixNano()),
		"idea":       "Test case ordering idea",
	}, &created)
	require.Equal(t, http.StatusCreated, status)

	firstID := createTestCaseWithScenario(t, client, created.ID, "zzz first by id")
	secondID := createTestCaseWithScenario(t, client, created.ID, "aaa second by id")
	thirdID := createTestCaseWithScenario(t, client, created.ID, "mmm third by id")

	var fetched detail
	status = client.Post(t, "/api/v1/change/get", map[string]any{"id": created.ID}, &fetched)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, fetched.TestCases, 3)
	assert.Equal(t, []int{firstID, secondID, thirdID}, []int{
		fetched.TestCases[0].ID,
		fetched.TestCases[1].ID,
		fetched.TestCases[2].ID,
	})
}

func TestChangeBooleanUpdatesRequireExplicitFields(t *testing.T) {
	client := shared.NewClient(t)

	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)

	var created change
	status := client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id": projectID,
		"title":      fmt.Sprintf("api-test-boolean-change-%d", time.Now().UnixNano()),
		"idea":       "Boolean update idea",
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
	status = client.Post(t, "/api/v1/change/update-idea-agent-edit", map[string]any{"id": created.ID, "idea": " "}, nil)
	require.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/change/get", map[string]any{"id": created.ID}, &fetched)
	require.Equal(t, http.StatusOK, status)
	assert.True(t, fetched.Change.AgentEdit)
}

func TestChangeCreateRejectsInvalidInput(t *testing.T) {
	client := shared.NewClient(t)

	status := client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id": 999999999,
		"title":      "orphan change",
		"idea":       "Orphan idea",
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	projectID := createProject(t, client)
	defer shared.CleanupProject(t, client, projectID)

	status = client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id": projectID,
		"title":      "   ",
		"idea":       "Blank title idea",
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/change/create", map[string]any{
		"project_id": projectID,
		"title":      "blank idea change",
		"idea":       "   ",
	}, nil)
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestChangeRejectsInvalidInputAndMissingRows(t *testing.T) {
	client := shared.NewClient(t)

	status := client.Post(t, "/api/v1/change/list", map[string]any{}, nil)
	assert.Equal(t, http.StatusBadRequest, status)

	status = client.Post(t, "/api/v1/change/get", map[string]any{"id": 999999999}, nil)
	assert.Equal(t, http.StatusNotFound, status)

	status = client.Post(t, "/api/v1/change/rendered-artifacts", map[string]any{"ids": []int{0}}, nil)
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

	status = client.Post(t, "/api/v1/change/update-idea-agent-edit", map[string]any{"id": 999999999, "idea": "missing"}, nil)
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

	return createTestCaseWithScenario(t, client, changeID, "Change delete removes this test case.")
}

func createTestCaseWithScenario(t *testing.T, client *shared.Client, changeID int, scenario string) int {
	t.Helper()

	var created testCaseMutation
	status := client.Post(t, "/api/v1/test-case/create", map[string]any{
		"change_id": changeID,
		"scenario":  scenario,
	}, &created)
	require.Equal(t, http.StatusCreated, status)
	require.NotNil(t, created.TestCase)
	require.NotEmpty(t, created.TestCase.ID)
	return created.TestCase.ID
}
