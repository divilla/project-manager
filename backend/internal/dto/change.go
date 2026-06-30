package dto

import "time"

type (
	// ChangeReferences defines ChangeReferences values.
	ChangeReferences struct {
		Phases []ChangePhase `json:"phases"`
		Types  []ChangePhase `json:"types"`
	}

	// Change defines Change values.
	Change struct {
		ID              int       `json:"id"`
		Version         int16     `json:"version"`
		Ref             int32     `json:"ref"`
		Slug            string    `json:"slug"`
		ProjectID       int       `json:"project_id"`
		EpicID          *int      `json:"epic_id"`
		ChangePhase     string    `json:"change_phase"`
		ChangeTypes     []string  `json:"change_types"`
		Title           string    `json:"title"`
		RequirementBody string    `json:"requirement_body"`
		RequirementHTML string    `json:"requirement_html"`
		PullRequestBody string    `json:"pull_request_body"`
		PullRequestHTML string    `json:"pull_request_html"`
		PullRequestURL  string    `json:"pull_request_url"`
		Closed          bool      `json:"closed"`
		DoneTC          int16     `json:"done_tc"`
		TotalTC         int16     `json:"total_tc"`
		Completed       int16     `json:"completed"`
		Created         time.Time `json:"created"`
		Modified        time.Time `json:"modified"`
	}

	// ChangeDetail defines ChangeDetail values.
	ChangeDetail struct {
		Change    Change     `json:"change"`
		TestCases []TestCase `json:"test_cases"`
	}

	// ChangeRenderedBodiesRequest defines ChangeRenderedBodiesRequest values.
	ChangeRenderedBodiesRequest struct {
		IDs []int `json:"ids"`
	}

	// ChangeRenderedBody defines ChangeRenderedBody values.
	ChangeRenderedBody struct {
		ID              int    `json:"id"`
		RequirementHTML string `json:"requirement_html"`
		PullRequestHTML string `json:"pull_request_html"`
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
		ProjectID       int      `json:"project_id"`
		ChangeTypes     []string `json:"change_types"`
		EpicID          *int     `json:"epic_id"`
		Title           string   `json:"title"`
		RequirementBody string   `json:"requirement_body"`
	}

	// ChangeUpdatePhaseRequest defines ChangeUpdatePhaseRequest values.
	ChangeUpdatePhaseRequest struct {
		ID          int    `json:"id"`
		ChangePhase string `json:"change_phase"`
	}

	// ChangeUpdateChangeTypesRequest defines ChangeUpdateChangeTypesRequest values.
	ChangeUpdateChangeTypesRequest struct {
		ID          int      `json:"id"`
		ChangeTypes []string `json:"change_types"`
	}

	// ChangeUpdateEpicRequest defines ChangeUpdateEpicRequest values.
	ChangeUpdateEpicRequest struct {
		ID     int  `json:"id"`
		EpicID *int `json:"epic_id"`
	}

	// ChangeUpdateTitleRequest defines ChangeUpdateTitleRequest values.
	ChangeUpdateTitleRequest struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	}

	// ChangeUpdateRequirementBodyRequest defines ChangeUpdateRequirementBodyRequest values.
	ChangeUpdateRequirementBodyRequest struct {
		ID              int    `json:"id"`
		RequirementBody string `json:"requirement_body"`
	}

	// ChangeUpdatePullRequestBodyRequest defines ChangeUpdatePullRequestBodyRequest values.
	ChangeUpdatePullRequestBodyRequest struct {
		ID              int    `json:"id"`
		PullRequestBody string `json:"pull_request_body"`
	}

	// ChangeUpdatePullRequestURLRequest defines ChangeUpdatePullRequestURLRequest values.
	ChangeUpdatePullRequestURLRequest struct {
		ID             int    `json:"id"`
		PullRequestURL string `json:"pull_request_url"`
	}

	// ChangeUpdateClosedRequest defines ChangeUpdateClosedRequest values.
	ChangeUpdateClosedRequest struct {
		ID     int  `json:"id"`
		Closed bool `json:"closed"`
	}
)
