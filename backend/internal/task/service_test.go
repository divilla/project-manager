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
	_, err = service.CreateTask(context.Background(), dto.TaskCreateRequest{ProjectID: 1, Name: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdateTask(context.Background(), dto.TaskUpdateRequest{ID: 2, Name: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdatePhase(context.Background(), dto.TaskUpdatePhaseRequest{ID: 2, TaskPhase: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
	err = service.DeleteTask(context.Background(), dto.TaskIDRequest{})
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestServiceNormalizesTaskRequests(t *testing.T) {
	repo := &fakeTaskRepository{}
	service := NewService(repo)
	parentID := 4

	_, err := service.ListTasks(context.Background(), dto.TaskListRequest{ProjectID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.projectID)
	_, err = service.GetTask(context.Background(), dto.TaskIDRequest{ID: 2})
	require.NoError(t, err)
	assert.Equal(t, 2, repo.id)

	_, err = service.CreateTask(context.Background(), dto.TaskCreateRequest{
		ProjectID: 1, Name: " Task Name ", Description: " Description ",
		TaskPhase: " backlog ", TaskType: " task ", ParentID: &parentID,
	})
	require.NoError(t, err)
	assert.Equal(t, "Task Name", repo.createReq.Name)
	assert.Equal(t, "Description", repo.createReq.Description)
	assert.Equal(t, "backlog", repo.createReq.TaskPhase)
	assert.Equal(t, "task", repo.createReq.TaskType)

	_, err = service.UpdateTask(context.Background(), dto.TaskUpdateRequest{
		ID: 2, Name: " Updated Task ", Description: " Updated Description ", TaskType: " feature ",
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Task", repo.updateReq.Name)
	assert.Equal(t, "Updated Description", repo.updateReq.Description)
	assert.Equal(t, "feature", repo.updateReq.TaskType)

	_, err = service.UpdatePhase(context.Background(), dto.TaskUpdatePhaseRequest{ID: 2, TaskPhase: " review "})
	require.NoError(t, err)
	assert.Equal(t, "review", repo.taskPhase)
	err = service.DeleteTask(context.Background(), dto.TaskIDRequest{ID: 2})
	require.NoError(t, err)
	assert.Equal(t, 2, repo.id)
}

type fakeTaskRepository struct {
	projectID int
	id        int
	taskPhase string
	createReq dto.TaskCreateRequest
	updateReq dto.TaskUpdateRequest
}

func (r *fakeTaskRepository) References(context.Context) (dto.TaskReferences, error) {
	return dto.TaskReferences{}, nil
}
func (r *fakeTaskRepository) List(_ context.Context, projectID int) ([]dto.Task, error) {
	r.projectID = projectID
	return []dto.Task{}, nil
}
func (r *fakeTaskRepository) Get(_ context.Context, id int) (dto.TaskDetail, error) {
	r.id = id
	return dto.TaskDetail{Task: dto.Task{ID: id}}, nil
}
func (r *fakeTaskRepository) Create(_ context.Context, req dto.TaskCreateRequest) (dto.Task, error) {
	r.createReq = req
	return dto.Task{ID: 2, ProjectID: req.ProjectID, Name: req.Name}, nil
}
func (r *fakeTaskRepository) Update(_ context.Context, req dto.TaskUpdateRequest) (dto.Task, error) {
	r.updateReq = req
	return dto.Task{ID: req.ID, Name: req.Name}, nil
}
func (r *fakeTaskRepository) UpdateDifficulty(_ context.Context, req dto.TaskUpdateDifficultyRequest) (dto.Task, error) {
	r.id = req.ID
	return dto.Task{ID: req.ID, Difficulty: req.Difficulty}, nil
}
func (r *fakeTaskRepository) UpdatePriority(_ context.Context, req dto.TaskUpdatePriorityRequest) (dto.Task, error) {
	r.id = req.ID
	return dto.Task{ID: req.ID, Priority: req.Priority}, nil
}
func (r *fakeTaskRepository) UpdateParent(_ context.Context, req dto.TaskUpdateParentRequest) (dto.Task, error) {
	r.id = req.ID
	return dto.Task{ID: req.ID, ParentID: req.ParentID}, nil
}
func (r *fakeTaskRepository) UpdatePhase(_ context.Context, req dto.TaskUpdatePhaseRequest) (dto.Task, error) {
	r.id, r.taskPhase = req.ID, req.TaskPhase
	return dto.Task{ID: req.ID, TaskPhase: req.TaskPhase}, nil
}
func (r *fakeTaskRepository) Delete(_ context.Context, req dto.TaskIDRequest) error {
	r.id = req.ID
	return nil
}
