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
		List(ctx context.Context, projectID int) ([]dto.ChangeListItem, error)
		Get(ctx context.Context, id int) (dto.ChangeDetail, error)
		Bodies(ctx context.Context, ids []int) ([]dto.Change, error)
		Create(ctx context.Context, req dto.ChangeCreateRequest) (dto.Change, error)
		UpdateChangeTypes(ctx context.Context, req dto.ChangeUpdateChangeTypesRequest) (dto.Change, error)
		UpdateTitle(ctx context.Context, req dto.ChangeUpdateTitleRequest) (dto.Change, error)
		UpdateBody(ctx context.Context, req dto.ChangeUpdateBodyRequest) (dto.Change, error)
		UpdatePRBody(ctx context.Context, req dto.ChangeUpdatePRBodyRequest) (dto.Change, error)
		UpdatePRUrl(ctx context.Context, req dto.ChangeUpdatePRUrlRequest) (dto.Change, error)
		UpdateAgentEdit(ctx context.Context, req dto.ChangeUpdateAgentEditRequest) (dto.Change, error)
		UpdateEpic(ctx context.Context, req dto.ChangeUpdateEpicRequest) (dto.Change, error)
		UpdatePhase(ctx context.Context, req dto.ChangeUpdatePhaseRequest) (dto.Change, error)
		UpdateOpen(ctx context.Context, req dto.ChangeUpdateOpenRequest) (dto.Change, error)
		Delete(ctx context.Context, req dto.ChangeIDRequest) error
	}
)

const changeDetailColumns = `
	id,
	ref,
	version,
	slug,
	project_id,
	change_phase,
	change_types,
	epic_id,
	epic_name,
	title,
	coalesce(body, ''),
	coalesce(pr_body, ''),
	coalesce(pr_url, ''),
	agent_edit,
	open,
	done_tc,
	total_tc,
	completed,
	created,
	modified`

const changeListColumns = `
	id,
	ref,
	slug,
	project_id,
	change_phase,
	change_types,
	epic_id,
	epic_name,
	title,
	agent_edit,
	open,
	done_tc,
	total_tc,
	completed,
	modified`

