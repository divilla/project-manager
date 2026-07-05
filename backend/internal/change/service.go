package change

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"aipm/internal/dto"
)

var (
	// ErrInvalidInput is a package-level value.
	ErrInvalidInput = errors.New("invalid change input")
	// ErrInvalidReference is returned when a change reference is invalid.
	ErrInvalidReference = errors.New("invalid change reference")
	// ErrNotFound is returned when a change cannot be found.
	ErrNotFound = errors.New("change not found")
)

// Service defines Service values.
type Service struct {
	repo     Repository
	renderer Renderer
}

// NewService initializes or executes NewService behavior.
func NewService(changeRepository Repository, renderer Renderer) *Service {
	return &Service{repo: changeRepository, renderer: renderer}
}

// ListChanges executes ListChanges behavior.
func (s *Service) ListChanges(ctx context.Context, req dto.ChangeListRequest) ([]dto.ChangeListItem, error) {
	if req.ProjectID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.List(ctx, req.ProjectID)
}

// GetChange executes GetChange behavior.
func (s *Service) GetChange(ctx context.Context, req dto.ChangeIDRequest) (dto.ChangeDetail, error) {
	if req.ID <= 0 {
		return dto.ChangeDetail{}, ErrInvalidInput
	}
	detail, err := s.repo.Get(ctx, req.ID)
	if err != nil {
		return dto.ChangeDetail{}, err
	}
	detail.Change = s.renderer.RenderChange(detail.Change)
	return detail, nil
}

// RenderedArtifacts executes RenderedArtifacts behavior.
func (s *Service) RenderedArtifacts(ctx context.Context, req dto.ChangeRenderedArtifactsRequest) (dto.ChangeRenderedArtifactsResponse, error) {
	ids, err := normalizeIDs(req.IDs)
	if err != nil {
		return dto.ChangeRenderedArtifactsResponse{}, err
	}
	if len(ids) == 0 {
		return dto.ChangeRenderedArtifactsResponse{Artifacts: []dto.ChangeRenderedArtifact{}}, nil
	}
	changes, err := s.repo.Artifacts(ctx, ids)
	if err != nil {
		return dto.ChangeRenderedArtifactsResponse{}, err
	}
	artifacts := make([]dto.ChangeRenderedArtifact, 0, len(changes))
	for _, item := range changes {
		item = s.renderer.RenderChange(item)
		specHTML := ""
		if item.SpecHTML != nil {
			specHTML = *item.SpecHTML
		}
		prHTML := ""
		if item.PRHtml != nil {
			prHTML = *item.PRHtml
		}
		artifacts = append(artifacts, dto.ChangeRenderedArtifact{
			ID:       item.ID,
			SpecHTML: specHTML,
			PRHtml:   prHTML,
		})
	}
	return dto.ChangeRenderedArtifactsResponse{Artifacts: artifacts}, nil
}

