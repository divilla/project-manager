package project

import (
	"context"
	"errors"
	"strings"

	"aipm/internal/dto"
)

var (
	ErrInvalidInput = errors.New("invalid project input")
	ErrNotFound     = errors.New("project not found")
)

type (
	Service struct {
		repo Repository
	}
)

func NewService(projectRepository Repository) *Service {
	return &Service{
		repo: projectRepository,
	}
}

func (s *Service) ListProjects(ctx context.Context, req dto.ProjectListRequest) ([]dto.Project, error) {
	limit := req.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, limit, offset)
}

func (s *Service) GetProject(ctx context.Context, req dto.ProjectIDRequest) (dto.Project, error) {
	id := strings.TrimSpace(req.ID)
	if id == "" {
		return dto.Project{}, ErrInvalidInput
	}
	return s.repo.Get(ctx, id)
}

func (s *Service) CreateProject(ctx context.Context, req dto.ProjectCreateRequest) (dto.Project, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return dto.Project{}, ErrInvalidInput
	}
	return s.repo.Create(ctx, name)
}

func (s *Service) UpdateProject(ctx context.Context, req dto.ProjectUpdateRequest) (dto.Project, error) {
	id := strings.TrimSpace(req.ID)
	name := strings.TrimSpace(req.Name)
	if id == "" || name == "" {
		return dto.Project{}, ErrInvalidInput
	}
	return s.repo.Update(ctx, id, name)
}

func (s *Service) DeleteProject(ctx context.Context, req dto.ProjectIDRequest) error {
	id := strings.TrimSpace(req.ID)
	if id == "" {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, id)
}
