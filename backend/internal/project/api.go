package project

import (
	"net/http"

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

	a.g.POST("/list", a.listProjects)

	return a
}

func (a *Api) listProjects(c *echo.Context) error {
	ctx := c.Request().Context()

	res, err := a.s.ListProjects(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &res)
}
