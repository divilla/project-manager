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
	ErrHasChildren      = errors.New("task has children")
)

type Service struct {
	repo Repository
}

func NewService(taskRepository Repository) *Service {
	return &Service{
		repo: taskRepository,
	}
}

func (s *Service) References(ctx context.Context) (dto.TaskReferences, error) {
	return s.repo.References(ctx)
}

func (s *Service) ListTasks(ctx context.Context, req dto.TaskListRequest) ([]dto.Task, error) {
	projectID := strings.TrimSpace(req.ProjectID)
	if projectID == "" {
		return nil, ErrInvalidInput
	}
	return s.repo.List(ctx, projectID)
}

func (s *Service) GetTask(ctx context.Context, req dto.TaskIDRequest) (dto.TaskDetail, error) {
	id := strings.TrimSpace(req.ID)
	if id == "" {
		return dto.TaskDetail{}, ErrInvalidInput
	}
	return s.repo.Get(ctx, id)
}

func (s *Service) CreateTask(ctx context.Context, req dto.TaskCreateRequest) (dto.Task, error) {
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.Phase = strings.TrimSpace(req.Phase)
	req.Type = strings.TrimSpace(req.Type)
	req.ParentID = strings.TrimSpace(req.ParentID)

	if req.ProjectID == "" || req.Name == "" {
		return dto.Task{}, ErrInvalidInput
	}

	return s.repo.Create(ctx, req)
}

func (s *Service) UpdateTask(ctx context.Context, req dto.TaskUpdateRequest) (dto.Task, error) {
	req.ID = strings.TrimSpace(req.ID)
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.Type = strings.TrimSpace(req.Type)

	if req.ID == "" || req.Name == "" {
		return dto.Task{}, ErrInvalidInput
	}

	return s.repo.Update(ctx, req)
}

func (s *Service) ChangePhase(ctx context.Context, req dto.TaskPhaseRequest) (dto.Task, error) {
	req.ID = strings.TrimSpace(req.ID)
	req.Phase = strings.TrimSpace(req.Phase)

	if req.ID == "" || req.Phase == "" {
		return dto.Task{}, ErrInvalidInput
	}

	return s.repo.ChangePhase(ctx, req.ID, req.Phase)
}

func (s *Service) DeleteTask(ctx context.Context, req dto.TaskIDRequest) error {
	req.ID = strings.TrimSpace(req.ID)
	if req.ID == "" {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, req.ID)
}
