package task

import (
	"context"
	"errors"

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
		References(ctx context.Context) (dto.TaskReferences, error)
		List(ctx context.Context, projectID string) ([]dto.Task, error)
		Get(ctx context.Context, id string) (dto.TaskDetail, error)
		Create(ctx context.Context, req dto.TaskCreateRequest) (dto.Task, error)
		Update(ctx context.Context, req dto.TaskUpdateRequest) (dto.Task, error)
		ChangePhase(ctx context.Context, id, phase string) (dto.Task, error)
		Delete(ctx context.Context, id string) error
	}
)

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{
		pool: pool,
	}
}

func (r *Repo) References(ctx context.Context) (dto.TaskReferences, error) {
	phases, err := r.referenceOptions(ctx, "task_phase")
	if err != nil {
		return dto.TaskReferences{}, err
	}
	types, err := r.referenceOptions(ctx, "task_type")
	if err != nil {
		return dto.TaskReferences{}, err
	}

	return dto.TaskReferences{
		Phases: phases,
		Types:  types,
	}, nil
}

func (r *Repo) List(ctx context.Context, projectID string) ([]dto.Task, error) {
	rows, err := r.pool.Query(ctx, `
		select
			t.id,
			t.project_id,
			t.parent_id,
			t.task_phase,
			t.task_type,
			t.name,
			coalesce(t.description, ''),
			t.difficulty,
			t.complete,
			t.priority,
			t.depth,
			t.created,
			t.modified
		from task t
		join task_phase tp on tp.slug = t.task_phase
		join task_type tt on tt.slug = t.task_type
		where t.project_id = $1
		order by tp.priority, t.priority, t.created
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]dto.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *Repo) Get(ctx context.Context, id string) (dto.TaskDetail, error) {
	task, err := r.getTask(ctx, r.pool, id)
	if err != nil {
		return dto.TaskDetail{}, err
	}

	requirements, err := r.requirements(ctx, r.pool, id)
	if err != nil {
		return dto.TaskDetail{}, err
	}

	return dto.TaskDetail{
		Task:         task,
		Requirements: requirements,
	}, nil
}

func (r *Repo) Create(ctx context.Context, req dto.TaskCreateRequest) (dto.Task, error) {
	if err := r.ensureProject(ctx, req.ProjectID); err != nil {
		return dto.Task{}, err
	}
	if req.Phase != "" {
		if err := r.ensureReference(ctx, "task_phase", req.Phase); err != nil {
			return dto.Task{}, err
		}
	}
	if req.Type != "" {
		if err := r.ensureReference(ctx, "task_type", req.Type); err != nil {
			return dto.Task{}, err
		}
	}
	if req.ParentID != "" {
		if err := r.ensureParentTask(ctx, req.ProjectID, req.ParentID); err != nil {
			return dto.Task{}, err
		}
	}

	return r.insertTask(ctx, req)
}

func (r *Repo) Update(ctx context.Context, req dto.TaskUpdateRequest) (dto.Task, error) {
	if req.Type != "" {
		if err := r.ensureReference(ctx, "task_type", req.Type); err != nil {
			return dto.Task{}, err
		}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.Task{}, err
	}
	defer tx.Rollback(ctx)

	if err := archiveTask(ctx, tx, req.ID, false); err != nil {
		return dto.Task{}, err
	}

	var task dto.Task
	row := tx.QueryRow(ctx, `
		update task
		set
			name = $2,
			description = nullif($3, ''),
			task_type = coalesce(nullif($4, ''), task_type),
			difficulty = case when $5 > 0 then $5 else difficulty end,
			priority = $6,
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
	`, req.ID, req.Name, req.Description, req.Type, req.Difficulty, req.Priority)
	task, err = scanTask(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Task{}, ErrNotFound
	}
	if err != nil {
		return dto.Task{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.Task{}, err
	}

	return task, nil
}

func (r *Repo) ChangePhase(ctx context.Context, id, phase string) (dto.Task, error) {
	if err := r.ensureReference(ctx, "task_phase", phase); err != nil {
		return dto.Task{}, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dto.Task{}, err
	}
	defer tx.Rollback(ctx)

	if err := archiveTask(ctx, tx, id, false); err != nil {
		return dto.Task{}, err
	}

	var task dto.Task
	row := tx.QueryRow(ctx, `
		update task
		set task_phase = $2, modified = now()
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
	`, id, phase)
	task, err = scanTask(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Task{}, ErrNotFound
	}
	if err != nil {
		return dto.Task{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.Task{}, err
	}

	return task, nil
}

func (r *Repo) Delete(ctx context.Context, id string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	taskIDs, err := taskTreeIDs(ctx, tx, id)
	if err != nil {
		return err
	}
	if len(taskIDs) == 0 {
		return ErrNotFound
	}

	if err := archiveRequirementsForTasks(ctx, tx, taskIDs, true); err != nil {
		return err
	}
	if err := archiveTasks(ctx, tx, taskIDs, true); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "delete from requirement where task_id = any($1::uuid[])", taskIDs); err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, "delete from task where id = any($1::uuid[])", taskIDs)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return tx.Commit(ctx)
}

func (r *Repo) referenceOptions(ctx context.Context, table string) ([]dto.ReferenceOption, error) {
	query := "select slug, priority from " + table + " order by priority, slug"
	rows, err := r.pool.Query(ctx, query)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return options, nil
}

func (r *Repo) ensureReference(ctx context.Context, table, slug string) error {
	query := "select exists(select 1 from " + table + " where slug = $1)"
	var exists bool
	if err := r.pool.QueryRow(ctx, query, slug).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrInvalidReference
	}
	return nil
}

func (r *Repo) ensureProject(ctx context.Context, id string) error {
	var exists bool
	if err := r.pool.QueryRow(ctx, "select exists(select 1 from project where id = $1)", id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrInvalidReference
	}
	return nil
}

func (r *Repo) ensureParentTask(ctx context.Context, projectID, parentID string) error {
	var exists bool
	if err := r.pool.QueryRow(ctx, `
		select exists(
			select 1
			from task
			where id = $1
				and project_id = $2
		)
	`, parentID, projectID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrInvalidReference
	}
	return nil
}

func (r *Repo) insertTask(ctx context.Context, req dto.TaskCreateRequest) (dto.Task, error) {
	columns := []string{"project_id", "name"}
	args := []any{req.ProjectID, req.Name}
	placeholders := []string{"$1", "$2"}

	if req.Description != "" {
		args = append(args, req.Description)
		columns = append(columns, "description")
		placeholders = append(placeholders, "$"+itoa(len(args)))
	}
	if req.Phase != "" {
		args = append(args, req.Phase)
		columns = append(columns, "task_phase")
		placeholders = append(placeholders, "$"+itoa(len(args)))
	}
	if req.Type != "" {
		args = append(args, req.Type)
		columns = append(columns, "task_type")
		placeholders = append(placeholders, "$"+itoa(len(args)))
	}
	if req.Difficulty > 0 {
		args = append(args, req.Difficulty)
		columns = append(columns, "difficulty")
		placeholders = append(placeholders, "$"+itoa(len(args)))
	}
	if req.Priority != 0 {
		args = append(args, req.Priority)
		columns = append(columns, "priority")
		placeholders = append(placeholders, "$"+itoa(len(args)))
	}
	if req.ParentID != "" {
		args = append(args, req.ParentID)
		columns = append(columns, "parent_id")
		placeholders = append(placeholders, "$"+itoa(len(args)))
	}

	query := `
		insert into task (` + join(columns) + `)
		values (` + join(placeholders) + `)
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
	`

	task, err := scanTask(r.pool.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Task{}, ErrNotFound
	}
	return task, err
}

func (r *Repo) getTask(ctx context.Context, q queryer, id string) (dto.Task, error) {
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

func (r *Repo) requirements(ctx context.Context, q queryer, taskID string) ([]dto.Requirement, error) {
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
		var requirement dto.Requirement
		if err := rows.Scan(
			&requirement.ID,
			&requirement.TaskID,
			&requirement.Definition,
			&requirement.Done,
			&requirement.Created,
			&requirement.Modified,
		); err != nil {
			return nil, err
		}
		requirements = append(requirements, requirement)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requirements, nil
}

type queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
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

func archiveRequirementsForTask(ctx context.Context, tx pgx.Tx, taskID string, deleted bool) error {
	_, err := tx.Exec(ctx, `
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
		where r.task_id = $1
	`, taskID, deleted)
	return err
}

func taskTreeIDs(ctx context.Context, tx pgx.Tx, taskID string) ([]string, error) {
	rows, err := tx.Query(ctx, `
		with recursive task_tree as (
			select id
			from task
			where id = $1
			union all
			select child.id
			from task child
			join task_tree parent on child.parent_id = parent.id
		)
		select id
		from task_tree
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func archiveTasks(ctx context.Context, tx pgx.Tx, taskIDs []string, deleted bool) error {
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
		where t.id = any($1::uuid[])
	`, taskIDs, deleted)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func archiveRequirementsForTasks(ctx context.Context, tx pgx.Tx, taskIDs []string, deleted bool) error {
	_, err := tx.Exec(ctx, `
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
		where r.task_id = any($1::uuid[])
	`, taskIDs, deleted)
	return err
}

func join(values []string) string {
	if len(values) == 0 {
		return ""
	}
	result := values[0]
	for _, value := range values[1:] {
		result += ", " + value
	}
	return result
}

func itoa(value int) string {
	const digits = "0123456789"
	if value == 0 {
		return "0"
	}
	var out [20]byte
	i := len(out)
	for value > 0 {
		i--
		out[i] = digits[value%10]
		value /= 10
	}
	return string(out[i:])
}
