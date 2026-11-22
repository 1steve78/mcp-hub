package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App    AppConfig    `yaml:"app"`
	Ollama OllamaConfig `yaml:"ollama"`
}

type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type OllamaConfig struct {
	BaseURL        string `yaml:"base_url"`
	DefaultModel   string `yaml:"default_model"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

func Load(path string) (*Config, error) {

	cfg := &Config{
		App: AppConfig{
			Name:    "mcp-hub",
			Version: "0.1.0-dev",
		},
		Ollama: OllamaConfig{
			BaseURL:        "http://localhost:11434",
			DefaultModel:   "qwen3:1.7b",
			TimeoutSeconds: 60,
		},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
		}
	} else {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	if val, ok := os.LookupEnv("MCP_HUB_OLLAMA_HOST"); ok {
		cfg.Ollama.BaseURL = strings.TrimSpace(val)
	}
	if val, ok := os.LookupEnv("MCP_HUB_MODEL"); ok {
		cfg.Ollama.DefaultModel = strings.TrimSpace(val)
	}
	if val, ok := os.LookupEnv("MCP_HUB_TIMEOUT"); ok {
		timeout, err := strconv.Atoi(strings.TrimSpace(val))
		if err != nil {
			return nil, fmt.Errorf("invalid MCP_HUB_TIMEOUT %q: must be an integer: %w", val, err)
		}
		cfg.Ollama.TimeoutSeconds = timeout
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.Ollama.BaseURL) == "" {
		return errors.New("ollama.base_url cannot be empty")
	}

	u, err := url.Parse(c.Ollama.BaseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("ollama.base_url %q must be a valid http or https URL", c.Ollama.BaseURL)
	}

	if strings.TrimSpace(c.Ollama.DefaultModel) == "" {
		return errors.New("ollama.default_model cannot be empty")
	}

	if c.Ollama.TimeoutSeconds <= 0 {
		return fmt.Errorf("ollama.timeout_seconds must be greater than 0, got %d", c.Ollama.TimeoutSeconds)
	}

	return nil
}
