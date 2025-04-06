package config

import (
	"encoding/json"
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

func Parse(path string) (*Config, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var config Config

	if err := json.Unmarshal(data, &config); err == nil {
		return &config, nil
	}

	if err := yaml.Unmarshal(data, &config); err == nil {
		return &config, nil
	}

	return nil, errors.New("failed to parse config file")
}

type Config struct {
	Servers map[string]Server `json:"servers" yaml:"servers"`
}

type Server struct {
	Type string `json:"type" yaml:"type"`

	URL     string            `json:"url" yaml:"url"`
	Headers map[string]string `json:"headers" yaml:"headers"`

	Command string            `json:"command" yaml:"command"`
	Env     map[string]string `json:"env" yaml:"env"`
	Args    []string          `json:"args" yaml:"args"`
}
