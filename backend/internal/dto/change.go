package dto

import "time"

type (
	// Change defines Change values.
	Change struct {
		ID          int       `json:"id"`
		Version     int16     `json:"version"`
		Ref         int32     `json:"ref"`
		Slug        string    `json:"slug"`
		ProjectID   int       `json:"project_id"`
		EpicID      *int      `json:"epic_id"`
		EpicName    *string   `json:"epic_name"`
		ChangePhase string    `json:"change_phase"`
		ChangeTypes []string  `json:"change_types"`
		Title       string    `json:"title"`
		Body        string    `json:"body"`
		HTML        string    `json:"html"`
		PRBody      string    `json:"pr_body"`
		PRHtml      string    `json:"pr_html"`
		PRUrl       string    `json:"pr_url"`
		AgentEdit   bool      `json:"agent_edit"`
		Open        bool      `json:"open"`
		DoneTC      int16     `json:"done_tc"`
		TotalTC     int16     `json:"total_tc"`
		Completed   int16     `json:"completed"`
		Created     time.Time `json:"created"`
		Modified    time.Time `json:"modified"`
	}

	// ChangeListItem defines ChangeListItem values.
	ChangeListItem struct {
		ID          int       `json:"id"`
		Ref         int32     `json:"ref"`
		Slug        string    `json:"slug"`
		ProjectID   int       `json:"project_id"`
		ChangePhase string    `json:"change_phase"`
		ChangeTypes []string  `json:"change_types"`
		EpicID      *int      `json:"epic_id"`
		EpicName    *string   `json:"epic_name"`
		Title       string    `json:"title"`
		AgentEdit   bool      `json:"agent_edit"`
		Open        bool      `json:"open"`
		DoneTC      int16     `json:"done_tc"`
		TotalTC     int16     `json:"total_tc"`
		Completed   int16     `json:"completed"`
		Modified    time.Time `json:"modified"`
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
		ID     int    `json:"id"`
		HTML   string `json:"html"`
		PRHtml string `json:"pr_html"`
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
		ChangeTypes []string `json:"change_types"`
		EpicID      *int     `json:"epic_id"`
		Title       string   `json:"title"`
		Body        string   `json:"body"`
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

	// ChangeUpdateBodyRequest defines ChangeUpdateBodyRequest values.
	ChangeUpdateBodyRequest struct {
		ID   int    `json:"id"`
		Body string `json:"body"`
	}

	// ChangeUpdatePRBodyRequest defines ChangeUpdatePRBodyRequest values.
	ChangeUpdatePRBodyRequest struct {
		ID     int    `json:"id"`
		PRBody string `json:"pr_body"`
	}

	// ChangeUpdatePRUrlRequest defines ChangeUpdatePRUrlRequest values.
	ChangeUpdatePRUrlRequest struct {
		ID    int    `json:"id"`
		PRUrl string `json:"pr_url"`
	}

	// ChangeUpdateAgentEditRequest defines ChangeUpdateAgentEditRequest values.
	ChangeUpdateAgentEditRequest struct {
		ID        int   `json:"id"`
		AgentEdit *bool `json:"agent_edit"`
	}

	// ChangeUpdateOpenRequest defines ChangeUpdateOpenRequest values.
	ChangeUpdateOpenRequest struct {
		ID   int   `json:"id"`
		Open *bool `json:"open"`
	}
)
