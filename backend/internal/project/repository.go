package project

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
		List(ctx context.Context, limit, offset int) ([]dto.Project, error)
		Get(ctx context.Context, id int) (dto.Project, error)
		Create(ctx context.Context, name string) (dto.Project, error)
		Update(ctx context.Context, id int, name string) (dto.Project, error)
		Delete(ctx context.Context, id int) error
	}
)

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) List(ctx context.Context, limit, offset int) ([]dto.Project, error) {
	rows, err := r.pool.Query(ctx, "select id, name from public.project order by name limit $1 offset $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := make([]dto.Project, 0)
	for rows.Next() {
		var project dto.Project
		if err := rows.Scan(&project.ID, &project.Name); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, rows.Err()
}

func (r *Repo) Get(ctx context.Context, id int) (dto.Project, error) {
	var project dto.Project
	err := r.pool.QueryRow(ctx, "select id, name from public.project where id = $1", id).Scan(&project.ID, &project.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Project{}, ErrNotFound
	}
	return project, err
}

func (r *Repo) Create(ctx context.Context, name string) (dto.Project, error) {
	var project dto.Project
	err := r.pool.QueryRow(ctx, "insert into public.project (name) values ($1) returning id, name", name).Scan(&project.ID, &project.Name)
	return project, err
}

func (r *Repo) Update(ctx context.Context, id int, name string) (dto.Project, error) {
	var project dto.Project
	err := r.pool.QueryRow(ctx, "update public.project set name = $2 where id = $1 returning id, name", id, name).Scan(&project.ID, &project.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Project{}, ErrNotFound
	}
	return project, err
}

func (r *Repo) Delete(ctx context.Context, id int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tasks, err := projectTasks(ctx, tx, id)
	if err != nil {
		return err
	}
	taskIDs := make([]int, 0, len(tasks))
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID)
	}
	requirements, err := projectRequirements(ctx, tx, id)
	if err != nil {
		return err
	}
	for _, requirement := range requirements {
		if _, err := tx.Exec(ctx, "call public.sp_requirement_to_history($1, true)", requirement.ID); err != nil {
			return err
		}
	}
	for _, taskID := range taskIDs {
		if _, err := tx.Exec(ctx, "call public.sp_task_to_history($1, true)", taskID); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `
		delete from public.requirement
		using public.task
		where requirement.task_id = task.id and task.project_id = $1
	`, id); err != nil {
		return err
	}
	for _, taskID := range uniqueTaskIDs(requirements) {
		if _, err := tx.Exec(ctx, "call public.sp_task_requirement_recalculate($1)", taskID); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, "delete from public.task where project_id = $1", id); err != nil {
		return err
	}
	for _, task := range tasks {
		if _, err := tx.Exec(ctx, "call public.sp_task_phase_recalculate($1)", task.ParentID); err != nil {
			return err
		}
	}
	tag, err := tx.Exec(ctx, "delete from public.project where id = $1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}

type taskRef struct {
	ID       int
	ParentID *int
}

func projectTasks(ctx context.Context, tx pgx.Tx, projectID int) ([]taskRef, error) {
	rows, err := tx.Query(ctx, "select id, parent_id from public.task where project_id = $1", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := make([]taskRef, 0)
	for rows.Next() {
		var task taskRef
		var parentID pgtype.Int8
		if err := rows.Scan(&task.ID, &parentID); err != nil {
			return nil, err
		}
		if parentID.Valid {
			value := int(parentID.Int64)
			task.ParentID = &value
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

type requirementRef struct {
	ID     int
	TaskID int
}

func projectRequirements(ctx context.Context, tx pgx.Tx, projectID int) ([]requirementRef, error) {
	rows, err := tx.Query(ctx, `
		select requirement.id, requirement.task_id
		from public.requirement
		join public.task on task.id = requirement.task_id
		where task.project_id = $1
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]requirementRef, 0)
	for rows.Next() {
		var item requirementRef
		if err := rows.Scan(&item.ID, &item.TaskID); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func uniqueTaskIDs(requirements []requirementRef) []int {
	seen := make(map[int]struct{})
	ids := make([]int, 0)
	for _, requirement := range requirements {
		if _, ok := seen[requirement.TaskID]; ok {
			continue
		}
		seen[requirement.TaskID] = struct{}{}
		ids = append(ids, requirement.TaskID)
	}
	return ids
}
