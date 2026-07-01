package change

import (
	"context"
	"errors"
	"slices"

	"aipm/internal/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	// Repo defines Repo values.
	Repo struct {
		pool *pgxpool.Pool
	}

	// Repository defines Repository values.
	Repository interface {
		References(ctx context.Context) (dto.ChangeReferences, error)
		List(ctx context.Context, projectID int) ([]dto.Change, error)
		Get(ctx context.Context, id int) (dto.ChangeDetail, error)
		Bodies(ctx context.Context, ids []int) ([]dto.Change, error)
		Create(ctx context.Context, req dto.ChangeCreateRequest) (dto.Change, error)
		UpdateChangeTypes(ctx context.Context, req dto.ChangeUpdateChangeTypesRequest) (dto.Change, error)
		UpdateTitle(ctx context.Context, req dto.ChangeUpdateTitleRequest) (dto.Change, error)
		UpdateRequirementBody(ctx context.Context, req dto.ChangeUpdateRequirementBodyRequest) (dto.Change, error)
		UpdatePullRequestBody(ctx context.Context, req dto.ChangeUpdatePullRequestBodyRequest) (dto.Change, error)
		UpdatePullRequestURL(ctx context.Context, req dto.ChangeUpdatePullRequestURLRequest) (dto.Change, error)
		UpdateEpic(ctx context.Context, req dto.ChangeUpdateEpicRequest) (dto.Change, error)
		UpdatePhase(ctx context.Context, req dto.ChangeUpdatePhaseRequest) (dto.Change, error)
		UpdateClosed(ctx context.Context, req dto.ChangeUpdateClosedRequest) (dto.Change, error)
		Delete(ctx context.Context, req dto.ChangeIDRequest) error
	}
)

const changeColumns = `
	id,
	ref,
	slug,
	version,
	project_id,
	epic_id,
	change_phase,
	change_types,
	title,
	coalesce(requirement_body, ''),
	coalesce(pull_request_body, ''),
	coalesce(pull_request_url, ''),
	closed,
	done_tc,
	total_tc,
	completed,
	created,
	modified`

// NewRepo initializes or executes NewRepo behavior.
func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// References executes References behavior.
func (r *Repo) References(ctx context.Context) (dto.ChangeReferences, error) {
	phases, err := r.referenceOptions(ctx, "change_phase")
	if err != nil {
		return dto.ChangeReferences{}, err
	}
	types, err := r.referenceOptions(ctx, "change_type")
	if err != nil {
		return dto.ChangeReferences{}, err
	}
	return dto.ChangeReferences{Phases: phases, Types: types}, nil
}

