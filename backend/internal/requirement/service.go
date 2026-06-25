package requirement

import (
	"context"
	"errors"
	"strings"

	"aipm/internal/dto"
	"aipm/internal/taskview"
)

var (
	ErrInvalidInput = errors.New("invalid requirement input")
	ErrNotFound     = errors.New("requirement not found")
)

type (
	Service struct {
		repo     Repository
		renderer taskview.TaskRenderer
	}

	Repository interface {
		List(ctx context.Context, taskID int) ([]dto.Requirement, error)
		Create(ctx context.Context, req dto.RequirementCreateRequest) (dto.RequirementMutationResponse, error)
		Update(ctx context.Context, req dto.RequirementUpdateRequest) (dto.RequirementMutationResponse, error)
		UpdateDone(ctx context.Context, req dto.RequirementUpdateDoneRequest) (dto.RequirementMutationResponse, error)
		UpdateTask(ctx context.Context, req dto.RequirementUpdateChangeRequest) (dto.RequirementMutationResponse, error)
		Delete(ctx context.Context, req dto.RequirementIDRequest) (dto.RequirementMutationResponse, error)
	}
)

func NewService(requirementRepository Repository, renderer taskview.TaskRenderer) *Service {
	return &Service{repo: requirementRepository, renderer: renderer}
}

func (s *Service) ListRequirements(ctx context.Context, req dto.RequirementListRequest) ([]dto.Requirement, error) {
	if req.ChangeID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.List(ctx, req.ChangeID)
}

func (s *Service) CreateRequirement(ctx context.Context, req dto.RequirementCreateRequest) (dto.RequirementMutationResponse, error) {
	req.Definition = strings.TrimSpace(req.Definition)
	if req.ChangeID <= 0 || req.Definition == "" {
		return dto.RequirementMutationResponse{}, ErrInvalidInput
	}
	mutation, err := s.repo.Create(ctx, req)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	return s.renderMutation(mutation), nil
}

func (s *Service) UpdateRequirement(ctx context.Context, req dto.RequirementUpdateRequest) (dto.RequirementMutationResponse, error) {
	req.Definition = strings.TrimSpace(req.Definition)
	if req.ID <= 0 || req.Definition == "" {
		return dto.RequirementMutationResponse{}, ErrInvalidInput
	}
	mutation, err := s.repo.Update(ctx, req)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	return s.renderMutation(mutation), nil
}

func (s *Service) UpdateRequirementDone(ctx context.Context, req dto.RequirementUpdateDoneRequest) (dto.RequirementMutationResponse, error) {
	if req.ID <= 0 {
		return dto.RequirementMutationResponse{}, ErrInvalidInput
	}
	mutation, err := s.repo.UpdateDone(ctx, req)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	return s.renderMutation(mutation), nil
}

func (s *Service) UpdateRequirementTask(ctx context.Context, req dto.RequirementUpdateChangeRequest) (dto.RequirementMutationResponse, error) {
	if req.ID <= 0 || req.TaskID <= 0 {
		return dto.RequirementMutationResponse{}, ErrInvalidInput
	}
	mutation, err := s.repo.UpdateTask(ctx, req)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	return s.renderMutation(mutation), nil
}

func (s *Service) DeleteRequirement(ctx context.Context, req dto.RequirementIDRequest) (dto.RequirementMutationResponse, error) {
	if req.ID <= 0 {
		return dto.RequirementMutationResponse{}, ErrInvalidInput
	}
	mutation, err := s.repo.Delete(ctx, req)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	return s.renderMutation(mutation), nil
}

func (s *Service) renderMutation(mutation dto.RequirementMutationResponse) dto.RequirementMutationResponse {
	return s.renderer.RenderMutation(mutation)
}
