package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
	JWTSecret   string
	GatewayURL  string
}

func Load() Config {
	return Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://weather:weather@localhost:5433/weather?sslmode=disable"),
		Port:        getEnv("PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", "super-secret-key"),
		GatewayURL:  getEnv("GATEWAY_URL", "http://localhost:8081"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func (c Config) Addr() string {
	return fmt.Sprintf(":%s", c.Port)
}
