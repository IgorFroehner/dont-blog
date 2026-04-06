package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type SiteConfig struct {
	Title        string            `yaml:"title"`
	Description  string            `yaml:"description"`
	Author       string            `yaml:"author"`
	URL          string            `yaml:"url"`
	Intro        string            `yaml:"intro"`
	Social       map[string]string `yaml:"social"`
	PostsPerPage int               `yaml:"postsPerPage"`
}

func Load(path string) (*SiteConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &SiteConfig{
		PostsPerPage: 10,
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if cfg.PostsPerPage <= 0 {
		cfg.PostsPerPage = 10
	}

	return cfg, nil
}
