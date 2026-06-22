package project

import (
	"errors"
	"net/http"

	"aipm/internal/dto"

	"github.com/labstack/echo/v5"
)

type Api struct {
	e *echo.Echo
	g *echo.Group
	s *Service
}

func NewAPI(e *echo.Echo, s *Service) *Api {
	a := &Api{
		e: e,
		g: e.Group("/api").Group("/v1").Group("/project"),
		s: s,
	}

	a.register(a.g)
	a.register(e.Group("/api").Group("/project"))

	return a
}

func (a *Api) register(g *echo.Group) {
	g.POST("/list", a.listProjects)
	g.POST("/get", a.getProject)
	g.POST("/create", a.createProject)
	g.POST("/update", a.updateProject)
	g.POST("/delete", a.deleteProject)
}

func (a *Api) listProjects(c *echo.Context) error {
	ctx := c.Request().Context()
	var req dto.ProjectListRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project list payload")
	}

	res, err := a.s.ListProjects(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &res)
}

func (a *Api) getProject(c *echo.Context) error {
	ctx := c.Request().Context()
	var req dto.ProjectIDRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project get payload")
	}

	res, err := a.s.GetProject(ctx, req)
	if err != nil {
		return projectError(err)
	}

	return c.JSON(http.StatusOK, &res)
}

func (a *Api) createProject(c *echo.Context) error {
	ctx := c.Request().Context()
	var req dto.ProjectCreateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project create payload")
	}

	res, err := a.s.CreateProject(ctx, req)
	if err != nil {
		return projectError(err)
	}

	return c.JSON(http.StatusCreated, &res)
}

func (a *Api) updateProject(c *echo.Context) error {
	ctx := c.Request().Context()
	var req dto.ProjectUpdateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project update payload")
	}

	res, err := a.s.UpdateProject(ctx, req)
	if err != nil {
		return projectError(err)
	}

	return c.JSON(http.StatusOK, &res)
}

func (a *Api) deleteProject(c *echo.Context) error {
	ctx := c.Request().Context()
	var req dto.ProjectIDRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project delete payload")
	}

	if err := a.s.DeleteProject(ctx, req); err != nil {
		return projectError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func projectError(err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project payload")
	case errors.Is(err, ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	default:
		return err
	}
}
