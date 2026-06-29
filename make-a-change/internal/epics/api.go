package epics

import "mch/internal/dto"

// API defines backend operations needed by epic screens.
type API interface {
	ListEpics(projectID string) ([]dto.Option, error)
}
