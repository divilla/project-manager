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

	a.g.POST("/list", a.listRequirements)
	a.g.POST("/create", a.createRequirement)
	a.g.POST("/update", a.updateRequirement)
	a.g.POST("/update-done", a.updateRequirementDone)
	a.g.POST("/update-task", a.updateRequirementTask)
	a.g.POST("/delete", a.deleteRequirement)

	return a
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

func (a *Api) updateRequirementDone(c *echo.Context) error {
	var req dto.RequirementUpdateDoneRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid requirement done payload")
	}
	res, err := a.s.UpdateRequirementDone(c.Request().Context(), req)
	if err != nil {
		return requirementError(err)
	}
	return c.JSON(http.StatusOK, &res)
}

func (a *Api) updateRequirementTask(c *echo.Context) error {
	var req dto.RequirementUpdateChangeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid requirement task payload")
	}
	res, err := a.s.UpdateRequirementTask(c.Request().Context(), req)
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
