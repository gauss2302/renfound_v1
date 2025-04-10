package validator

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"reflect"
	"strings"
)

type Validator struct {
	validate *validator.Validate
	logger   *zap.Logger
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func NewValidator(logger *zap.Logger) *Validator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		validate: v,
		logger:   logger.With(zap.String("component", "validator")),
	}
}

// validates a struct based on validator tags
func (v *Validator) Validate(data interface{}) ([]ValidationError, error) {
	if err := v.validate.Struct(data); err != nil {
		var validationErrors []ValidationError

		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			for _, err := range ve {
				// creates validation for each field
				validationError := ValidationError{
					Field:   err.Field(),
					Message: getErrorMessage(err),
				}
				validationErrors = append(validationErrors, validationError)
			}

			return validationErrors, nil
		}
		v.logger.Error("Validation error", zap.Error(err))
		return nil, err
	}
	return nil, nil
}

func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		if err.Type().Kind() == reflect.String {
			return fmt.Sprintf("Must be at least %s characters long", err.Param())
		}
		return fmt.Sprintf("Must be at least %s", err.Param())
	case "max":
		if err.Type().Kind() == reflect.String {
			return fmt.Sprintf("Must be at most %s characters long", err.Param())
		}
		return fmt.Sprintf("Must be at most %s", err.Param())
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", err.Param())
	default:
		return fmt.Sprintf("Failed validation for '%s'", err.Tag())
	}
}
