package main

import (
	"os"
)

// Config contains config
type Config struct {
	GoogleOrigin               string
	InjectParamsFromReqHeaders string
	Port                       string
}

// LoadConfig returns a new Config struct
func LoadConfig() *Config {
	return &Config{
		GoogleOrigin:               getEnv("GOOGLE_ORIGIN", "https://ssl.google-analytics.com"),
		InjectParamsFromReqHeaders: getEnv("INJECT_PARAMS_FROM_REQ_HEADERS", ""),
		Port:                       getEnv("PORT", "3000"),
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
