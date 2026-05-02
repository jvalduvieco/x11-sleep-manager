package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const defaultSocketName = "x11_sleep_manager.sock"

type Config struct {
	Socket    SocketConfig    `json:"socket"`
	Match     MatchConfig     `json:"match"`
	Reconcile ReconcileConfig `json:"reconcile"`
	Logging   LoggingConfig   `json:"logging"`
}

type SocketConfig struct {
	Path string `json:"path"`
}

type MatchConfig struct {
	UID     string   `json:"uid"`
	Who     []string `json:"who"`
	WhatAny []string `json:"what_any"`
}

type ReconcileConfig struct {
	Interval Duration `json:"interval"`
}

type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

type Duration struct {
	time.Duration
}

func Default() Config {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	socketPath := filepath.Join(runtimeDir, defaultSocketName)
	if runtimeDir == "" {
		socketPath = filepath.Join(os.TempDir(), defaultSocketName)
	}

	return Config{
		Socket: SocketConfig{
			Path: socketPath,
		},
		Match: MatchConfig{
			UID:     "self",
			Who:     []string{"OpenCode"},
			WhatAny: []string{"idle", "sleep"},
		},
		Reconcile: ReconcileConfig{
			Interval: Duration{Duration: 15 * time.Second},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

func Load(path string) (Config, error) {
	if path == "" {
		cfg := Default()
		return cfg, cfg.Validate()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	cfg := Default()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	return cfg, cfg.Validate()
}

func (c Config) Validate() error {
	if c.Socket.Path == "" {
		return errors.New("socket.path must not be empty")
	}
	if c.Reconcile.Interval.Duration <= 0 {
		return errors.New("reconcile.interval must be positive")
	}
	if c.Logging.Level == "" {
		return errors.New("logging.level must not be empty")
	}
	if c.Logging.Format == "" {
		return errors.New("logging.format must not be empty")
	}
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("duration must be a string: %w", err)
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fmt.Errorf("parse duration: %w", err)
	}

	d.Duration = parsed
	return nil
}
