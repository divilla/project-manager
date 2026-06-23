package project

import (
	"context"
	"errors"
	"strings"

	"aipm/internal/dto"
)

var (
	ErrInvalidInput    = errors.New("invalid project input")
	ErrNotFound        = errors.New("project not found")
	ErrProjectHasTasks = errors.New("project has tasks")
)

type Service struct {
	repo Repository
}

func NewService(projectRepository Repository) *Service {
	return &Service{repo: projectRepository}
}

func (s *Service) ListProjects(ctx context.Context, req dto.ProjectListRequest) ([]dto.Project, error) {
	limit := req.Limit
	if limit < 0 {
		limit = 0
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}

func (s *Service) GetProject(ctx context.Context, req dto.ProjectIDRequest) (dto.Project, error) {
	if req.ID <= 0 {
		return dto.Project{}, ErrInvalidInput
	}
	return s.repo.Get(ctx, req.ID)
}

func (s *Service) CreateProject(ctx context.Context, req dto.ProjectCreateRequest) (dto.Project, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return dto.Project{}, ErrInvalidInput
	}
	return s.repo.Create(ctx, name)
}

func (s *Service) UpdateProject(ctx context.Context, req dto.ProjectUpdateRequest) (dto.Project, error) {
	name := strings.TrimSpace(req.Name)
	if req.ID <= 0 || name == "" {
		return dto.Project{}, ErrInvalidInput
	}
	return s.repo.Update(ctx, req.ID, name)
}

func (s *Service) DeleteProject(ctx context.Context, req dto.ProjectIDRequest) error {
	if req.ID <= 0 {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, req.ID)
}
