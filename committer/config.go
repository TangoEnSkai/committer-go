package committer

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const configFileName = ".committer.yaml"

// Config holds the optional per-project configuration for committer-go.
// All fields have sensible defaults; a missing config file is not an error.
type Config struct {
	Extends string       `yaml:"extends"`
	Length  LengthConfig `yaml:"length"`
	Types   TypesConfig  `yaml:"types"`
	Body    BodyConfig   `yaml:"body"`
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
	data, err := os.ReadFile(configFileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}
		return DefaultConfig(), err
	}

	// Unmarshal into zero-value struct so we can detect which fields were set.
	var local Config
	if err := yaml.Unmarshal(data, &local); err != nil {
		return DefaultConfig(), err
	}

	// Handle extends: load base config and merge local on top.
	if local.Extends != "" {
		base, err := loadExtendsConfig(local.Extends)
		if err != nil {
			return DefaultConfig(), fmt.Errorf("extends: %w", err)
		}
		local = mergeConfigs(base, local)
	}

	// Apply zero-value fallbacks so partially-specified files keep defaults.
	if local.Length.Min == 0 {
		local.Length.Min = minLength
	}
	if local.Length.Max == 0 {
		local.Length.Max = maxLength
	}
	if local.Body.MaxLineLength == 0 {
		local.Body.MaxLineLength = 72
	}

	return local, nil
}

// loadExtendsConfig fetches or reads the config at the given extends path/URL.
func loadExtendsConfig(source string) (Config, error) {
	var data []byte
	var err error

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(source) //nolint:noctx
		if err != nil {
			return Config{}, fmt.Errorf("fetch %q: %w", source, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return Config{}, fmt.Errorf("fetch %q: HTTP %d", source, resp.StatusCode)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return Config{}, fmt.Errorf("read body %q: %w", source, err)
		}
	} else {
		data, err = os.ReadFile(source)
		if err != nil {
			return Config{}, fmt.Errorf("read file %q: %w", source, err)
		}
	}

	var base Config
	if err := yaml.Unmarshal(data, &base); err != nil {
		return Config{}, fmt.Errorf("parse %q: %w", source, err)
	}
	return base, nil
}

// mergeConfigs merges local on top of base. Zero values in local retain base values.
func mergeConfigs(base, local Config) Config {
	result := base
	// Extends is not inherited from base.
	result.Extends = ""

	if local.Length.Min != 0 {
		result.Length.Min = local.Length.Min
	}
	if local.Length.Max != 0 {
		result.Length.Max = local.Length.Max
	}
	if local.Body.MaxLineLength != 0 {
		result.Body.MaxLineLength = local.Body.MaxLineLength
	}
	// Extra types are additive.
	result.Types.Extra = append(base.Types.Extra, local.Types.Extra...)
	return result
}
