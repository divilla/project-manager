package changes

import "mch/internal/dto"

// API defines backend operations needed by change screens.
type API interface {
	ListChangeRows(projectID string) ([]dto.Change, error)
	GetChange(id int) (dto.Change, error)
	CreateChange(input dto.ChangeCreateInput) (dto.Change, error)
	UpdateChangeTitle(id int, title string) (dto.Change, error)
	UpdateChangeRequirementBody(id int, requirementBody string) (dto.Change, error)
	UpdateChangeTypes(id int, changeTypes []string) (dto.Change, error)
	UpdateChangeEpic(id int, epicID *int) (dto.Change, error)
	ListPhases() ([]dto.Option, error)
	ListTypes() ([]dto.Option, error)
}
