package cfg

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Apps map[string]Apps
}

type Apps []App

type App struct {
	Icon string
	Name string
	URL  string
}

// FromFile takes a yaml file and unmarshals the config
func FromFile(file string) (*Config, error) {
	var c Config

	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
