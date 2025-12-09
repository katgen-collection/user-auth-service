package http

import (
	"mikhailjbs/user-auth-service/internal/infra/http/handlers"
	"mikhailjbs/user-auth-service/internal/infra/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App, userHandler handlers.UserHandler, authHandler handlers.AuthHandler, authz *middleware.AuthMiddleware) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	users := v1.Group("/users", authz.Require(middleware.Policy{Roles: []string{"admin"}}))
	users.Post("/", userHandler.CreateUser)
	users.Get("/", userHandler.GetUsers)
	users.Get("/:id", userHandler.GetUser)
	users.Put("/:id", userHandler.UpdateUser)
	users.Delete("/:id", userHandler.DeleteUser)

	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Post("/logout", authz.Require(middleware.Policy{}), authHandler.Logout)
	auth.Get("/me", authz.Require(middleware.Policy{}), authHandler.Me)
}
