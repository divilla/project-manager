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

func (s *Service) References(ctx context.Context) (dto.ChangeReferences, error) {
	return s.repo.References(ctx)
}

func (s *Service) ListTasks(ctx context.Context, req dto.ChangeListRequest) ([]dto.Change, error) {
	if req.ProjectID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.List(ctx, req.ProjectID)
}

func (s *Service) GetTask(ctx context.Context, req dto.ChangeIDRequest) (dto.ChangeDetail, error) {
	if req.ID <= 0 {
		return dto.ChangeDetail{}, ErrInvalidInput
	}
	detail, err := s.repo.Get(ctx, req.ID)
	if err != nil {
		return dto.ChangeDetail{}, err
	}
	detail.Change = s.renderer.RenderTask(detail.Change)
	return detail, nil
}

func (s *Service) RenderedDescriptions(ctx context.Context, req dto.ChangeRenderedDescriptionsRequest) (dto.ChangeRenderedDescriptionsResponse, error) {
	ids, err := normalizeTaskIDs(req.IDs)
	if err != nil {
		return dto.ChangeRenderedDescriptionsResponse{}, err
	}
	if len(ids) == 0 {
		return dto.ChangeRenderedDescriptionsResponse{Descriptions: []dto.ChangeRenderedDescription{}}, nil
	}

	tasks, err := s.repo.Descriptions(ctx, ids)
	if err != nil {
		return dto.ChangeRenderedDescriptionsResponse{}, err
	}

	descriptions := make([]dto.ChangeRenderedDescription, 0, len(tasks))
	for _, task := range tasks {
		task = s.renderer.RenderTask(task)
		descriptions = append(descriptions, dto.ChangeRenderedDescription{
			ID:              task.ID,
			DescriptionHTML: task.BodyHTML,
		})
	}
	return dto.ChangeRenderedDescriptionsResponse{Descriptions: descriptions}, nil
}

func (s *Service) CreateTask(ctx context.Context, req dto.ChangeCreateRequest) (dto.Change, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	req.ChangePhase = strings.TrimSpace(req.ChangePhase)
	req.ChangeTypes = strings.TrimSpace(req.ChangeTypes)
	if req.ProjectID <= 0 || req.Title == "" || (req.ParentID != nil && *req.ParentID <= 0) {
		return dto.Change{}, ErrInvalidInput
	}
	task, err := s.repo.Create(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderTask(task), nil
}

func (s *Service) UpdateTask(ctx context.Context, req dto.ChangeUpdateRequest) (dto.Change, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.ChangeTypes = strings.TrimSpace(req.ChangeTypes)
	if req.ID <= 0 || req.Name == "" {
		return dto.Change{}, ErrInvalidInput
	}
	task, err := s.repo.Update(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderTask(task), nil
}

func (s *Service) UpdateDifficulty(ctx context.Context, req dto.TaskUpdateDifficultyRequest) (dto.Change, error) {
	if req.ID <= 0 || req.Difficulty <= 0 {
		return dto.Change{}, ErrInvalidInput
	}
	task, err := s.repo.UpdateDifficulty(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderTask(task), nil
}

func (s *Service) UpdatePriority(ctx context.Context, req dto.TaskUpdatePriorityRequest) (dto.Change, error) {
	if req.ID <= 0 {
		return dto.Change{}, ErrInvalidInput
	}
	task, err := s.repo.UpdatePriority(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderTask(task), nil
}

func (s *Service) UpdateParent(ctx context.Context, req dto.ChangeUpdateEpicRequest) (dto.Change, error) {
	if req.ID <= 0 || (req.EpicID != nil && *req.EpicID <= 0) {
		return dto.Change{}, ErrInvalidInput
	}
	task, err := s.repo.UpdateParent(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderTask(task), nil
}

func (s *Service) UpdatePhase(ctx context.Context, req dto.ChangeUpdatePhaseRequest) (dto.Change, error) {
	req.ChangePhase = strings.TrimSpace(req.ChangePhase)
	if req.ID <= 0 || req.ChangePhase == "" {
		return dto.Change{}, ErrInvalidInput
	}
	task, err := s.repo.UpdatePhase(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderTask(task), nil
}

func (s *Service) DeleteTask(ctx context.Context, req dto.ChangeIDRequest) error {
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
