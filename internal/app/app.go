package app

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"mikhailjbs/user-auth-service/internal/config"
	authdomain "mikhailjbs/user-auth-service/internal/domain/auth"
	sessiondomain "mikhailjbs/user-auth-service/internal/domain/session"
	"mikhailjbs/user-auth-service/internal/domain/user"
	"mikhailjbs/user-auth-service/internal/infra/http"
	"mikhailjbs/user-auth-service/internal/infra/http/handlers"
	"mikhailjbs/user-auth-service/internal/infra/logger"
	"mikhailjbs/user-auth-service/internal/infra/middleware"
	"mikhailjbs/user-auth-service/internal/infra/repository"
	"mikhailjbs/user-auth-service/internal/infra/security"
	authusecase "mikhailjbs/user-auth-service/internal/usecase/auth"
	usecase "mikhailjbs/user-auth-service/internal/usecase/user"
)

func Run() {
	// 1. Load Config
	cfg := config.Load()

	// 2. Init Logger
	logger.Init()
	logger.Log.Info("Starting User Auth Service...")

	// 3. Init Database
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require pool_mode=session search_path=user_service",
		cfg.DatabaseHost,
		cfg.DatabasePort,
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseName,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Fatalf("Failed to connect to database: %v", err)
	}

	// First, ensure the schema exists in the database
	if err := db.Exec("CREATE SCHEMA IF NOT EXISTS user_service").Error; err != nil {
		logger.Log.Fatalf("Failed to create schema 'user_service': %v", err)
	}

	// Second, switch the connection's focus to this schema
	if err := db.Exec("SET search_path TO user_service").Error; err != nil {
		logger.Log.Fatalf("Failed to set search_path: %v", err)
	}

	// Auto Migrate (for development simplicity, usually done via migration tools)
	if err := db.AutoMigrate(&user.User{}, &sessiondomain.Session{}); err != nil {
		logger.Log.Fatalf("Failed to migrate database: %v", err)
	}

	// 4. Init Repository
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// 5. Init Service (Domain)
	userService := user.NewService(userRepo)
	sessionService := sessiondomain.NewService(sessionRepo)
	authService := authdomain.NewService(userService, sessionService, userRepo)

	// 6. Init UseCases
	createUserUC := usecase.NewCreateUserUseCase(userService)
	getUsersUC := usecase.NewGetUsersUseCase(userService)
	getUserUC := usecase.NewGetUserUseCase(userService)
	updateUserUC := usecase.NewUpdateUserUseCase(userService)
	deleteUserUC := usecase.NewDeleteUserUseCase(userService)
	registerAuthUC := authusecase.NewRegisterUseCase(authService)
	loginAuthUC := authusecase.NewLoginUseCase(authService)
	meAuthUC := authusecase.NewGetMeUseCase(authService)

	// 7. Init Handlers
	userHandler := handlers.NewUserHandler(createUserUC, getUsersUC, getUserUC, updateUserUC, deleteUserUC)
	accessTTL := time.Duration(cfg.AccessTokenMinutes) * time.Minute
	refreshTTL := time.Duration(cfg.RefreshTokenDays) * 24 * time.Hour
	tokenManager, err := security.NewTokenManager(cfg.JWTSecret, cfg.JWTRefreshSecret, accessTTL, refreshTTL, logger.Log)
	if err != nil {
		logger.Log.Fatalf("Failed to initialize token manager: %v", err)
	}
	authHandler := handlers.NewAuthHandler(registerAuthUC, loginAuthUC, meAuthUC, sessionService, tokenManager, cfg.CookieDomain)
	authzMiddleware := middleware.NewAuthMiddleware(middleware.Config{
		TokenManager:      tokenManager,
		AccessTokenCookie: middleware.DefaultAccessTokenCookie,
		ContextKey:        middleware.DefaultClaimsContextKey,
	})

	// 8. Init Server
	app := http.NewServer()

	// 9. Register Routes
	http.RegisterRoutes(app, userHandler, authHandler, authzMiddleware)

	// 10. Start Server
	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Log.Infof("Server listening on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
