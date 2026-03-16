package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
)

const (
	DefaultEndpointURL = "https://app.loops.so/api/v1"
	keyringService     = "loops"
	keyringUser        = "api-key"
)

type Config struct {
	APIKey      string
	EndpointURL string
}

func Load() (*Config, error) {
	cfg := &Config{
		EndpointURL: DefaultEndpointURL,
	}

	apiKey, err := keyring.Get(keyringService, keyringUser)
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return nil, fmt.Errorf("could not read keyring: %w", err)
	}
	cfg.APIKey = apiKey

	if v := os.Getenv("LOOPS_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("LOOPS_ENDPOINT_URL"); v != "" {
		cfg.EndpointURL = v
	}

	if cfg.APIKey == "" {
		return nil, errors.New("LOOPS_API_KEY is not set and no stored API credentials were found")
	}

	return cfg, nil
}

func Save(apiKey string) error {
	return keyring.Set(keyringService, keyringUser, apiKey)
}
