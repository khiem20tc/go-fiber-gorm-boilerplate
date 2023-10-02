package userService

import (
	"botp-gateway/gorm"
	"botp-gateway/model"

	"github.com/gofiber/fiber/v2"
)

func Get(c *fiber.Ctx) error {
	var users model.User
	result := gorm.DB.Find(&users)

	// input := c.Locals("input").(*validator.GenNewPwd)

	// user := c.Locals("user").(*jsonWebToken.MapClaims)

	if result.Error != nil {
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"success": false,
			"message": "User not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    users,
	})
}
