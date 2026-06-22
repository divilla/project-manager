package requirement

import (
	"context"
	"errors"

	"aipm/internal/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{
		pool: pool,
	}
}

func (r *Repo) List(ctx context.Context, taskID string) ([]dto.Requirement, error) {
	if _, err := getTask(ctx, r.pool, taskID); err != nil {
		return nil, err
	}

	return listRequirements(ctx, r.pool, taskID)
}

func (r *Repo) Create(ctx context.Context, req dto.RequirementCreateRequest) (dto.RequirementMutationResponse, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := getTask(ctx, tx, req.TaskID); err != nil {
		return dto.RequirementMutationResponse{}, err
	}

	requirement, err := scanRequirement(tx.QueryRow(ctx, `
		insert into requirement (definition, task_id)
		values ($1, $2)
		returning id, task_id, definition, done, created, modified
	`, req.Definition, req.TaskID))
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}

	task, err := recalculateTaskCompleteness(ctx, tx, req.TaskID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}

	requirements, err := listRequirements(ctx, tx, req.TaskID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.RequirementMutationResponse{}, err
	}

	return dto.RequirementMutationResponse{
		Requirement:  &requirement,
		Task:         task,
		Requirements: requirements,
	}, nil
}

func (r *Repo) Update(ctx context.Context, req dto.RequirementUpdateRequest) (dto.RequirementMutationResponse, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	defer tx.Rollback(ctx)

	current, err := getRequirementForUpdate(ctx, tx, req.ID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}

	if err := archiveRequirement(ctx, tx, req.ID, false); err != nil {
		return dto.RequirementMutationResponse{}, err
	}

	done := current.Done
	if req.Done != nil {
		done = *req.Done
	}

	requirement, err := scanRequirement(tx.QueryRow(ctx, `
		update requirement
		set definition = $2,
			done = $3,
			modified = now()
		where id = $1
		returning id, task_id, definition, done, created, modified
	`, req.ID, req.Definition, done))
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.RequirementMutationResponse{}, ErrNotFound
	}
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}

	task, err := recalculateTaskCompleteness(ctx, tx, current.TaskID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}

	requirements, err := listRequirements(ctx, tx, current.TaskID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.RequirementMutationResponse{}, err
	}

	return dto.RequirementMutationResponse{
		Requirement:  &requirement,
		Task:         task,
		Requirements: requirements,
	}, nil
}

func (r *Repo) Delete(ctx context.Context, id string) (dto.RequirementMutationResponse, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	defer tx.Rollback(ctx)

	current, err := getRequirementForUpdate(ctx, tx, id)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}

	if err := archiveRequirement(ctx, tx, id, true); err != nil {
		return dto.RequirementMutationResponse{}, err
	}

	tag, err := tx.Exec(ctx, "delete from requirement where id = $1", id)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	if tag.RowsAffected() == 0 {
		return dto.RequirementMutationResponse{}, ErrNotFound
	}

	task, err := recalculateTaskCompleteness(ctx, tx, current.TaskID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}

	requirements, err := listRequirements(ctx, tx, current.TaskID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.RequirementMutationResponse{}, err
	}

	return dto.RequirementMutationResponse{
		Task:         task,
		Requirements: requirements,
	}, nil
}

