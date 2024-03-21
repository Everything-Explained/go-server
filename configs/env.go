package configs

import (
	"os"

	"github.com/Everything-Explained/go-server/internal"
	"github.com/joho/godotenv"
)

var env environment

type environment struct {
	InDev          bool
	ConfigFilePath string
}

func GetEnv() environment {
	if env.ConfigFilePath == "" {
		wd := internal.GetWorkingDir()
		if err := godotenv.Load(wd + "/configs/.env.dev"); err != nil {
			if err = godotenv.Load(wd + "/configs/.env.prod"); err != nil {
				panic(err)
			}
		}

		env = environment{
			InDev:          os.Getenv("ENV") == "dev",
			ConfigFilePath: wd + "/configs/" + os.Getenv("CONFIG_FILE"),
		}
	}
	return env
}
