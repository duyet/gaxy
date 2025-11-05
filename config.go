package main

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// Config contains application configuration loaded from environment variables
type Config struct {
	// RoutePrefix is the URL prefix for all endpoints (e.g., "/analytics")
	RoutePrefix string `env:"ROUTE_PREFIX"`

	// GoogleOrigin is the upstream Google Analytics/Tag Manager URL
	GoogleOrigin string `env:"GOOGLE_ORIGIN" default:"https://www.google-analytics.com"`

	// InjectParamsFromReqHeaders converts request headers to query parameters
	// Format: "header1,header2" or "header1__param1,header2__param2" for renaming
	InjectParamsFromReqHeaders string `env:"INJECT_PARAMS_FROM_REQ_HEADERS"`

	// SkipParamsFromReqHeaders removes specific query parameters from requests
	SkipParamsFromReqHeaders string `env:"SKIP_PARAMS_FROM_REQ_HEADERS"`

	// Port is the server listening port
	Port string `env:"PORT" default:"3000"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() Config {
	config := Config{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	return config
}
