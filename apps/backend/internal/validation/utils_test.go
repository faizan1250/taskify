package validation_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/faizan1250/taskify/internal/errs"
	"github.com/faizan1250/taskify/internal/validation"
)

type mockBinder struct {
	err error
}

func (m *mockBinder) Bind(_ interface{}, _ echo.Context) error {
	return m.err
}

type validPayload struct {
	Name string `json:"name"`
}

func (v *validPayload) Validate() error {
	return nil
}

type genericErrPayload struct{}

func (g *genericErrPayload) Validate() error {
	return errors.New("custom generic validation failure")
}

type customErrPayload struct{}

func (c *customErrPayload) Validate() error {
	return validation.CustomValidationErrors{
		{
			Field:   "custom_field",
			Message: "custom error occurred",
		},
	}
}

type comprehensiveStructPayload struct {
	RequiredStr string `validate:"required"`
	MinStr      string `validate:"min=5"`
	MaxStr      string `validate:"max=3"`
	MinNum      int    `validate:"min=10"`
	MaxNum      int    `validate:"max=100"`
	OneOfStr    string `validate:"oneof=red green blue"`
	EmailStr    string `validate:"email"`
	PhoneStr    string `validate:"e164"`
	UUIDStr     string `validate:"uuid"`
	UUIDListStr string `validate:"uuidList"`
	CustomTag   string `validate:"alphanum"`
}

func (s *comprehensiveStructPayload) Validate() error {
	v := validator.New()
	_ = v.RegisterValidation("uuidList", func(fl validator.FieldLevel) bool {
		return false
	})
	return v.Struct(s)
}

func newEchoContext(body string, contentType string) (echo.Context, *echo.Echo) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	if contentType != "" {
		req.Header.Set(echo.HeaderContentType, contentType)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), e
}

func TestBindAndValidate_NilPayload(t *testing.T) {
	c, _ := newEchoContext(`{"name":"test"}`, echo.MIMEApplicationJSON)

	err := validation.BindAndValidate(c, nil)
	require.Error(t, err)

	var httpErr *errs.HTTPError
	require.True(t, errors.As(err, &httpErr))
	assert.Equal(t, http.StatusBadRequest, httpErr.Status)
	assert.Equal(t, "payload cannot be nil", httpErr.Message)
	assert.False(t, httpErr.Override)
}

func TestBindAndValidate_BindingErrors(t *testing.T) {
	tests := []struct {
		name            string
		binderErr       error
		expectedMessage string
	}{
		{
			name:            "standard error without comma or message=",
			binderErr:       io.EOF,
			expectedMessage: "EOF",
		},
		{
			name:            "error with multiple commas",
			binderErr:       errors.New("invalid syntax, line 10, col 5"),
			expectedMessage: "invalid syntax, line 10, col 5",
		},
		{
			name:            "echo HTTPError with string message",
			binderErr:       echo.NewHTTPError(http.StatusBadRequest, "Malformed JSON input"),
			expectedMessage: "Malformed JSON input",
		},
		{
			name: "echo HTTPError with internal error",
			binderErr: &echo.HTTPError{
				Code:     http.StatusBadRequest,
				Message:  "Syntax error",
				Internal: errors.New("unexpected end of JSON input"),
			},
			expectedMessage: "Syntax error: unexpected end of JSON input",
		},
		{
			name: "echo HTTPError with non-string message",
			binderErr: &echo.HTTPError{
				Code:    http.StatusBadRequest,
				Message: 40012,
			},
			expectedMessage: "40012",
		},
		{
			name:            "empty error message fallback",
			binderErr:       errors.New(""),
			expectedMessage: "Invalid request payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, e := newEchoContext("", echo.MIMEApplicationJSON)
			e.Binder = &mockBinder{err: tt.binderErr}

			payload := &validPayload{}
			err := validation.BindAndValidate(c, payload)
			require.Error(t, err)

			var httpErr *errs.HTTPError
			require.True(t, errors.As(err, &httpErr))
			assert.Equal(t, http.StatusBadRequest, httpErr.Status)
			assert.Equal(t, tt.expectedMessage, httpErr.Message)
			assert.False(t, httpErr.Override)
		})
	}
}

