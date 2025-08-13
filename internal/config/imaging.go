package config

import "time"

type Imaging struct {
	Contrast      float64 `yaml:"contrast"`
	SharpenSigma  float64 `yaml:"sharpen_sigma"`
	SharpenAmount float64 `yaml:"sharpen_amount"`
	EdgeThreshold int     `yaml:"edge_threshold"`
	ResolutionDPI int     `yaml:"resolution_dpi"`
	TempDir       string  `yaml:"temp_dir"`
}

type Chroma struct {
	ColorThreshold float64 `yaml:"color_threshold"`
	SampleSize     int     `yaml:"sample_size"`
}

type Server struct {
	Address string        `yaml:"address"`
	Timeout TimeoutConfig `yaml:"timeout"`
}

type TimeoutConfig struct {
	Read  time.Duration `yaml:"read"`
	Write time.Duration `yaml:"write"`
	Idle  time.Duration `yaml:"idle"`
}

type Config struct {
	Imaging *Imaging `yaml:"imaging"`
	Chroma  *Chroma  `yaml:"chroma"`
	Server  *Server  `yaml:"server"`
}

func LoadConfig(path string) (*Config, error) {
	// Implementación de ejemplo con valores por defecto
	return &Config{
		Imaging: &Imaging{
			Contrast:      1.2,
			SharpenSigma:  1.0,
			SharpenAmount: 1.5,
			EdgeThreshold: 50,
			ResolutionDPI: 300,
		},
		Chroma: &Chroma{
			ColorThreshold: 0.05,
			SampleSize:     20,
		},
		Server: &Server{
			Address: ":8080",
			Timeout: TimeoutConfig{
				Read:  5 * time.Second,
				Write: 10 * time.Second,
				Idle:  15 * time.Second,
			},
		},
	}, nil
}
