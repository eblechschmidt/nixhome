package cfg

import (
	"html/template"
	"os"

	"github.com/eblechschmidt/nixhome/internal/theme"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Colors    *Colors
	Apps      map[string]Apps
	Bookmarks map[string]Bookmarks
}

type Apps []*App

type App struct {
	Icon          template.HTML
	Name          string
	URL           string
	ColorizedIcon template.HTML
}

type Bookmarks []*Bookmark

type Bookmark struct {
	Name string
	URL  string
}

type Colors struct {
	Dark  *ColorSet
	Light *ColorSet
	Icon  theme.Color
}

type ColorSet struct {
	Background theme.Color
	Text       theme.Color
	Accent     theme.Color
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

	// check colors

	return &c, nil
}
