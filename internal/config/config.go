package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config estructura principal de configuración
type Config struct {
	App      AppConfig      `yaml:"app"`
	Server   ServerConfig   `yaml:"server"`
	Logging  LoggingConfig  `yaml:"logging"`
	Python   PythonConfig   `yaml:"python"`
	Imaging  ImagingConfig  `yaml:"imaging"`
	API      APIConfig      `yaml:"api"`
	Chroma   ChromaConfig   `yaml:"chroma"`
	Cache    *CacheConfig   `yaml:"cache,omitempty"`
	Database DatabaseConfig `yaml:"database,omitempty"`
	Security SecurityConfig `yaml:"security,omitempty"`
}

// LoggingConfig configuración para logging
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"`
}

// AppConfig configuración de la aplicación
type AppConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
}

// ServerConfig configuración del servidor
type ServerConfig struct {
	Port string `yaml:"port"`
}

// CacheConfig configuración para el sistema de cache
type CacheConfig struct {
	Enabled         bool          `yaml:"enabled"`
	Directory       string        `yaml:"directory"`
	MemoryTTL       time.Duration `yaml:"memory_ttl"`
	DiskTTL         time.Duration `yaml:"disk_ttl"`
	MaxMemoryMB     int           `yaml:"max_memory_mb"`
	Compress        bool          `yaml:"compress"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
}

// Load carga la configuración desde un archivo YAML
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Establecer valores por defecto
	setDefaults(&cfg)

	// Convertir timeout de Python a duración
	if cfg.Python.Timeout == 0 {
		cfg.Python.Timeout = 30 * time.Second
	}

	return &cfg, nil
}

// setDefaults establece valores predeterminados para la configuración
func setDefaults(cfg *Config) {
	// App defaults
	if cfg.App.Environment == "" {
		cfg.App.Environment = "development"
	}

	// Logging defaults
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Output == "" {
		cfg.Logging.Output = "console"
	}

	// Python defaults
	if cfg.Python.Path == "" {
		cfg.Python.Path = "python3"
	}
	if cfg.Python.Script == "" {
		cfg.Python.Script = "py_services/feature_extractor.py"
	}
	if cfg.Python.Venv == "" {
		cfg.Python.Venv = "balistics-env"
	}
	cfg.Python.Enabled = true

	// Imaging defaults
	if cfg.Imaging.TempDir == "" {
		cfg.Imaging.TempDir = "/tmp/balistica"
	}
	if cfg.Imaging.MaxFileSize == 0 {
		cfg.Imaging.MaxFileSize = 10 * 1024 * 1024 // 10MB
	}
	if cfg.Imaging.DefaultResolution == 0 {
		cfg.Imaging.DefaultResolution = 300
	}
	if cfg.Imaging.MinQuality == 0 {
		cfg.Imaging.MinQuality = 75
	}

	// API defaults
	if cfg.API.Host == "" {
		cfg.API.Host = "0.0.0.0"
	}
	if cfg.API.Port == "" {
		cfg.API.Port = "8080"
	}

	// Chroma defaults
	if cfg.Chroma.ColorThreshold == 0 {
		cfg.Chroma.ColorThreshold = 0.05
	}
	if cfg.Chroma.SampleSize == 0 {
		cfg.Chroma.SampleSize = 1000
	}

	// Cache defaults
	if cfg.Cache == nil {
		cfg.Cache = &CacheConfig{}
	}
	if cfg.Cache.Directory == "" {
		cfg.Cache.Directory = "./cache"
	}
	if cfg.Cache.MemoryTTL == 0 {
		cfg.Cache.MemoryTTL = 5 * time.Minute
	}
	if cfg.Cache.DiskTTL == 0 {
		cfg.Cache.DiskTTL = 30 * time.Minute
	}
	if cfg.Cache.MaxMemoryMB == 0 {
		cfg.Cache.MaxMemoryMB = 100
	}
	if cfg.Cache.CleanupInterval == 0 {
		cfg.Cache.CleanupInterval = 10 * time.Minute
	}
	cfg.Cache.Enabled = true
	cfg.Cache.Compress = true
}
