// Package bootstrap contains utility functions to read environment variables, if a default is not provided and
// the value of an environment variable is not set then the application will panic.
package bootstrap

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

// GetIntEnvVar reads an environment variable and returns the value as an int, if the environment variable is not set
// then the application will panic.
func GetIntEnvVar(key string) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		panic(fmt.Sprintf("missing required env var %v", key))
	}

	var err error
	result, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf("cannot parse %v, error: %v", key, err))
	}

	slog.Info("Environment Variable Set", key, value)

	return result
}

// GetEnvVar reads an environment variable and returns the value as a string, if the environment variable is not set
// then the application will panic.
func GetEnvVar(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		panic(fmt.Sprintf("missing required env var %v", key))
	}

	slog.Info("Environment Variable Set", key, value)

	return value
}

// GetOptionalEnvVar reads an environment variable and returns the value as a string, or the default value if the
// environment variable is not set.
func GetOptionalEnvVar(key string, def string) string {
	strValue, exists := os.LookupEnv(key)
	result := def
	if exists {
		result = strValue
	}

	slog.Info("Environment Variable Set", key, strValue)

	return result
}

// GetOptionalBoolEnvVar reads an environment variable and returns the value as a bool, or the default value if the
// environment variable is not set. If the environment variable is set but cannot be parsed as a bool then the
// application will panic.
func GetOptionalBoolEnvVar(key string, def bool) bool {
	strValue, exists := os.LookupEnv(key)
	result := def
	if exists {
		var err error
		result, err = strconv.ParseBool(strValue)
		if err != nil {
			panic(fmt.Sprintf("cannot parse %v, error: %v", key, err))
		}
	}

	slog.Info("Environment Variable Set", key, strValue)

	return result
}

// GetBoolEnvVar reads an environment variable and returns the value as a bool, if the environment variable is not set
// then the application will panic.
func GetBoolEnvVar(key string) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		panic(fmt.Sprintf("missing required env var %v", key))
	}

	var err error
	result, err := strconv.ParseBool(value)
	if err != nil {
		panic(fmt.Sprintf("cannot parse %v, error: %v", key, err))
	}

	slog.Info("Environment Variable Set", key, value)

	return result
}

// GetOptionalIntEnvVar reads an environment variable and returns the value as an int, or the default value if the
// environment variable is not set. If the environment variable is set but cannot be parsed as an int then the
// application will panic.
func GetOptionalIntEnvVar(key string, def int) int {
	strValue, exists := os.LookupEnv(key)
	result := def
	if exists {
		var err error
		result, err = strconv.Atoi(strValue)
		if err != nil {
			panic(fmt.Sprintf("cannot parse %v, error: %v", key, err))
		}
	}

	slog.Info("Environment Variable Set", key, strValue)

	return result
}

// GetOptionalFloatEnvVar reads an environment variable and returns the value as a float64, if the environment variable is not set
// then the application will panic.  If the environment variable is set but cannot be parsed as a float64 then the
// application will panic.
func GetOptionalFloatEnvVar(key string, def float64) float64 {
	strValue, exists := os.LookupEnv(key)
	result := def
	if exists {
		var err error
		result, err = strconv.ParseFloat(strValue, 64)
		if err != nil {
			panic(fmt.Sprintf("cannot parse %v, error: %v", key, err))
		}
	}

	slog.Info("Environment Variable Set", key, result)

	return result
}
