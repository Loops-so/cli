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

func Save(apiKey string) error {
	configPath, err := xdg.ConfigFile("loops/loops.toml")
	if err != nil {
		return fmt.Errorf("could not determine config path: %w", err)
	}
	return saveToPath(configPath, apiKey)
}

func saveToPath(configPath string, apiKey string) error {
	f, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}
	defer f.Close()

	if err := f.Chmod(0600); err != nil {
		return fmt.Errorf("could not set config file permissions: %w", err)
	}

	return toml.NewEncoder(f).Encode(fileConfig{
		Config: struct {
			APIKey string `toml:"api-key"`
		}{APIKey: apiKey},
	})
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
