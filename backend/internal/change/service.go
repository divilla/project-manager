package change

import (
	"context"
	"errors"
	"strings"

	"aipm/internal/changeview"
	"aipm/internal/dto"
)

var (
	ErrInvalidInput     = errors.New("invalid change input")
	ErrInvalidReference = errors.New("invalid change reference")
	ErrNotFound         = errors.New("change not found")
)

type Service struct {
	repo     Repository
	renderer changeview.ChangeRenderer
}

func NewService(changeRepository Repository, renderer changeview.ChangeRenderer) *Service {
	return &Service{repo: changeRepository, renderer: renderer}
}

func (s *Service) References(ctx context.Context) (dto.ChangeReferences, error) {
	return s.repo.References(ctx)
}

func (s *Service) ListChanges(ctx context.Context, req dto.ChangeListRequest) ([]dto.Change, error) {
	if req.ProjectID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.List(ctx, req.ProjectID)
}

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
		bodies = append(bodies, dto.ChangeRenderedBody{ID: item.ID, BodyHTML: item.BodyHTML})
	}
	return dto.ChangeRenderedBodiesResponse{Bodies: bodies}, nil
}

func (s *Service) CreateChange(ctx context.Context, req dto.ChangeCreateRequest) (dto.Change, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	req.ChangePhase = strings.TrimSpace(req.ChangePhase)
	req.ChangeTypes = normalizeTypes(req.ChangeTypes)
	req.CodexSessionID = normalizeOptionalString(req.CodexSessionID)
	if req.ProjectID <= 0 || req.Title == "" || req.ChangePhase == "" || len(req.ChangeTypes) == 0 || invalidOptionalID(req.EpicID) {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.Create(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

func (s *Service) UpdateChange(ctx context.Context, req dto.ChangeUpdateRequest) (dto.Change, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	req.ChangeTypes = normalizeTypes(req.ChangeTypes)
	if req.ID <= 0 || req.Title == "" || len(req.ChangeTypes) == 0 {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.Update(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

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

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}
	return &normalized
}
