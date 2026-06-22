package requirement

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
		g: e.Group("/api").Group("/v1").Group("/requirement"),
		s: s,
	}

	a.register(a.g)
	a.register(e.Group("/api").Group("/requirement"))

	return a
}

func (a *Api) register(g *echo.Group) {
	g.POST("/list", a.listRequirements)
	g.POST("/create", a.createRequirement)
	g.POST("/update", a.updateRequirement)
	g.POST("/delete", a.deleteRequirement)
}

func (a *Api) listRequirements(c *echo.Context) error {
	var req dto.RequirementListRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid requirement list payload")
	}

	res, err := a.s.ListRequirements(c.Request().Context(), req)
	if err != nil {
		return requirementError(err)
	}

	return c.JSON(http.StatusOK, &res)
}

func (a *Api) createRequirement(c *echo.Context) error {
	var req dto.RequirementCreateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid requirement create payload")
	}

	res, err := a.s.CreateRequirement(c.Request().Context(), req)
	if err != nil {
		return requirementError(err)
	}

	return c.JSON(http.StatusCreated, &res)
}

func (a *Api) updateRequirement(c *echo.Context) error {
	var req dto.RequirementUpdateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid requirement update payload")
	}

	res, err := a.s.UpdateRequirement(c.Request().Context(), req)
	if err != nil {
		return requirementError(err)
	}

	return c.JSON(http.StatusOK, &res)
}

func (a *Api) deleteRequirement(c *echo.Context) error {
	var req dto.RequirementIDRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid requirement delete payload")
	}

	res, err := a.s.DeleteRequirement(c.Request().Context(), req)
	if err != nil {
		return requirementError(err)
	}

	return c.JSON(http.StatusOK, &res)
}

func requirementError(err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return echo.NewHTTPError(http.StatusBadRequest, "invalid requirement payload")
	case errors.Is(err, ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "requirement not found")
	default:
		return err
	}
}
