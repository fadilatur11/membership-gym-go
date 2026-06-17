package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	AppPort            string
	AppURL             string
	PublicMemberURL    string
	DBHost             string
	DBPort             string
	DBName             string
	DBUser             string
	DBPassword         string
	DBSSLMode          string
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	JWTAccessSecret    string
	JWTAccessTTL       time.Duration
	JWTRefreshTTL      time.Duration
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	DefaultTimezone    string
	DefaultCurrency    string
}

func Load() Config {
	_ = godotenv.Load()
	redisDB, _ := strconv.Atoi(env("REDIS_DB", "0"))
	accessTTL, _ := strconv.Atoi(env("JWT_ACCESS_TTL_MINUTES", "60"))
	refreshTTL, _ := strconv.Atoi(env("JWT_REFRESH_TTL_HOURS", "168"))

	return Config{
		AppEnv:             env("APP_ENV", "development"),
		AppPort:            env("APP_PORT", "8080"),
		AppURL:             env("APP_URL", "http://localhost:8080"),
		PublicMemberURL:    env("PUBLIC_MEMBER_URL", "http://localhost:3000/member/checkin"),
		DBHost:             env("DB_HOST", "localhost"),
		DBPort:             env("DB_PORT", "5432"),
		DBName:             env("DB_NAME", "gym_management"),
		DBUser:             env("DB_USER", "postgres"),
		DBPassword:         env("DB_PASSWORD", "postgres"),
		DBSSLMode:          env("DB_SSLMODE", "disable"),
		RedisAddr:          env("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      env("REDIS_PASSWORD", ""),
		RedisDB:            redisDB,
		JWTAccessSecret:    env("JWT_ACCESS_SECRET", "change_me_access_secret"),
		JWTAccessTTL:       time.Duration(accessTTL) * time.Minute,
		JWTRefreshTTL:      time.Duration(refreshTTL) * time.Hour,
		GoogleClientID:     env("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: env("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  env("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback"),
		DefaultTimezone:    env("DEFAULT_TIMEZONE", "Asia/Jakarta"),
		DefaultCurrency:    env("DEFAULT_CURRENCY", "IDR"),
	}
}

func (c Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
