package router

import (
	userRouter "go-fiber-gorm-boilerplate/router/user"

	"github.com/gofiber/fiber/v2"
)

func New(app *fiber.App) {
	userRouter.CreateRouter(app)
}