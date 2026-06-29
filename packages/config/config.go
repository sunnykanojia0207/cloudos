// Package config provides a unified configuration framework for CloudOS.
// It supports YAML files with environment variable interpolation, hot-reload
// via file watching, and a hierarchical configuration model where every
// subsystem reads from a single Config struct.
package config

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

// Config is the single, top-level configuration structure for CloudOS.
// Every subsystem reads its configuration from this struct.
type Config struct {
	Kernel  KernelConfig  `yaml:"kernel" json:"kernel"`
	API     APIConfig     `yaml:"api" json:"api"`
	Auth    AuthConfig    `yaml:"auth" json:"auth"`
	Logging LoggingConfig `yaml:"logging" json:"logging"`
}

// KernelConfig carries kernel-level settings.
type KernelConfig struct {
	LogLevel  string `yaml:"log_level" json:"logLevel"`
	DataDir   string `yaml:"data_dir" json:"dataDir"`
	PluginDir string `yaml:"plugin_dir" json:"pluginDir"`
}

// APIConfig carries HTTP API server settings.
type APIConfig struct {
	Host string `yaml:"host" json:"host"`
	Port int    `yaml:"port" json:"port"`
}

// AuthConfig carries authentication settings.
type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret" json:"jwtSecret"`
	TokenTTL  string `yaml:"token_ttl" json:"tokenTtl"`
}

// LoggingConfig carries logging settings.
type LoggingConfig struct {
	Level  string `yaml:"level" json:"level"`
	Format string `yaml:"format" json:"format"`
	Output string `yaml:"output" json:"output"`
	Path   string `yaml:"path" json:"path,omitempty"`
}

// DefaultConfig returns a Config with safe development defaults.
// All values can be overridden via environment variables in the YAML file.
func DefaultConfig() Config {
	return Config{
		Kernel: KernelConfig{
			LogLevel:  "info",
			DataDir:   "/var/lib/cloudos",
			PluginDir: "/var/lib/cloudos/plugins",
		},
		API: APIConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Auth: AuthConfig{
			JWTSecret: "change-me-in-production",
			TokenTTL:  "24h",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
	}
}

// Provider defines the interface for fetching configuration.
type Provider interface {
	Load(path string) (*Config, error)
	Watch(path string, onChange func(*Config)) error
	Close() error
}

// envPattern matches ${VAR_NAME} and ${VAR_NAME:-default} syntax.
var envPattern = regexp.MustCompile(`\$\{([^:}]+)(?::-([^}]*))?\}`)

// interpolateEnv replaces environment variable references in YAML content.
// Supports:
//   ${VAR}           — required variable, left as-is if unset
//   ${VAR:-default}  — optional variable with default
func interpolateEnv(data []byte) []byte {
	return envPattern.ReplaceAllFunc(data, func(match []byte) []byte {
		parts := envPattern.FindSubmatch(match)
		name := string(parts[1])
		def := ""
		if len(parts) > 2 && parts[2] != nil {
			def = string(parts[2])
		}

		val, ok := os.LookupEnv(name)
		if !ok {
			if def != "" {
				return []byte(def)
			}
			return match // leave unresolved
		}
		return []byte(val)
	})
}

// YAMLProvider implements Provider for YAML configuration files.
type YAMLProvider struct {
	mu      sync.RWMutex
	cfg     *Config
	watcher *fsnotify.Watcher
	done    chan struct{}
}

// NewYAMLProvider creates a new YAML-based configuration provider.
func NewYAMLProvider() *YAMLProvider {
	return &YAMLProvider{done: make(chan struct{})}
}

// Load reads a YAML file, interpolates environment variables, and parses the
// result into a Config struct.
func (p *YAMLProvider) Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	expanded := interpolateEnv(data)

	var cfg Config
	dec := yaml.NewDecoder(bytes.NewReader(expanded))
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	p.mu.Lock()
	p.cfg = &cfg
	p.mu.Unlock()

	return &cfg, nil
}

// Watch monitors the config file for changes. When a write or create event is
// detected it re-loads the file and invokes onChange with the new Config.
func (p *YAMLProvider) Watch(path string, onChange func(*Config)) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	p.watcher = w

	if err := w.Add(path); err != nil {
		w.Close()
		return fmt.Errorf("watch file: %w", err)
	}

	go func() {
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					cfg, err := p.Load(path)
					if err != nil {
						fmt.Fprintf(os.Stderr, "config: reload failed: %v\n", err)
						continue
					}
					if onChange != nil {
						onChange(cfg)
					}
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				fmt.Fprintf(os.Stderr, "config: watch error: %v\n", err)
			case <-p.done:
				return
			}
		}
	}()

	return nil
}

// Close stops the background file watcher.
func (p *YAMLProvider) Close() error {
	close(p.done)
	if p.watcher != nil {
		return p.watcher.Close()
	}
	return nil
}

// Get returns the currently loaded configuration. Safe for concurrent use.
func (p *YAMLProvider) Get() *Config {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.cfg
}
