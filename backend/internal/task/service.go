package task

import (
	"context"
	"errors"
	"strings"

	"aipm/internal/dto"
	"aipm/internal/taskview"
)

var (
	ErrInvalidInput     = errors.New("invalid task input")
	ErrInvalidReference = errors.New("invalid task reference")
	ErrNotFound         = errors.New("task not found")
)

type Service struct {
	repo     Repository
	renderer taskview.TaskRenderer
}

func NewService(taskRepository Repository, renderer taskview.TaskRenderer) *Service {
	return &Service{repo: taskRepository, renderer: renderer}
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
	detail, err := s.repo.Get(ctx, req.ID)
	if err != nil {
		return dto.TaskDetail{}, err
	}
	detail.Task = s.renderer.RenderTask(detail.Task)
	return detail, nil
}

func (s *Service) RenderedDescriptions(ctx context.Context, req dto.TaskRenderedDescriptionsRequest) (dto.TaskRenderedDescriptionsResponse, error) {
	ids, err := normalizeTaskIDs(req.IDs)
	if err != nil {
		return dto.TaskRenderedDescriptionsResponse{}, err
	}
	if len(ids) == 0 {
		return dto.TaskRenderedDescriptionsResponse{Descriptions: []dto.TaskRenderedDescription{}}, nil
	}

	tasks, err := s.repo.Descriptions(ctx, ids)
	if err != nil {
		return dto.TaskRenderedDescriptionsResponse{}, err
	}

	descriptions := make([]dto.TaskRenderedDescription, 0, len(tasks))
	for _, task := range tasks {
		task = s.renderer.RenderTask(task)
		descriptions = append(descriptions, dto.TaskRenderedDescription{
			ID:              task.ID,
			DescriptionHTML: task.DescriptionHTML,
		})
	}
	return dto.TaskRenderedDescriptionsResponse{Descriptions: descriptions}, nil
}

func (s *Service) CreateTask(ctx context.Context, req dto.TaskCreateRequest) (dto.Task, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.TaskPhase = strings.TrimSpace(req.TaskPhase)
	req.TaskType = strings.TrimSpace(req.TaskType)
	if req.ProjectID <= 0 || req.Name == "" || (req.ParentID != nil && *req.ParentID <= 0) {
		return dto.Task{}, ErrInvalidInput
	}
	task, err := s.repo.Create(ctx, req)
	if err != nil {
		return dto.Task{}, err
	}
	return s.renderer.RenderTask(task), nil
}

func (s *Service) UpdateTask(ctx context.Context, req dto.TaskUpdateRequest) (dto.Task, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.TaskType = strings.TrimSpace(req.TaskType)
	if req.ID <= 0 || req.Name == "" {
		return dto.Task{}, ErrInvalidInput
	}
	task, err := s.repo.Update(ctx, req)
	if err != nil {
		return dto.Task{}, err
	}
	return s.renderer.RenderTask(task), nil
}

func (s *Service) UpdateDifficulty(ctx context.Context, req dto.TaskUpdateDifficultyRequest) (dto.Task, error) {
	if req.ID <= 0 || req.Difficulty <= 0 {
		return dto.Task{}, ErrInvalidInput
	}
	task, err := s.repo.UpdateDifficulty(ctx, req)
	if err != nil {
		return dto.Task{}, err
	}
	return s.renderer.RenderTask(task), nil
}

func (s *Service) UpdatePriority(ctx context.Context, req dto.TaskUpdatePriorityRequest) (dto.Task, error) {
	if req.ID <= 0 {
		return dto.Task{}, ErrInvalidInput
	}
	task, err := s.repo.UpdatePriority(ctx, req)
	if err != nil {
		return dto.Task{}, err
	}
	return s.renderer.RenderTask(task), nil
}

func (s *Service) UpdateParent(ctx context.Context, req dto.TaskUpdateParentRequest) (dto.Task, error) {
	if req.ID <= 0 || (req.ParentID != nil && *req.ParentID <= 0) {
		return dto.Task{}, ErrInvalidInput
	}
	task, err := s.repo.UpdateParent(ctx, req)
	if err != nil {
		return dto.Task{}, err
	}
	return s.renderer.RenderTask(task), nil
}

func (s *Service) UpdatePhase(ctx context.Context, req dto.TaskUpdatePhaseRequest) (dto.Task, error) {
	req.TaskPhase = strings.TrimSpace(req.TaskPhase)
	if req.ID <= 0 || req.TaskPhase == "" {
		return dto.Task{}, ErrInvalidInput
	}
	task, err := s.repo.UpdatePhase(ctx, req)
	if err != nil {
		return dto.Task{}, err
	}
	return s.renderer.RenderTask(task), nil
}

func (s *Service) DeleteTask(ctx context.Context, req dto.TaskIDRequest) error {
	if req.ID <= 0 {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, req)
}

func normalizeTaskIDs(ids []int) ([]int, error) {
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
