package epic

import (
	"context"
	"errors"
	"strings"

	"aipm/internal/dto"
)

var (
	ErrInvalidInput   = errors.New("invalid epic input")
	ErrNotFound       = errors.New("epic not found")
	ErrEpicHasChanges = errors.New("epic has changes")
)

type (
	Service struct {
		repo Repository
	}

	Repository interface {
		List(ctx context.Context, projectID int) ([]dto.Epic, error)
		Get(ctx context.Context, id int) (dto.Epic, error)
		Create(ctx context.Context, req dto.EpicCreateRequest) (dto.Epic, error)
		Update(ctx context.Context, req dto.EpicUpdateRequest) (dto.Epic, error)
		Delete(ctx context.Context, id int) error
	}
)

func NewService(epicRepository Repository) *Service {
	return &Service{repo: epicRepository}
}

func (s *Service) ListEpics(ctx context.Context, req dto.EpicListRequest) ([]dto.Epic, error) {
	if req.ProjectID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.List(ctx, req.ProjectID)
}

func (s *Service) GetEpic(ctx context.Context, req dto.EpicIDRequest) (dto.Epic, error) {
	if req.ID <= 0 {
		return dto.Epic{}, ErrInvalidInput
	}
	return s.repo.Get(ctx, req.ID)
}

func (s *Service) CreateEpic(ctx context.Context, req dto.EpicCreateRequest) (dto.Epic, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.ProjectID <= 0 || req.Name == "" {
		return dto.Epic{}, ErrInvalidInput
	}
	return s.repo.Create(ctx, req)
}

func (s *Service) UpdateEpic(ctx context.Context, req dto.EpicUpdateRequest) (dto.Epic, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.ID <= 0 || req.Name == "" {
		return dto.Epic{}, ErrInvalidInput
	}
	return s.repo.Update(ctx, req)
}

func (s *Service) DeleteEpic(ctx context.Context, req dto.EpicIDRequest) error {
	if req.ID <= 0 {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, req.ID)
}
