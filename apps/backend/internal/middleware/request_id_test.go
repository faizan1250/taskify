package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/faizan1250/taskify/internal/middleware"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID_ValidatesClientValue(t *testing.T) {
	tests := []struct {
		name          string
		header        string
		wantPreserved bool
	}{
		{name: "empty generates new ID", header: "", wantPreserved: false},
		{name: "valid uuid is preserved", header: "4c90fc3f-39cc-4b04-af21-c83ee64aa67e", wantPreserved: true},
		{name: "alphanumeric and dash is preserved", header: "req-123-ABC", wantPreserved: true},
		{name: "64 chars is preserved", header: strings.Repeat("b", 64), wantPreserved: true},
		{name: "over 64 chars regenerates", header: strings.Repeat("a", 65), wantPreserved: false},
		{name: "newline regenerates", header: "abc\ndef", wantPreserved: false},
		{name: "space regenerates", header: "abc def", wantPreserved: false},
		{name: "underscore regenerates", header: "abc_def", wantPreserved: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set(middleware.RequestIDHeader, tt.header)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := middleware.RequestID()(func(c echo.Context) error { return nil })(c)
			require.NoError(t, err)

			got, ok := c.Get(middleware.RequestIDKey).(string)
			require.True(t, ok, "request ID must be stored in context")
			require.NotEmpty(t, got)
			assert.Equal(t, got, rec.Header().Get(middleware.RequestIDHeader), "response header must mirror context value")

			if tt.wantPreserved {
				assert.Equal(t, tt.header, got)
			} else {
				if tt.header != "" {
					assert.NotEqual(t, tt.header, got)
				}
				_, parseErr := uuid.Parse(got)
				assert.NoError(t, parseErr, "regenerated ID must be a valid UUID")
			}
		})
	}
}
