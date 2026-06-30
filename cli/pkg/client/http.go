package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"mch/internal/dto"
)

// Client defines backend API methods used by mch.
type Client interface {
	ListProjects() ([]dto.Option, error)
	ListProjectRows() ([]dto.Project, error)
	GetProject(id int) (dto.Project, error)
	CreateProject(name string) (dto.Project, error)
	UpdateProject(id int, name string) (dto.Project, error)
	ListEpics(projectID string) ([]dto.Option, error)
	ListPhases() ([]dto.Option, error)
	ListTypes() ([]dto.Option, error)
}

// HTTPClient calls the Project Manager backend over HTTP.
type HTTPClient struct {
	BaseURL string
	Client  *http.Client
}

// NewHTTPClient creates an HTTP backend client for a base URL.
func NewHTTPClient(baseURL string) HTTPClient {
	return HTTPClient{
		BaseURL: baseURL,
		Client:  http.DefaultClient,
	}
}

// ListProjects loads projects as selector options.
func (c HTTPClient) ListProjects() ([]dto.Option, error) {
	projects, err := c.ListProjectRows()
	if err != nil {
		return nil, err
	}
	options := make([]dto.Option, 0, len(projects))
	for _, project := range projects {
		label := project.Name
		if label == "" {
			label = project.ID
		}
		if label == "" {
			continue
		}
		options = append(options, dto.Option{ID: project.ID, Label: label})
	}
	return options, nil
}

// ListProjectRows loads projects with full table row fields.
func (c HTTPClient) ListProjectRows() ([]dto.Project, error) {
	return c.postProjects("/api/v1/project/list", map[string]any{}, "projects")
}

// GetProject loads a single project by numeric ID.
func (c HTTPClient) GetProject(id int) (dto.Project, error) {
	if id <= 0 {
		return dto.Project{}, fmt.Errorf("project ID must be a valid positive number")
	}
	return c.postProject("/api/v1/project/get", map[string]any{"id": id})
}

// CreateProject creates a project with the required name field.
func (c HTTPClient) CreateProject(name string) (dto.Project, error) {
	return c.postProject("/api/v1/project/create", map[string]any{"name": name})
}

// UpdateProject updates a project name by numeric ID.
func (c HTTPClient) UpdateProject(id int, name string) (dto.Project, error) {
	if id <= 0 {
		return dto.Project{}, fmt.Errorf("project ID must be a valid positive number")
	}
	return c.postProject("/api/v1/project/update", map[string]any{
		"id":   id,
		"name": name,
	})
}

// ListEpics loads epic selector options for a project.
func (c HTTPClient) ListEpics(projectID string) ([]dto.Option, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, fmt.Errorf("current project is required")
	}
	numericProjectID, err := strconv.Atoi(projectID)
	if err != nil {
		return nil, fmt.Errorf("current project must be numeric")
	}
	return c.postOptions("/api/v1/epic/list", map[string]any{"project_id": numericProjectID}, "epics")
}

// ListPhases loads change phase selector options.
func (c HTTPClient) ListPhases() ([]dto.Option, error) {
	options, err := c.postOptions("/api/v1/change/reference", map[string]any{}, "phases")
	if len(options) == 0 && err == nil {
		options, err = c.postOptions("/api/v1/change/reference", map[string]any{}, "phase")
	}
	return options, err
}

// ListTypes loads change type selector options.
func (c HTTPClient) ListTypes() ([]dto.Option, error) {
	options, err := c.postOptions("/api/v1/change/reference", map[string]any{}, "types")
	if len(options) == 0 && err == nil {
		options, err = c.postOptions("/api/v1/change/reference", map[string]any{}, "type")
	}
	return options, err
}

func (c HTTPClient) postOptions(path string, payload any, group string) ([]dto.Option, error) {
	data, err := c.postJSON(path, payload)
	if err != nil {
		return nil, err
	}
	return findOptions(data, group), nil
}

func (c HTTPClient) postProjects(path string, payload any, group string) ([]dto.Project, error) {
	data, err := c.postJSON(path, payload)
	if err != nil {
		return nil, err
	}
	return findProjects(data, group), nil
}

