package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	ServerPort      string
	PostgresHost    string
	PostgresPort    int
	PostgresUser    string
	PostgresPass    string
	PostgresDB      string
	PostgresSSL     string
	MigrationsPath  string
	LogLevel        string
	ShutdownTimeout time.Duration
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using environment variables")
	}

	cfg := &Config{
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		PostgresHost:    getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:    getEnvAsInt("POSTGRES_PORT", 5432),
		PostgresUser:    getEnv("POSTGRES_USER", "postgres"),
		PostgresPass:    getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:      getEnv("POSTGRES_DB", "subscription_db"),
		PostgresSSL:     getEnv("POSTGRES_SSL", "disable"),
		MigrationsPath:  getEnv("MIGRATIONS_PATH", "file://migrations"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		ShutdownTimeout: getEnvAsDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
	}

	return cfg, nil
}

func (c *Config) GetPostgresDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresPass, c.PostgresDB, c.PostgresSSL)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
