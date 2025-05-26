package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	OllamaServerUri string `yaml:"ollama-server-uri"`
	RedisServerUri  string `yaml:"redis-server-uri"`
	Model           string `yaml:"model"`
	LlmContext      string `yaml:"llm-context"`
}

var config *Config

func GetSystemConfig() (*Config, error) {
	if config != nil {
		return config, nil
	}

	filename := "config.yaml"
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config = &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return config, nil
}
