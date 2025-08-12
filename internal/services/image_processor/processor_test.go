package image_processor

import "testing"

func TestPythonEnabled(t *testing.T) {
	// Caso con Python desactivado
	proc := NewImageProcessor(&Config{}, nil)
	if proc.PythonEnabled() {
		t.Error("Expected Python to be disabled")
	}

	// Caso con Python activado (mock)
	mockPy := new(MockPythonService) // Implementa FeatureExtractor
	procWithPy := NewImageProcessor(&Config{}, mockPy)
	if !procWithPy.PythonEnabled() {
		t.Error("Expected Python to be enabled")
	}
}
