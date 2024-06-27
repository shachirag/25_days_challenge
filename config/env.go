package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func LoadENV() error {
	goEnv := os.Getenv("GO_ENV")
	if goEnv == "" || goEnv == "development" {
		if _, err := os.Stat(".env"); err == nil {
			// .env file exists, load it
			if loadErr := godotenv.Load(); loadErr != nil {
				return fmt.Errorf("Error loading .env file: %w", loadErr)
			}
		} else {
			// .env file does not exist, log or handle it as needed
			fmt.Println("No .env file found, skipping loading .env file")
		}
	}
	return nil
}
