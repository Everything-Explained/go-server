package configs

import (
	"os"

	"github.com/Everything-Explained/go-server/internal"
	"github.com/joho/godotenv"
)

var env Environment

type Environment struct {
	InDev          bool
	ConfigFilePath string
}

func GetEnv() Environment {
	if env.ConfigFilePath == "" {
		wd := internal.Getwd()
		if err := godotenv.Load(wd + "/configs/.env.dev"); err != nil {
			if err = godotenv.Load(wd + "/configs/.env.prod"); err != nil {
				panic(err)
			}
		}

		env = Environment{
			InDev:          os.Getenv("ENV") == "dev",
			ConfigFilePath: wd + "/configs/" + os.Getenv("CONFIG_FILE"),
		}
	}
	return env
}
