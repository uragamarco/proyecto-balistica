package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config representa la configuración global de la aplicación
type Config struct {
	Environment string         `yaml:"environment"`
	Logging     LoggingConfig  `yaml:"logging"`
	Imaging     ImagingConfig  `yaml:"imaging"`
	Chroma      ChromaConfig   `yaml:"chroma"`
	API         APIConfig      `yaml:"api"`
	Database    DatabaseConfig `yaml:"database,omitempty"`
	Security    SecurityConfig `yaml:"security,omitempty"`
}

// LoggingConfig configuración para el sistema de logging
type LoggingConfig struct {
	Level   string `yaml:"level"`    // debug, info, warn, error
	Output  string `yaml:"output"`   // stdout, file, both
	File    string `yaml:"file"`     // Ruta del archivo de log
	MaxSize int    `yaml:"max_size"` // Tamaño máximo en MB
}

// ImagingConfig configuración para procesamiento de imágenes
type ImagingConfig struct {
	TempDir           string   `yaml:"temp_dir"`
	MaxFileSize       int64    `yaml:"max_file_size"`      // Tamaño máximo en bytes
	DefaultResolution string   `yaml:"default_resolution"` // Ej: 1200dpi
	MinQuality        int      `yaml:"min_quality"`        // Calidad mínima aceptable
	Contrast          float64  `yaml:"contrast"`
	SharpenSigma      float64  `yaml:"sharpen_sigma"`
	SharpenAmount     float64  `yaml:"sharpen_amount"`
	EdgeThreshold     int      `yaml:"edge_threshold"`
	ResolutionDPI     int      `yaml:"resolution_dpi"`
	Formats           []string `yaml:"formats"` // Formatos soportados
}

// ChromaConfig configuración para análisis de color
type ChromaConfig struct {
	ColorThreshold float64 `yaml:"color_threshold"`
	SampleSize     int     `yaml:"sample_size"`
	ColorSpace     string  `yaml:"color_space"` // rgb, lab, hsv
	MinContrast    float64 `yaml:"min_contrast"`
}

// APIConfig configuración para el servidor API
type APIConfig struct {
	Host    string        `yaml:"host"`
	Port    int           `yaml:"port"`
	Timeout TimeoutConfig `yaml:"timeout"`
	CORS    CORSConfig    `yaml:"cors"`
}

// TimeoutConfig configuración de tiempos de espera
type TimeoutConfig struct {
	Read     time.Duration `yaml:"read"`
	Write    time.Duration `yaml:"write"`
	Idle     time.Duration `yaml:"idle"`
	Shutdown time.Duration `yaml:"shutdown"`
}

// CORSConfig configuración para CORS
type CORSConfig struct {
	Enabled          bool     `yaml:"enabled"`
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

// DatabaseConfig configuración para base de datos
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
}

// SecurityConfig configuración de seguridad
type SecurityConfig struct {
	JWTSecret       string `yaml:"jwt_secret"`
	TokenExpiration int    `yaml:"token_expiration"` // en horas
	RateLimit       int    `yaml:"rate_limit"`       // peticiones por segundo
}

// LoadConfig carga la configuración desde un archivo YAML
func LoadConfig(path string) (*Config, error) {
	// Resolver ruta absoluta
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Establecer valores por defecto
	cfg = setDefaults(cfg)

	return &cfg, nil
}

// setDefaults establece valores por defecto para configuraciones faltantes
func setDefaults(cfg Config) Config {
	// Entorno
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}

	// Logging
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Output == "" {
		cfg.Logging.Output = "both"
	}
	if cfg.Logging.File == "" {
		cfg.Logging.File = "balistica.log"
	}
	if cfg.Logging.MaxSize == 0 {
		cfg.Logging.MaxSize = 100 // 100MB
	}

	// Imaging
	if cfg.Imaging.TempDir == "" {
		cfg.Imaging.TempDir = os.TempDir()
	}
	if cfg.Imaging.MaxFileSize == 0 {
		cfg.Imaging.MaxFileSize = 10 << 20 // 10MB
	}
	if cfg.Imaging.DefaultResolution == "" {
		cfg.Imaging.DefaultResolution = "1200dpi"
	}
	if cfg.Imaging.MinQuality == 0 {
		cfg.Imaging.MinQuality = 90
	}
	if cfg.Imaging.Formats == nil {
		cfg.Imaging.Formats = []string{"image/jpeg", "image/png", "image/tiff"}
	}

	// API
	if cfg.API.Host == "" {
		cfg.API.Host = "0.0.0.0"
	}
	if cfg.API.Port == 0 {
		cfg.API.Port = 8080
	}
	if cfg.API.Timeout.Read == 0 {
		cfg.API.Timeout.Read = 15 * time.Second
	}
	if cfg.API.Timeout.Write == 0 {
		cfg.API.Timeout.Write = 30 * time.Second
	}
	if cfg.API.Timeout.Idle == 0 {
		cfg.API.Timeout.Idle = 60 * time.Second
	}
	if cfg.API.Timeout.Shutdown == 0 {
		cfg.API.Timeout.Shutdown = 30 * time.Second
	}

	// Chroma
	if cfg.Chroma.ColorSpace == "" {
		cfg.Chroma.ColorSpace = "rgb"
	}
	if cfg.Chroma.MinContrast == 0 {
		cfg.Chroma.MinContrast = 0.3
	}

	return cfg
}
