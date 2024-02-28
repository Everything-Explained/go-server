package lib

import (
	"os"

	"github.com/Everything-Explained/go-server/internal/utils"
	"gopkg.in/yaml.v3"
)

type config struct {
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

var Config config = config{}

func init() {
	content, err := os.ReadFile(utils.Env.ConfigFilePath)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(content, &Config)
	if err != nil {
		panic(err)
	}
}
