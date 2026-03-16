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

func EndpointURL() string {
	if v := os.Getenv("LOOPS_ENDPOINT_URL"); v != "" {
		return v
	}
	return DefaultEndpointURL
}

func Load() (*Config, error) {
	cfg := &Config{
		EndpointURL: EndpointURL(),
	}

	apiKey, err := keyring.Get(keyringService, keyringUser)
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return nil, fmt.Errorf("could not read keyring: %w", err)
	}
	cfg.APIKey = apiKey

	if v := os.Getenv("LOOPS_API_KEY"); v != "" {
		cfg.APIKey = v
	}

	if cfg.APIKey == "" {
		return nil, errors.New("LOOPS_API_KEY is not set and no stored API credentials were found")
	}

	return cfg, nil
}

func Save(apiKey string) error {
	return keyring.Set(keyringService, keyringUser, apiKey)
}

func Delete() error {
	err := keyring.Delete(keyringService, keyringUser)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}
