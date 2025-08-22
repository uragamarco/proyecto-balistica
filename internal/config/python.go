package config

import (
	"time"
)

// PythonConfig configuración para el servicio Python
type PythonConfig struct {
	Timeout time.Duration `yaml:"timeout"`
	Enabled bool          `yaml:"enabled"`
	Path    string        `yaml:"path"`
	Script  string        `yaml:"script"`
	Venv    string        `yaml:"venv"`
}
