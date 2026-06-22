package task

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
		g: e.Group("/api").Group("/v1").Group("/task"),
		s: s,
	}

	a.g.POST("/reference", a.references)
	a.g.POST("/list", a.listTasks)
	a.g.POST("/get", a.getTask)
	a.g.POST("/create", a.createTask)
	a.g.POST("/update", a.updateTask)
	a.g.POST("/update-difficulty", a.updateDifficulty)
	a.g.POST("/update-priority", a.updatePriority)
	a.g.POST("/update-parent", a.updateParent)
	a.g.POST("/update-phase", a.updatePhase)
	a.g.POST("/delete", a.deleteTask)

	return a
}

func (a *Api) references(c *echo.Context) error {
	res, err := a.s.References(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, &res)
}

func (a *Api) listTasks(c *echo.Context) error {
	var req dto.TaskListRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task list payload")
	}

	res, err := a.s.ListTasks(c.Request().Context(), req)
	if err != nil {
		return taskError(err)
	}

	return c.JSON(http.StatusOK, &res)
}

func (a *Api) getTask(c *echo.Context) error {
	var req dto.TaskIDRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task get payload")
	}

	res, err := a.s.GetTask(c.Request().Context(), req)
	if err != nil {
		return taskError(err)
	}

	return c.JSON(http.StatusOK, &res)
}

func (a *Api) createTask(c *echo.Context) error {
	var req dto.TaskCreateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task create payload")
	}

	res, err := a.s.CreateTask(c.Request().Context(), req)
	if err != nil {
		return taskError(err)
	}

	return c.JSON(http.StatusCreated, &res)
}

func (a *Api) updateTask(c *echo.Context) error {
	var req dto.TaskUpdateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task update payload")
	}

	res, err := a.s.UpdateTask(c.Request().Context(), req)
	if err != nil {
		return taskError(err)
	}

	return c.JSON(http.StatusOK, &res)
}

func (a *Api) updateDifficulty(c *echo.Context) error {
	var req dto.TaskUpdateDifficultyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task difficulty payload")
	}
	res, err := a.s.UpdateDifficulty(c.Request().Context(), req)
	if err != nil {
		return taskError(err)
	}
	return c.JSON(http.StatusOK, &res)
}

func (a *Api) updatePriority(c *echo.Context) error {
	var req dto.TaskUpdatePriorityRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task priority payload")
	}
	res, err := a.s.UpdatePriority(c.Request().Context(), req)
	if err != nil {
		return taskError(err)
	}
	return c.JSON(http.StatusOK, &res)
}

func (a *Api) updateParent(c *echo.Context) error {
	var req dto.TaskUpdateParentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task parent payload")
	}
	res, err := a.s.UpdateParent(c.Request().Context(), req)
	if err != nil {
		return taskError(err)
	}
	return c.JSON(http.StatusOK, &res)
}

func (a *Api) updatePhase(c *echo.Context) error {
	var req dto.TaskUpdatePhaseRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task phase payload")
	}

	res, err := a.s.UpdatePhase(c.Request().Context(), req)
	if err != nil {
		return taskError(err)
	}

	return c.JSON(http.StatusOK, &res)
}

func (a *Api) deleteTask(c *echo.Context) error {
	var req dto.TaskIDRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task delete payload")
	}

	if err := a.s.DeleteTask(c.Request().Context(), req); err != nil {
		return taskError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func taskError(err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task payload")
	case errors.Is(err, ErrInvalidReference):
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task reference")
	case errors.Is(err, ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	default:
		return err
	}
}
