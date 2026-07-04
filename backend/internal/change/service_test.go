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
		ProjectID: 1, Title: "   ",
	})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdateTitle(context.Background(), dto.ChangeUpdateTitleRequest{ID: 2, Title: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdatePhase(context.Background(), dto.ChangeUpdatePhaseRequest{ID: 2, ChangePhase: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdateAgentEdit(context.Background(), dto.ChangeUpdateAgentEditRequest{ID: 2})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdateIdeaAgentEdit(context.Background(), dto.ChangeUpdateIdeaAgentEditRequest{ID: 2, Idea: "   "})
	require.ErrorIs(t, err, ErrInvalidInput)
	_, err = service.UpdateOpen(context.Background(), dto.ChangeUpdateOpenRequest{ID: 2})
	require.ErrorIs(t, err, ErrInvalidInput)
	badURL := "javascript:alert(1)"
	_, err = service.UpdatePRUrl(context.Background(), dto.ChangeUpdatePRUrlRequest{ID: 2, PRUrl: &badURL})
	require.ErrorIs(t, err, ErrInvalidInput)
	missingHostURL := "https:///missing-host"
	_, err = service.UpdatePRUrl(context.Background(), dto.ChangeUpdatePRUrlRequest{ID: 2, PRUrl: &missingHostURL})
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

	_, err = service.CreateChange(context.Background(), dto.ChangeCreateRequest{ProjectID: 1, Title: " Change Title ", Idea: " Idea "})
	require.NoError(t, err)
	assert.Equal(t, "Change Title", repo.createReq.Title)
	assert.Equal(t, "Idea", repo.createReq.Idea)

	_, err = service.UpdateChangeTypes(context.Background(), dto.ChangeUpdateChangeTypesRequest{ID: 2, ChangeTypes: []string{" fix ", "fix "}})
	require.NoError(t, err)
	assert.Equal(t, []string{"fix"}, repo.updateTypesReq.ChangeTypes)
	_, err = service.UpdateTitle(context.Background(), dto.ChangeUpdateTitleRequest{ID: 2, Title: " Focused Title "})
	require.NoError(t, err)
	assert.Equal(t, "Focused Title", repo.updateTitleReq.Title)
	_, err = service.UpdateIdea(context.Background(), dto.ChangeUpdateIdeaRequest{ID: 2, Idea: " Focused Idea "})
	require.NoError(t, err)
	assert.Equal(t, "Focused Idea", repo.updateIdeaReq.Idea)
	_, err = service.UpdateIdeaAgentEdit(context.Background(), dto.ChangeUpdateIdeaAgentEditRequest{ID: 2, Idea: " Agent Idea "})
	require.NoError(t, err)
	assert.Equal(t, "Agent Idea", repo.updateIdeaAgentEditReq.Idea)
	spec := " Focused Spec "
	_, err = service.UpdateSpec(context.Background(), dto.ChangeUpdateSpecRequest{ID: 2, Spec: &spec})
	require.NoError(t, err)
	require.NotNil(t, repo.updateSpecReq.Spec)
	assert.Equal(t, "Focused Spec", *repo.updateSpecReq.Spec)
	_, err = service.UpdateSpec(context.Background(), dto.ChangeUpdateSpecRequest{ID: 2})
	require.NoError(t, err)
	assert.Nil(t, repo.updateSpecReq.Spec)
	prBody := " PR Body "
	_, err = service.UpdatePRBody(context.Background(), dto.ChangeUpdatePRBodyRequest{ID: 2, PRBody: &prBody})
	require.NoError(t, err)
	require.NotNil(t, repo.updatePRBodyReq.PRBody)
	assert.Equal(t, "PR Body", *repo.updatePRBodyReq.PRBody)
	prURL := " https://example.test/pr "
	_, err = service.UpdatePRUrl(context.Background(), dto.ChangeUpdatePRUrlRequest{ID: 2, PRUrl: &prURL})
	require.NoError(t, err)
	require.NotNil(t, repo.updatePRUrlReq.PRUrl)
	assert.Equal(t, "https://example.test/pr", *repo.updatePRUrlReq.PRUrl)
	blankPRURL := ""
	_, err = service.UpdatePRUrl(context.Background(), dto.ChangeUpdatePRUrlRequest{ID: 2, PRUrl: &blankPRURL})
	require.NoError(t, err)
	require.NotNil(t, repo.updatePRUrlReq.PRUrl)
	assert.Empty(t, *repo.updatePRUrlReq.PRUrl)
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

func TestServiceRendersChangeSpecHTML(t *testing.T) {
	repo := &fakeChangeRepository{}
	service := NewService(repo, NewRenderer(fakeMarkdownParser{}, fakeMarkdownSanitizer{}))

	detail, err := service.GetChange(context.Background(), dto.ChangeIDRequest{ID: 2})
	require.NoError(t, err)
	require.NotNil(t, detail.Change.SpecHTML)
	assert.Equal(t, "clean(parsed(**Change**))", *detail.Change.SpecHTML)
}

func TestServiceRendersBatchChangeSpecs(t *testing.T) {
	repo := &fakeChangeRepository{}
	service := NewService(repo, NewRenderer(fakeMarkdownParser{}, fakeMarkdownSanitizer{}))

	response, err := service.RenderedArtifacts(context.Background(), dto.ChangeRenderedArtifactsRequest{
		IDs: []int{3, 2, 3},
	})
	require.NoError(t, err)
	assert.Equal(t, []int{3, 2}, repo.specIDs)
	require.Equal(t, 2, len(response.Artifacts))
	assert.Equal(t, 3, response.Artifacts[0].ID)
	assert.Equal(t, "clean(parsed(**Change 3**))", response.Artifacts[0].SpecHTML)
	assert.Equal(t, 2, response.Artifacts[1].ID)
	assert.Equal(t, "clean(parsed(**Change 2**))", response.Artifacts[1].SpecHTML)
}

func TestServiceRejectsInvalidRenderedSpecIDs(t *testing.T) {
	service := &Service{}
	_, err := service.RenderedArtifacts(context.Background(), dto.ChangeRenderedArtifactsRequest{IDs: []int{1, 0}})
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
	projectID              int
	id                     int
	phase                  string
	open                   *bool
	specIDs                []int
	createReq              dto.ChangeCreateRequest
	updateTypesReq         dto.ChangeUpdateChangeTypesRequest
	updateTitleReq         dto.ChangeUpdateTitleRequest
	updateIdeaReq          dto.ChangeUpdateIdeaRequest
	updateIdeaAgentEditReq dto.ChangeUpdateIdeaAgentEditRequest
	updateSpecReq          dto.ChangeUpdateSpecRequest
	updatePRBodyReq        dto.ChangeUpdatePRBodyRequest
	updatePRUrlReq         dto.ChangeUpdatePRUrlRequest
	updateAgentEditReq     dto.ChangeUpdateAgentEditRequest
}

func (r *fakeChangeRepository) List(_ context.Context, projectID int) ([]dto.ChangeListItem, error) {
	r.projectID = projectID
	return []dto.ChangeListItem{}, nil
}

func (r *fakeChangeRepository) Get(_ context.Context, id int) (dto.ChangeDetail, error) {
	r.id = id
	spec := "**Change**"
	return dto.ChangeDetail{Change: dto.Change{ID: id, Spec: &spec}}, nil
}

func (r *fakeChangeRepository) Artifacts(_ context.Context, ids []int) ([]dto.Change, error) {
	r.specIDs = ids
	changes := make([]dto.Change, 0, len(ids))
	for _, id := range ids {
		spec := "**Change " + strconv.Itoa(id) + "**"
		changes = append(changes, dto.Change{ID: id, Spec: &spec})
	}
	return changes, nil
}

func (r *fakeChangeRepository) Create(_ context.Context, req dto.ChangeCreateRequest) (dto.Change, error) {
	r.createReq = req
	return dto.Change{ID: 2, ProjectID: req.ProjectID, Title: req.Title, Idea: req.Idea}, nil
}

func (r *fakeChangeRepository) UpdateChangeTypes(_ context.Context, req dto.ChangeUpdateChangeTypesRequest) (dto.Change, error) {
	r.updateTypesReq = req
	return dto.Change{ID: req.ID, ChangeTypes: req.ChangeTypes}, nil
}

func (r *fakeChangeRepository) UpdateTitle(_ context.Context, req dto.ChangeUpdateTitleRequest) (dto.Change, error) {
	r.updateTitleReq = req
	return dto.Change{ID: req.ID, Title: req.Title}, nil
}

func (r *fakeChangeRepository) UpdateIdea(_ context.Context, req dto.ChangeUpdateIdeaRequest) (dto.Change, error) {
	r.updateIdeaReq = req
	return dto.Change{ID: req.ID, Idea: req.Idea}, nil
}

func (r *fakeChangeRepository) UpdateIdeaAgentEdit(_ context.Context, req dto.ChangeUpdateIdeaAgentEditRequest) (dto.Change, error) {
	r.updateIdeaAgentEditReq = req
	return dto.Change{ID: req.ID, Idea: req.Idea, AgentEdit: true}, nil
}

func (r *fakeChangeRepository) UpdateSpec(_ context.Context, req dto.ChangeUpdateSpecRequest) (dto.Change, error) {
	r.updateSpecReq = req
	return dto.Change{ID: req.ID, Spec: req.Spec}, nil
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

func (r *fakeChangeRepository) Reference(_ context.Context, req dto.ChangeIDRequest) (dto.Change, error) {
	r.id = req.ID
	return dto.Change{ID: req.ID}, nil
}
