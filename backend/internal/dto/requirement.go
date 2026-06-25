package dto

import "time"

type (
	Requirement struct {
		ID         int       `json:"id"`
		Version    int16     `json:"version"`
		Definition string    `json:"definition"`
		Done       bool      `json:"done"`
		ChangeID   int       `json:"change_id"`
		Created    time.Time `json:"created"`
		Modified   time.Time `json:"modified"`
	}

	RequirementListRequest struct {
		ChangeID int `json:"change_id"`
	}

	RequirementIDRequest struct {
		ID int `json:"id"`
	}

	RequirementCreateRequest struct {
		Definition string `json:"definition"`
		ChangeID   int    `json:"change_id"`
	}

	RequirementUpdateRequest struct {
		ID         int    `json:"id"`
		Definition string `json:"definition"`
	}

	RequirementUpdateDoneRequest struct {
		ID   int  `json:"id"`
		Done bool `json:"done"`
	}

	RequirementUpdateChangeRequest struct {
		ID     int `json:"id"`
		TaskID int `json:"task_id"`
	}

	RequirementMutationResponse struct {
		Requirement  *Requirement  `json:"requirement,omitempty"`
		Change       Change        `json:"change"`
		Requirements []Requirement `json:"requirements"`
	}
)
