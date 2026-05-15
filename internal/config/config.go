package config

import (
	"os"
	"strconv"
)

type Config struct {
	Addr        string
	DatabaseURL string
	AIProvider  string
	AIAPIKey    string
}

func Load() Config {
	return Config{
		Addr:        env("APP_ADDR", ":8080"),
		DatabaseURL: env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/reqsplit?sslmode=disable"),
		AIProvider:  env("AI_PROVIDER", "stub"),
		AIAPIKey:    os.Getenv("AI_API_KEY"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func EnvInt(key string, fallback int) int {
	v, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return v
}
