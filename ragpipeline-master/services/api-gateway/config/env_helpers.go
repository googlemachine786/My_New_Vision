package config

import (
	"os"
	"strconv"
	"time"
)

// getEnvString returns the environment variable value or a default
func getEnvString(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return defaultVal
}

// getEnvInt returns the environment variable as an integer or a default
func getEnvInt(key string, defaultVal int) int {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

// getEnvDuration returns the environment variable as a duration or a default
func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
		// Try parsing as integer seconds
		if i, err := strconv.Atoi(val); err == nil {
			return time.Duration(i) * time.Second
		}
	}
	return defaultVal
}
