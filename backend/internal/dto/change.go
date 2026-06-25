package dto

import "time"

type (
	ChangeReferences struct {
		Phases []ReferenceOption `json:"phases"`
		Types  []ReferenceOption `json:"types"`
	}

	Change struct {
		ID           int       `json:"id"`
		Version      int16     `json:"version"`
		ProjectID    int       `json:"project_id"`
		EpicID       *int      `json:"epic_id"`
		ChangesPhase string    `json:"changes_phase"`
		ChangeTypes  []string  `json:"change_types"`
		Name         string    `json:"name"`
		Body         string    `json:"body"`
		BodyHTML     string    `json:"body_html"`
		Closed       bool      `json:"closed"`
		DoneReq      int16     `json:"done_req"`
		TotalReq     int16     `json:"total_req"`
		Completed    int16     `json:"completed"`
		Created      time.Time `json:"created"`
		Modified     time.Time `json:"modified"`
	}

	ChangeDetail struct {
		Change       Change        `json:"change"`
		Requirements []Requirement `json:"requirements"`
	}

	ChangeRenderedDescriptionsRequest struct {
		IDs []int `json:"ids"`
	}

	ChangeRenderedDescription struct {
		ID              int    `json:"id"`
		DescriptionHTML string `json:"description_html"`
	}

	ChangeRenderedDescriptionsResponse struct {
		Descriptions []ChangeRenderedDescription `json:"descriptions"`
	}

	ChangeListRequest struct {
		ProjectID int `json:"project_id"`
	}

	ChangeIDRequest struct {
		ID int `json:"id"`
	}

	ChangeCreateRequest struct {
		ProjectID   int      `json:"project_id"`
		EpicID      int      `json:"epic_id"`
		ChangePhase string   `json:"change_phase"`
		ChangeTypes []string `json:"change_types"`
		Title       string   `json:"title"`
		Body        string   `json:"body"`
	}

	ChangeUpdateRequest struct {
		ID          int      `json:"id"`
		ChangeTypes []string `json:"change_types"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
	}

	ChangeUpdateEpicRequest struct {
		ID     int  `json:"id"`
		EpicID *int `json:"epic_id"`
	}

	ChangeUpdatePhaseRequest struct {
		ID          int    `json:"id"`
		ChangePhase string `json:"change_phase"`
	}

	ChangeUpdateClosedRequest struct {
		ID     int  `json:"id"`
		Closed bool `json:"closed"`
	}
)
