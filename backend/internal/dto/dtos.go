package dto

import "time"

type (
	ReferenceOption struct {
		Slug     string `json:"slug"`
		Priority int    `json:"priority"`
	}

	Requirement struct {
		ID         string    `json:"id"`
		TaskID     string    `json:"task_id"`
		Definition string    `json:"definition"`
		Done       bool      `json:"done"`
		Created    time.Time `json:"created"`
		Modified   time.Time `json:"modified"`
	}
)
