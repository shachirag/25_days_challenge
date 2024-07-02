package app

import (
	"challenge/database"
)

func SetupAndRunApp() error {
	// loaded the env config
	// err := config.LoadENV()
	// if err != nil {
	// 	return err
	// }

	err := database.SetupAWSClient()
	if err != nil {
		return err
	}

	return nil
}
