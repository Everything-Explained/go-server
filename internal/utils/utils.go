package utils

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type environment struct {
	InDev          bool
	ConfigFilePath string
}

var (
	Env        environment
	WorkingDir string
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	WorkingDir = wd

	if err = godotenv.Load(WorkingDir + "\\.env.dev"); err != nil {
		if err = godotenv.Load(WorkingDir + "\\.env.prod"); err != nil {
			panic(err)
		}
	}

	Env = environment{
		InDev:          os.Getenv("ENV") == "dev",
		ConfigFilePath: WorkingDir + "\\" + os.Getenv("CONFIG_FILE"),
	}
}

func GetISODateNow() string {
	const ISODate8601Format = "2006-01-02T15:04:05.000Z07:00:00"
	return time.Now().UTC().Format(ISODate8601Format)
}
