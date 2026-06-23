package dto

import "time"

type (
	Project struct {
		ID        int       `json:"id"`
		Name      string    `json:"name"`
		Created   time.Time `json:"created"`
		Modified  time.Time `json:"modified"`
		TaskCount int       `json:"task_count"`
	}

	ProjectListRequest struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	}

	ProjectIDRequest struct {
		ID int `json:"id"`
	}

	ProjectCreateRequest struct {
		Name string `json:"name"`
	}

	ProjectUpdateRequest struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
)
