package changes

import "mch/internal/dto"

// API defines backend operations needed by change screens.
type API interface {
	ListPhases() ([]dto.Option, error)
	ListTypes() ([]dto.Option, error)
}