func TestBindAndValidate_RealEchoJSONBindingError(t *testing.T) {
	c, _ := newEchoContext("{not valid json", echo.MIMEApplicationJSON)
	payload := &validPayload{}

	err := validation.BindAndValidate(c, payload)
	require.Error(t, err)

	var httpErr *errs.HTTPError
	require.True(t, errors.As(err, &httpErr))
	assert.Equal(t, http.StatusBadRequest, httpErr.Status)
	assert.Contains(t, httpErr.Message, "Syntax error")
}

func TestBindAndValidate_GenericValidationErrorNotSwallowed(t *testing.T) {
	c, _ := newEchoContext("{}", echo.MIMEApplicationJSON)
	payload := &genericErrPayload{}

	err := validation.BindAndValidate(c, payload)
	require.Error(t, err, "generic validation errors must not be silently swallowed")

	var httpErr *errs.HTTPError
	require.True(t, errors.As(err, &httpErr))
	assert.Equal(t, http.StatusBadRequest, httpErr.Status)
	assert.Equal(t, "Validation failed", httpErr.Message)
	assert.True(t, httpErr.Override)
	require.Len(t, httpErr.Errors, 1)
	assert.Equal(t, "payload", httpErr.Errors[0].Field)
	assert.Equal(t, "custom generic validation failure", httpErr.Errors[0].Error)
}

func TestBindAndValidate_CustomValidationErrors(t *testing.T) {
	c, _ := newEchoContext("{}", echo.MIMEApplicationJSON)
	payload := &customErrPayload{}

	err := validation.BindAndValidate(c, payload)
	require.Error(t, err)

	var httpErr *errs.HTTPError
	require.True(t, errors.As(err, &httpErr))
	assert.Equal(t, http.StatusBadRequest, httpErr.Status)
	assert.Equal(t, "Validation failed", httpErr.Message)
	assert.True(t, httpErr.Override)
	require.Len(t, httpErr.Errors, 1)
	assert.Equal(t, "custom_field", httpErr.Errors[0].Field)
	assert.Equal(t, "custom error occurred", httpErr.Errors[0].Error)
}

func TestBindAndValidate_ValidatorTags(t *testing.T) {
	c, _ := newEchoContext("{}", echo.MIMEApplicationJSON)
	payload := &comprehensiveStructPayload{
		RequiredStr: "",
		MinStr:      "abc",
		MaxStr:      "toolong",
		MinNum:      5,
		MaxNum:      150,
		OneOfStr:    "yellow",
		EmailStr:    "notanemail",
		PhoneStr:    "12345",
		UUIDStr:     "not-a-uuid",
		UUIDListStr: "invalid",
		CustomTag:   "!!!",
	}

	err := validation.BindAndValidate(c, payload)
	require.Error(t, err)

	var httpErr *errs.HTTPError
	require.True(t, errors.As(err, &httpErr))
	assert.Equal(t, http.StatusBadRequest, httpErr.Status)
	assert.Equal(t, "Validation failed", httpErr.Message)
	assert.True(t, httpErr.Override)

	errMap := make(map[string]string)
	for _, fe := range httpErr.Errors {
		errMap[fe.Field] = fe.Error
	}

	assert.Equal(t, "is required", errMap["requiredstr"])
	assert.Equal(t, "must be at least 5 characters", errMap["minstr"])
	assert.Equal(t, "must not exceed 3 characters", errMap["maxstr"])
	assert.Equal(t, "must be at least 10", errMap["minnum"])
	assert.Equal(t, "must not exceed 100", errMap["maxnum"])
	assert.Equal(t, "must be one of: red green blue", errMap["oneofstr"])
	assert.Equal(t, "must be a valid email address", errMap["emailstr"])
	assert.Equal(t, "must be a valid phone number with country code", errMap["phonestr"])
	assert.Equal(t, "must be a valid UUID", errMap["uuidstr"])
	assert.Equal(t, "must be a comma-separated list of valid UUIDs", errMap["uuidliststr"])
	assert.Equal(t, "customtag: alphanum", errMap["customtag"])
}

