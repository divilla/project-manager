package dto

type (
	RequirementListRequest struct {
		TaskID string `json:"task_id"`
	}

	RequirementIDRequest struct {
		ID string `json:"id"`
	}

	RequirementCreateRequest struct {
		TaskID     string `json:"task_id"`
		Definition string `json:"definition"`
	}

	RequirementUpdateRequest struct {
		ID         string `json:"id"`
		Definition string `json:"definition"`
		Done       *bool  `json:"done"`
	}

	RequirementMutationResponse struct {
		Requirement  *Requirement  `json:"requirement,omitempty"`
		Task         Task          `json:"task"`
		Requirements []Requirement `json:"requirements"`
	}
)
