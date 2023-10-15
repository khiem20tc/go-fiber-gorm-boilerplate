package middleware

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

func GetValidate() *validator.Validate {
	return validate
}

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Email is not valid or has keywords not allow office|admin."
	}
	return fe.Error() // default error
}

func ValidateInput(inputStruct interface{}, isBody bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Create a new instance of the provided input struct
		input := reflect.New(reflect.TypeOf(inputStruct)).Interface()

		if isBody {
			// Parse the request body into the provided input struct
			if err := c.BodyParser(input); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": "Invalid input data",
				})
			}
		} else {
			// Parse query parameters into the provided input struct
			if err := c.QueryParser(input); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": "Invalid query parameters",
				})
			}
		}

		// Perform validation using the validator
		if err := validate.Struct(input); err != nil {
			// Handle validation errors
			var validationErrors []string
			for _, err := range err.(validator.ValidationErrors) {
				validationErrors = append(validationErrors, fmt.Sprintf("%s: %s - %s", err.StructField(), err.Tag(), msgForTag(err)))
			}
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Validation error",
				"errors":  strings.Join(validationErrors, ", "),
			})
		}

		// Set the validated input data in the context for the next handler
		c.Locals("input", input)

		// If validation is successful, continue to the next handler
		return c.Next()
	}
}
