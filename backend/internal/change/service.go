package change

import (
	"context"
	"errors"
	"strings"

	"aipm/internal/dto"
)

var (
	// ErrInvalidInput is a package-level value.
	ErrInvalidInput = errors.New("invalid change input")
	// ErrInvalidReference is returned when a change reference is invalid.
	ErrInvalidReference = errors.New("invalid change reference")
	// ErrNotFound is returned when a change cannot be found.
	ErrNotFound = errors.New("change not found")
)

// Service defines Service values.
type Service struct {
	repo     Repository
	renderer Renderer
}

// NewService initializes or executes NewService behavior.
func NewService(changeRepository Repository, renderer Renderer) *Service {
	return &Service{repo: changeRepository, renderer: renderer}
}

// References executes References behavior.
func (s *Service) References(ctx context.Context) (dto.ChangeReferences, error) {
	return s.repo.References(ctx)
}

// ListChanges executes ListChanges behavior.
func (s *Service) ListChanges(ctx context.Context, req dto.ChangeListRequest) ([]dto.Change, error) {
	if req.ProjectID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.List(ctx, req.ProjectID)
}

// GetChange executes GetChange behavior.
func (s *Service) GetChange(ctx context.Context, req dto.ChangeIDRequest) (dto.ChangeDetail, error) {
	if req.ID <= 0 {
		return dto.ChangeDetail{}, ErrInvalidInput
	}
	detail, err := s.repo.Get(ctx, req.ID)
	if err != nil {
		return dto.ChangeDetail{}, err
	}
	detail.Change = s.renderer.RenderChange(detail.Change)
	return detail, nil
}

// RenderedBodies executes RenderedBodies behavior.
func (s *Service) RenderedBodies(ctx context.Context, req dto.ChangeRenderedBodiesRequest) (dto.ChangeRenderedBodiesResponse, error) {
	ids, err := normalizeIDs(req.IDs)
	if err != nil {
		return dto.ChangeRenderedBodiesResponse{}, err
	}
	if len(ids) == 0 {
		return dto.ChangeRenderedBodiesResponse{Bodies: []dto.ChangeRenderedBody{}}, nil
	}
	changes, err := s.repo.Bodies(ctx, ids)
	if err != nil {
		return dto.ChangeRenderedBodiesResponse{}, err
	}
	bodies := make([]dto.ChangeRenderedBody, 0, len(changes))
	for _, item := range changes {
		item = s.renderer.RenderChange(item)
		bodies = append(bodies, dto.ChangeRenderedBody{
			ID:              item.ID,
			RequirementHTML: item.RequirementHTML,
			PullRequestHTML: item.PullRequestHTML,
		})
	}
	return dto.ChangeRenderedBodiesResponse{Bodies: bodies}, nil
}

// CreateChange executes CreateChange behavior.
func (s *Service) CreateChange(ctx context.Context, req dto.ChangeCreateRequest) (dto.Change, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.RequirementBody = strings.TrimSpace(req.RequirementBody)
	req.PullRequestBody = strings.TrimSpace(req.PullRequestBody)
	req.PullRequestURL = strings.TrimSpace(req.PullRequestURL)
	req.ChangePhase = strings.TrimSpace(req.ChangePhase)
	req.ChangeTypes = normalizeTypes(req.ChangeTypes)
	if req.ProjectID <= 0 || req.Title == "" || req.ChangePhase == "" || len(req.ChangeTypes) == 0 || invalidOptionalID(req.EpicID) {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.Create(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdateChangeTypes executes UpdateChangeTypes behavior.
func (s *Service) UpdateChangeTypes(ctx context.Context, req dto.ChangeUpdateChangeTypesRequest) (dto.Change, error) {
	req.ChangeTypes = normalizeTypes(req.ChangeTypes)
	if req.ID <= 0 || len(req.ChangeTypes) == 0 {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdateChangeTypes(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdateTitle executes UpdateTitle behavior.
func (s *Service) UpdateTitle(ctx context.Context, req dto.ChangeUpdateTitleRequest) (dto.Change, error) {
	req.Title = strings.TrimSpace(req.Title)
	if req.ID <= 0 || req.Title == "" {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdateTitle(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdateRequirementBody executes UpdateRequirementBody behavior.
func (s *Service) UpdateRequirementBody(ctx context.Context, req dto.ChangeUpdateRequirementBodyRequest) (dto.Change, error) {
	req.RequirementBody = strings.TrimSpace(req.RequirementBody)
	if req.ID <= 0 {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdateRequirementBody(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdatePullRequestBody executes UpdatePullRequestBody behavior.
func (s *Service) UpdatePullRequestBody(ctx context.Context, req dto.ChangeUpdatePullRequestBodyRequest) (dto.Change, error) {
	req.PullRequestBody = strings.TrimSpace(req.PullRequestBody)
	if req.ID <= 0 {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdatePullRequestBody(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdateEpic executes UpdateEpic behavior.
func (s *Service) UpdateEpic(ctx context.Context, req dto.ChangeUpdateEpicRequest) (dto.Change, error) {
	if req.ID <= 0 || invalidOptionalID(req.EpicID) {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdateEpic(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdatePhase executes UpdatePhase behavior.
func (s *Service) UpdatePhase(ctx context.Context, req dto.ChangeUpdatePhaseRequest) (dto.Change, error) {
	req.ChangePhase = strings.TrimSpace(req.ChangePhase)
	if req.ID <= 0 || req.ChangePhase == "" {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdatePhase(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdateClosed executes UpdateClosed behavior.
func (s *Service) UpdateClosed(ctx context.Context, req dto.ChangeUpdateClosedRequest) (dto.Change, error) {
	if req.ID <= 0 {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdateClosed(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// DeleteChange executes DeleteChange behavior.
func (s *Service) DeleteChange(ctx context.Context, req dto.ChangeIDRequest) error {
	if req.ID <= 0 {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, req)
}

func normalizeIDs(ids []int) ([]int, error) {
	normalized := make([]int, 0, len(ids))
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, ErrInvalidInput
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	return normalized, nil
}

func normalizeTypes(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func invalidOptionalID(value *int) bool {
	return value != nil && *value <= 0
}
