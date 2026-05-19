package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort  string
	DBPath      string
	JWTKeyPath  string
}

func Load() *Config {
	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8083"),
		DBPath:     getEnv("DB_PATH", "system.db"),
		JWTKeyPath: getEnv("JWT_KEY_PATH", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
