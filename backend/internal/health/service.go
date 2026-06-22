package health

import (
	"context"
	"time"

	"aipm/internal/dto"

	"github.com/rs/zerolog/log"
)

type (
	Service struct {
		repo *Repository
	}
)

func NewService(healthRepository *Repository) *Service {
	return &Service{
		repo: healthRepository,
	}
}

func (s *Service) Check(ctx context.Context) dto.Health {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	res := dto.Health{
		Status:   "ok",
		API:      "ok",
		Database: "ok",
	}

	if err := s.repo.Ping(ctx); err != nil {
		log.Warn().Err(err).Msg("database health check failed")
		res.Status = "degraded"
		res.Database = "error"
		res.Error = "database unavailable"
	}

	return res
}
