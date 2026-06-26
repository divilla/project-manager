package dto

import "time"

type (
	Epic struct {
		ID        int       `json:"id"`
		Version   int16     `json:"version"`
		ProjectID int       `json:"project_id"`
		Name      string    `json:"name"`
		DoneReq   int16     `json:"done_req"`
		TotalReq  int16     `json:"total_req"`
		Completed int16     `json:"completed"`
		Created   time.Time `json:"created"`
		Modified  time.Time `json:"modified"`
	}

	EpicListRequest struct {
		ProjectID int `json:"project_id"`
	}

	EpicIDRequest struct {
		ID int `json:"id"`
	}

	EpicCreateRequest struct {
		ProjectID int    `json:"project_id"`
		Name      string `json:"name"`
	}

	EpicUpdateRequest struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
)
