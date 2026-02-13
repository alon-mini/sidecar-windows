package keymap

import (
	"encoding/json"
	"os"
)

// Config represents user key binding configuration.
type Config struct {
	Bindings map[string]string `json:"bindings"` // key -> command ID
}

// LoadConfig loads key binding overrides from a JSON file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Bindings: make(map[string]string)}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Bindings == nil {
		cfg.Bindings = make(map[string]string)
	}

	return &cfg, nil
}

// ApplyConfig applies user configuration overrides to the registry.
func ApplyConfig(r *Registry, cfg *Config) {
	for key, cmdID := range cfg.Bindings {
		r.SetUserOverride(key, cmdID)
	}
}