func TestBindAndValidate_Success(t *testing.T) {
	c, _ := newEchoContext(`{"name":"Taskify"}`, echo.MIMEApplicationJSON)
	payload := &validPayload{}

	err := validation.BindAndValidate(c, payload)
	require.NoError(t, err)
	assert.Equal(t, "Taskify", payload.Name)
}

func TestValidateStruct_Nil(t *testing.T) {
	msg, fieldErrors := validation.ExportValidateStruct(nil)
	assert.Equal(t, "Validation failed", msg)
	require.Len(t, fieldErrors, 1)
	assert.Equal(t, "payload", fieldErrors[0].Field)
	assert.Equal(t, "payload cannot be nil", fieldErrors[0].Error)
}

func TestValidateStruct_Success(t *testing.T) {
	msg, fieldErrors := validation.ExportValidateStruct(&validPayload{Name: "ok"})
	assert.Empty(t, msg)
	assert.Empty(t, fieldErrors)
}

type stubFieldError struct {
	field string
	tag   string
	param string
	typ   reflect.Type
	kind  reflect.Kind
}

func (s stubFieldError) Tag() string                      { return s.tag }
func (s stubFieldError) ActualTag() string                { return s.tag }
func (s stubFieldError) Namespace() string                { return s.field }
func (s stubFieldError) StructNamespace() string          { return s.field }
func (s stubFieldError) Field() string                    { return s.field }
func (s stubFieldError) StructField() string              { return s.field }
func (s stubFieldError) Value() interface{}               { return nil }
func (s stubFieldError) Param() string                    { return s.param }
func (s stubFieldError) Kind() reflect.Kind               { return s.kind }
func (s stubFieldError) Type() reflect.Type               { return s.typ }
func (s stubFieldError) Translate(_ ut.Translator) string { return "" }
func (s stubFieldError) Error() string                    { return s.tag }

func TestFormatValidatorFieldError_EdgeCasesAndNilType(t *testing.T) {
	tests := []struct {
		name     string
		err      validator.FieldError
		expected string
	}{
		{
			name: "dive tag",
			err: stubFieldError{
				field: "items",
				tag:   "dive",
			},
			expected: "some items are invalid",
		},
		{
			name: "min tag with nil Type (avoids panic)",
			err: stubFieldError{
				field: "age",
				tag:   "min",
				param: "18",
				typ:   nil,
			},
			expected: "must be at least 18",
		},
		{
			name: "max tag with nil Type (avoids panic)",
			err: stubFieldError{
				field: "age",
				tag:   "max",
				param: "65",
				typ:   nil,
			},
			expected: "must not exceed 65",
		},
		{
			name: "default tag with param",
			err: stubFieldError{
				field: "custom",
				tag:   "customRule",
				param: "foo",
			},
			expected: "custom: customRule:foo",
		},
		{
			name: "default tag without param",
			err: stubFieldError{
				field: "custom",
				tag:   "customRule",
			},
			expected: "custom: customRule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				msg := validation.ExportFormatValidatorFieldError(tt.err)
				assert.Equal(t, tt.expected, msg)
			})
		})
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		uuid  string
		valid bool
	}{
		{"123e4567-e89b-12d3-a456-426614174000", true},
		{"c9b4e857-8df7-4648-936d-96ce1922c0f6", true},
		{"C9B4E857-8DF7-4648-936D-96CE1922C0F6", true},
		{"invalid-uuid", false},
		{"", false},
		{"123e4567-e89b-12d3-a456-42661417400", false},
		{"123e4567-e89b-12d3-a456-4266141740000", false},
		{"123e4567_e89b_12d3_a456_426614174000", false},
	}

	for _, tt := range tests {
		t.Run(tt.uuid, func(t *testing.T) {
			assert.Equal(t, tt.valid, validation.IsValidUUID(tt.uuid))
		})
	}
}
