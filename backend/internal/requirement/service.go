package requirement

import (
	"context"
	"errors"
	"strings"

	"aipm/internal/dto"
)

var (
	ErrInvalidInput = errors.New("invalid requirement input")
	ErrNotFound     = errors.New("requirement not found")
)

type (
	Service struct {
		repo Repository
	}

	Repository interface {
		List(ctx context.Context, taskID string) ([]dto.Requirement, error)
		Create(ctx context.Context, req dto.RequirementCreateRequest) (dto.RequirementMutationResponse, error)
		Update(ctx context.Context, req dto.RequirementUpdateRequest) (dto.RequirementMutationResponse, error)
		Delete(ctx context.Context, id string) (dto.RequirementMutationResponse, error)
	}
)

func NewService(requirementRepository Repository) *Service {
	return &Service{
		repo: requirementRepository,
	}
}

func (s *Service) ListRequirements(ctx context.Context, req dto.RequirementListRequest) ([]dto.Requirement, error) {
	taskID := strings.TrimSpace(req.TaskID)
	if taskID == "" {
		return nil, ErrInvalidInput
	}

	return s.repo.List(ctx, taskID)
}

func (s *Service) CreateRequirement(ctx context.Context, req dto.RequirementCreateRequest) (dto.RequirementMutationResponse, error) {
	req.TaskID = strings.TrimSpace(req.TaskID)
	req.Definition = strings.TrimSpace(req.Definition)
	if req.TaskID == "" || req.Definition == "" {
		return dto.RequirementMutationResponse{}, ErrInvalidInput
	}

	return s.repo.Create(ctx, req)
}

func (s *Service) UpdateRequirement(ctx context.Context, req dto.RequirementUpdateRequest) (dto.RequirementMutationResponse, error) {
	req.ID = strings.TrimSpace(req.ID)
	req.Definition = strings.TrimSpace(req.Definition)
	if req.ID == "" || req.Definition == "" {
		return dto.RequirementMutationResponse{}, ErrInvalidInput
	}

	return s.repo.Update(ctx, req)
}

func (s *Service) DeleteRequirement(ctx context.Context, req dto.RequirementIDRequest) (dto.RequirementMutationResponse, error) {
	id := strings.TrimSpace(req.ID)
	if id == "" {
		return dto.RequirementMutationResponse{}, ErrInvalidInput
	}

	return s.repo.Delete(ctx, id)
}
