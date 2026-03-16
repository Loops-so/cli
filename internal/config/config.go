package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
)

const DefaultEndpointURL = "https://app.loops.so/api/v1"

type Config struct {
	APIKey      string
	EndpointURL string
}

type fileConfig struct {
	Config struct {
		APIKey string `toml:"api-key"`
	} `toml:"config"`
}

func Load() (*Config, error) {
	configPath, _ := xdg.SearchConfigFile("loops/loops.toml")
	return load(configPath)
}

func load(configPath string) (*Config, error) {
	cfg := &Config{
		EndpointURL: DefaultEndpointURL,
	}

	if configPath != "" {
		var fc fileConfig
		if _, err := toml.DecodeFile(configPath, &fc); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		cfg.APIKey = fc.Config.APIKey
	}

	if v := os.Getenv("LOOPS_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("LOOPS_ENDPOINT_URL"); v != "" {
		cfg.EndpointURL = v
	}

	if cfg.APIKey == "" {
		return nil, errors.New("LOOPS_API_KEY is not set")
	}

	return cfg, nil
}
