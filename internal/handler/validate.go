package handler

import (
	"fmt"
	"strings"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/go-playground/validator/v10"
)

var v = validator.New()

func validate(input any) error {
	err := v.Struct(input)
	if err == nil {
		return nil
	}

	var msgs []string
	for _, fe := range err.(validator.ValidationErrors) {
		msgs = append(msgs, fieldError(fe))
	}
	return apperr.Validation(strings.Join(msgs, ";"))
}

func fieldError(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	case "gtefield":
		return fmt.Sprintf("%s must be >= %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
