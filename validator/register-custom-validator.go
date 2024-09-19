package validator

import (
	"fiber-gateway/middleware"
	"reflect"
	"regexp"
	"strings"
	"time"

	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/go-playground/validator/v10"
)

var (
	verifier = emailverifier.NewVerifier()
)

func RegisterCustomValidator() {
	validate := middleware.GetValidate()
	validate.RegisterValidation("page", PageValidator)
	validate.RegisterValidation("page_size", PageSizeValidator)
	validate.RegisterValidation("sortOrder", func(fl validator.FieldLevel) bool {
		return EnumValidator(fl, []string{string("desc"), string("asc")})
	})
	validate.RegisterValidation("bool_value_validator", func(fl validator.FieldLevel) bool {
		value := fl.Field()
		return isBool(value)
	})
	validate.RegisterValidation("email", func(fl validator.FieldLevel) bool {
		email := fl.Field().String()

		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

		if strings.Contains(email, "office") || strings.Contains(email, "admin") {
			return false
		}

		return emailRegex.MatchString(email)
	})
	validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()

		// Regular expressions to enforce password requirements
		var (
			hasUpperCase = regexp.MustCompile(`[A-Z]`).MatchString
			hasLowerCase = regexp.MustCompile(`[a-z]`).MatchString
			hasNumber    = regexp.MustCompile(`[0-9]`).MatchString
			hasSpecial   = regexp.MustCompile(`[^A-Za-z0-9]`).MatchString
		)

		return len(password) >= 8 &&
			hasUpperCase(password) &&
			hasLowerCase(password) &&
			hasNumber(password) &&
			hasSpecial(password)
	})

	validate.RegisterValidation("timeUTC", func(fl validator.FieldLevel) bool {
		timeUTC := fl.Field().String()
		layout := "2006-01-02T15:04:05.000Z"

		t, err := time.Parse(layout, timeUTC)
		if err != nil {
			return false
		}

		return t.Location() == time.UTC
	})
}

func isBool(value reflect.Value) bool {
	if value.IsNil() {
		return false
	} else {
		return value.Kind() == reflect.Bool
	}
}
