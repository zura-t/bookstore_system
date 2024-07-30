package pkg

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func MsgForTag(validation_err validator.FieldError, fieldName string) string {
	switch validation_err.Tag() {
	case "required":
		return fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", fieldName, validation_err.Tag())
	case "email":
		return fmt.Sprintf("Invalid email")
	case "min":
		return fmt.Sprintf("min value for '%s' field is %s", fieldName, validation_err.Param())
	case "max":
		return fmt.Sprintf("max value for '%s' field is %s", fieldName, validation_err.Param())
	}
	return ""
}
