package dto

type (
	Project struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}

	ProjectListRequest struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	}

	ProjectIDRequest struct {
		ID string `json:"id"`
	}

	ProjectCreateRequest struct {
		Name string `json:"name"`
	}

	ProjectUpdateRequest struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
)
