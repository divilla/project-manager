package requirement

import (
	"context"
	"testing"

	"aipm/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceRejectsInvalidRequirementInput(t *testing.T) {
	service := &Service{}

	_, err := service.ListRequirements(context.Background(), dto.RequirementListRequest{})
	require.ErrorIs(t, err, ErrInvalidInput)

	_, err = service.CreateRequirement(context.Background(), dto.RequirementCreateRequest{TaskID: "task-id", Definition: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)

	_, err = service.CreateRequirement(context.Background(), dto.RequirementCreateRequest{TaskID: "   ", Definition: "Add test"})
	require.ErrorIs(t, err, ErrInvalidInput)

	_, err = service.UpdateRequirement(context.Background(), dto.RequirementUpdateRequest{ID: "requirement-id", Definition: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)

	_, err = service.DeleteRequirement(context.Background(), dto.RequirementIDRequest{ID: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestServiceTrimsRequirementRequests(t *testing.T) {
	repo := &fakeRequirementRepository{}
	service := NewService(repo)
	done := true

	_, err := service.ListRequirements(context.Background(), dto.RequirementListRequest{TaskID: " task-id "})
	require.NoError(t, err)
	assert.Equal(t, "task-id", repo.taskID)

	_, err = service.CreateRequirement(context.Background(), dto.RequirementCreateRequest{
		TaskID:     " task-id ",
		Definition: " Add API test ",
	})
	require.NoError(t, err)
	assert.Equal(t, dto.RequirementCreateRequest{
		TaskID:     "task-id",
		Definition: "Add API test",
	}, repo.createReq)

	_, err = service.UpdateRequirement(context.Background(), dto.RequirementUpdateRequest{
		ID:         " requirement-id ",
		Definition: " Mark test green ",
		Done:       &done,
	})
	require.NoError(t, err)
	assert.Equal(t, dto.RequirementUpdateRequest{
		ID:         "requirement-id",
		Definition: "Mark test green",
		Done:       &done,
	}, repo.updateReq)

	_, err = service.DeleteRequirement(context.Background(), dto.RequirementIDRequest{ID: " requirement-id "})
	require.NoError(t, err)
	assert.Equal(t, "requirement-id", repo.id)
}

type fakeRequirementRepository struct {
	id        string
	taskID    string
	createReq dto.RequirementCreateRequest
	updateReq dto.RequirementUpdateRequest
}

func (r *fakeRequirementRepository) List(_ context.Context, taskID string) ([]dto.Requirement, error) {
	r.taskID = taskID
	return []dto.Requirement{}, nil
}

func (r *fakeRequirementRepository) Create(_ context.Context, req dto.RequirementCreateRequest) (dto.RequirementMutationResponse, error) {
	r.createReq = req
	requirement := dto.Requirement{ID: "requirement-id", TaskID: req.TaskID, Definition: req.Definition}
	return dto.RequirementMutationResponse{Requirement: &requirement}, nil
}

func (r *fakeRequirementRepository) Update(_ context.Context, req dto.RequirementUpdateRequest) (dto.RequirementMutationResponse, error) {
	r.updateReq = req
	done := false
	if req.Done != nil {
		done = *req.Done
	}
	requirement := dto.Requirement{ID: req.ID, Definition: req.Definition, Done: done}
	return dto.RequirementMutationResponse{Requirement: &requirement}, nil
}

func (r *fakeRequirementRepository) Delete(_ context.Context, id string) (dto.RequirementMutationResponse, error) {
	r.id = id
	return dto.RequirementMutationResponse{}, nil
}
