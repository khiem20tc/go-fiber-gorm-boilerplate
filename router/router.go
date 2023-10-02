package router

import (
	userRouter "botp-gateway/router/user"

	"github.com/gofiber/fiber/v2"
)

func New(app *fiber.App) {
	userRouter.CreateRouter(app)
}