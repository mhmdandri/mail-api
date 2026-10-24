package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	APIKey    string
	SMTPHost  string
	SMTPPort  int
	SMTPUser  string
	SMTPPass  string
	FromEmail string
	FromName  string
}

// Load loads configuration from environment variables and an optional .env file.
func Load() *Config {
	// Attempt to load .env file if it exists, ignore error if missing (e.g. in production/docker)
	if err := godotenv.Load(); err != nil {
		slog.Debug("No .env file found, using system environment variables")
	}

	smtpPort := getEnvAsInt("SMTP_PORT", 587)

	cfg := &Config{
		Port:      getEnv("PORT", "8082"),
		APIKey:    getEnv("API_KEY", ""),
		SMTPHost:  getEnv("SMTP_HOST", "smtp.mail.me.com"),
		SMTPPort:  smtpPort,
		SMTPUser:  getEnv("SMTP_USER", "mhmdandri03@icloud.com"),
		SMTPPass:  getEnv("SMTP_PASS", ""),
		FromEmail: getEnv("FROM_EMAIL", "no-reply@weddingofandricica.me"),
		FromName:  getEnv("FROM_NAME", "Andri & Cica"),
	}

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		slog.Warn("Invalid integer for environment variable, using default", "key", key, "value", valStr, "default", defaultVal)
		return defaultVal
	}
	return val
}
