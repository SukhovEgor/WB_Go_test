package config

import "os"

type Config struct {
	DBPassword string
	DBUser     string
	DBName     string
}

// 	connStr := "postgres://postgres:qwerty@localhost:5433/WB_ordersDB"
func LoadConfig() (*Config, error) {
	return &Config{
		DBPassword: getEnv("POSTGRES_PASSWORD", "qwerty"),
		DBUser:     getEnv("POSTGRES_USER", "postgres"),
		DBName:     getEnv("POSTGRES_DB", "WB_ordersDB"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
