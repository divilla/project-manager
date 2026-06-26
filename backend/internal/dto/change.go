package dto

import "time"

type (
	ChangeReferences struct {
		Phases []ReferenceOption `json:"phases"`
		Types  []ReferenceOption `json:"types"`
	}

	Change struct {
		ID          int       `json:"id"`
		Version     int16     `json:"version"`
		ProjectID   int       `json:"project_id"`
		EpicID      *int      `json:"epic_id"`
		ChangePhase string    `json:"change_phase"`
		ChangeTypes []string  `json:"change_types"`
		Title       string    `json:"title"`
		Body        string    `json:"body"`
		BodyHTML    string    `json:"body_html"`
		Closed      bool      `json:"closed"`
		DoneReq     int16     `json:"done_req"`
		TotalReq    int16     `json:"total_req"`
		Completed   int16     `json:"completed"`
		Created     time.Time `json:"created"`
		Modified    time.Time `json:"modified"`
	}

	ChangeDetail struct {
		Change       Change        `json:"change"`
		Requirements []Requirement `json:"requirements"`
	}

	ChangeRenderedBodiesRequest struct {
		IDs []int `json:"ids"`
	}

	ChangeRenderedBody struct {
		ID       int    `json:"id"`
		BodyHTML string `json:"body_html"`
	}

	ChangeRenderedBodiesResponse struct {
		Bodies []ChangeRenderedBody `json:"bodies"`
	}

	ChangeListRequest struct {
		ProjectID int `json:"project_id"`
	}

	ChangeIDRequest struct {
		ID int `json:"id"`
	}

	ChangeCreateRequest struct {
		ProjectID   int      `json:"project_id"`
		EpicID      *int     `json:"epic_id"`
		ChangePhase string   `json:"change_phase"`
		ChangeTypes []string `json:"change_types"`
		Title       string   `json:"title"`
		Body        string   `json:"body"`
	}

	ChangeUpdateRequest struct {
		ID          int      `json:"id"`
		ChangeTypes []string `json:"change_types"`
		Title       string   `json:"title"`
		Body        string   `json:"body"`
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