func getRequirement(ctx context.Context, q queryer, id string) (dto.Requirement, error) {
	requirement, err := scanRequirement(q.QueryRow(ctx, `
		select id, task_id, definition, done, created, modified
		from requirement
		where id = $1
	`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Requirement{}, ErrNotFound
	}
	return requirement, err
}

func getRequirementForUpdate(ctx context.Context, tx pgx.Tx, id string) (dto.Requirement, error) {
	requirement, err := scanRequirement(tx.QueryRow(ctx, `
		select id, task_id, definition, done, created, modified
		from requirement
		where id = $1
		for update
	`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Requirement{}, ErrNotFound
	}
	return requirement, err
}

func listRequirements(ctx context.Context, q queryer, taskID string) ([]dto.Requirement, error) {
	rows, err := q.Query(ctx, `
		select id, task_id, definition, done, created, modified
		from requirement
		where task_id = $1
		order by created, definition
	`, taskID)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requirements, nil
}

func scanRequirement(row pgx.Row) (dto.Requirement, error) {
	var requirement dto.Requirement
	err := row.Scan(
		&requirement.ID,
		&requirement.TaskID,
		&requirement.Definition,
		&requirement.Done,
		&requirement.Created,
		&requirement.Modified,
	)
	if err != nil {
		return dto.Requirement{}, err
	}

	return requirement, nil
}

func archiveRequirement(ctx context.Context, tx pgx.Tx, id string, deleted bool) error {
	tag, err := tx.Exec(ctx, `
		insert into requirement_history (
			id,
			version,
			definition,
			done,
			task_id,
			created,
			deleted
		)
		select
			r.id,
			coalesce((select max(rh.version) + 1 from requirement_history rh where rh.id = r.id), 1),
			r.definition,
			r.done,
			r.task_id,
			r.created,
			$2
		from requirement r
		where r.id = $1
	`, id, deleted)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func recalculateTaskCompleteness(ctx context.Context, tx pgx.Tx, taskID string) (dto.Task, error) {
	task, err := getTaskForUpdate(ctx, tx, taskID)
	if err != nil {
		return dto.Task{}, err
	}

	var complete int
	if err := tx.QueryRow(ctx, `
		select
			case
				when count(*) = 0 then 0
				else (count(*) filter (where done) * 100 / count(*))::int
			end
		from requirement
		where task_id = $1
	`, taskID).Scan(&complete); err != nil {
		return dto.Task{}, err
	}

	if task.Complete == complete {
		return task, nil
	}

	if err := archiveTask(ctx, tx, taskID, false); err != nil {
		return dto.Task{}, err
	}

	task, err = scanTask(tx.QueryRow(ctx, `
		update task
		set complete = $2,
			modified = now()
		where id = $1
		returning
			id,
			project_id,
			parent_id,
			task_phase,
			task_type,
			name,
			coalesce(description, ''),
			difficulty,
			complete,
			priority,
			depth,
			created,
			modified
	`, taskID, complete))
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Task{}, ErrNotFound
	}

	return task, err
}

func getTask(ctx context.Context, q queryer, id string) (dto.Task, error) {
	task, err := scanTask(q.QueryRow(ctx, `
		select
			id,
			project_id,
			parent_id,
			task_phase,
			task_type,
			name,
			coalesce(description, ''),
			difficulty,
			complete,
			priority,
			depth,
			created,
			modified
		from task
		where id = $1
	`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Task{}, ErrNotFound
	}

	return task, err
}

func getTaskForUpdate(ctx context.Context, tx pgx.Tx, id string) (dto.Task, error) {
	task, err := scanTask(tx.QueryRow(ctx, `
		select
			id,
			project_id,
			parent_id,
			task_phase,
			task_type,
			name,
			coalesce(description, ''),
			difficulty,
			complete,
			priority,
			depth,
			created,
			modified
		from task
		where id = $1
		for update
	`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Task{}, ErrNotFound
	}

	return task, err
}

func scanTask(row pgx.Row) (dto.Task, error) {
	var task dto.Task
	var parentID pgtype.UUID
	err := row.Scan(
		&task.ID,
		&task.ProjectID,
		&parentID,
		&task.Phase,
		&task.Type,
		&task.Name,
		&task.Description,
		&task.Difficulty,
		&task.Complete,
		&task.Priority,
		&task.Depth,
		&task.Created,
		&task.Modified,
	)
	if err != nil {
		return dto.Task{}, err
	}
	if parentID.Valid {
		value := parentID.String()
		task.ParentID = &value
	}

	return task, nil
}

func archiveTask(ctx context.Context, tx pgx.Tx, id string, deleted bool) error {
	tag, err := tx.Exec(ctx, `
		insert into task_history (
			id,
			version,
			task_phase,
			task_type,
			name,
			description,
			difficulty,
			complete,
			parent_id,
			project_id,
			priority,
			depth,
			created,
			deleted
		)
		select
			t.id,
			coalesce((select max(th.version) + 1 from task_history th where th.id = t.id), 1),
			t.task_phase,
			t.task_type,
			t.name,
			t.description,
			t.difficulty,
			t.complete,
			t.parent_id,
			t.project_id,
			t.priority,
			t.depth,
			t.created,
			$2
		from task t
		where t.id = $1
	`, id, deleted)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

type queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}
