package pkg

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

func ListValidationErrors(req interface{}, validationErrors validator.ValidationErrors) []string {
	var validation_errs = make([]string, len(validationErrors))
	for l, validation_err := range validationErrors {
		field, ok := reflect.TypeOf(req).Elem().FieldByName(validation_err.StructField())
		fieldName := field.Tag.Get("json")
		if !ok {
			panic("Field not found")
		}
		validation_errs[l] = MsgForTag(validation_err, fieldName)
	}
	return validation_errs
}

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
