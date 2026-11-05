package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/faizan1250/taskify/internal/config"
	"github.com/faizan1250/taskify/internal/middleware"
	"github.com/faizan1250/taskify/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCORSMiddleware() *middleware.GlobalMiddlewares {
	cfg := &config.Config{}
	cfg.Server.CORSAllowedOrigins = []string{"https://app.example.com"}
	return middleware.NewGlobalMiddlewares(&server.Server{Config: cfg})
}

func TestCORS_PreflightAllowsAuthorizationHeader(t *testing.T) {
	e := echo.New()
	e.Use(newCORSMiddleware().CORS())
	e.GET("/todos", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/todos", nil)
	req.Header.Set(echo.HeaderOrigin, "https://app.example.com")
	req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodGet)
	req.Header.Set(echo.HeaderAccessControlRequestHeaders, "authorization")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "https://app.example.com", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.Contains(t, rec.Header().Get(echo.HeaderAccessControlAllowHeaders), echo.HeaderAuthorization)
	allowMethods := rec.Header().Get(echo.HeaderAccessControlAllowMethods)
	for _, m := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		assert.Contains(t, allowMethods, m)
	}
}

func TestCORS_PreflightRejectsDisallowedOrigin(t *testing.T) {
	e := echo.New()
	e.Use(newCORSMiddleware().CORS())
	e.GET("/todos", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/todos", nil)
	req.Header.Set(echo.HeaderOrigin, "https://evil.example.com")
	req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodGet)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
}
