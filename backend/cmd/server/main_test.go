package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONErrorHandlerPreservesEchoStatusCode(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	jsonErrorHandler(c, echo.ErrNotFound)

	require.Equal(t, http.StatusNotFound, rec.Code)

	var body errorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	assert.Equal(t, http.StatusText(http.StatusNotFound), body.Message)
}

func TestJSONErrorHandlerDefaultsUnknownErrorsToInternalServerError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	jsonErrorHandler(c, errors.New("boom"))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestJSONErrorHandlerUsesHTTPErrorMessage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	jsonErrorHandler(c, echo.NewHTTPError(http.StatusBadRequest, "invalid payload"))

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body errorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	assert.Equal(t, "invalid payload", body.Message)
}
