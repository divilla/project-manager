package project

import (
	"context"
	"errors"

	"aipm/internal/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Repo struct {
		pool *pgxpool.Pool
	}

	Repository interface {
		List(ctx context.Context, limit, offset int) ([]dto.Project, error)
		Get(ctx context.Context, id string) (dto.Project, error)
		Create(ctx context.Context, name string) (dto.Project, error)
		Update(ctx context.Context, id, name string) (dto.Project, error)
		Delete(ctx context.Context, id string) error
	}
)

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{
		pool: pool,
	}
}

func (r *Repo) List(ctx context.Context, limit, offset int) ([]dto.Project, error) {
	ps := make([]dto.Project, 0)
	rows, err := r.pool.Query(ctx, "select id, name from project order by name limit $1 offset $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p dto.Project
		err = rows.Scan(&p.Id, &p.Name)
		if err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ps, nil
}

func (r *Repo) Get(ctx context.Context, id string) (dto.Project, error) {
	var p dto.Project
	err := r.pool.QueryRow(ctx, "select id, name from project where id = $1", id).Scan(&p.Id, &p.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Project{}, ErrNotFound
	}
	return p, err
}

func (r *Repo) Create(ctx context.Context, name string) (dto.Project, error) {
	var p dto.Project
	err := r.pool.QueryRow(ctx, "insert into project (name) values ($1) returning id, name", name).Scan(&p.Id, &p.Name)
	return p, err
}

func (r *Repo) Update(ctx context.Context, id, name string) (dto.Project, error) {
	var p dto.Project
	err := r.pool.QueryRow(ctx, "update project set name = $2 where id = $1 returning id, name", id, name).Scan(&p.Id, &p.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Project{}, ErrNotFound
	}
	return p, err
}

func (r *Repo) Delete(ctx context.Context, id string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := archiveRequirementsForProject(ctx, tx, id, true); err != nil {
		return err
	}
	if err := archiveTasksForProject(ctx, tx, id, true); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		delete from requirement
		using task
		where requirement.task_id = task.id
			and task.project_id = $1
	`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "delete from task where project_id = $1", id); err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, "delete from project where id = $1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return tx.Commit(ctx)
}

func archiveTasksForProject(ctx context.Context, tx pgx.Tx, projectID string, deleted bool) error {
	_, err := tx.Exec(ctx, `
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
		where t.project_id = $1
	`, projectID, deleted)
	return err
}

func archiveRequirementsForProject(ctx context.Context, tx pgx.Tx, projectID string, deleted bool) error {
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
		join task t on t.id = r.task_id
		where t.project_id = $1
	`, projectID, deleted)
	return err
}
