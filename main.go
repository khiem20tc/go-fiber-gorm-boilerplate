package main

import (
	"botp-gateway/config"
	"botp-gateway/router"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/swagger"
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

	// Gen swagger API documents
	app.Get("/swagger/*", swagger.HandlerDefault) // default

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
