package config

import (
	"log"
	"os"
	"strings"
)

type Config struct {
	Database struct {
		Host     string
		Port     string
		User     string
		Password string
		Name     string
	}
	Server struct {
		Host string
		Port string
	}
	Kafka struct {
		Brokers      []string
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnv returns the environment variable value or default if not set.
func GetEnv(key, defaultValue string) string {
	return getEnv(key, defaultValue)
}

// SplitAndTrim splits a comma separated list and trims spaces.
func SplitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func InitConfig() *Config {
	config := &Config{}

	// Database config from environment variables
	config.Database.Host = getEnv("DB_HOST", "postgres-user")
	config.Database.Port = getEnv("DB_PORT", "5432")
	config.Database.User = getEnv("DB_USER", "postgres")
	config.Database.Password = getEnv("DB_PASSWORD", "")
	config.Database.Name = getEnv("DB_NAME", "user_service_db")

	config.Kafka.Brokers = SplitAndTrim(getEnv("KAFKA_BROKERS", ""), ",")

	// Server config from environment variables
	config.Server.Host = getEnv("SERVER_HOST", "0.0.0.0")
	config.Server.Port = getEnv("SERVER_PORT", "8095")

	log.Println("Configuration loaded from environment variables")
	return config
}
