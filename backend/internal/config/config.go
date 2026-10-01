package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
	AppTimezone string
	CorsOrigin  string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://app:app@localhost:5432/evaluation?sslmode=disable"),
		Port:        getEnv("PORT", "8080"),
		AppTimezone: getEnv("APP_TIMEZONE", "America/Sao_Paulo"),
		CorsOrigin:  getEnv("CORS_ORIGIN", "http://localhost:5173"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
