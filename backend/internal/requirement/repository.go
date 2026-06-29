package requirement

import (
	"context"
	"errors"

	"aipm/internal/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo defines Repo values.
type Repo struct {
	pool *pgxpool.Pool
}

const changeColumns = `
	id,
	version,
	project_id,
	epic_id,
	change_phase,
	change_types,
	title,
	coalesce(body, ''),
	closed,
	done_req,
	total_req,
	completed,
	created,
	modified`

// NewRepo initializes or executes NewRepo behavior.
func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// List executes List behavior.
func (r *Repo) List(ctx context.Context, changeID int) ([]dto.Requirement, error) {
	if err := ensureChangeExists(ctx, r.pool, changeID); err != nil {
		return nil, err
	}
	return listRequirements(ctx, r.pool, changeID)
}

// Create executes Create behavior.
func (r *Repo) Create(ctx context.Context, req dto.RequirementCreateRequest) (dto.RequirementMutationResponse, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	defer tx.Rollback(ctx)

	if err := ensureChangeExists(ctx, tx, req.ChangeID); err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	requirement, err := scanRequirement(tx.QueryRow(ctx, `
		insert into public.requirement (definition, change_id)
		values ($1, $2)
		returning id, version, definition, done, change_id, created, modified
	`, req.Definition, req.ChangeID))
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	return finishMutation(ctx, tx, requirement.ChangeID, []int{requirement.ChangeID}, &requirement)
}

// Update executes Update behavior.
func (r *Repo) Update(ctx context.Context, req dto.RequirementUpdateRequest) (dto.RequirementMutationResponse, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	defer tx.Rollback(ctx)

	current, err := getRequirement(ctx, tx, req.ID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	if req.Definition == current.Definition {
		return finishMutation(ctx, tx, current.ChangeID, nil, &current)
	}
	if _, err := tx.Exec(ctx, "call public.sp_requirement_to_history($1, false)", req.ID); err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	requirement, err := scanRequirement(tx.QueryRow(ctx, `
		update public.requirement
		set definition = $2,
			version = version + 1,
			modified = now()
		where id = $1
		returning id, version, definition, done, change_id, created, modified
	`, req.ID, req.Definition))
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.RequirementMutationResponse{}, ErrNotFound
	}
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	return finishMutation(ctx, tx, current.ChangeID, nil, &requirement)
}

// UpdateDone executes UpdateDone behavior.
func (r *Repo) UpdateDone(ctx context.Context, req dto.RequirementUpdateDoneRequest) (dto.RequirementMutationResponse, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	defer tx.Rollback(ctx)

	current, err := getRequirement(ctx, tx, req.ID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	if req.Done == current.Done {
		return finishMutation(ctx, tx, current.ChangeID, nil, &current)
	}
	requirement, err := scanRequirement(tx.QueryRow(ctx, `
		update public.requirement
		set done = $2,
			modified = now()
		where id = $1
		returning id, version, definition, done, change_id, created, modified
	`, req.ID, req.Done))
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	return finishMutation(ctx, tx, current.ChangeID, []int{current.ChangeID}, &requirement)
}

// UpdateChange executes UpdateChange behavior.
func (r *Repo) UpdateChange(ctx context.Context, req dto.RequirementUpdateChangeRequest) (dto.RequirementMutationResponse, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	defer tx.Rollback(ctx)

	current, err := getRequirement(ctx, tx, req.ID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	if err := ensureChangeExists(ctx, tx, req.ChangeID); err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	if req.ChangeID == current.ChangeID {
		return finishMutation(ctx, tx, current.ChangeID, nil, &current)
	}
	requirement, err := scanRequirement(tx.QueryRow(ctx, `
		update public.requirement
		set change_id = $2,
			modified = now()
		where id = $1
		returning id, version, definition, done, change_id, created, modified
	`, req.ID, req.ChangeID))
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	return finishMutation(ctx, tx, req.ChangeID, []int{current.ChangeID, req.ChangeID}, &requirement)
}

// Delete executes Delete behavior.
func (r *Repo) Delete(ctx context.Context, req dto.RequirementIDRequest) (dto.RequirementMutationResponse, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	defer tx.Rollback(ctx)

	current, err := getRequirement(ctx, tx, req.ID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	if _, err := tx.Exec(ctx, "call public.sp_requirement_to_history($1, true)", req.ID); err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	tag, err := tx.Exec(ctx, "delete from public.requirement where id = $1", req.ID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	if tag.RowsAffected() == 0 {
		return dto.RequirementMutationResponse{}, ErrNotFound
	}
	return finishMutation(ctx, tx, current.ChangeID, []int{current.ChangeID}, nil)
}

func finishMutation(ctx context.Context, tx pgx.Tx, responseChangeID int, affectedChangeIDs []int, requirement *dto.Requirement) (dto.RequirementMutationResponse, error) {
	for _, changeID := range uniqueIDs(affectedChangeIDs) {
		if _, err := tx.Exec(ctx, "call public.sp_change_requirement_recalculate($1)", changeID); err != nil {
			return dto.RequirementMutationResponse{}, err
		}
	}
	change, err := getChange(ctx, tx, responseChangeID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	requirements, err := listRequirements(ctx, tx, responseChangeID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	return dto.RequirementMutationResponse{Requirement: requirement, Change: change, Requirements: requirements}, nil
}

func getRequirement(ctx context.Context, tx pgx.Tx, id int) (dto.Requirement, error) {
	requirement, err := scanRequirement(tx.QueryRow(ctx, `
		select id, version, definition, done, change_id, created, modified
		from public.requirement
		where id = $1
	`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Requirement{}, ErrNotFound
	}
	if err != nil {
		return dto.Requirement{}, err
	}
	return requirement, nil
}

func ensureChangeExists(ctx context.Context, q queryer, id int) error {
	var exists bool
	if err := q.QueryRow(ctx, "select exists(select 1 from public.change where id = $1)", id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return nil
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
		requirement, err := scanRequirement(rows)
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, requirement)
	}
	return requirements, rows.Err()
}

func scanRequirement(row pgx.Row) (dto.Requirement, error) {
	var requirement dto.Requirement
	err := row.Scan(
		&requirement.ID, &requirement.Version, &requirement.Definition, &requirement.Done,
		&requirement.ChangeID, &requirement.Created, &requirement.Modified,
	)
	return requirement, err
}

func getChange(ctx context.Context, q queryer, id int) (dto.Change, error) {
	change, err := scanChange(q.QueryRow(ctx, "select "+changeColumns+" from public.change where id = $1", id))
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Change{}, ErrNotFound
	}
	return change, err
}

func scanChange(row pgx.Row) (dto.Change, error) {
	var change dto.Change
	var epicID pgtype.Int8
	err := row.Scan(
		&change.ID, &change.Version, &change.ProjectID, &epicID, &change.ChangePhase,
		&change.ChangeTypes, &change.Title, &change.Body, &change.Closed, &change.DoneReq,
		&change.TotalReq, &change.Completed, &change.Created, &change.Modified,
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

func uniqueIDs(values []int) []int {
	seen := make(map[int]struct{}, len(values))
	result := make([]int, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

type queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}
