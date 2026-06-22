package project

import (
	"context"
	"testing"

	"aipm/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceRejectsInvalidProjectInput(t *testing.T) {
	service := &Service{}

	_, err := service.GetProject(context.Background(), dto.ProjectIDRequest{})
	require.ErrorIs(t, err, ErrInvalidInput)

	_, err = service.CreateProject(context.Background(), dto.ProjectCreateRequest{Name: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)

	_, err = service.UpdateProject(context.Background(), dto.ProjectUpdateRequest{ID: "project-id", Name: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)

	err = service.DeleteProject(context.Background(), dto.ProjectIDRequest{ID: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestServiceNormalizesProjectListRequest(t *testing.T) {
	repo := &fakeProjectRepository{}
	service := NewService(repo)

	_, err := service.ListProjects(context.Background(), dto.ProjectListRequest{Limit: 999, Offset: -5})
	require.NoError(t, err)

	assert.Equal(t, 100, repo.limit)
	assert.Equal(t, 0, repo.offset)
}

func TestServiceTrimsProjectRequests(t *testing.T) {
	repo := &fakeProjectRepository{}
	service := NewService(repo)

	_, err := service.GetProject(context.Background(), dto.ProjectIDRequest{ID: " project-id "})
	require.NoError(t, err)
	assert.Equal(t, "project-id", repo.id)

	_, err = service.CreateProject(context.Background(), dto.ProjectCreateRequest{Name: " Project Name "})
	require.NoError(t, err)
	assert.Equal(t, "Project Name", repo.name)

	_, err = service.UpdateProject(context.Background(), dto.ProjectUpdateRequest{ID: " project-id ", Name: " Updated Name "})
	require.NoError(t, err)
	assert.Equal(t, "project-id", repo.id)
	assert.Equal(t, "Updated Name", repo.name)

	err = service.DeleteProject(context.Background(), dto.ProjectIDRequest{ID: " project-id "})
	require.NoError(t, err)
	assert.Equal(t, "project-id", repo.id)
}

type fakeProjectRepository struct {
	limit  int
	offset int
	id     string
	name   string
}

func (r *fakeProjectRepository) List(_ context.Context, limit, offset int) ([]dto.Project, error) {
	r.limit = limit
	r.offset = offset
	return []dto.Project{}, nil
}

func (r *fakeProjectRepository) Get(_ context.Context, id string) (dto.Project, error) {
	r.id = id
	return dto.Project{Id: id, Name: "Project"}, nil
}

func (r *fakeProjectRepository) Create(_ context.Context, name string) (dto.Project, error) {
	r.name = name
	return dto.Project{Id: "project-id", Name: name}, nil
}

func (r *fakeProjectRepository) Update(_ context.Context, id, name string) (dto.Project, error) {
	r.id = id
	r.name = name
	return dto.Project{Id: id, Name: name}, nil
}

func (r *fakeProjectRepository) Delete(_ context.Context, id string) error {
	r.id = id
	return nil
}
