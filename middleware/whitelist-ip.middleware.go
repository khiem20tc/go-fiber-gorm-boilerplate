package middleware

import (
	"net"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func CheckWhitelistIPs(allowedIPs []string) fiber.Handler {
	allowedIPSet := make(map[string]struct{})
	for _, ip := range allowedIPs {
		allowedIPSet[ip] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		clientIP, _, err := net.SplitHostPort(c.IP())
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Internal Server Error",
			})
		}

		_, allowed := allowedIPSet[clientIP]
		if !allowed {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"message": "Forbidden",
			})
		}

		return c.Next()
	}
}
