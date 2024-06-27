package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func LoadENV() error {
	goEnv := os.Getenv("GO_ENV")
	if goEnv == "" || goEnv == "prod" {
		// Attempt to load the .env file
		fmt.Println("Attempting to load .env file...")
		err := godotenv.Load()
		if err != nil {
			fmt.Println("Error loading .env file:", err)
			return err
		}
		// Print environment variables loaded from .env file for confirmation
		fmt.Println(".env file loaded successfully:")
		for _, env := range os.Environ() {
			fmt.Println(env)
		}
	} else {
		fmt.Println("Skipping .env file loading in production mode")
	}
	return nil
}
