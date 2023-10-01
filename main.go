package main

import (
	"go-fiber-gorm-boilerplate/config"
	"go-fiber-gorm-boilerplate/router"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
)

func main() {
	//Default
	//AllowOrigins: "*",
	//AllowMethods: "GET, POST, PATCH, HEAD, PUT, DELETE,"
	//AllowHeaders: ""
	//AllowCredentials: false
	app := fiber.New()

	// Initialize default config (Assign the middleware to /metrics)
	app.Get("/metrics", monitor.New())

	// Initialize default config
	app.Use(logger.New(logger.Config{}))

	//Initialize connect to database
	// gorm.Connection()
	// gorm.AutoMigration()
	// defer gorm.Disconnection()

	// Initialize router
	router.New(app)

	app.Listen(":" + config.Env("PORT", "3002"))
}
