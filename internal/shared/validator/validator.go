package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidationError format error per field
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validate memvalidasi struct dan mengembalikan list error per field
func Validate(s interface{}) []ValidationError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var errors []ValidationError
	for _, e := range err.(validator.ValidationErrors) {
		errors = append(errors, ValidationError{
			Field:   toSnakeCase(e.Field()),
			Message: toMessage(e),
		})
	}
	return errors
}

// toMessage mengubah validation tag menjadi pesan yang readable
func toMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s wajib diisi", toSnakeCase(e.Field()))
	case "email":
		return fmt.Sprintf("%s harus berupa email yang valid", toSnakeCase(e.Field()))
	case "min":
		return fmt.Sprintf("%s minimal %s karakter", toSnakeCase(e.Field()), e.Param())
	case "max":
		return fmt.Sprintf("%s maksimal %s karakter", toSnakeCase(e.Field()), e.Param())
	case "oneof":
		return fmt.Sprintf("%s harus salah satu dari: %s", toSnakeCase(e.Field()), e.Param())
	case "gte":
		return fmt.Sprintf("%s minimal bernilai %s", toSnakeCase(e.Field()), e.Param())
	case "lte":
		return fmt.Sprintf("%s maksimal bernilai %s", toSnakeCase(e.Field()), e.Param())
	default:
		return fmt.Sprintf("%s tidak valid", toSnakeCase(e.Field()))
	}
}

// toSnakeCase mengubah FieldName menjadi field_name
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
