package cmdutil

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func ParseKeyValuePairs(flag string, pairs []string, dst map[string]any) (map[string]any, error) {
	for _, pair := range pairs {
		idx := strings.IndexByte(pair, '=')
		if idx < 0 {
			return nil, fmt.Errorf("--%s %q: expected KEY=value", flag, pair)
		}
		if dst == nil {
			dst = make(map[string]any)
		}
		dst[pair[:idx]] = pair[idx+1:]
	}
	return dst, nil
}

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
