package middleware_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/faizan1250/taskify/internal/errs"
	"github.com/faizan1250/taskify/internal/middleware"
	"github.com/faizan1250/taskify/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAuthMiddleware() *middleware.AuthMiddleware {
	logger := zerolog.New(io.Discard)
	return middleware.NewAuthMiddleware(&server.Server{Logger: &logger})
}

func authContextWithClaims(t *testing.T, claims *clerk.SessionClaims) echo.Context {
	t.Helper()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if claims != nil {
		req = req.WithContext(clerk.ContextWithSessionClaims(req.Context(), claims))
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func requireUnauthorized(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var httpErr *errs.HTTPError
	require.True(t, errors.As(err, &httpErr), "expected *errs.HTTPError, got %T", err)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Status)
}

func TestRequireAuth_RejectsMissingClaims(t *testing.T) {
	auth := newAuthMiddleware()
	c := authContextWithClaims(t, nil)

	nextCalled := false
	err := auth.RequireAuth(func(c echo.Context) error {
		nextCalled = true
		return nil
	})(c)

	requireUnauthorized(t, err)
	assert.False(t, nextCalled)
}

func TestRequireAuth_RejectsEmptySubject(t *testing.T) {
	auth := newAuthMiddleware()
	c := authContextWithClaims(t, &clerk.SessionClaims{})

	nextCalled := false
	err := auth.RequireAuth(func(c echo.Context) error {
		nextCalled = true
		return nil
	})(c)

	requireUnauthorized(t, err)
	assert.False(t, nextCalled)
	assert.Empty(t, c.Get("user_id"))
}

func TestRequireAuth_AcceptsValidSubject(t *testing.T) {
	auth := newAuthMiddleware()
	claims := &clerk.SessionClaims{}
	claims.Subject = "user_123"
	c := authContextWithClaims(t, claims)

	nextCalled := false
	err := auth.RequireAuth(func(c echo.Context) error {
		nextCalled = true
		assert.Equal(t, "user_123", c.Get("user_id"))
		return nil
	})(c)

	require.NoError(t, err)
	assert.True(t, nextCalled)
}
