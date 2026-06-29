package projects

import "mch/internal/dto"

// API defines backend operations needed by project screens.
type API interface {
	ListProjects() ([]dto.Option, error)
	ListProjectRows() ([]dto.Project, error)
}