// CreateChange executes CreateChange behavior.
func (s *Service) CreateChange(ctx context.Context, req dto.ChangeCreateRequest) (dto.Change, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Idea = strings.TrimSpace(req.Idea)
	if req.ProjectID <= 0 || req.Title == "" || req.Idea == "" {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.Create(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdateChangeTypes executes UpdateChangeTypes behavior.
func (s *Service) UpdateChangeTypes(ctx context.Context, req dto.ChangeUpdateChangeTypesRequest) (dto.Change, error) {
	req.ChangeTypes = normalizeTypes(req.ChangeTypes)
	if req.ID <= 0 {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdateChangeTypes(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdateTitle executes UpdateTitle behavior.
func (s *Service) UpdateTitle(ctx context.Context, req dto.ChangeUpdateTitleRequest) (dto.Change, error) {
	req.Title = strings.TrimSpace(req.Title)
	if req.ID <= 0 || req.Title == "" {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdateTitle(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdateIdea executes UpdateIdea behavior.
func (s *Service) UpdateIdea(ctx context.Context, req dto.ChangeUpdateIdeaRequest) (dto.Change, error) {
	req.Idea = strings.TrimSpace(req.Idea)
	if req.ID <= 0 || req.Idea == "" {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdateIdea(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdateIdeaAgentEdit executes UpdateIdeaAgentEdit behavior.
func (s *Service) UpdateIdeaAgentEdit(ctx context.Context, req dto.ChangeUpdateIdeaAgentEditRequest) (dto.Change, error) {
	req.Idea = strings.TrimSpace(req.Idea)
	if req.ID <= 0 || req.Idea == "" {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdateIdeaAgentEdit(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdateSpec executes UpdateSpec behavior.
func (s *Service) UpdateSpec(ctx context.Context, req dto.ChangeUpdateSpecRequest) (dto.Change, error) {
	if req.Spec != nil {
		trimmed := strings.TrimSpace(*req.Spec)
		req.Spec = &trimmed
	}
	if req.ID <= 0 {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdateSpec(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdatePRBody executes UpdatePRBody behavior.
func (s *Service) UpdatePRBody(ctx context.Context, req dto.ChangeUpdatePRBodyRequest) (dto.Change, error) {
	if req.PRBody != nil {
		trimmed := strings.TrimSpace(*req.PRBody)
		req.PRBody = &trimmed
	}
	if req.ID <= 0 {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdatePRBody(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdatePRUrl executes UpdatePRUrl behavior.
func (s *Service) UpdatePRUrl(ctx context.Context, req dto.ChangeUpdatePRUrlRequest) (dto.Change, error) {
	if req.PRUrl != nil {
		trimmed := strings.TrimSpace(*req.PRUrl)
		req.PRUrl = &trimmed
	}
	if req.ID <= 0 || invalidPRURL(req.PRUrl) {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdatePRUrl(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdateAgentEdit executes UpdateAgentEdit behavior.
func (s *Service) UpdateAgentEdit(ctx context.Context, req dto.ChangeUpdateAgentEditRequest) (dto.Change, error) {
	if req.ID <= 0 || req.AgentEdit == nil {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdateAgentEdit(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdateEpic executes UpdateEpic behavior.
func (s *Service) UpdateEpic(ctx context.Context, req dto.ChangeUpdateEpicRequest) (dto.Change, error) {
	if req.ID <= 0 || invalidOptionalID(req.EpicID) {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdateEpic(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdatePhase executes UpdatePhase behavior.
func (s *Service) UpdatePhase(ctx context.Context, req dto.ChangeUpdatePhaseRequest) (dto.Change, error) {
	req.ChangePhase = strings.TrimSpace(req.ChangePhase)
	if req.ID <= 0 || req.ChangePhase == "" {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdatePhase(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// UpdateOpen executes UpdateOpen behavior.
func (s *Service) UpdateOpen(ctx context.Context, req dto.ChangeUpdateOpenRequest) (dto.Change, error) {
	if req.ID <= 0 || req.Open == nil {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.UpdateOpen(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// AssignFlow executes AssignFlow behavior.
func (s *Service) AssignFlow(ctx context.Context, req dto.ChangeIDRequest) (dto.Change, error) {
	if req.ID <= 0 {
		return dto.Change{}, ErrInvalidInput
	}
	change, err := s.repo.AssignFlow(ctx, req)
	if err != nil {
		return dto.Change{}, err
	}
	return s.renderer.RenderChange(change), nil
}

// StartRun executes StartRun behavior.
func (s *Service) StartRun(ctx context.Context, req dto.ChangeIDRequest) (dto.ChangeRunClaimResponse, error) {
	if req.ID <= 0 {
		return dto.ChangeRunClaimResponse{}, ErrInvalidInput
	}
	return s.repo.StartRun(ctx, req)
}

// UpdateRun executes UpdateRun behavior.
func (s *Service) UpdateRun(ctx context.Context, req dto.ChangeUpdateRunRequest) (dto.ChangeRunUpdateResponse, error) {
	req.RunClaimID = strings.TrimSpace(req.RunClaimID)
	req.RunFlowStage = strings.TrimSpace(req.RunFlowStage)
	req.RunTaskStep = strings.TrimSpace(req.RunTaskStep)
	req.RunTaskStatus = strings.TrimSpace(req.RunTaskStatus)
	req.RunError = strings.TrimSpace(req.RunError)
	if req.ID <= 0 || req.RunClaimID == "" {
		return dto.ChangeRunUpdateResponse{}, ErrInvalidInput
	}
	return s.repo.UpdateRun(ctx, req)
}

// ResetClaim executes ResetClaim behavior.
func (s *Service) ResetClaim(ctx context.Context, req dto.ChangeIDRequest) (dto.ChangeRunClaimResponse, error) {
	if req.ID <= 0 {
		return dto.ChangeRunClaimResponse{}, ErrInvalidInput
	}
	return s.repo.ResetClaim(ctx, req)
}

// DeleteChange executes DeleteChange behavior.
func (s *Service) DeleteChange(ctx context.Context, req dto.ChangeIDRequest) error {
	if req.ID <= 0 {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, req)
}

func normalizeIDs(ids []int) ([]int, error) {
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

func normalizeTypes(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func invalidOptionalID(value *int) bool {
	return value != nil && *value <= 0
}

func invalidPRURL(value *string) bool {
	if value == nil || *value == "" {
		return false
	}
	parsed, err := url.Parse(*value)
	if err != nil {
		return true
	}
	if parsed.Host == "" {
		return true
	}
	return !strings.EqualFold(parsed.Scheme, "https") && !strings.EqualFold(parsed.Scheme, "http")
}
