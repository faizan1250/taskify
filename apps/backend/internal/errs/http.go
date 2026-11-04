package errs

import (
	"strings"
)

type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

type ActionType string

const (
	ActionTypeRedirect ActionType = "redirect"
)

type Action struct {
	Type    ActionType `json:"type"`
	Message string     `json:"message"`
	Value   string     `json:"value"`
}

//nolint:recvcheck // HTTPError implements Error() on value to satisfy error interface for value comparison
type HTTPError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Status   int    `json:"status"`
	Override bool   `json:"override"`
	// field level errors
	Errors []FieldError `json:"errors"`
	// action to be taken
	Action *Action `json:"action"`
}

func (e HTTPError) Error() string {
	return e.Message
}

func (e *HTTPError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}

	var t *HTTPError
	switch err := target.(type) {
	case *HTTPError:
		if err == nil {
			return false
		}
		t = err
	case HTTPError:
		t = &err
	default:
		return false
	}

	if t.Status == 0 && t.Code == "" {
		return e.Status == 0 && e.Code == ""
	}

	if t.Status != 0 && e.Status != t.Status {
		return false
	}
	if t.Code != "" && e.Code != t.Code {
		return false
	}

	return true
}

func (e *HTTPError) WithMessage(message string) *HTTPError {
	return &HTTPError{
		Code:     e.Code,
		Message:  message,
		Status:   e.Status,
		Override: e.Override,
		Errors:   e.Errors,
		Action:   e.Action,
	}
}

func MakeUpperCaseWithUnderscores(str string) string {
	return strings.ToUpper(strings.ReplaceAll(str, " ", "_"))
}
