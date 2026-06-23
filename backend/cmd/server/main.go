package main

import (
	"context"
	"errors"
	"net/http"
	"os"

	"aipm/internal/health"
	"aipm/internal/project"
	"aipm/internal/requirement"
	"aipm/internal/task"
	"aipm/internal/taskview"
	"aipm/pkg/config"
	"aipm/pkg/db"
	"aipm/pkg/markdown"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	config.New()
	cfg := config.Get()

	ctx := context.Background()
	pool := db.Pool(ctx, cfg.ConnectionString)
	defer pool.Close()

	defaultCORSConfig := middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: cfg.AllowedOrigins(),
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAccessControlAllowOrigin, echo.HeaderAccessControlAllowMethods},
	}
	e := echo.New()
	e.HTTPErrorHandler = jsonErrorHandler
	e.Pre(middleware.RemoveTrailingSlash())

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	log.Logger = logger
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info().
				Str("URI", v.URI).
				Int("status", v.Status).
				Msg("request")
			return nil
		},
	}))

	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(defaultCORSConfig))

	markdownParser := markdown.NewGoldmarkParser()
	htmlSanitizer := markdown.NewBluemondaySanitizer()
	taskRenderer := taskview.NewTaskRenderer(markdownParser, htmlSanitizer)

	healthRepository := health.NewRepo(pool)
	healthService := health.NewService(healthRepository)
	health.NewAPI(e, healthService)

	projectRepository := project.NewRepo(pool)
	projectService := project.NewService(projectRepository)
	project.NewAPI(e, projectService)

	taskRepository := task.NewRepo(pool)
	taskService := task.NewService(taskRepository, taskRenderer)
	task.NewAPI(e, taskService)

	requirementRepository := requirement.NewRepo(pool)
	requirementService := requirement.NewService(requirementRepository, taskRenderer)
	requirement.NewAPI(e, requirementService)

	if err := e.Start(cfg.Addr()); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error().Err(err).Msg("server stopped")
	}
}

type errorResponse struct {
	Message string `json:"message"`
}

func jsonErrorHandler(c *echo.Context, err error) {
	code := http.StatusInternalServerError
	message := http.StatusText(code)

	if status := echo.StatusCode(err); status != 0 {
		code = status
		message = http.StatusText(code)
	}

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if he.Message != "" {
			message = he.Message
		}
	}

	if code >= http.StatusInternalServerError {
		log.Error().Err(err).Msg("request failed")
	}

	if writeErr := c.JSON(code, errorResponse{Message: message}); writeErr != nil {
		log.Error().Err(writeErr).Msg("failed to write error response")
	}
}
