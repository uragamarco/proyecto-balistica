package integration

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

// PythonService proporciona una interfaz para interactuar con servicios Python
type PythonService struct {
	Logger  *zap.Logger
	Timeout time.Duration
	bridge  FeatureExtractor
}

// NewPythonService crea una nueva instancia del servicio Python
func NewPythonService(logger *zap.Logger, timeout time.Duration) *PythonService {
	// Inicializar el extractor RPC
	bridge := NewRPCExtractor()

	return &PythonService{
		Logger:  logger,
		Timeout: timeout,
		bridge:  bridge,
	}
}

// ExtractFeatures extrae características de una imagen utilizando Python
func (s *PythonService) ExtractFeatures(imagePath string) (*PythonResponse, error) {
	s.Logger.Debug("Extrayendo características con Python",
		zap.String("imagePath", imagePath))

	// Verificar que el archivo existe antes de enviarlo a Python
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		s.Logger.Error("El archivo de imagen no existe",
			zap.String("imagePath", imagePath))
		return nil, fmt.Errorf("el archivo de imagen no existe: %s", imagePath)
	}

	// Establecer un timeout para la operación
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	// Crear un canal para recibir el resultado
	ch := make(chan struct {
		resp PythonResponse
		err  error
	})

	// Ejecutar la extracción en una goroutine
	go func() {
		resp, err := s.bridge.ExtractFeatures(imagePath)
		ch <- struct {
			resp PythonResponse
			err  error
		}{resp, err}
	}()

	// Esperar el resultado o timeout
	select {
	case result := <-ch:
		if result.err != nil {
			s.Logger.Error("Error al extraer características con Python",
				zap.String("imagePath", imagePath),
				zap.Error(result.err))
			return nil, result.err
		}
		return &result.resp, nil
	case <-ctx.Done():
		s.Logger.Error("Timeout al extraer características con Python",
			zap.String("imagePath", imagePath),
			zap.Duration("timeout", s.Timeout))
		return nil, fmt.Errorf("timeout al procesar la imagen después de %v", s.Timeout)
	}
}

// HealthCheck verifica si el servicio Python está disponible
func (s *PythonService) HealthCheck() error {
	s.Logger.Debug("Verificando disponibilidad del servicio Python")

	// Establecer un timeout para la operación
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	// Crear un canal para recibir el resultado
	ch := make(chan error)

	// Ejecutar la verificación en una goroutine
	go func() {
		err := s.bridge.HealthCheck()
		ch <- err
	}()

	// Esperar el resultado o timeout
	select {
	case err := <-ch:
		if err != nil {
			s.Logger.Error("El servicio Python no está disponible", zap.Error(err))
			return fmt.Errorf("el servicio Python no está disponible: %w", err)
		}
		s.Logger.Debug("El servicio Python está disponible")
		return nil
	case <-ctx.Done():
		s.Logger.Error("Timeout al verificar disponibilidad del servicio Python",
			zap.Duration("timeout", s.Timeout))
		return fmt.Errorf("timeout al verificar disponibilidad del servicio Python después de %v", s.Timeout)
	}
}

// Close cierra los recursos utilizados por el servicio
func (s *PythonService) Close() error {
	s.Logger.Debug("Cerrando servicio Python")
	// Aquí podríamos liberar recursos si fuera necesario
	return nil
}
