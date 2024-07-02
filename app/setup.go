package app

import (
	"challenge/config"
	"challenge/database"
)

func SetupAndRunApp() error {
	err := config.LoadENV()
	if err != nil {
		return err
	}

	err = database.SetupAWSClient()
	if err != nil {
		return err
	}

	return nil
}
