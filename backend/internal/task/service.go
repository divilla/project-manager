package task

import (
	"context"
	"errors"
	"strings"

	"aipm/internal/dto"
)

var (
	ErrInvalidInput     = errors.New("invalid task input")
	ErrInvalidReference = errors.New("invalid task reference")
	ErrNotFound         = errors.New("task not found")
)

type Service struct {
	repo Repository
}

func NewService(taskRepository Repository) *Service {
	return &Service{repo: taskRepository}
}

func (s *Service) References(ctx context.Context) (dto.TaskReferences, error) {
	return s.repo.References(ctx)
}

func (s *Service) ListTasks(ctx context.Context, req dto.TaskListRequest) ([]dto.Task, error) {
	if req.ProjectID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.List(ctx, req.ProjectID)
}

func (s *Service) GetTask(ctx context.Context, req dto.TaskIDRequest) (dto.TaskDetail, error) {
	if req.ID <= 0 {
		return dto.TaskDetail{}, ErrInvalidInput
	}
	return s.repo.Get(ctx, req.ID)
}

func (s *Service) CreateTask(ctx context.Context, req dto.TaskCreateRequest) (dto.Task, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.TaskPhase = strings.TrimSpace(req.TaskPhase)
	req.TaskType = strings.TrimSpace(req.TaskType)
	if req.ProjectID <= 0 || req.Name == "" || (req.ParentID != nil && *req.ParentID <= 0) {
		return dto.Task{}, ErrInvalidInput
	}
	return s.repo.Create(ctx, req)
}

func (s *Service) UpdateTask(ctx context.Context, req dto.TaskUpdateRequest) (dto.Task, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.TaskType = strings.TrimSpace(req.TaskType)
	if req.ID <= 0 || req.Name == "" {
		return dto.Task{}, ErrInvalidInput
	}
	return s.repo.Update(ctx, req)
}

func (s *Service) UpdateDifficulty(ctx context.Context, req dto.TaskUpdateDifficultyRequest) (dto.Task, error) {
	if req.ID <= 0 || req.Difficulty <= 0 {
		return dto.Task{}, ErrInvalidInput
	}
	return s.repo.UpdateDifficulty(ctx, req)
}

func (s *Service) UpdatePriority(ctx context.Context, req dto.TaskUpdatePriorityRequest) (dto.Task, error) {
	if req.ID <= 0 {
		return dto.Task{}, ErrInvalidInput
	}
	return s.repo.UpdatePriority(ctx, req)
}

func (s *Service) UpdateParent(ctx context.Context, req dto.TaskUpdateParentRequest) (dto.Task, error) {
	if req.ID <= 0 || (req.ParentID != nil && *req.ParentID <= 0) {
		return dto.Task{}, ErrInvalidInput
	}
	return s.repo.UpdateParent(ctx, req)
}

func (s *Service) UpdatePhase(ctx context.Context, req dto.TaskUpdatePhaseRequest) (dto.Task, error) {
	req.TaskPhase = strings.TrimSpace(req.TaskPhase)
	if req.ID <= 0 || req.TaskPhase == "" {
		return dto.Task{}, ErrInvalidInput
	}
	return s.repo.UpdatePhase(ctx, req)
}

func (s *Service) DeleteTask(ctx context.Context, req dto.TaskIDRequest) error {
	if req.ID <= 0 {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, req)
}
