package dto

import (
	"time"
)

type (
	// Change defines Change values.
	Change struct {
		ID             int        `json:"id"`
		Version        int16      `json:"version"`
		Ref            *int32     `json:"ref"`
		Slug           *string    `json:"slug"`
		ProjectID      int        `json:"project_id"`
		EpicID         *int       `json:"epic_id"`
		EpicName       *string    `json:"epic_name"`
		ChangePhase    string     `json:"change_phase"`
		ChangeTypes    []string   `json:"change_types"`
		Title          string     `json:"title"`
		Idea           string     `json:"idea"`
		Spec           *string    `json:"spec"`
		SpecHTML       *string    `json:"spec_html"`
		PRBody         *string    `json:"pr_body"`
		PRHtml         *string    `json:"pr_html"`
		PRUrl          *string    `json:"pr_url"`
		AgentEdit      bool       `json:"agent_edit"`
		FlowStages     []string   `json:"flow_stages"`
		FlowStageModes []string   `json:"flow_stage_modes"`
		RunClaimID     *string    `json:"run_claim_id"`
		RunFlowStage   string     `json:"run_flow_stage"`
		RunTaskStep    string     `json:"run_task_step"`
		RunTaskStatus  string     `json:"run_task_status"`
		RunError       string     `json:"run_error"`
		RunIsCompleted bool       `json:"run_is_completed"`
		RunStartedAt   *time.Time `json:"run_started_at"`
		RunUpdatedAt   *time.Time `json:"run_updated_at"`
		Open           bool       `json:"open"`
		DoneTC         int16      `json:"done_tc"`
		TotalTC        int16      `json:"total_tc"`
		Completed      int16      `json:"completed"`
		Created        time.Time  `json:"created"`
		Modified       time.Time  `json:"modified"`
	}

	// ChangeListItem defines ChangeListItem values.
	ChangeListItem struct {
		ID          int       `json:"id"`
		Ref         *int32    `json:"ref"`
		Slug        *string   `json:"slug"`
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

	// ChangeRenderedArtifactsRequest defines ChangeRenderedArtifactsRequest values.
	ChangeRenderedArtifactsRequest struct {
		IDs []int `json:"ids"`
	}

	// ChangeRenderedArtifact defines ChangeRenderedArtifact values.
	ChangeRenderedArtifact struct {
		ID       int    `json:"id"`
		SpecHTML string `json:"spec_html"`
		PRHtml   string `json:"pr_html"`
	}

	// ChangeRenderedArtifactsResponse defines ChangeRenderedArtifactsResponse values.
	ChangeRenderedArtifactsResponse struct {
		Artifacts []ChangeRenderedArtifact `json:"artifacts"`
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
		ProjectID int    `json:"project_id"`
		Title     string `json:"title"`
		Idea      string `json:"idea"`
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

	// ChangeUpdateIdeaRequest defines ChangeUpdateIdeaRequest values.
	ChangeUpdateIdeaRequest struct {
		ID   int    `json:"id"`
		Idea string `json:"idea"`
	}

	// ChangeUpdateIdeaAgentEditRequest defines ChangeUpdateIdeaAgentEditRequest values.
	ChangeUpdateIdeaAgentEditRequest struct {
		ID   int    `json:"id"`
		Idea string `json:"idea"`
	}

	// ChangeUpdateSpecRequest defines ChangeUpdateSpecRequest values.
	ChangeUpdateSpecRequest struct {
		ID   int     `json:"id"`
		Spec *string `json:"spec"`
	}

	// ChangeUpdatePRBodyRequest defines ChangeUpdatePRBodyRequest values.
	ChangeUpdatePRBodyRequest struct {
		ID     int     `json:"id"`
		PRBody *string `json:"pr_body"`
	}

	// ChangeUpdatePRUrlRequest defines ChangeUpdatePRUrlRequest values.
	ChangeUpdatePRUrlRequest struct {
		ID    int     `json:"id"`
		PRUrl *string `json:"pr_url"`
	}

	// ChangeUpdateAgentEditRequest defines ChangeUpdateAgentEditRequest values.
	ChangeUpdateAgentEditRequest struct {
		ID        int   `json:"id"`
		AgentEdit *bool `json:"agent_edit"`
	}

	// ChangeUpdateRunRequest defines ChangeUpdateRunRequest values.
	ChangeUpdateRunRequest struct {
		ID             int    `json:"id"`
		RunClaimID     string `json:"run_claim_id"`
		RunFlowStage   string `json:"run_flow_stage"`
		RunTaskStep    string `json:"run_task_step"`
		RunTaskStatus  string `json:"run_task_status"`
		RunError       string `json:"run_error"`
		RunIsCompleted bool   `json:"run_is_completed"`
	}

	// ChangeRunClaimResponse defines ChangeRunClaimResponse values.
	ChangeRunClaimResponse struct {
		ClaimID *string `json:"claim_id"`
	}

	// ChangeRunUpdateResponse defines ChangeRunUpdateResponse values.
	ChangeRunUpdateResponse struct {
		ChangeID *int `json:"change_id"`
	}

	// ChangeUpdateOpenRequest defines ChangeUpdateOpenRequest values.
	ChangeUpdateOpenRequest struct {
		ID   int   `json:"id"`
		Open *bool `json:"open"`
	}
)
