package dto

import "time"

type (
	// ChangeReferences defines ChangeReferences values.
	ChangeReferences struct {
		Phases []ReferenceOption `json:"phases"`
		Types  []ReferenceOption `json:"types"`
	}

	// Change defines Change values.
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

	// ChangeDetail defines ChangeDetail values.
	ChangeDetail struct {
		Change       Change        `json:"change"`
		Requirements []Requirement `json:"requirements"`
	}

	// ChangeRenderedBodiesRequest defines ChangeRenderedBodiesRequest values.
	ChangeRenderedBodiesRequest struct {
		IDs []int `json:"ids"`
	}

	// ChangeRenderedBody defines ChangeRenderedBody values.
	ChangeRenderedBody struct {
		ID       int    `json:"id"`
		BodyHTML string `json:"body_html"`
	}

	// ChangeRenderedBodiesResponse defines ChangeRenderedBodiesResponse values.
	ChangeRenderedBodiesResponse struct {
		Bodies []ChangeRenderedBody `json:"bodies"`
	}

	// ChangeListRequest defines ChangeListRequest values.
	ChangeListRequest struct {
		ProjectID int `json:"project_id"`
	}

	// ChangeIDRequest defines ChangeIDRequest values.
	ChangeIDRequest struct {
		ID int `json:"id"`
	}

	// ChangeCreateRequest defines ChangeCreateRequest values.
	ChangeCreateRequest struct {
		ProjectID   int      `json:"project_id"`
		EpicID      *int     `json:"epic_id"`
		ChangePhase string   `json:"change_phase"`
		ChangeTypes []string `json:"change_types"`
		Title       string   `json:"title"`
		Body        string   `json:"body"`
	}

	// ChangeUpdateRequest defines ChangeUpdateRequest values.
	ChangeUpdateRequest struct {
		ID          int      `json:"id"`
		ChangeTypes []string `json:"change_types"`
		Title       string   `json:"title"`
		Body        string   `json:"body"`
	}

	// ChangeUpdateEpicRequest defines ChangeUpdateEpicRequest values.
	ChangeUpdateEpicRequest struct {
		ID     int  `json:"id"`
		EpicID *int `json:"epic_id"`
	}

	// ChangeUpdatePhaseRequest defines ChangeUpdatePhaseRequest values.
	ChangeUpdatePhaseRequest struct {
		ID          int    `json:"id"`
		ChangePhase string `json:"change_phase"`
	}

	// ChangeUpdateClosedRequest defines ChangeUpdateClosedRequest values.
	ChangeUpdateClosedRequest struct {
		ID     int  `json:"id"`
		Closed bool `json:"closed"`
	}
)
