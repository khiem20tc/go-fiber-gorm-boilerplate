package router

import (
	userRouter "fiber-gateway/router/user"

	"github.com/gofiber/fiber/v2"
)

func New(app *fiber.App) {
	userRouter.CreateRouter(app)
}
