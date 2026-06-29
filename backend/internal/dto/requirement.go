package dto

import "time"

type (
	// Requirement defines Requirement values.
	Requirement struct {
		ID         int       `json:"id"`
		Version    int16     `json:"version"`
		Definition string    `json:"definition"`
		Done       bool      `json:"done"`
		ChangeID   int       `json:"change_id"`
		Created    time.Time `json:"created"`
		Modified   time.Time `json:"modified"`
	}

	// RequirementListRequest defines RequirementListRequest values.
	RequirementListRequest struct {
		ChangeID int `json:"change_id"`
	}

	// RequirementIDRequest defines RequirementIDRequest values.
	RequirementIDRequest struct {
		ID int `json:"id"`
	}

	// RequirementCreateRequest defines RequirementCreateRequest values.
	RequirementCreateRequest struct {
		Definition string `json:"definition"`
		ChangeID   int    `json:"change_id"`
	}

	// RequirementUpdateRequest defines RequirementUpdateRequest values.
	RequirementUpdateRequest struct {
		ID         int    `json:"id"`
		Definition string `json:"definition"`
	}

	// RequirementUpdateDoneRequest defines RequirementUpdateDoneRequest values.
	RequirementUpdateDoneRequest struct {
		ID   int  `json:"id"`
		Done bool `json:"done"`
	}

	// RequirementUpdateChangeRequest defines RequirementUpdateChangeRequest values.
	RequirementUpdateChangeRequest struct {
		ID       int `json:"id"`
		ChangeID int `json:"change_id"`
	}

	// RequirementMutationResponse defines RequirementMutationResponse values.
	RequirementMutationResponse struct {
		Requirement  *Requirement  `json:"requirement,omitempty"`
		Change       Change        `json:"change"`
		Requirements []Requirement `json:"requirements"`
	}
)
