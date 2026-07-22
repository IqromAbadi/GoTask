package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateStruct validates a struct and returns a map of field errors.
// Error messages are in Indonesian for mobile app readability.
func ValidateStruct(s any) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, err := range err.(validator.ValidationErrors) {
		field := toSnakeCase(err.Field())
		errors[field] = validationMessage(err)
	}
	return errors
}

// Var validates a single variable.
func Var(field any, tag string) error {
	return validate.Var(field, tag)
}

func validationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "Wajib diisi"
	case "email":
		return "Format email tidak valid"
	case "min":
		return fmt.Sprintf("Minimal %s karakter", err.Param())
	case "max":
		return fmt.Sprintf("Maksimal %s karakter", err.Param())
	case "len":
		return fmt.Sprintf("Harus %s karakter", err.Param())
	case "oneof":
		return fmt.Sprintf("Harus salah satu dari: %s", err.Param())
	case "gte":
		return fmt.Sprintf("Harus lebih besar atau sama dengan %s", err.Param())
	case "lte":
		return fmt.Sprintf("Harus lebih kecil atau sama dengan %s", err.Param())
	case "gt":
		return fmt.Sprintf("Harus lebih besar dari %s", err.Param())
	case "lt":
		return fmt.Sprintf("Harus lebih kecil dari %s", err.Param())
	case "uuid":
		return "Format UUID tidak valid"
	default:
		return fmt.Sprintf("Tidak valid (%s)", err.Tag())
	}
}

var snakeCaseRe = regexp.MustCompile(`([a-z0-9])([A-Z])`)

func toSnakeCase(s string) string {
	snake := snakeCaseRe.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(snake)
}
