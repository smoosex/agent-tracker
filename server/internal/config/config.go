package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	DataDir string `toml:"data_dir"`
	LogPath string `toml:"log_path"`
	Port    string `toml:"port"`
}

func Load(configPath string) (Config, error) {
	if configPath == "" {
		var err error
		configPath, err = findConfigPath()
		if err != nil {
			return Config{}, err
		}
	} else {
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			return Config{}, fmt.Errorf("resolve config path: %w", err)
		}
		configPath = absPath
	}

	if _, err := os.Stat(configPath); err != nil {
		return Config{}, fmt.Errorf("config not found: %w", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if cfg.DataDir == "" {
		return Config{}, errors.New("config.data_dir is required")
	}
	if cfg.LogPath == "" {
		return Config{}, errors.New("config.log_path is required")
	}
	if cfg.Port == "" {
		return Config{}, errors.New("config.port is required")
	}

	if !filepath.IsAbs(cfg.DataDir) {
		cfg.DataDir = filepath.Clean(filepath.Join(filepath.Dir(configPath), cfg.DataDir))
	}
	if !filepath.IsAbs(cfg.LogPath) {
		cfg.LogPath = filepath.Clean(filepath.Join(filepath.Dir(configPath), cfg.LogPath))
	}

	return cfg, nil
}

func LoadFromArgs(args []string) (Config, error) {
	configPath, err := parseConfigPath(args)
	if err != nil {
		return Config{}, err
	}

	return Load(configPath)
}

func findConfigPath() (string, error) {
	candidates := []string{
		"config.toml",
		filepath.Join("server", "config.toml"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return filepath.Abs(candidate)
		}
	}

	return "", errors.New("config.toml not found")
}

func parseConfigPath(args []string) (string, error) {
	fs := flag.NewFlagSet("agent-tracker", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var configPath string
	fs.StringVar(&configPath, "config", "", "path to config.toml")
	fs.StringVar(&configPath, "c", "", "path to config.toml")

	if err := fs.Parse(args); err != nil {
		return "", err
	}

	return configPath, nil
}
