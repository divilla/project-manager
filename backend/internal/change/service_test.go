package change

import (
	"context"
	"strconv"
	"testing"

	"aipm/internal/changeview"
	"aipm/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceRejectsInvalidChangeInput(t *testing.T) {
	service := &Service{}
	_, err := service.ListChanges(context.Background(), dto.ChangeListRequest{})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.GetChange(context.Background(), dto.ChangeIDRequest{})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.CreateChange(context.Background(), dto.ChangeCreateRequest{
		ProjectID: 1, Title: "   ", ChangePhase: "backlog", ChangeTypes: []string{"feature"},
	})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdateChange(context.Background(), dto.ChangeUpdateRequest{ID: 2, Title: "   ", ChangeTypes: []string{"feature"}})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdatePhase(context.Background(), dto.ChangeUpdatePhaseRequest{ID: 2, ChangePhase: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
	err = service.DeleteChange(context.Background(), dto.ChangeIDRequest{})
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestServiceNormalizesChangeRequests(t *testing.T) {
	repo := &fakeChangeRepository{}
	service := NewService(repo, changeview.NewChangeRenderer(fakeMarkdownParser{}, fakeMarkdownSanitizer{}))
	epicID := 4

	_, err := service.ListChanges(context.Background(), dto.ChangeListRequest{ProjectID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.projectID)
	_, err = service.GetChange(context.Background(), dto.ChangeIDRequest{ID: 2})
	require.NoError(t, err)
	assert.Equal(t, 2, repo.id)

	_, err = service.CreateChange(context.Background(), dto.ChangeCreateRequest{
		ProjectID: 1, Title: " Change Title ", Body: " Body ",
		ChangePhase: " backlog ", ChangeTypes: []string{" feature ", "feature", " fix "}, EpicID: &epicID,
	})
	require.NoError(t, err)
	assert.Equal(t, "Change Title", repo.createReq.Title)
	assert.Equal(t, "Body", repo.createReq.Body)
	assert.Equal(t, "backlog", repo.createReq.ChangePhase)
	assert.Equal(t, []string{"feature", "fix"}, repo.createReq.ChangeTypes)

	_, err = service.UpdateChange(context.Background(), dto.ChangeUpdateRequest{
		ID: 2, Title: " Updated Change ", Body: " Updated Body ", ChangeTypes: []string{" docs "},
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Change", repo.updateReq.Title)
	assert.Equal(t, "Updated Body", repo.updateReq.Body)
	assert.Equal(t, []string{"docs"}, repo.updateReq.ChangeTypes)

	_, err = service.UpdateEpic(context.Background(), dto.ChangeUpdateEpicRequest{ID: 2, EpicID: &epicID})
	require.NoError(t, err)
	assert.Equal(t, 2, repo.id)
	_, err = service.UpdatePhase(context.Background(), dto.ChangeUpdatePhaseRequest{ID: 2, ChangePhase: " review "})
	require.NoError(t, err)
	assert.Equal(t, "review", repo.phase)
	_, err = service.UpdateClosed(context.Background(), dto.ChangeUpdateClosedRequest{ID: 2, Closed: true})
	require.NoError(t, err)
	assert.True(t, repo.closed)
	err = service.DeleteChange(context.Background(), dto.ChangeIDRequest{ID: 2})
	require.NoError(t, err)
	assert.Equal(t, 2, repo.id)
}

func TestServiceRendersChangeBodyHTML(t *testing.T) {
	repo := &fakeChangeRepository{}
	service := NewService(repo, changeview.NewChangeRenderer(fakeMarkdownParser{}, fakeMarkdownSanitizer{}))

	detail, err := service.GetChange(context.Background(), dto.ChangeIDRequest{ID: 2})
	require.NoError(t, err)
	assert.Equal(t, "clean(parsed(**Change**))", detail.Change.BodyHTML)
}

func TestServiceRendersBatchChangeBodies(t *testing.T) {
	repo := &fakeChangeRepository{}
	service := NewService(repo, changeview.NewChangeRenderer(fakeMarkdownParser{}, fakeMarkdownSanitizer{}))

	response, err := service.RenderedBodies(context.Background(), dto.ChangeRenderedBodiesRequest{
		IDs: []int{3, 2, 3},
	})
	require.NoError(t, err)
	assert.Equal(t, []int{3, 2}, repo.bodyIDs)
	require.Equal(t, 2, len(response.Bodies))
	assert.Equal(t, 3, response.Bodies[0].ID)
	assert.Equal(t, "clean(parsed(**Change 3**))", response.Bodies[0].BodyHTML)
	assert.Equal(t, 2, response.Bodies[1].ID)
	assert.Equal(t, "clean(parsed(**Change 2**))", response.Bodies[1].BodyHTML)
}

func TestServiceRejectsInvalidRenderedBodyIDs(t *testing.T) {
	service := &Service{}
	_, err := service.RenderedBodies(context.Background(), dto.ChangeRenderedBodiesRequest{IDs: []int{1, 0}})
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

type fakeChangeRepository struct {
	projectID int
	id        int
	phase     string
	closed    bool
	bodyIDs   []int
	createReq dto.ChangeCreateRequest
	updateReq dto.ChangeUpdateRequest
}

func (r *fakeChangeRepository) References(context.Context) (dto.ChangeReferences, error) {
	return dto.ChangeReferences{}, nil
}

func (r *fakeChangeRepository) List(_ context.Context, projectID int) ([]dto.Change, error) {
	r.projectID = projectID
	return []dto.Change{}, nil
}

func (r *fakeChangeRepository) Get(_ context.Context, id int) (dto.ChangeDetail, error) {
	r.id = id
	return dto.ChangeDetail{Change: dto.Change{ID: id, Body: "**Change**"}}, nil
}

func (r *fakeChangeRepository) Bodies(_ context.Context, ids []int) ([]dto.Change, error) {
	r.bodyIDs = ids
	changes := make([]dto.Change, 0, len(ids))
	for _, id := range ids {
		changes = append(changes, dto.Change{ID: id, Body: "**Change " + strconv.Itoa(id) + "**"})
	}
	return changes, nil
}

func (r *fakeChangeRepository) Create(_ context.Context, req dto.ChangeCreateRequest) (dto.Change, error) {
	r.createReq = req
	return dto.Change{ID: 2, ProjectID: req.ProjectID, Title: req.Title, Body: req.Body}, nil
}

func (r *fakeChangeRepository) Update(_ context.Context, req dto.ChangeUpdateRequest) (dto.Change, error) {
	r.updateReq = req
	return dto.Change{ID: req.ID, Title: req.Title, Body: req.Body}, nil
}

func (r *fakeChangeRepository) UpdateEpic(_ context.Context, req dto.ChangeUpdateEpicRequest) (dto.Change, error) {
	r.id = req.ID
	return dto.Change{ID: req.ID, EpicID: req.EpicID}, nil
}

func (r *fakeChangeRepository) UpdatePhase(_ context.Context, req dto.ChangeUpdatePhaseRequest) (dto.Change, error) {
	r.id, r.phase = req.ID, req.ChangePhase
	return dto.Change{ID: req.ID, ChangePhase: req.ChangePhase}, nil
}

func (r *fakeChangeRepository) UpdateClosed(_ context.Context, req dto.ChangeUpdateClosedRequest) (dto.Change, error) {
	r.id, r.closed = req.ID, req.Closed
	return dto.Change{ID: req.ID, Closed: req.Closed}, nil
}

func (r *fakeChangeRepository) Delete(_ context.Context, req dto.ChangeIDRequest) error {
	r.id = req.ID
	return nil
}
