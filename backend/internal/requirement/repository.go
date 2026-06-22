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

const taskColumns = `
	id,
	version,
	task_type,
	name,
	coalesce(description, ''),
	difficulty,
	priority,
	task_phase,
	parent_id,
	project_id,
	done_req,
	total_req,
	completed,
	created,
	modified`

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) List(ctx context.Context, taskID int) ([]dto.Requirement, error) {
	if err := ensureTaskExists(ctx, r.pool, taskID); err != nil {
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

	if err := ensureTaskExists(ctx, tx, req.TaskID); err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	requirement, err := scanRequirement(tx.QueryRow(ctx, `
		insert into public.requirement (definition, task_id)
		values ($1, $2)
		returning id, version, definition, done, task_id, created, modified
	`, req.Definition, req.TaskID))
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	return finishMutation(ctx, tx, requirement.TaskID, []int{requirement.TaskID}, &requirement)
}

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
		return finishMutation(ctx, tx, current.TaskID, nil, &current)
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
		returning id, version, definition, done, task_id, created, modified
	`, req.ID, req.Definition))
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.RequirementMutationResponse{}, ErrNotFound
	}
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	return finishMutation(ctx, tx, current.TaskID, nil, &requirement)
}

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
		return finishMutation(ctx, tx, current.TaskID, nil, &current)
	}
	requirement, err := scanRequirement(tx.QueryRow(ctx, `
		update public.requirement
		set done = $2, modified = now()
		where id = $1
		returning id, version, definition, done, task_id, created, modified
	`, req.ID, req.Done))
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	return finishMutation(ctx, tx, current.TaskID, []int{current.TaskID}, &requirement)
}

func (r *Repo) UpdateTask(ctx context.Context, req dto.RequirementUpdateTaskRequest) (dto.RequirementMutationResponse, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	defer tx.Rollback(ctx)
	current, err := getRequirement(ctx, tx, req.ID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	if err := ensureTaskExists(ctx, tx, req.TaskID); err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	if req.TaskID == current.TaskID {
		return finishMutation(ctx, tx, current.TaskID, nil, &current)
	}
	requirement, err := scanRequirement(tx.QueryRow(ctx, `
		update public.requirement
		set task_id = $2, modified = now()
		where id = $1
		returning id, version, definition, done, task_id, created, modified
	`, req.ID, req.TaskID))
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	return finishMutation(ctx, tx, req.TaskID, []int{current.TaskID, req.TaskID}, &requirement)
}

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
	return finishMutation(ctx, tx, current.TaskID, []int{current.TaskID}, nil)
}

func finishMutation(ctx context.Context, tx pgx.Tx, responseTaskID int, affectedTaskIDs []int, requirement *dto.Requirement) (dto.RequirementMutationResponse, error) {
	for _, taskID := range affectedTaskIDs {
		if _, err := tx.Exec(ctx, "call public.sp_task_requirement_recalculate($1)", taskID); err != nil {
			return dto.RequirementMutationResponse{}, err
		}
	}
	task, err := getTask(ctx, tx, responseTaskID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	requirements, err := listRequirements(ctx, tx, responseTaskID)
	if err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return dto.RequirementMutationResponse{}, err
	}
	return dto.RequirementMutationResponse{Requirement: requirement, Task: task, Requirements: requirements}, nil
}

func getRequirement(ctx context.Context, tx pgx.Tx, id int) (dto.Requirement, error) {
	requirement, err := scanRequirement(tx.QueryRow(ctx, `
		select id, version, definition, done, task_id, created, modified
		from public.requirement where id = $1
	`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Requirement{}, ErrNotFound
	}
	if err != nil {
		return dto.Requirement{}, err
	}
	return requirement, nil
}

func ensureTaskExists(ctx context.Context, q queryer, id int) error {
	var exists bool
	if err := q.QueryRow(ctx, "select exists(select 1 from public.task where id = $1)", id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}

func listRequirements(ctx context.Context, q queryer, taskID int) ([]dto.Requirement, error) {
	rows, err := q.Query(ctx, `
		select id, version, definition, done, task_id, created, modified
		from public.requirement where task_id = $1 order by created, definition
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
	return requirements, rows.Err()
}

func scanRequirement(row pgx.Row) (dto.Requirement, error) {
	var requirement dto.Requirement
	err := row.Scan(
		&requirement.ID, &requirement.Version, &requirement.Definition, &requirement.Done,
		&requirement.TaskID, &requirement.Created, &requirement.Modified,
	)
	return requirement, err
}

func getTask(ctx context.Context, q queryer, id int) (dto.Task, error) {
	task, err := scanTask(q.QueryRow(ctx, "select "+taskColumns+" from public.vw_task where id = $1", id))
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Task{}, ErrNotFound
	}
	return task, err
}

func scanTask(row pgx.Row) (dto.Task, error) {
	var task dto.Task
	var parentID pgtype.Int8
	err := row.Scan(
		&task.ID, &task.Version, &task.TaskType, &task.Name, &task.Description,
		&task.Difficulty, &task.Priority, &task.TaskPhase, &parentID, &task.ProjectID,
		&task.DoneReq, &task.TotalReq, &task.Completed, &task.Created, &task.Modified,
	)
	if err != nil {
		return dto.Task{}, err
	}
	if parentID.Valid {
		value := int(parentID.Int64)
		task.ParentID = &value
	}
	return task, nil
}

type queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}
