package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/faizan1250/taskify/internal/handler"
	"github.com/faizan1250/taskify/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIHandler_ServeOpenAPIUI_DirectoryIndependence(t *testing.T) {
	// Ensure template is loaded strictly from embedded FS regardless of working directory
	t.Chdir(t.TempDir())

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := handler.NewOpenAPIHandler(&server.Server{})
	err := h.ServeOpenAPIUI(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	assert.Contains(t, rec.Body.String(), `<script id="api-reference" data-url="/static/openapi.json"></script>`)
}
