package change

import (
	"context"
	"strconv"
	"testing"

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
		ProjectID: 1, Title: "   ", ChangeTypes: []string{"feature"},
	})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdateTitle(context.Background(), dto.ChangeUpdateTitleRequest{ID: 2, Title: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdatePhase(context.Background(), dto.ChangeUpdatePhaseRequest{ID: 2, ChangePhase: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdateAgentEdit(context.Background(), dto.ChangeUpdateAgentEditRequest{ID: 2})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdateOpen(context.Background(), dto.ChangeUpdateOpenRequest{ID: 2})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdatePRUrl(context.Background(), dto.ChangeUpdatePRUrlRequest{ID: 2, PRUrl: "javascript:alert(1)"})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdatePRUrl(context.Background(), dto.ChangeUpdatePRUrlRequest{ID: 2, PRUrl: "https:///missing-host"})
	require.ErrorIs(t, err, ErrInvalidInput)
	err = service.DeleteChange(context.Background(), dto.ChangeIDRequest{})
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestServiceNormalizesChangeRequests(t *testing.T) {
	repo := &fakeChangeRepository{}
	service := NewService(repo, NewRenderer(fakeMarkdownParser{}, fakeMarkdownSanitizer{}))
	epicID := 4

	_, err := service.ListChanges(context.Background(), dto.ChangeListRequest{ProjectID: 1})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.projectID)
	_, err = service.GetChange(context.Background(), dto.ChangeIDRequest{ID: 2})
	require.NoError(t, err)
	assert.Equal(t, 2, repo.id)

	_, err = service.CreateChange(context.Background(), dto.ChangeCreateRequest{
		ProjectID: 1, Title: " Change Title ", Body: " Body ",
		ChangeTypes: []string{" feature ", "feature", " fix "}, EpicID: &epicID,
	})
	require.NoError(t, err)
	assert.Equal(t, "Change Title", repo.createReq.Title)
	assert.Equal(t, "Body", repo.createReq.Body)
	assert.Equal(t, []string{"feature", "fix"}, repo.createReq.ChangeTypes)

	_, err = service.UpdateChangeTypes(context.Background(), dto.ChangeUpdateChangeTypesRequest{ID: 2, ChangeTypes: []string{" fix ", "fix "}})
	require.NoError(t, err)
	assert.Equal(t, []string{"fix"}, repo.updateTypesReq.ChangeTypes)
	_, err = service.UpdateTitle(context.Background(), dto.ChangeUpdateTitleRequest{ID: 2, Title: " Focused Title "})
	require.NoError(t, err)
	assert.Equal(t, "Focused Title", repo.updateTitleReq.Title)
	_, err = service.UpdateBody(context.Background(), dto.ChangeUpdateBodyRequest{ID: 2, Body: " Focused Body "})
	require.NoError(t, err)
	assert.Equal(t, "Focused Body", repo.updateBodyReq.Body)
	_, err = service.UpdatePRBody(context.Background(), dto.ChangeUpdatePRBodyRequest{ID: 2, PRBody: " PR Body "})
	require.NoError(t, err)
	assert.Equal(t, "PR Body", repo.updatePRBodyReq.PRBody)
	_, err = service.UpdatePRUrl(context.Background(), dto.ChangeUpdatePRUrlRequest{ID: 2, PRUrl: " https://example.test/pr "})
	require.NoError(t, err)
	assert.Equal(t, "https://example.test/pr", repo.updatePRUrlReq.PRUrl)
	_, err = service.UpdatePRUrl(context.Background(), dto.ChangeUpdatePRUrlRequest{ID: 2, PRUrl: ""})
	require.NoError(t, err)
	assert.Empty(t, repo.updatePRUrlReq.PRUrl)
	agentEdit := true
	_, err = service.UpdateAgentEdit(context.Background(), dto.ChangeUpdateAgentEditRequest{ID: 2, AgentEdit: &agentEdit})
	require.NoError(t, err)
	require.NotNil(t, repo.updateAgentEditReq.AgentEdit)
	assert.True(t, *repo.updateAgentEditReq.AgentEdit)

	_, err = service.UpdateEpic(context.Background(), dto.ChangeUpdateEpicRequest{ID: 2, EpicID: &epicID})
	require.NoError(t, err)
	assert.Equal(t, 2, repo.id)
	_, err = service.UpdatePhase(context.Background(), dto.ChangeUpdatePhaseRequest{ID: 2, ChangePhase: " review "})
	require.NoError(t, err)
	assert.Equal(t, "review", repo.phase)
	open := true
	_, err = service.UpdateOpen(context.Background(), dto.ChangeUpdateOpenRequest{ID: 2, Open: &open})
	require.NoError(t, err)
	require.NotNil(t, repo.open)
	assert.True(t, *repo.open)
	err = service.DeleteChange(context.Background(), dto.ChangeIDRequest{ID: 2})
	require.NoError(t, err)
	assert.Equal(t, 2, repo.id)
}

func TestServiceRendersChangeBodyHTML(t *testing.T) {
	repo := &fakeChangeRepository{}
	service := NewService(repo, NewRenderer(fakeMarkdownParser{}, fakeMarkdownSanitizer{}))

	detail, err := service.GetChange(context.Background(), dto.ChangeIDRequest{ID: 2})
	require.NoError(t, err)
	assert.Equal(t, "clean(parsed(**Change**))", detail.Change.HTML)
}

func TestServiceRendersBatchChangeBodies(t *testing.T) {
	repo := &fakeChangeRepository{}
	service := NewService(repo, NewRenderer(fakeMarkdownParser{}, fakeMarkdownSanitizer{}))

	response, err := service.RenderedBodies(context.Background(), dto.ChangeRenderedBodiesRequest{
		IDs: []int{3, 2, 3},
	})
	require.NoError(t, err)
	assert.Equal(t, []int{3, 2}, repo.bodyIDs)
	require.Equal(t, 2, len(response.Bodies))
	assert.Equal(t, 3, response.Bodies[0].ID)
	assert.Equal(t, "clean(parsed(**Change 3**))", response.Bodies[0].HTML)
	assert.Equal(t, 2, response.Bodies[1].ID)
	assert.Equal(t, "clean(parsed(**Change 2**))", response.Bodies[1].HTML)
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
	projectID          int
	id                 int
	phase              string
	open               *bool
	bodyIDs            []int
	createReq          dto.ChangeCreateRequest
	updateTypesReq     dto.ChangeUpdateChangeTypesRequest
	updateTitleReq     dto.ChangeUpdateTitleRequest
	updateBodyReq      dto.ChangeUpdateBodyRequest
	updatePRBodyReq    dto.ChangeUpdatePRBodyRequest
	updatePRUrlReq     dto.ChangeUpdatePRUrlRequest
	updateAgentEditReq dto.ChangeUpdateAgentEditRequest
}

func (r *fakeChangeRepository) List(_ context.Context, projectID int) ([]dto.ChangeListItem, error) {
	r.projectID = projectID
	return []dto.ChangeListItem{}, nil
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

func (r *fakeChangeRepository) UpdateChangeTypes(_ context.Context, req dto.ChangeUpdateChangeTypesRequest) (dto.Change, error) {
	r.updateTypesReq = req
	return dto.Change{ID: req.ID, ChangeTypes: req.ChangeTypes}, nil
}

func (r *fakeChangeRepository) UpdateTitle(_ context.Context, req dto.ChangeUpdateTitleRequest) (dto.Change, error) {
	r.updateTitleReq = req
	return dto.Change{ID: req.ID, Title: req.Title}, nil
}

func (r *fakeChangeRepository) UpdateBody(_ context.Context, req dto.ChangeUpdateBodyRequest) (dto.Change, error) {
	r.updateBodyReq = req
	return dto.Change{ID: req.ID, Body: req.Body}, nil
}

func (r *fakeChangeRepository) UpdatePRBody(_ context.Context, req dto.ChangeUpdatePRBodyRequest) (dto.Change, error) {
	r.updatePRBodyReq = req
	return dto.Change{ID: req.ID, PRBody: req.PRBody}, nil
}

func (r *fakeChangeRepository) UpdatePRUrl(_ context.Context, req dto.ChangeUpdatePRUrlRequest) (dto.Change, error) {
	r.updatePRUrlReq = req
	return dto.Change{ID: req.ID, PRUrl: req.PRUrl}, nil
}

func (r *fakeChangeRepository) UpdateAgentEdit(_ context.Context, req dto.ChangeUpdateAgentEditRequest) (dto.Change, error) {
	r.updateAgentEditReq = req
	return dto.Change{ID: req.ID, AgentEdit: *req.AgentEdit}, nil
}

func (r *fakeChangeRepository) UpdateEpic(_ context.Context, req dto.ChangeUpdateEpicRequest) (dto.Change, error) {
	r.id = req.ID
	return dto.Change{ID: req.ID, EpicID: req.EpicID}, nil
}

func (r *fakeChangeRepository) UpdatePhase(_ context.Context, req dto.ChangeUpdatePhaseRequest) (dto.Change, error) {
	r.id, r.phase = req.ID, req.ChangePhase
	return dto.Change{ID: req.ID, ChangePhase: req.ChangePhase}, nil
}

func (r *fakeChangeRepository) UpdateOpen(_ context.Context, req dto.ChangeUpdateOpenRequest) (dto.Change, error) {
	r.id, r.open = req.ID, req.Open
	return dto.Change{ID: req.ID, Open: *req.Open}, nil
}

func (r *fakeChangeRepository) Delete(_ context.Context, req dto.ChangeIDRequest) error {
	r.id = req.ID
	return nil
}
