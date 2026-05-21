package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr        string
	DatabaseURL string
	AIProvider  string
	AIAPIKey    string
	AIAPIUrl    string
	AIModel     string
}

func Load() Config {
	// 尝试加载 .env 文件
	_ = godotenv.Load()

	return Config{
		Addr:        env("APP_ADDR", ":8080"),
		DatabaseURL: env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/reqsplit?sslmode=disable"),
		AIProvider:  env("AI_PROVIDER", "stub"),
		AIAPIKey:    os.Getenv("AI_API_KEY"),
		AIAPIUrl:    env("AI_API_URL", "https://api.openai.com"),
		AIModel:     env("AI_MODEL", "gpt-4"),
	}
}

// Validate 验证配置的有效性
func (c Config) Validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	if c.AIProvider != "stub" && c.AIProvider != "openai" {
		return fmt.Errorf("invalid AI_PROVIDER: %s (must be 'stub' or 'openai')", c.AIProvider)
	}

	if c.AIProvider == "openai" && c.AIAPIKey == "" {
		return errors.New("AI_API_KEY is required when using openai provider")
	}

	return nil
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
