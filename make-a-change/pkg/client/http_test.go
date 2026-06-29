package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"mch/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPClientPostsToSelectorEndpoints(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/api/v1/project/list":
			writeJSON(t, w, map[string]any{"projects": []map[string]any{{
				"id":           7,
				"name":         "Project Seven",
				"change_count": 3,
				"created":      "2026-06-29T08:15:00Z",
				"modified":     "2026-06-29T10:45:00Z",
			}}})
		case "/api/v1/change/reference":
			writeJSON(t, w, map[string]any{
				"phases": []map[string]any{{"slug": "backlog"}},
				"types":  []map[string]any{{"slug": "feature"}},
			})
		case "/api/v1/epic/list":
			var payload map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			assert.Equal(t, float64(7), payload["project_id"])
			writeJSON(t, w, map[string]any{"epics": []map[string]any{{"id": 3, "title": "Epic Three"}}})
		default:
			require.Failf(t, "unexpected path", "path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)

	projects, err := client.ListProjects()
	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, dto.Option{ID: "7", Label: "Project Seven"}, projects[0])

	phases, err := client.ListPhases()
	require.NoError(t, err)
	require.Len(t, phases, 1)
	assert.Equal(t, dto.Option{ID: "backlog", Label: "backlog"}, phases[0])

	types, err := client.ListTypes()
	require.NoError(t, err)
	require.Len(t, types, 1)
	assert.Equal(t, dto.Option{ID: "feature", Label: "feature"}, types[0])

	epics, err := client.ListEpics("7")
	require.NoError(t, err)
	require.Len(t, epics, 1)
	assert.Equal(t, dto.Option{ID: "3", Label: "Epic Three"}, epics[0])

	wantPaths := []string{
		"/api/v1/project/list",
		"/api/v1/change/reference",
		"/api/v1/change/reference",
		"/api/v1/epic/list",
	}
	assert.Equal(t, wantPaths, paths)
}

func TestHTTPClientListsProjectRows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/v1/project/list", r.URL.Path)
		writeJSON(t, w, map[string]any{"projects": []map[string]any{{
			"id":           7,
			"name":         "Project Seven",
			"change_count": "3",
			"created_at":   "2026-06-29T08:15:00Z",
			"updated_at":   "2026-06-29T10:45:00Z",
		}}})
	}))
	defer server.Close()

	projects, err := NewHTTPClient(server.URL).ListProjectRows()

	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, dto.Project{
		ID:          "7",
		Name:        "Project Seven",
		ChangeCount: 3,
		Created:     "2026-06-29T08:15:00Z",
		Modified:    "2026-06-29T10:45:00Z",
	}, projects[0])
}

func TestHTTPClientProjectCreateUpdateAndGetPayloads(t *testing.T) {
	var paths []string
	var payloads []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		paths = append(paths, r.URL.Path)

		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		payloads = append(payloads, payload)

		switch r.URL.Path {
		case "/api/v1/project/create":
			writeJSON(t, w, map[string]any{"project": map[string]any{"id": 7}})
		case "/api/v1/project/update":
			writeJSON(t, w, map[string]any{"project": map[string]any{"id": payload["id"], "name": payload["name"]}})
		case "/api/v1/project/get":
			writeJSON(t, w, map[string]any{"project": map[string]any{
				"id":           payload["id"],
				"name":         fmt.Sprintf("Project %.0f", payload["id"]),
				"change_count": 2,
				"created":      "2026-06-29T08:15:00Z",
				"modified":     "2026-06-29T10:45:00Z",
			}})
		default:
			require.Failf(t, "unexpected path", "path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)

	created, err := client.CreateProject("New\nProject")
	require.NoError(t, err)
	assert.Equal(t, dto.Project{ID: "7"}, created)

	updated, err := client.UpdateProject(7, "Renamed\nProject")
	require.NoError(t, err)
	assert.Equal(t, dto.Project{ID: "7", Name: "Renamed\nProject"}, updated)

	got, err := client.GetProject(7)
	require.NoError(t, err)
	assert.Equal(t, dto.Project{
		ID:          "7",
		Name:        "Project 7",
		ChangeCount: 2,
		Created:     "2026-06-29T08:15:00Z",
		Modified:    "2026-06-29T10:45:00Z",
	}, got)

	assert.Equal(t, []string{
		"/api/v1/project/create",
		"/api/v1/project/update",
		"/api/v1/project/get",
	}, paths)
	assert.Equal(t, map[string]any{"name": "New\nProject"}, payloads[0])
	assert.Equal(t, map[string]any{"id": float64(7), "name": "Renamed\nProject"}, payloads[1])
	assert.Equal(t, map[string]any{"id": float64(7)}, payloads[2])
}

func TestHTTPClientProjectMutationValidationAndBackendErrors(t *testing.T) {
	client := NewHTTPClient("http://example.invalid")

	_, err := client.GetProject(0)
	require.Error(t, err)

	_, err = client.UpdateProject(0, "Name")
	require.Error(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, map[string]any{"message": "invalid project payload"})
	}))
	defer server.Close()

	_, err = NewHTTPClient(server.URL).CreateProject("Name")
	require.Error(t, err)
	assert.Equal(t, "invalid project payload", err.Error())
}

func TestListEpicsRequiresCurrentProject(t *testing.T) {
	client := NewHTTPClient("http://example.invalid")

	_, err := client.ListEpics("")
	require.Error(t, err)

	_, err = client.ListEpics("not-a-number")
	require.Error(t, err)
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(value))
}
