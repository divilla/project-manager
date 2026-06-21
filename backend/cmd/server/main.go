package main

import (
	"context"
	"net/http"

	"github.com/divilla/project-manager/internal/project"
	"github.com/divilla/project-manager/pkg/config"
	"github.com/divilla/project-manager/pkg/db"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	config.New()

	ctx := context.Background()
	pool := db.Pool(ctx, config.Get().ConnectionString)
	defer pool.Close()

	defaultCORSConfig := middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAccessControlAllowOrigin, echo.HeaderAccessControlAllowMethods},
	}
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(defaultCORSConfig))

	projectRepository := project.NewRepository(pool)
	projectService := project.NewService(projectRepository)
	project.NewAPI(e, projectService)

	e.Logger.Error(e.Start(":8080").Error())
}
