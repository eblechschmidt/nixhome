package cfg

import (
	"html/template"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Colors *Colors
	Apps   map[string]Apps
}

type Apps []*App

type App struct {
	Icon          template.HTML
	Name          string
	URL           string
	ColorizedIcon template.HTML
}

type Colors struct {
	Dark  *ColorSet
	Light *ColorSet
	Icon  string
}

type ColorSet struct {
	Background string
	Text       string
	Accent     string
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
