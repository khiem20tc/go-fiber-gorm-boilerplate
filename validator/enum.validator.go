package validator

import (
	"reflect"
	"slices"

	"github.com/go-playground/validator/v10"
)

func EnumValidator(fl validator.FieldLevel, values []string) bool {
	switch v := fl.Field(); v.Kind() {
	case reflect.String:
		if v.String() == "" {
			return true
		}
		return slices.Contains(values, v.String())
	default:
		return false
	}

}
