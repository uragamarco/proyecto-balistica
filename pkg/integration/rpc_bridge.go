package integration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// PythonResponse estructura para parsear la respuesta de Python
type PythonResponse struct {
	HuMoments    []float64 `json:"hu_moments"`
	Striations   []float64 `json:"striations"`
	ContourArea  float64   `json:"contour_area"`
	ContourLen   float64   `json:"contour_len"`
	ErrorMessage string    `json:"error,omitempty"`
}

// FeatureExtractor interfaz para abstraer la extracción de features
type FeatureExtractor interface {
	ExtractFeatures(imagePath string) (PythonResponse, error)
}

// RPCExtractor implementación concreta usando RPC
type RPCExtractor struct {
	PythonPath string
	ScriptPath string
	VirtualEnv string
}

// NewRPCExtractor constructor con valores por defecto
func NewRPCExtractor() *RPCExtractor {
	return &RPCExtractor{
		PythonPath: "python3",
		ScriptPath: "py_services/feature_extractor.py",
		VirtualEnv: "balistics-env",
	}
}

// ExtractFeatures llama al script Python y parsea la respuesta
func (r *RPCExtractor) ExtractFeatures(imagePath string) (PythonResponse, error) {
	// Validaciones de entrada
	if !filepath.IsAbs(imagePath) {
		return PythonResponse{}, fmt.Errorf("image path must be absolute: %s", imagePath)
	}

	// Construir el comando Python
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command(
			filepath.Join(r.VirtualEnv, "Scripts", "python"),
			r.ScriptPath,
			imagePath,
		)
	} else {
		cmd = exec.Command(
			"source",
			filepath.Join(r.VirtualEnv, "bin", "activate"),
			"&&",
			r.PythonPath,
			r.ScriptPath,
			imagePath,
		)
		cmd.Shell = true
	}

	// Capturar salida y errores
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Ejecutar
	err := cmd.Run()
	if err != nil {
		return PythonResponse{}, fmt.Errorf(
			"python execution failed: %v\nStderr: %s",
			err,
			stderr.String(),
		)
	}

	// Parsear respuesta JSON
	var response PythonResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return PythonResponse{}, fmt.Errorf(
			"failed to parse python response: %v\nOutput: %s",
			err,
			stdout.String(),
		)
	}

	// Verificar si Python reportó error interno
	if response.ErrorMessage != "" {
		return PythonResponse{}, errors.New(response.ErrorMessage)
	}

	// Validar datos esenciales
	if len(response.HuMoments) != 7 {
		return PythonResponse{}, fmt.Errorf(
			"invalid hu moments count: expected 7, got %d",
			len(response.HuMoments),
		)
	}

	return response, nil
}

// HealthCheck verifica que el sistema Python esté disponible
func (r *RPCExtractor) HealthCheck() error {
	testCmd := exec.Command(r.PythonPath, "--version")
	if err := testCmd.Run(); err != nil {
		return fmt.Errorf("python not available: %v", err)
	}

	// Verificar que el script existe
	if _, err := os.Stat(r.ScriptPath); os.IsNotExist(err) {
		return fmt.Errorf("python script not found at %s", r.ScriptPath)
	}

	return nil
}