// NewRepo initializes or executes NewRepo behavior.
func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// List executes List behavior.
func (r *Repo) List(ctx context.Context, projectID int) ([]dto.ChangeListItem, error) {
	rows, err := r.pool.Query(ctx, `
		select `+changeListColumns+`
		from public.vw_change_list
		where project_id = $1
		order by modified desc, id
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	changes := make([]dto.ChangeListItem, 0)
	for rows.Next() {
		change, err := scanChangeList(rows)
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
		select requested.id::integer, coalesce(c.body, ''), coalesce(c.pr_body, '')
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
		if err := rows.Scan(&change.ID, &change.Body, &change.PRBody); err != nil {
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
	`, req.ProjectID, req.ChangeTypes, req.EpicID, req.Title, req.Body).Scan(&id)
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

// UpdateBody executes UpdateBody behavior.
func (r *Repo) UpdateBody(ctx context.Context, req dto.ChangeUpdateBodyRequest) (dto.Change, error) {
	return r.updateField(ctx, req.ID, func(current state) bool {
		return current.Body == req.Body
	}, `
		update public.change
		set body = nullif($2, ''),
			version = version + 1,
			modified = now()
		where id = $1
	`, req.Body)
}

// UpdatePRBody executes UpdatePRBody behavior.
func (r *Repo) UpdatePRBody(ctx context.Context, req dto.ChangeUpdatePRBodyRequest) (dto.Change, error) {
	return r.updateField(ctx, req.ID, func(current state) bool {
		return current.PRBody == req.PRBody
	}, `
		update public.change
		set pr_body = nullif($2, ''),
			version = version + 1,
			modified = now()
		where id = $1
	`, req.PRBody)
}

// UpdatePRUrl executes UpdatePRUrl behavior.
func (r *Repo) UpdatePRUrl(ctx context.Context, req dto.ChangeUpdatePRUrlRequest) (dto.Change, error) {
	return r.updateField(ctx, req.ID, func(current state) bool {
		return current.PRUrl == req.PRUrl
	}, `
		update public.change
		set pr_url = nullif($2, ''),
			version = version + 1,
			modified = now()
		where id = $1
	`, req.PRUrl)
}

// UpdateAgentEdit executes UpdateAgentEdit behavior.
func (r *Repo) UpdateAgentEdit(ctx context.Context, req dto.ChangeUpdateAgentEditRequest) (dto.Change, error) {
	agentEdit := *req.AgentEdit
	return r.updateField(ctx, req.ID, func(current state) bool {
		return current.AgentEdit == agentEdit
	}, `
		update public.change
		set agent_edit = $2,
			version = version + 1,
			modified = now()
		where id = $1
	`, agentEdit)
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

// UpdateOpen executes UpdateOpen behavior.
func (r *Repo) UpdateOpen(ctx context.Context, req dto.ChangeUpdateOpenRequest) (dto.Change, error) {
	open := *req.Open
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.Change{}, err
	}
	defer tx.Rollback(ctx)

	current, err := getState(ctx, tx, req.ID)
	if err != nil {
		return dto.Change{}, err
	}
	if current.Open == open {
		return finishMutation(ctx, tx, req.ID)
	}
	if _, err := tx.Exec(ctx, "call public.sp_change_to_history($1, false)", req.ID); err != nil {
		return dto.Change{}, err
	}
	tag, err := tx.Exec(ctx, `
		update public.change
		set open = $2,
			version = version + 1,
			modified = now()
		where id = $1
	`, req.ID, open)
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
	change, err := scanChange(q.QueryRow(ctx, "select "+changeDetailColumns+" from public.vw_change_details where id = $1", id))
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
	var epicName pgtype.Text
	err := row.Scan(
		&change.ID, &change.Ref, &change.Version, &change.Slug, &change.ProjectID,
		&change.ChangePhase, &change.ChangeTypes, &epicID, &epicName, &change.Title,
		&change.Body, &change.PRBody, &change.PRUrl, &change.AgentEdit,
		&change.Open, &change.DoneTC,
		&change.TotalTC, &change.Completed, &change.Created, &change.Modified,
	)
	if err != nil {
		return dto.Change{}, err
	}
	if epicID.Valid {
		value := int(epicID.Int64)
		change.EpicID = &value
	}
	if epicName.Valid {
		value := epicName.String
		change.EpicName = &value
	}
	return change, nil
}

func scanChangeList(row pgx.Row) (dto.ChangeListItem, error) {
	var change dto.ChangeListItem
	var epicID pgtype.Int8
	var epicName pgtype.Text
	err := row.Scan(
		&change.ID, &change.Ref, &change.Slug, &change.ProjectID, &change.ChangePhase,
		&change.ChangeTypes, &epicID, &epicName, &change.Title, &change.AgentEdit,
		&change.Open, &change.DoneTC, &change.TotalTC, &change.Completed, &change.Modified,
	)
	if err != nil {
		return dto.ChangeListItem{}, err
	}
	if epicID.Valid {
		value := int(epicID.Int64)
		change.EpicID = &value
	}
	if epicName.Valid {
		value := epicName.String
		change.EpicName = &value
	}
	return change, nil
}

type state struct {
	ProjectID   int
	EpicID      *int
	ChangePhase string
	ChangeTypes []string
	Title       string
	Body        string
	PRBody      string
	PRUrl       string
	AgentEdit   bool
	Open        bool
}

func getState(ctx context.Context, tx pgx.Tx, id int) (state, error) {
	var item state
	var epicID pgtype.Int8
	err := tx.QueryRow(ctx, `
		select project_id, epic_id, change_phase, change_types, title, coalesce(body, ''), coalesce(pr_body, ''), coalesce(pr_url, ''), agent_edit, open
		from public.change
		where id = $1
	`, id).Scan(&item.ProjectID, &epicID, &item.ChangePhase, &item.ChangeTypes, &item.Title, &item.Body, &item.PRBody, &item.PRUrl, &item.AgentEdit, &item.Open)
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
