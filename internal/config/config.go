package config

import (
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration.
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Cache   CacheConfig   `yaml:"cache"`
	Scraper ScraperConfig `yaml:"scraper"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// CacheConfig holds caching settings.
type CacheConfig struct {
	TTL time.Duration `yaml:"ttl"`
}

// ScraperConfig holds scraper HTTP client settings.
type ScraperConfig struct {
	UserAgent      string        `yaml:"user_agent"`
	RequestTimeout time.Duration `yaml:"request_timeout"`
}

// Defaults returns a Config with sensible default values.
func Defaults() Config {
	return Config{
		Server: ServerConfig{
			Port:         8080,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		Cache: CacheConfig{
			TTL: 30 * time.Minute,
		},
		Scraper: ScraperConfig{
			UserAgent:      "rssify/1.0",
			RequestTimeout: 15 * time.Second,
		},
	}
}

// Load reads config from a YAML file (if it exists) and applies environment variable overrides.
func Load(path string) (Config, error) {
	cfg := Defaults()

	data, err := os.ReadFile(path)
	if err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, err
		}
	}

	applyEnvOverrides(&cfg)
	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("RSS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("RSS_CACHE_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Cache.TTL = d
		}
	}
	if v := os.Getenv("RSS_USER_AGENT"); v != "" {
		cfg.Scraper.UserAgent = v
	}
	if v := os.Getenv("RSS_REQUEST_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Scraper.RequestTimeout = d
		}
	}
}