// List executes List behavior.
func (r *Repo) List(ctx context.Context, projectID int) ([]dto.Change, error) {
	rows, err := r.pool.Query(ctx, `
		select `+changeColumns+`
		from public.change
		where project_id = $1
		order by modified desc, id desc
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	changes := make([]dto.Change, 0)
	for rows.Next() {
		change, err := scanChange(rows)
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	return changes, rows.Err()
}

// Get executes Get behavior.
func (r *Repo) Get(ctx context.Context, id int) (dto.ChangeDetail, error) {
	change, err := getChange(ctx, r.pool, id)
	if err != nil {
		return dto.ChangeDetail{}, err
	}
	testCases, err := listTestCases(ctx, r.pool, id)
	if err != nil {
		return dto.ChangeDetail{}, err
	}
	return dto.ChangeDetail{Change: change, TestCases: testCases}, nil
}

// Bodies executes Bodies behavior.
func (r *Repo) Bodies(ctx context.Context, ids []int) ([]dto.Change, error) {
	rows, err := r.pool.Query(ctx, `
		select requested.id::integer, coalesce(c.requirement_body, ''), coalesce(c.pull_request_body, '')
		from unnest($1::bigint[]) with ordinality as requested(id, ord)
		join public.change c on c.id = requested.id
		order by requested.ord
	`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	changes := make([]dto.Change, 0, len(ids))
	for rows.Next() {
		var change dto.Change
		if err := rows.Scan(&change.ID, &change.RequirementBody, &change.PullRequestBody); err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	return changes, rows.Err()
}

// Create executes Create behavior.
func (r *Repo) Create(ctx context.Context, req dto.ChangeCreateRequest) (dto.Change, error) {
	if err := r.ensureProject(ctx, req.ProjectID); err != nil {
		return dto.Change{}, err
	}
	if err := r.ensureReferences(ctx, "change_type", req.ChangeTypes); err != nil {
		return dto.Change{}, err
	}
	if req.EpicID != nil {
		if err := r.ensureEpic(ctx, req.ProjectID, *req.EpicID); err != nil {
			return dto.Change{}, err
		}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.Change{}, err
	}
	defer tx.Rollback(ctx)

	var id int
	err = tx.QueryRow(ctx, `
		select public.fn_change_insert($1, $2, $3, $4, nullif($5, ''))
	`, req.ProjectID, req.ChangeTypes, req.EpicID, req.Title, req.RequirementBody).Scan(&id)
	if err != nil {
		return dto.Change{}, err
	}

	return finishMutation(ctx, tx, id)
}

// UpdateChangeTypes executes UpdateChangeTypes behavior.
func (r *Repo) UpdateChangeTypes(ctx context.Context, req dto.ChangeUpdateChangeTypesRequest) (dto.Change, error) {
	if err := r.ensureReferences(ctx, "change_type", req.ChangeTypes); err != nil {
		return dto.Change{}, err
	}
	return r.updateField(ctx, req.ID, func(current state) bool {
		return slices.Equal(current.ChangeTypes, req.ChangeTypes)
	}, `
		update public.change
		set change_types = $2,
			version = version + 1,
			modified = now()
		where id = $1
	`, req.ChangeTypes)
}

// UpdateTitle executes UpdateTitle behavior.
func (r *Repo) UpdateTitle(ctx context.Context, req dto.ChangeUpdateTitleRequest) (dto.Change, error) {
	return r.updateField(ctx, req.ID, func(current state) bool {
		return current.Title == req.Title
	}, `
		update public.change
		set title = $2,
			version = version + 1,
			modified = now()
		where id = $1
	`, req.Title)
}

// UpdateRequirementBody executes UpdateRequirementBody behavior.
func (r *Repo) UpdateRequirementBody(ctx context.Context, req dto.ChangeUpdateRequirementBodyRequest) (dto.Change, error) {
	return r.updateField(ctx, req.ID, func(current state) bool {
		return current.RequirementBody == req.RequirementBody
	}, `
		update public.change
		set requirement_body = nullif($2, ''),
			version = version + 1,
			modified = now()
		where id = $1
	`, req.RequirementBody)
}

// UpdatePullRequestBody executes UpdatePullRequestBody behavior.
func (r *Repo) UpdatePullRequestBody(ctx context.Context, req dto.ChangeUpdatePullRequestBodyRequest) (dto.Change, error) {
	return r.updateField(ctx, req.ID, func(current state) bool {
		return current.PullRequestBody == req.PullRequestBody
	}, `
		update public.change
		set pull_request_body = nullif($2, ''),
			version = version + 1,
			modified = now()
		where id = $1
	`, req.PullRequestBody)
}

// UpdatePullRequestURL executes UpdatePullRequestURL behavior.
func (r *Repo) UpdatePullRequestURL(ctx context.Context, req dto.ChangeUpdatePullRequestURLRequest) (dto.Change, error) {
	return r.updateField(ctx, req.ID, func(current state) bool {
		return current.PullRequestURL == req.PullRequestURL
	}, `
		update public.change
		set pull_request_url = nullif($2, ''),
			version = version + 1,
			modified = now()
		where id = $1
	`, req.PullRequestURL)
}

// UpdateEpic executes UpdateEpic behavior.
func (r *Repo) UpdateEpic(ctx context.Context, req dto.ChangeUpdateEpicRequest) (dto.Change, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.Change{}, err
	}
	defer tx.Rollback(ctx)

	current, err := getState(ctx, tx, req.ID)
	if err != nil {
		return dto.Change{}, err
	}
	if req.EpicID != nil {
		if err := ensureEpic(ctx, tx, current.ProjectID, *req.EpicID); err != nil {
			return dto.Change{}, err
		}
	}
	if equalIntPointers(current.EpicID, req.EpicID) {
		return finishMutation(ctx, tx, req.ID)
	}
	if _, err := tx.Exec(ctx, "call public.sp_change_to_history($1, false)", req.ID); err != nil {
		return dto.Change{}, err
	}
	tag, err := tx.Exec(ctx, `
		update public.change
		set epic_id = $2,
			version = version + 1,
			modified = now()
		where id = $1
	`, req.ID, req.EpicID)
	if err != nil {
		return dto.Change{}, err
	}
	if tag.RowsAffected() == 0 {
		return dto.Change{}, ErrNotFound
	}
	if err := recalculateEpics(ctx, tx, current.EpicID, req.EpicID); err != nil {
		return dto.Change{}, err
	}
	return finishMutation(ctx, tx, req.ID)
}

// UpdatePhase executes UpdatePhase behavior.
func (r *Repo) UpdatePhase(ctx context.Context, req dto.ChangeUpdatePhaseRequest) (dto.Change, error) {
	if err := r.ensureReference(ctx, "change_phase", req.ChangePhase); err != nil {
		return dto.Change{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.Change{}, err
	}
	defer tx.Rollback(ctx)

	current, err := getState(ctx, tx, req.ID)
	if err != nil {
		return dto.Change{}, err
	}
	if current.ChangePhase == req.ChangePhase {
		return finishMutation(ctx, tx, req.ID)
	}
	if _, err := tx.Exec(ctx, "call public.sp_change_to_history($1, false)", req.ID); err != nil {
		return dto.Change{}, err
	}
	tag, err := tx.Exec(ctx, `
		update public.change
		set change_phase = $2,
			version = version + 1,
			modified = now()
		where id = $1
	`, req.ID, req.ChangePhase)
	if err != nil {
		return dto.Change{}, err
	}
	if tag.RowsAffected() == 0 {
		return dto.Change{}, ErrNotFound
	}
	return finishMutation(ctx, tx, req.ID)
}

// UpdateClosed executes UpdateClosed behavior.
func (r *Repo) UpdateClosed(ctx context.Context, req dto.ChangeUpdateClosedRequest) (dto.Change, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.Change{}, err
	}
	defer tx.Rollback(ctx)

	current, err := getState(ctx, tx, req.ID)
	if err != nil {
		return dto.Change{}, err
	}
	if current.Closed == req.Closed {
		return finishMutation(ctx, tx, req.ID)
	}
	if _, err := tx.Exec(ctx, "call public.sp_change_to_history($1, false)", req.ID); err != nil {
		return dto.Change{}, err
	}
	tag, err := tx.Exec(ctx, `
		update public.change
		set closed = $2,
			version = version + 1,
			modified = now()
		where id = $1
	`, req.ID, req.Closed)
	if err != nil {
		return dto.Change{}, err
	}
	if tag.RowsAffected() == 0 {
		return dto.Change{}, ErrNotFound
	}
	return finishMutation(ctx, tx, req.ID)
}

// Delete executes Delete behavior.
func (r *Repo) Delete(ctx context.Context, req dto.ChangeIDRequest) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	current, err := getState(ctx, tx, req.ID)
	if err != nil {
		return err
	}
	testCases, err := testCasesForChange(ctx, tx, req.ID)
	if err != nil {
		return err
	}
	for _, testCase := range testCases {
		if _, err := tx.Exec(ctx, "call public.sp_test_case_to_history($1, true)", testCase.ID); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, "call public.sp_change_to_history($1, true)", req.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "delete from public.test_case where change_id = $1", req.ID); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, "delete from public.change where id = $1", req.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	if _, err := tx.Exec(ctx, "call public.sp_epic_test_case_recalculate($1)", current.EpicID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repo) referenceOptions(ctx context.Context, table string) ([]dto.ChangePhase, error) {
	rows, err := r.pool.Query(ctx, "select slug, priority from public."+table+" order by priority, slug")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	options := make([]dto.ChangePhase, 0)
	for rows.Next() {
		var option dto.ChangePhase
		if err := rows.Scan(&option.Slug, &option.Priority); err != nil {
			return nil, err
		}
		options = append(options, option)
	}
	return options, rows.Err()
}

func (r *Repo) ensureReference(ctx context.Context, table, slug string) error {
	var exists bool
	if err := r.pool.QueryRow(ctx, "select exists(select 1 from public."+table+" where slug = $1)", slug).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrInvalidReference
	}
	return nil
}

func (r *Repo) ensureReferences(ctx context.Context, table string, slugs []string) error {
	for _, slug := range slugs {
		if err := r.ensureReference(ctx, table, slug); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) ensureProject(ctx context.Context, id int) error {
	var exists bool
	if err := r.pool.QueryRow(ctx, "select exists(select 1 from public.project where id = $1)", id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrInvalidReference
	}
	return nil
}

func (r *Repo) ensureEpic(ctx context.Context, projectID, id int) error {
	return ensureEpic(ctx, r.pool, projectID, id)
}

func ensureEpic(ctx context.Context, q queryer, projectID, id int) error {
	var exists bool
	if err := q.QueryRow(ctx, `
		select exists(select 1 from public.epic where id = $1 and project_id = $2)
	`, id, projectID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrInvalidReference
	}
	return nil
}

func getChange(ctx context.Context, q queryer, id int) (dto.Change, error) {
	change, err := scanChange(q.QueryRow(ctx, "select "+changeColumns+" from public.change where id = $1", id))
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Change{}, ErrNotFound
	}
	return change, err
}

func listTestCases(ctx context.Context, q queryer, changeID int) ([]dto.TestCase, error) {
	rows, err := q.Query(ctx, `
		select id, version, scenario, done, change_id, created, modified
		from public.test_case
		where change_id = $1
		order by created, scenario
	`, changeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	testCases := make([]dto.TestCase, 0)
	for rows.Next() {
		var item dto.TestCase
		if err := rows.Scan(&item.ID, &item.Version, &item.Scenario, &item.Done, &item.ChangeID, &item.Created, &item.Modified); err != nil {
			return nil, err
		}
		testCases = append(testCases, item)
	}
	return testCases, rows.Err()
}

func scanChange(row pgx.Row) (dto.Change, error) {
	var change dto.Change
	var epicID pgtype.Int8
	err := row.Scan(
		&change.ID, &change.Ref, &change.Slug, &change.Version, &change.ProjectID, &epicID, &change.ChangePhase,
		&change.ChangeTypes, &change.Title, &change.RequirementBody, &change.PullRequestBody,
		&change.PullRequestURL, &change.Closed, &change.DoneTC,
		&change.TotalTC, &change.Completed, &change.Created, &change.Modified,
	)
	if err != nil {
		return dto.Change{}, err
	}
	if epicID.Valid {
		value := int(epicID.Int64)
		change.EpicID = &value
	}
	return change, nil
}

type state struct {
	ProjectID       int
	EpicID          *int
	ChangePhase     string
	ChangeTypes     []string
	Title           string
	RequirementBody string
	PullRequestBody string
	PullRequestURL  string
	Closed          bool
}

func getState(ctx context.Context, tx pgx.Tx, id int) (state, error) {
	var item state
	var epicID pgtype.Int8
	err := tx.QueryRow(ctx, `
		select project_id, epic_id, change_phase, change_types, title, coalesce(requirement_body, ''), coalesce(pull_request_body, ''), coalesce(pull_request_url, ''), closed
		from public.change
		where id = $1
	`, id).Scan(&item.ProjectID, &epicID, &item.ChangePhase, &item.ChangeTypes, &item.Title, &item.RequirementBody, &item.PullRequestBody, &item.PullRequestURL, &item.Closed)
	if errors.Is(err, pgx.ErrNoRows) {
		return state{}, ErrNotFound
	}
	if err != nil {
		return state{}, err
	}
	if epicID.Valid {
		value := int(epicID.Int64)
		item.EpicID = &value
	}
	return item, nil
}

func finishMutation(ctx context.Context, tx pgx.Tx, id int) (dto.Change, error) {
	change, err := getChange(ctx, tx, id)
	if err != nil {
		return dto.Change{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return dto.Change{}, err
	}
	return change, nil
}

func (r *Repo) updateField(ctx context.Context, id int, unchanged func(state) bool, query string, args ...any) (dto.Change, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.Change{}, err
	}
	defer tx.Rollback(ctx)

	current, err := getState(ctx, tx, id)
	if err != nil {
		return dto.Change{}, err
	}
	if unchanged(current) {
		return finishMutation(ctx, tx, id)
	}
	if _, err := tx.Exec(ctx, "call public.sp_change_to_history($1, false)", id); err != nil {
		return dto.Change{}, err
	}
	queryArgs := append([]any{id}, args...)
	tag, err := tx.Exec(ctx, query, queryArgs...)
	if err != nil {
		return dto.Change{}, err
	}
	if tag.RowsAffected() == 0 {
		return dto.Change{}, ErrNotFound
	}
	return finishMutation(ctx, tx, id)
}

type testCaseRef struct {
	ID int
}

func testCasesForChange(ctx context.Context, tx pgx.Tx, changeID int) ([]testCaseRef, error) {
	rows, err := tx.Query(ctx, "select id from public.test_case where change_id = $1", changeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]testCaseRef, 0)
	for rows.Next() {
		var item testCaseRef
		if err := rows.Scan(&item.ID); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func recalculateEpics(ctx context.Context, tx pgx.Tx, values ...*int) error {
	seen := make(map[int]struct{}, len(values))
	for _, value := range values {
		if value == nil {
			continue
		}
		if _, ok := seen[*value]; ok {
			continue
		}
		seen[*value] = struct{}{}
		if _, err := tx.Exec(ctx, "call public.sp_epic_test_case_recalculate($1)", *value); err != nil {
			return err
		}
	}
	return nil
}

func equalIntPointers(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

type queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}
