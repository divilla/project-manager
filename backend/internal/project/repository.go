package project

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/divilla/project-manager/internal/dto"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]dto.Project, error) {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var ps []dto.Project
	rows, err := conn.Query(ctx, "select * from project order by name limit $1 offset $2", limit, offset)
	fmt.Println(rows, err)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p dto.Project
		err = rows.Scan(&p.Id, &p.Name)
		if err != nil {
			slog.Error("failed to scan product search", slog.String("error", err.Error()))
		}
		ps = append(ps, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ps, nil
}
