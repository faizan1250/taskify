package validation

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/faizan1250/taskify/internal/errs"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type Validatable interface {
	Validate() error
}

type CustomValidationError struct {
	Field   string
	Message string
}

type CustomValidationErrors []CustomValidationError

func (c CustomValidationErrors) Error() string {
	return "Validation failed"
}

func BindAndValidate(c echo.Context, payload Validatable) error {
	if payload == nil {
		return errs.NewBadRequestError("payload cannot be nil", false, nil, nil, nil)
	}

	if err := c.Bind(payload); err != nil {
		message := extractBindErrorMessage(err)
		return errs.NewBadRequestError(message, false, nil, nil, nil)
	}

	if msg, fieldErrors := validateStruct(payload); len(fieldErrors) > 0 {
		return errs.NewBadRequestError(msg, true, nil, fieldErrors, nil)
	}

	return nil
}

func extractBindErrorMessage(err error) string {
	var echoErr *echo.HTTPError
	if errors.As(err, &echoErr) {
		message := fmt.Sprintf("%v", echoErr.Message)
		if echoErr.Internal != nil {
			return fmt.Sprintf("%s: %v", message, echoErr.Internal)
		}
		return message
	}

	if err.Error() != "" {
		return err.Error()
	}

	return "Invalid request payload"
}

func validateStruct(v Validatable) (string, []errs.FieldError) {
	if v == nil {
		return "Validation failed", []errs.FieldError{{
			Field: "payload",
			Error: "payload cannot be nil",
		}}
	}
	if err := v.Validate(); err != nil {
		return extractValidationErrors(err)
	}
	return "", nil
}

func extractValidationErrors(err error) (string, []errs.FieldError) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		fieldErrors := make([]errs.FieldError, 0, len(validationErrors))
		for _, vErr := range validationErrors {
			fieldErrors = append(fieldErrors, errs.FieldError{
				Field: strings.ToLower(vErr.Field()),
				Error: formatValidatorFieldError(vErr),
			})
		}
		return "Validation failed", fieldErrors
	}

	var customValidationErrors CustomValidationErrors
	if errors.As(err, &customValidationErrors) {
		fieldErrors := make([]errs.FieldError, 0, len(customValidationErrors))
		for _, cErr := range customValidationErrors {
			fieldErrors = append(fieldErrors, errs.FieldError{
				Field: cErr.Field,
				Error: cErr.Message,
			})
		}
		return "Validation failed", fieldErrors
	}

	return "Validation failed", []errs.FieldError{
		{
			Field: "payload",
			Error: err.Error(),
		},
	}
}

func formatValidatorFieldError(err validator.FieldError) string {
	field := strings.ToLower(err.Field())

	switch err.Tag() {
	case "required":
		return "is required"
	case "min":
		if err.Type() != nil && err.Type().Kind() == reflect.String {
			return fmt.Sprintf("must be at least %s characters", err.Param())
		}
		return fmt.Sprintf("must be at least %s", err.Param())
	case "max":
		if err.Type() != nil && err.Type().Kind() == reflect.String {
			return fmt.Sprintf("must not exceed %s characters", err.Param())
		}
		return fmt.Sprintf("must not exceed %s", err.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", err.Param())
	case "email":
		return "must be a valid email address"
	case "e164":
		return "must be a valid phone number with country code"
	case "uuid":
		return "must be a valid UUID"
	case "uuidList":
		return "must be a comma-separated list of valid UUIDs"
	case "dive":
		return "some items are invalid"
	default:
		if err.Param() != "" {
			return fmt.Sprintf("%s: %s:%s", field, err.Tag(), err.Param())
		}
		return fmt.Sprintf("%s: %s", field, err.Tag())
	}
}

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func IsValidUUID(uuid string) bool {
	return uuidRegex.MatchString(uuid)
}
