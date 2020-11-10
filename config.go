package main

import (
	"fmt"
	"os"
)

// Config contains config
type Config struct {
	RoutePrefix                string
	GoogleOrigin               string
	InjectParamsFromReqHeaders string
	SkipParamsFromReqHeaders   string
	Port                       string
}

// LoadConfig returns a new Config struct
func LoadConfig() Config {
	config := Config{
		RoutePrefix:                getEnv("ROUTE_PREFIX", ""),
		GoogleOrigin:               getEnv("GOOGLE_ORIGIN", "https://www.google-analytics.com"),
		InjectParamsFromReqHeaders: getEnv("INJECT_PARAMS_FROM_REQ_HEADERS", ""),
		SkipParamsFromReqHeaders:   getEnv("SKIP_PARAMS_FROM_REQ_HEADERS", ""),
		Port:                       getEnv("PORT", "3000"),
	}

	fmt.Printf("Loaded config: %+v\n\n", config)

	return config
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
