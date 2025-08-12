package python_features

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/uragamarco/proyecto-balistica/pkg/integration"
)

type Service struct {
	bridge integration.FeatureExtractor
}

func NewService(bridge integration.FeatureExtractor) *Service {
	return &Service{bridge: bridge}
}

func (s *Service) ExtractAdvancedFeatures(imagePath string) (*integration.PythonResponse, error) {
	if s.bridge == nil {
		return nil, ErrPythonNotAvailable
	}

	// Obtener ruta absoluta para compatibilidad multiplataforma
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo ruta absoluta: %w", err)
	}

	return s.bridge.ExtractFeatures(absPath)
}

// HealthCheck verifica la disponibilidad del servicio Python
func (s *Service) HealthCheck() error {
	if s.bridge == nil {
		return ErrPythonNotAvailable
	}
	return s.bridge.HealthCheck()
}

// GetFeatureNames devuelve los nombres de las características para referencia
func (s *Service) GetFeatureNames() []string {
	return []string{
		"hu_moment_1", "hu_moment_2", "hu_moment_3",
		"hu_moment_4", "hu_moment_5", "hu_moment_6", "hu_moment_7",
		"striation_1", "striation_2", "striation_3", "striation_4",
		"striation_5", "striation_6", "striation_7", "striation_8",
		"striation_9", "striation_10",
		"contour_area", "contour_length",
	}
}

var ErrPythonNotAvailable = errors.New("python integration not available")
