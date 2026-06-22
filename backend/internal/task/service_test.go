package task

import (
	"context"
	"testing"

	"aipm/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceRejectsInvalidTaskInput(t *testing.T) {
	service := &Service{}

	_, err := service.ListTasks(context.Background(), dto.TaskListRequest{})
	require.ErrorIs(t, err, ErrInvalidInput)

	_, err = service.GetTask(context.Background(), dto.TaskIDRequest{})
	require.ErrorIs(t, err, ErrInvalidInput)

	_, err = service.CreateTask(context.Background(), dto.TaskCreateRequest{ProjectID: "project-id", Name: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)

	_, err = service.UpdateTask(context.Background(), dto.TaskUpdateRequest{ID: "task-id", Name: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)

	_, err = service.ChangePhase(context.Background(), dto.TaskPhaseRequest{ID: "task-id", Phase: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)

	err = service.DeleteTask(context.Background(), dto.TaskIDRequest{ID: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestServiceTrimsTaskRequests(t *testing.T) {
	repo := &fakeTaskRepository{}
	service := NewService(repo)

	_, err := service.ListTasks(context.Background(), dto.TaskListRequest{ProjectID: " project-id "})
	require.NoError(t, err)
	assert.Equal(t, "project-id", repo.projectID)

	_, err = service.GetTask(context.Background(), dto.TaskIDRequest{ID: " task-id "})
	require.NoError(t, err)
	assert.Equal(t, "task-id", repo.id)

	_, err = service.CreateTask(context.Background(), dto.TaskCreateRequest{
		ProjectID:   " project-id ",
		Name:        " Task Name ",
		Description: " Description ",
		Phase:       " backlog ",
		Type:        " task ",
		ParentID:    " parent-id ",
	})
	require.NoError(t, err)
	assert.Equal(t, dto.TaskCreateRequest{
		ProjectID:   "project-id",
		Name:        "Task Name",
		Description: "Description",
		Phase:       "backlog",
		Type:        "task",
		ParentID:    "parent-id",
	}, repo.createReq)

	_, err = service.UpdateTask(context.Background(), dto.TaskUpdateRequest{
		ID:          " task-id ",
		Name:        " Updated Task ",
		Description: " Updated Description ",
		Type:        " feature ",
	})
	require.NoError(t, err)
	assert.Equal(t, dto.TaskUpdateRequest{
		ID:          "task-id",
		Name:        "Updated Task",
		Description: "Updated Description",
		Type:        "feature",
	}, repo.updateReq)

	_, err = service.ChangePhase(context.Background(), dto.TaskPhaseRequest{ID: " task-id ", Phase: " review "})
	require.NoError(t, err)
	assert.Equal(t, "task-id", repo.id)
	assert.Equal(t, "review", repo.phase)

	err = service.DeleteTask(context.Background(), dto.TaskIDRequest{ID: " task-id "})
	require.NoError(t, err)
	assert.Equal(t, "task-id", repo.id)
}

type fakeTaskRepository struct {
	projectID string
	id        string
	phase     string
	createReq dto.TaskCreateRequest
	updateReq dto.TaskUpdateRequest
}

func (r *fakeTaskRepository) References(context.Context) (dto.TaskReferences, error) {
	return dto.TaskReferences{}, nil
}

func (r *fakeTaskRepository) List(_ context.Context, projectID string) ([]dto.Task, error) {
	r.projectID = projectID
	return []dto.Task{}, nil
}

func (r *fakeTaskRepository) Get(_ context.Context, id string) (dto.TaskDetail, error) {
	r.id = id
	return dto.TaskDetail{Task: dto.Task{ID: id}}, nil
}

func (r *fakeTaskRepository) Create(_ context.Context, req dto.TaskCreateRequest) (dto.Task, error) {
	r.createReq = req
	return dto.Task{ID: "task-id", ProjectID: req.ProjectID, Name: req.Name}, nil
}

func (r *fakeTaskRepository) Update(_ context.Context, req dto.TaskUpdateRequest) (dto.Task, error) {
	r.updateReq = req
	return dto.Task{ID: req.ID, Name: req.Name}, nil
}

func (r *fakeTaskRepository) ChangePhase(_ context.Context, id, phase string) (dto.Task, error) {
	r.id = id
	r.phase = phase
	return dto.Task{ID: id, Phase: phase}, nil
}

func (r *fakeTaskRepository) Delete(_ context.Context, id string) error {
	r.id = id
	return nil
}
