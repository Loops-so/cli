package config

import (
	"os"
	"path/filepath"
)

func ConfigDir() (string, error) {
	if dir := os.Getenv("LOOPS_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "loops"), nil
}
