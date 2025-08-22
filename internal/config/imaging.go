package config

import (
	"time"
)

// ImagingConfig configuración para procesamiento de imágenes
type ImagingConfig struct {
	TempDir           string   `yaml:"temp_dir"`
	MaxFileSize       int64    `yaml:"max_file_size"`      // Tamaño máximo en bytes
	DefaultResolution int      `yaml:"default_resolution"` // Resolución en DPI
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

// APIConfig configuración para la API REST
type APIConfig struct {
	Host    string        `yaml:"host"`
	Port    string        `yaml:"port"`
	Timeout TimeoutConfig `yaml:"timeout"`
	CORS    CORSConfig    `yaml:"cors"`
}

// TimeoutConfig configuración de timeouts para la API
type TimeoutConfig struct {
	Read     time.Duration `yaml:"read"`
	Write    time.Duration `yaml:"write"`
	Idle     time.Duration `yaml:"idle"`
	Shutdown time.Duration `yaml:"shutdown"`
}

// CORSConfig configuración de CORS
type CORSConfig struct {
	AllowOrigins     []string `yaml:"allow_origins"`
	AllowMethods     []string `yaml:"allow_methods"`
	AllowHeaders     []string `yaml:"allow_headers"`
	ExposeHeaders    []string `yaml:"expose_headers"`
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
