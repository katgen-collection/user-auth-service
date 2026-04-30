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
	RSAPrivateKeyPath  string
	JWTRefreshSecret   string
	RedisURL           string
	CookieDomain       string
	CorsAllowedOrigins string
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
		RSAPrivateKeyPath:  getEnv("RSA_PRIVATE_KEY_PATH", "../.secrets/private.pem"),
		JWTRefreshSecret:   getEnv("JWT_REFRESH_SECRET", ""),
		RedisURL:           getEnv("REDIS_URL", "redis://localhost:6379"),
		CookieDomain:       getEnv("AUTH_COOKIE_DOMAIN", "localhost"),
		CorsAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3002,http://localhost:8080,https://chat.mikhailjbs.my.id"),
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
