package task

import (
	"context"
	"strconv"
	"testing"

	"aipm/internal/dto"
	"aipm/internal/taskview"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceRejectsInvalidTaskInput(t *testing.T) {
	service := &Service{}
	_, err := service.ListTasks(context.Background(), dto.ChangeListRequest{})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.GetTask(context.Background(), dto.ChangeIDRequest{})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.CreateTask(context.Background(), dto.ChangeCreateRequest{ProjectID: 1, Title: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdateTask(context.Background(), dto.ChangeUpdateRequest{ID: 2, Name: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdatePhase(context.Background(), dto.ChangeUpdatePhaseRequest{ID: 2, ChangePhase: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
	err = service.DeleteTask(context.Background(), dto.ChangeIDRequest{})
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestServiceNormalizesTaskRequests(t *testing.T) {
	repo := &fakeTaskRepository{}
	service := NewService(repo, taskview.NewTaskRenderer(fakeMarkdownParser{}, fakeMarkdownSanitizer{}))
	parentID := 4

	_, err := service.ListTasks(context.Background(), dto.ChangeListRequest{ProjectID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.projectID)
	_, err = service.GetTask(context.Background(), dto.ChangeIDRequest{ID: 2})
	require.NoError(t, err)
	assert.Equal(t, 2, repo.id)

	_, err = service.CreateTask(context.Background(), dto.ChangeCreateRequest{
		ProjectID: 1, Title: " Task Name ", Body: " Description ",
		ChangePhase: " backlog ", ChangeTypes: " task ", ParentID: &parentID,
	})
	require.NoError(t, err)
	assert.Equal(t, "Task Name", repo.createReq.Title)
	assert.Equal(t, "Description", repo.createReq.Body)
	assert.Equal(t, "backlog", repo.createReq.ChangePhase)
	assert.Equal(t, "task", repo.createReq.ChangeTypes)

	_, err = service.UpdateTask(context.Background(), dto.ChangeUpdateRequest{
		ID: 2, Name: " Updated Task ", Description: " Updated Description ", ChangeTypes: " feature ",
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Task", repo.updateReq.Name)
	assert.Equal(t, "Updated Description", repo.updateReq.Description)
	assert.Equal(t, "feature", repo.updateReq.ChangeTypes)

	_, err = service.UpdatePhase(context.Background(), dto.ChangeUpdatePhaseRequest{ID: 2, ChangePhase: " review "})
	require.NoError(t, err)
	assert.Equal(t, "review", repo.taskPhase)
	err = service.DeleteTask(context.Background(), dto.ChangeIDRequest{ID: 2})
	require.NoError(t, err)
	assert.Equal(t, 2, repo.id)
}

func TestServiceRendersTaskDescriptionHTML(t *testing.T) {
	repo := &fakeTaskRepository{}
	service := NewService(repo, taskview.NewTaskRenderer(fakeMarkdownParser{}, fakeMarkdownSanitizer{}))

	detail, err := service.GetTask(context.Background(), dto.ChangeIDRequest{ID: 2})
	require.NoError(t, err)
	assert.Equal(t, "clean(parsed(**Task**))", detail.Change.BodyHTML)
}

func TestServiceRendersBatchTaskDescriptions(t *testing.T) {
	repo := &fakeTaskRepository{}
	service := NewService(repo, taskview.NewTaskRenderer(fakeMarkdownParser{}, fakeMarkdownSanitizer{}))

	response, err := service.RenderedDescriptions(context.Background(), dto.ChangeRenderedDescriptionsRequest{
		IDs: []int{3, 2, 3},
	})
	require.NoError(t, err)
	assert.Equal(t, []int{3, 2}, repo.descriptionIDs)
	require.Equal(t, 2, len(response.Descriptions))
	assert.Equal(t, 3, response.Descriptions[0].ID)
	assert.Equal(t, "clean(parsed(**Task 3**))", response.Descriptions[0].DescriptionHTML)
	assert.Equal(t, 2, response.Descriptions[1].ID)
	assert.Equal(t, "clean(parsed(**Task 2**))", response.Descriptions[1].DescriptionHTML)
}

func TestServiceRejectsInvalidRenderedDescriptionIDs(t *testing.T) {
	service := &Service{}

	_, err := service.RenderedDescriptions(context.Background(), dto.ChangeRenderedDescriptionsRequest{
		IDs: []int{1, 0},
	})
	require.ErrorIs(t, err, ErrInvalidInput)
}

type fakeMarkdownParser struct{}

func (fakeMarkdownParser) Parse(source string) string {
	return "parsed(" + source + ")"
}

type fakeMarkdownSanitizer struct{}

func (fakeMarkdownSanitizer) Parse(source string) string {
	return "clean(" + source + ")"
}

type fakeTaskRepository struct {
	projectID      int
	id             int
	taskPhase      string
	descriptionIDs []int
	createReq      dto.ChangeCreateRequest
	updateReq      dto.ChangeUpdateRequest
}

func (r *fakeTaskRepository) References(context.Context) (dto.ChangeReferences, error) {
	return dto.ChangeReferences{}, nil
}
func (r *fakeTaskRepository) List(_ context.Context, projectID int) ([]dto.Change, error) {
	r.projectID = projectID
	return []dto.Change{}, nil
}
func (r *fakeTaskRepository) Get(_ context.Context, id int) (dto.ChangeDetail, error) {
	r.id = id
	return dto.ChangeDetail{Change: dto.Change{ID: id, Body: "**Task**"}}, nil
}
func (r *fakeTaskRepository) Descriptions(_ context.Context, ids []int) ([]dto.Change, error) {
	r.descriptionIDs = ids
	tasks := make([]dto.Change, 0, len(ids))
	for _, id := range ids {
		tasks = append(tasks, dto.Change{ID: id, Body: "**Task " + strconv.Itoa(id) + "**"})
	}
	return tasks, nil
}
func (r *fakeTaskRepository) Create(_ context.Context, req dto.ChangeCreateRequest) (dto.Change, error) {
	r.createReq = req
	return dto.Change{ID: 2, ProjectID: req.ProjectID, Name: req.Title, Body: req.Body}, nil
}
func (r *fakeTaskRepository) Update(_ context.Context, req dto.ChangeUpdateRequest) (dto.Change, error) {
	r.updateReq = req
	return dto.Change{ID: req.ID, Name: req.Name, Body: req.Description}, nil
}
func (r *fakeTaskRepository) UpdateDifficulty(_ context.Context, req dto.TaskUpdateDifficultyRequest) (dto.Change, error) {
	r.id = req.ID
	return dto.Change{ID: req.ID, Difficulty: req.Difficulty}, nil
}
func (r *fakeTaskRepository) UpdatePriority(_ context.Context, req dto.TaskUpdatePriorityRequest) (dto.Change, error) {
	r.id = req.ID
	return dto.Change{ID: req.ID, Priority: req.Priority}, nil
}
func (r *fakeTaskRepository) UpdateParent(_ context.Context, req dto.ChangeUpdateEpicRequest) (dto.Change, error) {
	r.id = req.ID
	return dto.Change{ID: req.ID, EpicID: req.EpicID}, nil
}
func (r *fakeTaskRepository) UpdatePhase(_ context.Context, req dto.ChangeUpdatePhaseRequest) (dto.Change, error) {
	r.id, r.taskPhase = req.ID, req.ChangePhase
	return dto.Change{ID: req.ID, ChangesPhase: req.ChangePhase}, nil
}
func (r *fakeTaskRepository) Delete(_ context.Context, req dto.ChangeIDRequest) error {
	r.id = req.ID
	return nil
}
