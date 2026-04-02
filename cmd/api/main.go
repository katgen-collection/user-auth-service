package main

import (
	"mikhailjbs/user-auth-service/internal/app"
	_ "mikhailjbs/user-auth-service/docs" // required for swaggo
)

// @title Katgen User Auth Service API
// @version 1.0
// @description This is the core identity and auth service.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@katgen.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
func main() {
	app.Run()
}
