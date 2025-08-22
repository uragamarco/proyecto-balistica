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
	HuMoments       []float64 `json:"hu_moments"`
	ContourArea     float64   `json:"contour_area"`
	ContourLen      float64   `json:"contour_len"`
	LBPUniformity   float64   `json:"lbp_uniformity"`
	FiringPinMarks  []struct {
		X      float64 `json:"x"`
		Y      float64 `json:"y"`
		Radius float64 `json:"radius"`
	} `json:"firing_pin_marks"`
	StriationPatterns []struct {
		Angle    float64 `json:"angle"`
		Length   float64 `json:"length"`
		Strength float64 `json:"strength"`
	} `json:"striation_patterns"`
	ErrorMessage string `json:"error,omitempty"`
	// Metadatos adicionales
	Filename     string `json:"filename,omitempty"`
	ContentType  string `json:"content_type,omitempty"`
	FileSize     int64  `json:"file_size,omitempty"`
}

// FeatureExtractor interfaz para abstraer la extracción de features
type FeatureExtractor interface {
	ExtractFeatures(imagePath string) (PythonResponse, error)
	HealthCheck() error
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

// HealthCheck verifica si el extractor Python está disponible
func (r *RPCExtractor) HealthCheck() error {
	// Verificar si el script existe
	if _, err := os.Stat(r.ScriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script Python no encontrado: %s", r.ScriptPath)
	}

	// Verificar si el entorno virtual existe
	activatePath := ""
	if runtime.GOOS == "windows" {
		activatePath = filepath.Join(r.VirtualEnv, "Scripts", "activate")
	} else {
		activatePath = filepath.Join(r.VirtualEnv, "bin", "activate")
	}

	if _, err := os.Stat(activatePath); os.IsNotExist(err) {
		return fmt.Errorf("entorno virtual no encontrado: %s", activatePath)
	}

	return nil
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
		// En sistemas Unix, usamos bash para ejecutar comandos con source
		cmd = exec.Command(
			"bash",
			"-c",
			"source "+filepath.Join(r.VirtualEnv, "bin", "activate")+" && "+
				r.PythonPath+" "+r.ScriptPath+" "+imagePath,
		)
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
	var errorResponse struct {
		Error string `json:"error"`
	}

	// Primero intentamos parsear como respuesta normal
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		// Si falla, intentamos parsear como respuesta de error
		if errJson := json.Unmarshal(stdout.Bytes(), &errorResponse); errJson == nil && errorResponse.Error != "" {
			return PythonResponse{}, errors.New(errorResponse.Error)
		}

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
