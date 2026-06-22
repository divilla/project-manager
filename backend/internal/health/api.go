package health

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
		g: e.Group("/api").Group("/v1").Group("/health"),
		s: s,
	}

	a.g.GET("", a.check)
	e.GET("/api/health", a.check)

	return a
}

func (a *Api) check(c *echo.Context) error {
	ctx := c.Request().Context()

	res := a.s.Check(ctx)
	if res.Status != "ok" {
		return c.JSON(http.StatusServiceUnavailable, &res)
	}

	return c.JSON(http.StatusOK, &res)
}
