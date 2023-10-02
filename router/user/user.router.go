package userRouter

import (
	userService "botp-gateway/service/user"

	"github.com/gofiber/fiber/v2"
)

func CreateRouter(app *fiber.App) {
	r := app.Group("/user")
	{
		r.Get("", userService.Get)
	}

	r2 := app.Group("/user2")
	{
		r2.Get("/", func(c *fiber.Ctx) error {
			return c.SendString("Hello, World 2!")
		})
	}
}
