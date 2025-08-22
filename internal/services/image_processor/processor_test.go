package image_processor

import (
	"testing"
	"go.uber.org/zap"
)

func TestPythonFeaturesStatus(t *testing.T) {
	// Crear logger para las pruebas
	logger := zap.NewNop()
	
	// Caso con Python desactivado
	proc := NewImageProcessor(&Config{Logger: logger}, nil)
	enabled, status := proc.PythonFeaturesStatus()
	if enabled {
		t.Error("Expected Python to be disabled")
	}
	if status == "" {
		t.Error("Expected status message when Python is disabled")
	}

	// Caso con Python activado (nil service simula servicio no disponible)
	// En un caso real, aquí se pasaría un servicio Python válido
	procWithPy := NewImageProcessor(&Config{Logger: logger}, nil)
	enabled2, status2 := procWithPy.PythonFeaturesStatus()
	if enabled2 {
		t.Error("Expected Python to be disabled when service is nil")
	}
	if status2 == "" {
		t.Error("Expected status message")
	}
}
