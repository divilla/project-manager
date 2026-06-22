package dto

type (
	Health struct {
		Status   string `json:"status"`
		API      string `json:"api"`
		Database string `json:"database"`
		Error    string `json:"error,omitempty"`
	}
)
