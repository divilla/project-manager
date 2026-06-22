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
		List(ctx context.Context, taskID int) ([]dto.Requirement, error)
		Create(ctx context.Context, req dto.RequirementCreateRequest) (dto.RequirementMutationResponse, error)
		Update(ctx context.Context, req dto.RequirementUpdateRequest) (dto.RequirementMutationResponse, error)
		UpdateDone(ctx context.Context, req dto.RequirementUpdateDoneRequest) (dto.RequirementMutationResponse, error)
		UpdateTask(ctx context.Context, req dto.RequirementUpdateTaskRequest) (dto.RequirementMutationResponse, error)
		Delete(ctx context.Context, req dto.RequirementIDRequest) (dto.RequirementMutationResponse, error)
	}
)

func NewService(requirementRepository Repository) *Service {
	return &Service{repo: requirementRepository}
}

func (s *Service) ListRequirements(ctx context.Context, req dto.RequirementListRequest) ([]dto.Requirement, error) {
	if req.TaskID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.List(ctx, req.TaskID)
}

func (s *Service) CreateRequirement(ctx context.Context, req dto.RequirementCreateRequest) (dto.RequirementMutationResponse, error) {
	req.Definition = strings.TrimSpace(req.Definition)
	if req.TaskID <= 0 || req.Definition == "" {
		return dto.RequirementMutationResponse{}, ErrInvalidInput
	}
	return s.repo.Create(ctx, req)
}

func (s *Service) UpdateRequirement(ctx context.Context, req dto.RequirementUpdateRequest) (dto.RequirementMutationResponse, error) {
	req.Definition = strings.TrimSpace(req.Definition)
	if req.ID <= 0 || req.Definition == "" {
		return dto.RequirementMutationResponse{}, ErrInvalidInput
	}
	return s.repo.Update(ctx, req)
}

func (s *Service) UpdateRequirementDone(ctx context.Context, req dto.RequirementUpdateDoneRequest) (dto.RequirementMutationResponse, error) {
	if req.ID <= 0 {
		return dto.RequirementMutationResponse{}, ErrInvalidInput
	}
	return s.repo.UpdateDone(ctx, req)
}

func (s *Service) UpdateRequirementTask(ctx context.Context, req dto.RequirementUpdateTaskRequest) (dto.RequirementMutationResponse, error) {
	if req.ID <= 0 || req.TaskID <= 0 {
		return dto.RequirementMutationResponse{}, ErrInvalidInput
	}
	return s.repo.UpdateTask(ctx, req)
}

func (s *Service) DeleteRequirement(ctx context.Context, req dto.RequirementIDRequest) (dto.RequirementMutationResponse, error) {
	if req.ID <= 0 {
		return dto.RequirementMutationResponse{}, ErrInvalidInput
	}
	return s.repo.Delete(ctx, req)
}
