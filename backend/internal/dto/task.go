package dto

import "time"

type (
	TaskReferences struct {
		Phases []ReferenceOption `json:"phases"`
		Types  []ReferenceOption `json:"types"`
	}

	Task struct {
		ID              int       `json:"id"`
		Version         int16     `json:"version"`
		TaskType        string    `json:"task_type"`
		Name            string    `json:"name"`
		Description     string    `json:"description"`
		DescriptionHTML string    `json:"description_html"`
		Difficulty      int16     `json:"difficulty"`
		Priority        int16     `json:"priority"`
		TaskPhase       string    `json:"task_phase"`
		ParentID        *int      `json:"parent_id,omitempty"`
		ProjectID       int       `json:"project_id"`
		DoneReq         int16     `json:"done_req"`
		TotalReq        int16     `json:"total_req"`
		Completed       int16     `json:"completed"`
		Created         time.Time `json:"created"`
		Modified        time.Time `json:"modified"`
	}

	TaskDetail struct {
		Task         Task          `json:"task"`
		Requirements []Requirement `json:"requirements"`
	}

	TaskRenderedDescriptionsRequest struct {
		IDs []int `json:"ids"`
	}

	TaskRenderedDescription struct {
		ID              int    `json:"id"`
		DescriptionHTML string `json:"description_html"`
	}

	TaskRenderedDescriptionsResponse struct {
		Descriptions []TaskRenderedDescription `json:"descriptions"`
	}

	TaskListRequest struct {
		ProjectID int `json:"project_id"`
	}

	TaskIDRequest struct {
		ID int `json:"id"`
	}

	TaskCreateRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		TaskPhase   string `json:"task_phase"`
		TaskType    string `json:"task_type"`
		Difficulty  int16  `json:"difficulty"`
		Priority    int16  `json:"priority"`
		ParentID    *int   `json:"parent_id"`
		ProjectID   int    `json:"project_id"`
	}

	TaskUpdateRequest struct {
		ID          int    `json:"id"`
		TaskType    string `json:"task_type"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	TaskUpdateDifficultyRequest struct {
		ID         int   `json:"id"`
		Difficulty int16 `json:"difficulty"`
	}

	TaskUpdatePriorityRequest struct {
		ID       int   `json:"id"`
		Priority int16 `json:"priority"`
	}

	TaskUpdateParentRequest struct {
		ID       int  `json:"id"`
		ParentID *int `json:"parent_id"`
	}

	TaskUpdatePhaseRequest struct {
		ID        int    `json:"id"`
		TaskPhase string `json:"task_phase"`
	}
)
