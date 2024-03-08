package lib

import (
	"os"

	"github.com/Everything-Explained/go-server/internal"
	"gopkg.in/yaml.v3"
)

type ConfigData struct {
	Red33mPassword string `yaml:"red33m-password"`
	SecurityHeader string `yaml:"security-header"`
	Port           int
	AllowedOrigins []string `yaml:"allowed-origins"`
	Mail           struct {
		Address    string
		Host       string
		Port       int
		RequireTLS bool `yaml:"require-tls"`
		Username   string
		Password   string
	}
}

var config ConfigData = ConfigData{}

func GetConfig() ConfigData {
	return copyConfig(config)
}

func init() {
	content, err := os.ReadFile(internal.GetEnv().ConfigFilePath)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(content, &config)
	if err != nil {
		panic(err)
	}
}

func copyConfig(c ConfigData) ConfigData {
	return c
}
