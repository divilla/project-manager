package project

import (
	"context"

	"aipm/internal/dto"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
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
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p dto.Project
		err = rows.Scan(&p.Id, &p.Name)
		if err != nil {
			log.Error().Err(err).Msg("failed to scan project")
		}
		ps = append(ps, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ps, nil
}
