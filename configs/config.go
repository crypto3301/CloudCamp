package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Server   ServerSettings `json:"server"`
	Backends []BackendsPool `json:"backends"`
}

type ServerSettings struct {
	Port         int `json:"port"`
	ReadTimeout  int `json:"read_timeout"`
	WriteTimeout int `json:"write_timeout"`
}

type BackendsPool struct {
	URL    string `json:"url"`
	Weight int    `json:"weight"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
