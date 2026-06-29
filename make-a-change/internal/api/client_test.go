package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
			writeJSON(t, w, map[string]any{"projects": []map[string]any{{"id": 7, "name": "Project Seven"}}})
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
	assert.Equal(t, Option{ID: "7", Label: "Project Seven"}, projects[0])

	phases, err := client.ListPhases()
	require.NoError(t, err)
	require.Len(t, phases, 1)
	assert.Equal(t, Option{ID: "backlog", Label: "backlog"}, phases[0])

	types, err := client.ListTypes()
	require.NoError(t, err)
	require.Len(t, types, 1)
	assert.Equal(t, Option{ID: "feature", Label: "feature"}, types[0])

	epics, err := client.ListEpics("7")
	require.NoError(t, err)
	require.Len(t, epics, 1)
	assert.Equal(t, Option{ID: "3", Label: "Epic Three"}, epics[0])

	wantPaths := []string{
		"/api/v1/project/list",
		"/api/v1/change/reference",
		"/api/v1/change/reference",
		"/api/v1/epic/list",
	}
	assert.Equal(t, wantPaths, paths)
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
