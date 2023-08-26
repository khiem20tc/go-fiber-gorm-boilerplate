package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

/**
 * Get env variable
 */
 func Env(key string, fallback ...string) string {
	err := godotenv.Load(".env")

	value := os.Getenv(key)
	
	if err != nil || value == "" {
		if len(fallback) > 0 && fallback[0] != "" {
			return fallback[0]
		}else {
			log.Fatalf("Error loading .env file or variable %s is not set", key)
		}
	}

	return value
}