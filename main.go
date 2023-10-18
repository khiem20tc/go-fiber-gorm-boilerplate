package main

import (
	"botp-gateway/config"
	"botp-gateway/router"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/swagger"
)

// @title BOTP Gateway API
// @version 1.0
// @description Documentation for BOTP Gateway API
// @termsOfService http://swagger.io/terms/

// @contact.name BOTP Gateway Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
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
