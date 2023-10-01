package validator

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

func PageValidator(fl validator.FieldLevel) bool {
	min := decimal.NewFromInt(1)
	switch v := fl.Field(); v.Kind() {
	case reflect.String:
		val, err := decimal.NewFromString(v.String())
		if err == nil && val.GreaterThanOrEqual(min) {
			return true
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val := decimal.NewFromInt(v.Int())
		if val.GreaterThanOrEqual(min) {
			return true
		}
	case reflect.Float32, reflect.Float64:
		val := decimal.NewFromFloat(v.Float())
		if val.GreaterThanOrEqual(min) {
			return true
		}
	default:
		return false
	}

	return false
}

func PageSizeValidator(fl validator.FieldLevel) bool {
	min := decimal.NewFromInt(1)
	max := decimal.NewFromInt(100)

	switch v := fl.Field(); v.Kind() {
	case reflect.String:
		val, err := decimal.NewFromString(v.String())
		if err == nil && val.GreaterThanOrEqual(min) && val.LessThanOrEqual(max) {
			return true
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val := decimal.NewFromInt(v.Int())
		if val.GreaterThanOrEqual(min) && val.LessThanOrEqual(max) {
			return true
		}
	case reflect.Float32, reflect.Float64:
		val := decimal.NewFromFloat(v.Float())
		if val.GreaterThanOrEqual(min) && val.LessThanOrEqual(max) {
			return true
		}
	default:
		return false
	}

	return false
}
