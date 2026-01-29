package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PostgresDSN string
	RedisAddr   string
	AppEnv      string
}

// func LoadConfig() *Config {
// 	return &Config{
// 		PostgresDSN: os.Getenv("POSTGRES_DSN"),
// 		RedisAddr:   os.Getenv("REDIS_ADDR"),
// 	}
// }

func LoadConfig() *Config {
	// Load .env only in local environment
	_ = godotenv.Load() // no panic if missing in production

	cfg := &Config{
		PostgresDSN: os.Getenv("POSTGRES_DSN"),
		RedisAddr:   os.Getenv("REDIS_ADDR"),
		AppEnv:      os.Getenv("APP_ENV"),
	}

	if cfg.PostgresDSN == "" {
		log.Fatal("POSTGRES_DSN is required")
	}

	return cfg
}
