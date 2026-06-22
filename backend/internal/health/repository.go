package health

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Repo struct {
		pool *pgxpool.Pool
	}

	Repository interface {
		Ping(ctx context.Context) error
	}
)

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{
		pool: pool,
	}
}

func (r *Repo) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
