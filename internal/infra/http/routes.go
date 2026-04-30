package http

import (
	"mikhailjbs/user-auth-service/internal/infra/http/handlers"
	"mikhailjbs/user-auth-service/internal/infra/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func RegisterRoutes(app *fiber.App, userHandler handlers.UserHandler, authHandler handlers.AuthHandler, authz *middleware.AuthMiddleware) {
	app.Get("/swagger/*", swagger.HandlerDefault)

	// UI Routes
	app.Get("/login", func(c *fiber.Ctx) error {
		return c.SendFile("./views/login.html")
	})
	app.Get("/register", func(c *fiber.Ctx) error {
		return c.SendFile("./views/register.html")
	})
	app.Get("/profile", func(c *fiber.Ctx) error {
		return c.SendFile("./views/profile.html")
	})

	api := app.Group("/api")
	v1 := api.Group("/v1")

	users := v1.Group("/users", authz.Require(middleware.Policy{Roles: []string{"user", "admin"}}))
	users.Post("/", authz.Require(middleware.Policy{Roles: []string{"admin"}}), userHandler.CreateUser)
	users.Get("/", userHandler.GetUsers)
	users.Get("/:id", userHandler.GetUser)
	users.Put("/:id", userHandler.UpdateUser)
	users.Delete("/:id", authz.Require(middleware.Policy{Roles: []string{"admin"}}), userHandler.DeleteUser)

	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Post("/logout", authz.Require(middleware.Policy{}), authHandler.Logout)
	auth.Get("/me", authz.Require(middleware.Policy{}), authHandler.Me)
}
