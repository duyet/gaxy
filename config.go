package main

import (
	"errors"
	"log"
	"os"
	"reflect"
)

// Config contains config
type Config struct {
	RoutePrefix                string `env:"ROUTE_PREFIX"`
	GoogleOrigin               string `env:"GOOGLE_ORIGIN"`
	InjectParamsFromReqHeaders string `env:"INJECT_PARAMS_FROM_REQ_HEADERS"`
	SkipParamsFromReqHeaders   string `env:"SKIP_PARAMS_FROM_REQ_HEADERS"`
	Port                       string `env:"PORT"`
}

func (config *Config) Set(key string, value string) {
	s := reflect.Indirect(reflect.ValueOf(config))
	s.FieldByName(key).SetString(value)
}

func (config *Config) GetEnvKey(key string) (string, error) {
	t := reflect.TypeOf(*config)
	field, found := t.FieldByName(key)

	if found == false {
		return "", errors.New("field not found")
	}

	return field.Tag.Get("env"), nil
}

func LoadDefaultConfig() Config {
	config := Config{
		RoutePrefix:                "",
		GoogleOrigin:               "https://www.google-analytics.com",
		InjectParamsFromReqHeaders: "",
		SkipParamsFromReqHeaders:   "",
		Port:                       "3000",
	}

	return config
}

// LoadConfig returns a new Config struct
func LoadConfig() Config {
	config := LoadDefaultConfig()

	t := reflect.TypeOf(config)
	v := reflect.Indirect(reflect.ValueOf(&config))

	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i).Name
		envName, err := config.GetEnvKey(key)
		if err != nil {
			continue
		}

		currentValue := v.FieldByName(key).String()
		valueFromEnv := getEnv(envName, currentValue)

		config.Set(key, valueFromEnv)
	}

	log.Printf("Loaded config: %+v\n\n", config)

	return config
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
