package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	MQTT        MQTTConfig
	JWT         JWTConfig
}

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host string
	Port int
}

type MQTTConfig struct {
	BrokerURL string
}

type JWTConfig struct {
	Secret          string
	ExpiryDuration  time.Duration
	RefreshDuration time.Duration
}

func Load() (*Config, error) {
	config := &Config{
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port: getEnvAsIntOrDefault("SERVER_PORT", 50051),
		},
		Database: DatabaseConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvAsIntOrDefault("DB_PORT", 5432),
			User:     getEnvOrDefault("DB_USER", "syncwrite"),
			Password: getEnvOrDefault("DB_PASSWORD", "syncwrite_password"),
			DBName:   getEnvOrDefault("DB_NAME", "syncwrite"),
			SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host: getEnvOrDefault("REDIS_HOST", "localhost"),
			Port: getEnvAsIntOrDefault("REDIS_PORT", 6379),
		},
		MQTT: MQTTConfig{
			BrokerURL: getEnvOrDefault("MQTT_BROKER", "mqtt://localhost:1883"),
		},
		JWT: JWTConfig{
			Secret:          getEnvOrDefault("JWT_SECRET", "your_jwt_secret_here"),
			ExpiryDuration:  getEnvAsDurationOrDefault("JWT_EXPIRY", 24*time.Hour),
			RefreshDuration: getEnvAsDurationOrDefault("REFRESH_TOKEN_EXPIRY", 720*time.Hour),
		},
	}

	return config, nil
}

func (c *DatabaseConfig) GetConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
