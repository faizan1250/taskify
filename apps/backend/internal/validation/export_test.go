package validation

import (
	"github.com/faizan1250/taskify/internal/errs"
	"github.com/go-playground/validator/v10"
)

func ExportValidateStruct(v Validatable) (string, []errs.FieldError) {
	return validateStruct(v)
}

func ExportExtractValidationErrors(err error) (string, []errs.FieldError) {
	return extractValidationErrors(err)
}

func ExportFormatValidatorFieldError(err validator.FieldError) string {
	return formatValidatorFieldError(err)
}
