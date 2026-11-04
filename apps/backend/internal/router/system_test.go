package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/faizan1250/taskify/internal/handler"
	"github.com/faizan1250/taskify/internal/router"
	"github.com/faizan1250/taskify/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSystemRoutes_DirectoryIndependence(t *testing.T) {
	// Ensure routes and static files are served correctly from embedded FS regardless of working directory
	t.Chdir(t.TempDir())

	e := echo.New()
	h := &handler.Handlers{
		OpenAPI: handler.NewOpenAPIHandler(&server.Server{}),
	}
	router.ExportRegisterSystemRoutes(e, h)

	t.Run("serve openapi docs", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `<script id="api-reference" data-url="/static/openapi.json"></script>`)
	})

	t.Run("serve static openapi json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/static/openapi.json", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "openapi")
	})

	t.Run("serve static openapi html", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/static/openapi.html", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "API Reference")
	})

	t.Run("static 404 for missing file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/static/nonexistent.js", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
