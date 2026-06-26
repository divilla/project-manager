package change

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
		g: e.Group("/api").Group("/v1").Group("/change"),
		s: s,
	}

	a.g.POST("/reference", a.references)
	a.g.POST("/list", a.listChanges)
	a.g.POST("/get", a.getChange)
	a.g.POST("/rendered-bodies", a.renderedBodies)
	a.g.POST("/create", a.createChange)
	a.g.POST("/update", a.updateChange)
	a.g.POST("/update-epic", a.updateEpic)
	a.g.POST("/update-phase", a.updatePhase)
	a.g.POST("/update-closed", a.updateClosed)
	a.g.POST("/delete", a.deleteChange)

	return a
}

func (a *Api) references(c *echo.Context) error {
	res, err := a.s.References(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, &res)
}

func (a *Api) listChanges(c *echo.Context) error {
	var req dto.ChangeListRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid change list payload")
	}
	res, err := a.s.ListChanges(c.Request().Context(), req)
	if err != nil {
		return changeError(err)
	}
	return c.JSON(http.StatusOK, &res)
}

func (a *Api) getChange(c *echo.Context) error {
	var req dto.ChangeIDRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid change get payload")
	}
	res, err := a.s.GetChange(c.Request().Context(), req)
	if err != nil {
		return changeError(err)
	}
	return c.JSON(http.StatusOK, &res)
}

func (a *Api) renderedBodies(c *echo.Context) error {
	var req dto.ChangeRenderedBodiesRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid change rendered bodies payload")
	}
	res, err := a.s.RenderedBodies(c.Request().Context(), req)
	if err != nil {
		return changeError(err)
	}
	return c.JSON(http.StatusOK, &res)
}

func (a *Api) createChange(c *echo.Context) error {
	var req dto.ChangeCreateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid change create payload")
	}
	res, err := a.s.CreateChange(c.Request().Context(), req)
	if err != nil {
		return changeError(err)
	}
	return c.JSON(http.StatusCreated, &res)
}

func (a *Api) updateChange(c *echo.Context) error {
	var req dto.ChangeUpdateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid change update payload")
	}
	res, err := a.s.UpdateChange(c.Request().Context(), req)
	if err != nil {
		return changeError(err)
	}
	return c.JSON(http.StatusOK, &res)
}

func (a *Api) updateEpic(c *echo.Context) error {
	var req dto.ChangeUpdateEpicRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid change epic payload")
	}
	res, err := a.s.UpdateEpic(c.Request().Context(), req)
	if err != nil {
		return changeError(err)
	}
	return c.JSON(http.StatusOK, &res)
}

func (a *Api) updatePhase(c *echo.Context) error {
	var req dto.ChangeUpdatePhaseRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid change phase payload")
	}
	res, err := a.s.UpdatePhase(c.Request().Context(), req)
	if err != nil {
		return changeError(err)
	}
	return c.JSON(http.StatusOK, &res)
}

func (a *Api) updateClosed(c *echo.Context) error {
	var req dto.ChangeUpdateClosedRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid change closed payload")
	}
	res, err := a.s.UpdateClosed(c.Request().Context(), req)
	if err != nil {
		return changeError(err)
	}
	return c.JSON(http.StatusOK, &res)
}

func (a *Api) deleteChange(c *echo.Context) error {
	var req dto.ChangeIDRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid change delete payload")
	}
	if err := a.s.DeleteChange(c.Request().Context(), req); err != nil {
		return changeError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func changeError(err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return echo.NewHTTPError(http.StatusBadRequest, "invalid change payload")
	case errors.Is(err, ErrInvalidReference):
		return echo.NewHTTPError(http.StatusBadRequest, "invalid change reference")
	case errors.Is(err, ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "change not found")
	default:
		return err
	}
}
