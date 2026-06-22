package dto

import "time"

type (
	Task struct {
		ID          string    `json:"id"`
		ProjectID   string    `json:"project_id"`
		ParentID    *string   `json:"parent_id,omitempty"`
		Phase       string    `json:"phase"`
		Type        string    `json:"type"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Difficulty  int       `json:"difficulty"`
		Complete    int       `json:"complete"`
		Priority    int       `json:"priority"`
		Depth       int       `json:"depth"`
		Created     time.Time `json:"created"`
		Modified    time.Time `json:"modified"`
	}

	TaskDetail struct {
		Task         Task          `json:"task"`
		Requirements []Requirement `json:"requirements"`
	}

	TaskReferences struct {
		Phases []ReferenceOption `json:"phases"`
		Types  []ReferenceOption `json:"types"`
	}

	TaskListRequest struct {
		ProjectID string `json:"project_id"`
	}

	TaskIDRequest struct {
		ID string `json:"id"`
	}

	TaskCreateRequest struct {
		ProjectID   string `json:"project_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Phase       string `json:"phase"`
		Type        string `json:"type"`
		Difficulty  int    `json:"difficulty"`
		Priority    int    `json:"priority"`
		ParentID    string `json:"parent_id"`
	}

	TaskUpdateRequest struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Difficulty  int    `json:"difficulty"`
		Priority    int    `json:"priority"`
	}

	TaskPhaseRequest struct {
		ID    string `json:"id"`
		Phase string `json:"phase"`
	}
)
