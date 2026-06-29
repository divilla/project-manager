package change

import (
	"errors"
	"net/http"

	"aipm/internal/dto"

	"github.com/labstack/echo/v5"
)

// API defines API values.
type API struct {
	e *echo.Echo
	g *echo.Group
	s *Service
}

// NewAPI initializes or executes NewAPI behavior.
func NewAPI(e *echo.Echo, s *Service) *API {
	a := &API{
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

func (a *API) references(c *echo.Context) error {
	res, err := a.s.References(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, &res)
}

func (a *API) listChanges(c *echo.Context) error {
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

func (a *API) getChange(c *echo.Context) error {
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

func (a *API) renderedBodies(c *echo.Context) error {
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

func (a *API) createChange(c *echo.Context) error {
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

func (a *API) updateChange(c *echo.Context) error {
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

func (a *API) updateEpic(c *echo.Context) error {
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

func (a *API) updatePhase(c *echo.Context) error {
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

func (a *API) updateClosed(c *echo.Context) error {
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

func (a *API) deleteChange(c *echo.Context) error {
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
