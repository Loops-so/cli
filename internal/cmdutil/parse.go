package cmdutil

import (
	"encoding/json"
	"fmt"
	"os"
)

func ParseJSONFile(flag, path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("--%s: %w", flag, err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("--%s must be a valid JSON object: %w", flag, err)
	}
	return m, nil
}
