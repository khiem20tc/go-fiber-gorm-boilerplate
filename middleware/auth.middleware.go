package middleware

import (
	"fiber-gateway/model"
	jsonWebToken "fiber-gateway/utils/jwt"
	"slices"

	"github.com/gofiber/fiber/v2"
)

func ExtractToken(c *fiber.Ctx) (string, error) {
	tokenString := c.Get("Authorization")
	if tokenString == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: No token provided")
	}

	if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: Invalid token format")
	}

	return tokenString[7:], nil
}

func Authen() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := ExtractToken(c)
		if err != nil {
			return err
		}

		claims, err := jsonWebToken.ParseToken(token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: Invalid token")
		}

		c.Locals("user", claims)

		return c.Next()
	}
}

func Author(roles ...model.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := ExtractToken(c)
		if err != nil {
			return err
		}

		claims, err := jsonWebToken.ParseToken(token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: "+err.Error())
		}

		if !slices.Contains(roles, claims.Role) {
			return fiber.NewError(fiber.StatusForbidden, "Forbidden: Insufficient privileges")
		}

		c.Locals("user", claims)

		return c.Next()
	}
}

func AuthenEmail() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := ExtractToken(c)
		if err != nil {
			return err
		}

		claims, err := jsonWebToken.ParseTokenEmail(token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: Invalid token")
		}

		isInWhitelist, err := jsonWebToken.IsInWhitelist(claims.Email, token)

		if err != nil || !isInWhitelist {
			return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: Token is not valid in whitelist")
		}

		jsonWebToken.RemoveFromWhitelist(claims.Email)

		c.Locals("user", claims)

		return c.Next()
	}
}
