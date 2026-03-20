package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App    AppConfig    `yaml:"app"`
	Site   SiteConfig   `yaml:"site"`
	SQLite SQLiteConfig `yaml:"sqlite"`
}

type AppConfig struct {
	Name          string `yaml:"name"`
	Addr          string `yaml:"addr"`
	BaseURL       string `yaml:"base_url"`
	IsProduction  bool   `yaml:"is_production"`
	SessionSecret string `yaml:"session_secret"`
}

type SiteConfig struct {
	Title       string `yaml:"title"`
	Keywords    string `yaml:"keywords"`
	Description string `yaml:"description"`
}

type SQLiteConfig struct {
	Path string `yaml:"path"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
