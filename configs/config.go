package configs

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigData struct {
	Red33mPassword string `yaml:"red33m-password"`
	SecurityHeader string `yaml:"security-header"`
	Port           int
	ClientPath     string   `yaml:"client-path"`
	DataPath       string   `yaml:"data-path"`
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

// IMPORTANT: Implement embedding config.yml files
func GetConfig(dir string) (*ConfigData, error) {
	env, err := GetEnv(dir)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(env.ConfigFilePath)
	if err != nil {
		return nil, err
	}

	var config ConfigData
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
