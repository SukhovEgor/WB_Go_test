package config

import "os"

type Config struct {
	DBpassword string
	DBuser     string
	DBname     string
}

// 	connStr := "postgres://postgres:qwerty@localhost:5433/WB_ordersDB"
func LoadConfig() (*Config, error) {
	return &Config{
		DBpassword: getEnv("POSTGRES_PASSWORD", "qwerty"),
		DBuser:     getEnv("POSTGRES_USER", "postgres"),
		DBname:     getEnv("POSTGRES_DB", "WB_ordersDB"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
