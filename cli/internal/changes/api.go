package changes

import "mch/internal/dto"

// API defines backend operations needed by change screens.
type API interface {
	ListChangeRows(projectID string) ([]dto.Change, error)
	GetChange(id int) (dto.Change, error)
	CreateChange(input dto.ChangeCreateInput) (dto.Change, error)
	UpdateChangeTitle(id int, title string) (dto.Change, error)
	UpdateChangeRequirementBody(id int, requirementBody string) (dto.Change, error)
	UpdateChangePullRequestBody(id int, pullRequestBody string) (dto.Change, error)
	UpdateChangeTypes(id int, changeTypes []string) (dto.Change, error)
	UpdateChangePhase(id int, changePhase string) (dto.Change, error)
	UpdateChangeEpic(id int, epicID *int) (dto.Change, error)
	DeleteChange(id int) error
	ListPhases() ([]dto.Option, error)
	ListTypes() ([]dto.Option, error)
}
