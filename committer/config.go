package committer

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

const configFileName = ".committer.yaml"

// Config holds the optional per-project configuration for committer-go.
// All fields have sensible defaults; a missing config file is not an error.
type Config struct {
	Length LengthConfig `yaml:"length"`
	Types  TypesConfig  `yaml:"types"`
	Body   BodyConfig   `yaml:"body"`
}

// BodyConfig controls validation of the commit message body.
type BodyConfig struct {
	// MaxLineLength is the maximum allowed length for body lines.
	// Default is 72. Set to 0 to disable.
	MaxLineLength int `yaml:"max_line_length"`
}

// LengthConfig controls the minimum and maximum commit header length.
type LengthConfig struct {
	Min int `yaml:"min"`
	Max int `yaml:"max"`
}

// TypesConfig allows projects to extend the built-in type list.
type TypesConfig struct {
	Extra []string `yaml:"extra"`
}

// DefaultConfig returns a Config populated with the built-in defaults.
func DefaultConfig() Config {
	return Config{
		Length: LengthConfig{
			Min: minLength,
			Max: maxLength,
		},
		Body: BodyConfig{
			MaxLineLength: 72,
		},
	}
}

// LoadConfig reads .committer.yaml from the current working directory.
// If the file does not exist the default config is returned without error.
func LoadConfig() (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(configFileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	// Apply zero-value fallbacks so partially-specified files keep defaults.
	if cfg.Length.Min == 0 {
		cfg.Length.Min = minLength
	}
	if cfg.Length.Max == 0 {
		cfg.Length.Max = maxLength
	}
	if cfg.Body.MaxLineLength == 0 {
		cfg.Body.MaxLineLength = 72
	}

	return cfg, nil
}
