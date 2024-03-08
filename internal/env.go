package internal

import (
	"os"

	"github.com/Everything-Explained/go-server/internal/utils"
	"github.com/joho/godotenv"
)

var env environment

type environment struct {
	InDev          bool
	ConfigFilePath string
}

func GetEnv() environment {
	if env.ConfigFilePath == "" {
		if err := godotenv.Load(utils.GetWorkingDir() + "\\.env.dev"); err != nil {
			if err = godotenv.Load(utils.GetWorkingDir() + "\\.env.prod"); err != nil {
				panic(err)
			}
		}

		env = environment{
			InDev:          os.Getenv("ENV") == "dev",
			ConfigFilePath: utils.GetWorkingDir() + "\\" + os.Getenv("CONFIG_FILE"),
		}
	}
	return env
}
