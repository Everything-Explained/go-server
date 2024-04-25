package configs

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Environment struct {
	InDev          bool
	ConfigFilePath string
}

func GetEnv(dir string) (*Environment, error) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("directory not found::%s", dir)
	}

	if err != nil {
		return nil, err
	}

	if err := godotenv.Load(dir + "/.env.dev"); err != nil {
		if err = godotenv.Load(dir + "/.env.prod"); err != nil {
			return nil, err
		}
	}

	return &Environment{
		InDev:          os.Getenv("ENV") == "dev",
		ConfigFilePath: dir + "/" + os.Getenv("CONFIG_FILE"),
	}, nil
}
