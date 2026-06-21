package project

import (
	"context"

	"github.com/divilla/project-manager/internal/dto"
)

type (
	Service struct {
		repo *Repository
	}
)

func NewService(projectRepository *Repository) *Service {
	return &Service{
		repo: projectRepository,
	}
}

func (s *Service) ListProjects(ctx context.Context) ([]dto.Project, error) {
	return s.repo.List(ctx, 100, 0)
}
