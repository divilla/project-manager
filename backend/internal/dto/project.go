package dto

import "time"

type (
	// Project defines Project values.
	Project struct {
		ID          int       `json:"id"`
		Name        string    `json:"name"`
		Created     time.Time `json:"created"`
		Modified    time.Time `json:"modified"`
		ChangeCount int       `json:"change_count"`
	}

	// ProjectListRequest defines ProjectListRequest values.
	ProjectListRequest struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	}

	// ProjectIDRequest defines ProjectIDRequest values.
	ProjectIDRequest struct {
		ID int `json:"id"`
	}

	// ProjectCreateRequest defines ProjectCreateRequest values.
	ProjectCreateRequest struct {
		Name string `json:"name"`
	}

	// ProjectUpdateRequest defines ProjectUpdateRequest values.
	ProjectUpdateRequest struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
)