func (c HTTPClient) postProject(path string, payload any) (dto.Project, error) {
	data, err := c.postJSON(path, payload)
	if err != nil {
		return dto.Project{}, err
	}
	project, ok := findProject(data)
	if !ok {
		return dto.Project{}, fmt.Errorf("project response missing project")
	}
	return project, nil
}

func (c HTTPClient) postJSON(path string, payload any) (any, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var data map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
			if message := firstString(data, "message", "error"); message != "" {
				return nil, fmt.Errorf("%s", message)
			}
		}
		return nil, fmt.Errorf("backend returned %s", resp.Status)
	}

	var data any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func findOptions(value any, group string) []dto.Option {
	switch typed := value.(type) {
	case []any:
		return optionsFromArray(typed)
	case map[string]any:
		for key, candidate := range typed {
			if key == group {
				if list, ok := candidate.([]any); ok {
					return optionsFromArray(list)
				}
			}
		}
		for _, candidate := range typed {
			if nested := findOptions(candidate, group); len(nested) > 0 {
				return nested
			}
		}
	}
	return nil
}

func findProjects(value any, group string) []dto.Project {
	switch typed := value.(type) {
	case []any:
		return projectsFromArray(typed)
	case map[string]any:
		for key, candidate := range typed {
			if key == group {
				if list, ok := candidate.([]any); ok {
					return projectsFromArray(list)
				}
			}
		}
		for _, candidate := range typed {
			if nested := findProjects(candidate, group); len(nested) > 0 {
				return nested
			}
		}
	}
	return nil
}

func findProject(value any) (dto.Project, bool) {
	switch typed := value.(type) {
	case []any:
		projects := projectsFromArray(typed)
		if len(projects) > 0 {
			return projects[0], true
		}
	case map[string]any:
		for _, key := range []string{"project", "data", "result"} {
			if candidate, ok := typed[key]; ok {
				if project, found := findProject(candidate); found {
					return project, true
				}
			}
		}
		project := projectFromMap(typed)
		if project.ID != "" || project.Name != "" {
			return project, true
		}
		for _, candidate := range typed {
			if project, found := findProject(candidate); found {
				return project, true
			}
		}
	}
	return dto.Project{}, false
}

func optionsFromArray(values []any) []dto.Option {
	options := make([]dto.Option, 0, len(values))
	for _, value := range values {
		switch typed := value.(type) {
		case string:
			options = append(options, dto.Option{ID: typed, Label: typed})
		case map[string]any:
			option := dto.Option{
				ID:    firstString(typed, "id", "project_id", "epic_id", "slug", "value"),
				Label: firstString(typed, "name", "title", "slug", "label", "value", "id"),
			}
			if option.Label == "" {
				continue
			}
			if option.ID == "" {
				option.ID = option.Label
			}
			options = append(options, option)
		}
	}
	return options
}

func projectsFromArray(values []any) []dto.Project {
	projects := make([]dto.Project, 0, len(values))
	for _, value := range values {
		typed, ok := value.(map[string]any)
		if !ok {
			continue
		}
		project := projectFromMap(typed)
		if project.ID == "" && project.Name == "" {
			continue
		}
		projects = append(projects, project)
	}
	return projects
}

func projectFromMap(values map[string]any) dto.Project {
	return dto.Project{
		ID:          firstString(values, "id", "project_id"),
		Name:        firstString(values, "name", "title"),
		LastRef:     firstInt(values, "last_ref"),
		ChangeCount: firstInt(values, "change_count"),
		Created:     firstString(values, "created", "created_at"),
		Modified:    firstString(values, "modified", "updated", "updated_at"),
	}
}

func firstString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := values[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case string:
			return typed
		case float64:
			return strconv.FormatInt(int64(typed), 10)
		case int:
			return strconv.Itoa(typed)
		}
	}
	return ""
}

func firstInt(values map[string]any, keys ...string) int {
	for _, key := range keys {
		value, ok := values[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return int(typed)
		case int:
			return typed
		case string:
			parsed, err := strconv.Atoi(strings.TrimSpace(typed))
			if err == nil {
				return parsed
			}
		}
	}
	return 0
}
