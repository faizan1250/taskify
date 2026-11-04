package errs_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/faizan1250/taskify/internal/errs"
)

func TestHTTPError_Is(t *testing.T) {
	t.Parallel()

	customCode := "CUSTOM_NOT_FOUND"

	tests := []struct {
		name     string
		err      error
		target   error
		expected bool
	}{
		{
			name:     "matching status and code returns true",
			err:      errs.ErrNotFound,
			target:   errs.ErrNotFound,
			expected: true,
		},
		{
			name:     "matching status and code from constructor returns true",
			err:      errs.NewNotFoundError("resource missing", false, nil),
			target:   errs.ErrNotFound,
			expected: true,
		},
		{
			name:     "matching status with custom code when target has status only returns true",
			err:      errs.NewNotFoundError("custom not found", false, &customCode),
			target:   &errs.HTTPError{Status: http.StatusNotFound},
			expected: true,
		},
		{
			name:     "matching status with default code against status-only target returns true",
			err:      errs.ErrBadRequest,
			target:   &errs.HTTPError{Status: http.StatusBadRequest},
			expected: true,
		},
		{
			name:     "different status returns false (ErrNotFound vs ErrBadRequest)",
			err:      errs.ErrNotFound,
			target:   errs.ErrBadRequest,
			expected: false,
		},
		{
			name:     "different status returns false (500 vs 404)",
			err:      errs.ErrInternalServerError,
			target:   errs.ErrNotFound,
			expected: false,
		},
		{
			name:     "different code returns false",
			err:      errs.NewNotFoundError("custom not found", false, &customCode),
			target:   errs.ErrNotFound,
			expected: false,
		},
		{
			name:     "matching code when target has code only returns true",
			err:      errs.NewNotFoundError("custom not found", false, &customCode),
			target:   &errs.HTTPError{Code: customCode},
			expected: true,
		},
		{
			name:     "target is nil returns false",
			err:      errs.ErrBadRequest,
			target:   nil,
			expected: false,
		},
		{
			name:     "target is typed nil returns false",
			err:      errs.ErrBadRequest,
			target:   (*errs.HTTPError)(nil),
			expected: false,
		},
		{
			name:     "target is non-HTTPError returns false",
			err:      errs.ErrBadRequest,
			target:   errors.New("standard error"),
			expected: false,
		},
		{
			name:     "target passed by value matches",
			err:      errs.ErrNotFound,
			target:   *errs.ErrNotFound,
			expected: true,
		},
		{
			name:     "target passed by value with status only matches",
			err:      errs.ErrNotFound,
			target:   errs.HTTPError{Status: http.StatusNotFound},
			expected: true,
		},
		{
			name:     "target passed by value with different status returns false",
			err:      errs.ErrNotFound,
			target:   *errs.ErrBadRequest,
			expected: false,
		},
		{
			name:     "wrapped error matches via errors.Is",
			err:      fmt.Errorf("wrapped context: %w", errs.ErrNotFound),
			target:   errs.ErrNotFound,
			expected: true,
		},
		{
			name:     "deeply wrapped error matches via errors.Is",
			err:      fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", errs.ErrForbidden)),
			target:   errs.ErrForbidden,
			expected: true,
		},
		{
			name:     "uninitialized empty target does not match initialized error",
			err:      errs.ErrBadRequest,
			target:   &errs.HTTPError{},
			expected: false,
		},
		{
			name:     "uninitialized empty target matches uninitialized empty error",
			err:      &errs.HTTPError{},
			target:   &errs.HTTPError{},
			expected: true,
		},
		{
			name:     "uninitialized empty value target matches uninitialized empty error",
			err:      &errs.HTTPError{},
			target:   errs.HTTPError{},
			expected: true,
		},
		{
			name:     "WithMessage preserves matching",
			err:      errs.ErrUnauthorized.WithMessage("session expired"),
			target:   errs.ErrUnauthorized,
			expected: true,
		},
		{
			name:     "standard ErrUnauthorized matches NewUnauthorizedError",
			err:      errs.NewUnauthorizedError("please login", false),
			target:   errs.ErrUnauthorized,
			expected: true,
		},
		{
			name:     "standard ErrForbidden matches NewForbiddenError",
			err:      errs.NewForbiddenError("access denied", false),
			target:   errs.ErrForbidden,
			expected: true,
		},
		{
			name:     "standard ErrInternalServerError matches NewInternalServerError",
			err:      errs.NewInternalServerError(),
			target:   errs.ErrInternalServerError,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, errors.Is(tt.err, tt.target))
		})
	}
}

func TestHTTPError_Is_NilReceiver(t *testing.T) {
	t.Parallel()

	var nilErr *errs.HTTPError
	assert.False(t, nilErr.Is(errs.ErrBadRequest))
	assert.False(t, errors.Is(nilErr, errs.ErrBadRequest))
}

func TestHTTPError_Error(t *testing.T) {
	t.Parallel()

	err := &errs.HTTPError{Message: "something went wrong"}
	assert.Equal(t, "something went wrong", err.Error())

	errVal := errs.HTTPError{Message: "value error message"}
	assert.Equal(t, "value error message", errVal.Error())
}
