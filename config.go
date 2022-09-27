package main

import (
	"github.com/kelseyhightower/envconfig"
)

// Config contains config
type Config struct {
	RoutePrefix                string `env:"ROUTE_PREFIX"`
	GoogleOrigin               string `env:"GOOGLE_ORIGIN" default:"https://www.google-analytics.com"`
	InjectParamsFromReqHeaders string `env:"INJECT_PARAMS_FROM_REQ_HEADERS"`
	SkipParamsFromReqHeaders   string `env:"SKIP_PARAMS_FROM_REQ_HEADERS"`
	Port                       string `env:"PORT" default:"3000"`
}

func LoadConfig() Config {
	config := Config{}
	envconfig.Process("", &config)

	return config
}
