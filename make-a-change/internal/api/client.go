package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Option struct {
	ID    string
	Label string
}

type Client interface {
	ListProjects() ([]Option, error)
	ListEpics(projectID string) ([]Option, error)
	ListPhases() ([]Option, error)
	ListTypes() ([]Option, error)
}

type HTTPClient struct {
	BaseURL string
	Client  *http.Client
}

func NewHTTPClient(baseURL string) HTTPClient {
	return HTTPClient{
		BaseURL: baseURL,
		Client:  http.DefaultClient,
	}
}

func (c HTTPClient) ListProjects() ([]Option, error) {
	return c.postOptions("/api/v1/project/list", map[string]any{}, "projects")
}

func (c HTTPClient) ListEpics(projectID string) ([]Option, error) {
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

func (c HTTPClient) ListPhases() ([]Option, error) {
	options, err := c.postOptions("/api/v1/change/reference", map[string]any{}, "phases")
	if len(options) == 0 && err == nil {
		options, err = c.postOptions("/api/v1/change/reference", map[string]any{}, "phase")
	}
	return options, err
}

func (c HTTPClient) ListTypes() ([]Option, error) {
	options, err := c.postOptions("/api/v1/change/reference", map[string]any{}, "types")
	if len(options) == 0 && err == nil {
		options, err = c.postOptions("/api/v1/change/reference", map[string]any{}, "type")
	}
	return options, err
}

func (c HTTPClient) postOptions(path string, payload any, group string) ([]Option, error) {
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
		return nil, fmt.Errorf("backend returned %s", resp.Status)
	}

	var data any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return findOptions(data, group), nil
}

func findOptions(value any, group string) []Option {
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

func optionsFromArray(values []any) []Option {
	options := make([]Option, 0, len(values))
	for _, value := range values {
		switch typed := value.(type) {
		case string:
			options = append(options, Option{ID: typed, Label: typed})
		case map[string]any:
			option := Option{
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
