package utility

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

func ParseValidationError(err error) *APIError {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return NewBadRequestError(err.Error(), "BAD_REQUEST")
	}

	details := make([]ValidationError, len(ve))
	for i, fe := range ve {
		details[i] = ValidationError{
			Field:   fe.Field(),
			Message: validationMessage(fe),
		}
	}

	return NewValidationError("Validasi gagal", "VALIDATION_ERROR", details)
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " wajib diisi"
	case "email":
		return fe.Field() + " format email tidak valid"
	case "min":
		return fe.Field() + " minimal " + fe.Param() + " karakter"
	case "max":
		return fe.Field() + " maksimal " + fe.Param() + " karakter"
	case "uuid":
		return fe.Field() + " harus berupa UUID"
	default:
		return fe.Field() + " tidak valid"
	}
}
