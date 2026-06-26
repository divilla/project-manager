package requirement

import (
	"context"
	"testing"

	"aipm/internal/changeview"
	"aipm/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceRejectsInvalidRequirementInput(t *testing.T) {
	service := &Service{}
	_, err := service.ListRequirements(context.Background(), dto.RequirementListRequest{})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.CreateRequirement(context.Background(), dto.RequirementCreateRequest{ChangeID: 2, Definition: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdateRequirement(context.Background(), dto.RequirementUpdateRequest{ID: 3, Definition: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdateRequirementChange(context.Background(), dto.RequirementUpdateChangeRequest{ID: 3})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.DeleteRequirement(context.Background(), dto.RequirementIDRequest{})
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestServiceNormalizesRequirementRequests(t *testing.T) {
	repo := &fakeRequirementRepository{}
	service := NewService(repo, changeview.NewChangeRenderer(fakeMarkdownParser{}, fakeMarkdownSanitizer{}))

	_, err := service.ListRequirements(context.Background(), dto.RequirementListRequest{ChangeID: 2})
	require.NoError(t, err)
	assert.Equal(t, 2, repo.changeID)
	_, err = service.CreateRequirement(context.Background(), dto.RequirementCreateRequest{ChangeID: 2, Definition: " Add API test "})
	require.NoError(t, err)
	assert.Equal(t, "Add API test", repo.createReq.Definition)
	_, err = service.UpdateRequirement(context.Background(), dto.RequirementUpdateRequest{
		ID: 3, Definition: " Mark test green ",
	})
	require.NoError(t, err)
	assert.Equal(t, "Mark test green", repo.updateReq.Definition)
	_, err = service.UpdateRequirementChange(context.Background(), dto.RequirementUpdateChangeRequest{ID: 3, ChangeID: 4})
	require.NoError(t, err)
	assert.Equal(t, 4, repo.changeID)
	_, err = service.DeleteRequirement(context.Background(), dto.RequirementIDRequest{ID: 3})
	require.NoError(t, err)
	assert.Equal(t, 3, repo.id)
}

func TestServiceRendersMutationChangeBodyHTML(t *testing.T) {
	repo := &fakeRequirementRepository{}
	service := NewService(repo, changeview.NewChangeRenderer(fakeMarkdownParser{}, fakeMarkdownSanitizer{}))

	mutation, err := service.CreateRequirement(context.Background(), dto.RequirementCreateRequest{
		ChangeID:   2,
		Definition: "Requirement",
	})
	require.NoError(t, err)
	assert.Equal(t, "clean(parsed(**Change**))", mutation.Change.BodyHTML)
}

type fakeMarkdownParser struct{}

func (fakeMarkdownParser) Parse(source string) string {
	return "parsed(" + source + ")"
}

type fakeMarkdownSanitizer struct{}

func (fakeMarkdownSanitizer) Parse(source string) string {
	return "clean(" + source + ")"
}

type fakeRequirementRepository struct {
	id        int
	changeID  int
	createReq dto.RequirementCreateRequest
	updateReq dto.RequirementUpdateRequest
}

func (r *fakeRequirementRepository) List(_ context.Context, changeID int) ([]dto.Requirement, error) {
	r.changeID = changeID
	return []dto.Requirement{}, nil
}

func (r *fakeRequirementRepository) Create(_ context.Context, req dto.RequirementCreateRequest) (dto.RequirementMutationResponse, error) {
	r.createReq = req
	requirement := dto.Requirement{ID: 3, ChangeID: req.ChangeID, Definition: req.Definition}
	return dto.RequirementMutationResponse{
		Requirement: &requirement,
		Change:      dto.Change{ID: req.ChangeID, Body: "**Change**"},
	}, nil
}

func (r *fakeRequirementRepository) Update(_ context.Context, req dto.RequirementUpdateRequest) (dto.RequirementMutationResponse, error) {
	r.updateReq = req
	return dto.RequirementMutationResponse{}, nil
}

func (r *fakeRequirementRepository) UpdateDone(_ context.Context, req dto.RequirementUpdateDoneRequest) (dto.RequirementMutationResponse, error) {
	r.id = req.ID
	return dto.RequirementMutationResponse{}, nil
}

func (r *fakeRequirementRepository) UpdateChange(_ context.Context, req dto.RequirementUpdateChangeRequest) (dto.RequirementMutationResponse, error) {
	r.id, r.changeID = req.ID, req.ChangeID
	return dto.RequirementMutationResponse{}, nil
}

func (r *fakeRequirementRepository) Delete(_ context.Context, req dto.RequirementIDRequest) (dto.RequirementMutationResponse, error) {
	r.id = req.ID
	return dto.RequirementMutationResponse{}, nil
}
