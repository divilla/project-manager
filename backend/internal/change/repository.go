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
	Repo struct {
		pool *pgxpool.Pool
	}

	Repository interface {
		References(ctx context.Context) (dto.ChangeReferences, error)
		List(ctx context.Context, projectID int) ([]dto.Change, error)
		Get(ctx context.Context, id int) (dto.ChangeDetail, error)
		Bodies(ctx context.Context, ids []int) ([]dto.Change, error)
		Create(ctx context.Context, req dto.ChangeCreateRequest) (dto.Change, error)
		Update(ctx context.Context, req dto.ChangeUpdateRequest) (dto.Change, error)
		UpdateEpic(ctx context.Context, req dto.ChangeUpdateEpicRequest) (dto.Change, error)
		UpdatePhase(ctx context.Context, req dto.ChangeUpdatePhaseRequest) (dto.Change, error)
		UpdateClosed(ctx context.Context, req dto.ChangeUpdateClosedRequest) (dto.Change, error)
		Delete(ctx context.Context, req dto.ChangeIDRequest) error
	}
)

const changeColumns = `
	id,
	version,
	project_id,
	epic_id,
	change_phase,
	change_types,
	title,
	coalesce(body, ''),
	codex_session_id,
	closed,
	done_req,
	total_req,
	completed,
	created,
	modified`

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

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

func (r *Repo) List(ctx context.Context, projectID int) ([]dto.Change, error) {
	rows, err := r.pool.Query(ctx, `
		select `+changeColumns+`
		from public.change
		where project_id = $1
		order by (select priority from public.change_phase where slug = change.change_phase), created, id
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

func (r *Repo) Get(ctx context.Context, id int) (dto.ChangeDetail, error) {
	change, err := getChange(ctx, r.pool, id)
	if err != nil {
		return dto.ChangeDetail{}, err
	}
	requirements, err := listRequirements(ctx, r.pool, id)
	if err != nil {
		return dto.ChangeDetail{}, err
	}
	return dto.ChangeDetail{Change: change, Requirements: requirements}, nil
}

func (r *Repo) Bodies(ctx context.Context, ids []int) ([]dto.Change, error) {
	rows, err := r.pool.Query(ctx, `
		select requested.id::integer, coalesce(c.body, '')
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
		if err := rows.Scan(&change.ID, &change.Body); err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	return changes, rows.Err()
}

func (r *Repo) Create(ctx context.Context, req dto.ChangeCreateRequest) (dto.Change, error) {
	if err := r.ensureProject(ctx, req.ProjectID); err != nil {
		return dto.Change{}, err
	}
	if err := r.ensureReference(ctx, "change_phase", req.ChangePhase); err != nil {
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
		insert into public.change (
			project_id, epic_id, change_phase, change_types, title, body, codex_session_id
		)
		values ($1, $2, $3, $4, $5, nullif($6, ''), $7)
		returning id
	`, req.ProjectID, req.EpicID, req.ChangePhase, req.ChangeTypes, req.Title, req.Body, req.CodexSessionID).Scan(&id)
	if err != nil {
		return dto.Change{}, err
	}

	return finishMutation(ctx, tx, id)
}

func (r *Repo) Update(ctx context.Context, req dto.ChangeUpdateRequest) (dto.Change, error) {
	if err := r.ensureReferences(ctx, "change_type", req.ChangeTypes); err != nil {
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
	if slices.Equal(current.ChangeTypes, req.ChangeTypes) && current.Title == req.Title && current.Body == req.Body {
		return finishMutation(ctx, tx, req.ID)
	}
	if _, err := tx.Exec(ctx, "call public.sp_change_to_history($1, false)", req.ID); err != nil {
		return dto.Change{}, err
	}
	tag, err := tx.Exec(ctx, `
		update public.change
		set title = $2,
			body = nullif($3, ''),
			change_types = $4,
			version = version + 1,
			modified = now()
		where id = $1
	`, req.ID, req.Title, req.Body, req.ChangeTypes)
	if err != nil {
		return dto.Change{}, err
	}
	if tag.RowsAffected() == 0 {
		return dto.Change{}, ErrNotFound
	}
	return finishMutation(ctx, tx, req.ID)
}

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
	requirements, err := requirementsForChange(ctx, tx, req.ID)
	if err != nil {
		return err
	}
	for _, requirement := range requirements {
		if _, err := tx.Exec(ctx, "call public.sp_requirement_to_history($1, true)", requirement.ID); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, "call public.sp_change_to_history($1, true)", req.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "delete from public.requirement where change_id = $1", req.ID); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, "delete from public.change where id = $1", req.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	if _, err := tx.Exec(ctx, "call public.sp_epic_requirement_recalculate($1)", current.EpicID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repo) referenceOptions(ctx context.Context, table string) ([]dto.ReferenceOption, error) {
	rows, err := r.pool.Query(ctx, "select slug, priority from public."+table+" order by priority, slug")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	options := make([]dto.ReferenceOption, 0)
	for rows.Next() {
		var option dto.ReferenceOption
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

func listRequirements(ctx context.Context, q queryer, changeID int) ([]dto.Requirement, error) {
	rows, err := q.Query(ctx, `
		select id, version, definition, done, change_id, created, modified
		from public.requirement
		where change_id = $1
		order by created, definition
	`, changeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	requirements := make([]dto.Requirement, 0)
	for rows.Next() {
		var item dto.Requirement
		if err := rows.Scan(&item.ID, &item.Version, &item.Definition, &item.Done, &item.ChangeID, &item.Created, &item.Modified); err != nil {
			return nil, err
		}
		requirements = append(requirements, item)
	}
	return requirements, rows.Err()
}

func scanChange(row pgx.Row) (dto.Change, error) {
	var change dto.Change
	var epicID pgtype.Int8
	var codexSessionID pgtype.Text
	err := row.Scan(
		&change.ID, &change.Version, &change.ProjectID, &epicID, &change.ChangePhase,
		&change.ChangeTypes, &change.Title, &change.Body, &codexSessionID, &change.Closed, &change.DoneReq,
		&change.TotalReq, &change.Completed, &change.Created, &change.Modified,
	)
	if err != nil {
		return dto.Change{}, err
	}
	if epicID.Valid {
		value := int(epicID.Int64)
		change.EpicID = &value
	}
	if codexSessionID.Valid {
		value := codexSessionID.String
		change.CodexSessionID = &value
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
	Closed      bool
}

func getState(ctx context.Context, tx pgx.Tx, id int) (state, error) {
	var item state
	var epicID pgtype.Int8
	err := tx.QueryRow(ctx, `
		select project_id, epic_id, change_phase, change_types, title, coalesce(body, ''), closed
		from public.change
		where id = $1
	`, id).Scan(&item.ProjectID, &epicID, &item.ChangePhase, &item.ChangeTypes, &item.Title, &item.Body, &item.Closed)
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

type requirementRef struct {
	ID int
}

func requirementsForChange(ctx context.Context, tx pgx.Tx, changeID int) ([]requirementRef, error) {
	rows, err := tx.Query(ctx, "select id from public.requirement where change_id = $1", changeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]requirementRef, 0)
	for rows.Next() {
		var item requirementRef
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
		if _, err := tx.Exec(ctx, "call public.sp_epic_requirement_recalculate($1)", *value); err != nil {
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
