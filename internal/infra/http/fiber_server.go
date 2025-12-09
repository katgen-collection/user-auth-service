package http

import (
	"mikhailjbs/user-auth-service/internal/infra/http/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func NewServer() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "User Auth Service",
	})

	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(recover.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return handlers.SendSuccess(c, fiber.StatusOK, "Service is healthy", nil)
	})

	return app
}
