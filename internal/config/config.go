package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseHost       string
	DatabasePort       string
	DatabaseUser       string
	DatabasePassword   string
	DatabaseName       string
	Port               string
	JWTSecret          string
	JWTRefreshSecret   string
	CookieDomain       string
	SessionExpiryHours int
	AccessTokenMinutes int
	RefreshTokenDays   int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it, using system environment variables")
	}

	return &Config{
		DatabaseHost:       getEnv("DB_HOST", ""),
		DatabasePort:       getEnv("DB_PORT", ""),
		DatabaseUser:       getEnv("DB_USER", ""),
		DatabasePassword:   getEnv("DB_PASSWORD", ""),
		DatabaseName:       getEnv("DB_NAME", ""),
		Port:               getEnv("PORT", "3000"),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTRefreshSecret:   getEnv("JWT_REFRESH_SECRET", ""),
		CookieDomain:       getEnv("AUTH_COOKIE_DOMAIN", "localhost"),
		SessionExpiryHours: getEnvAsInt("SESSION_EXPIRY_HOURS", 72),
		AccessTokenMinutes: getEnvAsInt("ACCESS_TOKEN_TTL_MINUTES", 30),
		RefreshTokenDays:   getEnvAsInt("REFRESH_TOKEN_TTL_DAYS", 30),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return fallback
}
