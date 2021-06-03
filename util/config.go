package util

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

// LoadConfig reads configuration from file or environment variables.
func GetEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
