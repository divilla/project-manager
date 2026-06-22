package dto

type (
	Project struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
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
