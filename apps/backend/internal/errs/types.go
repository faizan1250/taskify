package errs

import (
	"net/http"
)

var (
	ErrBadRequest          = &HTTPError{Status: http.StatusBadRequest, Code: "BAD_REQUEST", Message: "Bad Request"}
	ErrUnauthorized        = &HTTPError{Status: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: "Unauthorized"}
	ErrForbidden           = &HTTPError{Status: http.StatusForbidden, Code: "FORBIDDEN", Message: "Forbidden"}
	ErrNotFound            = &HTTPError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: "Not Found"}
	ErrInternalServerError = &HTTPError{Status: http.StatusInternalServerError, Code: "INTERNAL_SERVER_ERROR", Message: "Internal Server Error"}
)

func NewUnauthorizedError(message string, override bool) *HTTPError {
	return &HTTPError{
		Code:     MakeUpperCaseWithUnderscores(http.StatusText(http.StatusUnauthorized)),
		Message:  message,
		Status:   http.StatusUnauthorized,
		Override: override,
	}
}

func NewForbiddenError(message string, override bool) *HTTPError {
	return &HTTPError{
		Code:     MakeUpperCaseWithUnderscores(http.StatusText(http.StatusForbidden)),
		Message:  message,
		Status:   http.StatusForbidden,
		Override: override,
	}
}

func NewBadRequestError(message string, override bool, code *string, errors []FieldError, action *Action) *HTTPError {
	formattedCode := MakeUpperCaseWithUnderscores(http.StatusText(http.StatusBadRequest))

	if code != nil {
		formattedCode = *code
	}

	return &HTTPError{
		Code:     formattedCode,
		Message:  message,
		Status:   http.StatusBadRequest,
		Override: override,
		Errors:   errors,
		Action:   action,
	}
}

func NewNotFoundError(message string, override bool, code *string) *HTTPError {
	formattedCode := MakeUpperCaseWithUnderscores(http.StatusText(http.StatusNotFound))

	if code != nil {
		formattedCode = *code
	}

	return &HTTPError{
		Code:     formattedCode,
		Message:  message,
		Status:   http.StatusNotFound,
		Override: override,
	}
}

func NewInternalServerError() *HTTPError {
	return &HTTPError{
		Code:     MakeUpperCaseWithUnderscores(http.StatusText(http.StatusInternalServerError)),
		Message:  http.StatusText(http.StatusInternalServerError),
		Status:   http.StatusInternalServerError,
		Override: false,
	}
}

func ValidationError(err error) *HTTPError {
	return NewBadRequestError("Validation failed: "+err.Error(), false, nil, nil, nil)
}
