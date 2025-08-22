package services

import (
	"image"

	"github.com/uragamarco/proyecto-balistica/internal/config"
	"github.com/uragamarco/proyecto-balistica/internal/services/image_processor"
	"github.com/uragamarco/proyecto-balistica/internal/services/python_features"
	"github.com/uragamarco/proyecto-balistica/pkg/integration"
	"go.uber.org/zap"
)

// ImageProcessingService es un servicio que encapsula la funcionalidad de procesamiento de imágenes
type ImageProcessingService struct {
	Processor     *image_processor.ImageProcessor
	PythonService *python_features.Service
	Logger        *zap.Logger
}

// NewImageProcessingService crea una nueva instancia del servicio de procesamiento de imágenes
func NewImageProcessingService(logger *zap.Logger, cfg *config.Config) *ImageProcessingService {
	// Inicializar el servicio Python
	pythonBridge := integration.NewRPCExtractor()
	pythonBridge.PythonPath = "python3"
	pythonBridge.ScriptPath = "py_services/feature_extractor.py"
	pythonBridge.VirtualEnv = "balistics-env"

	pythonService := python_features.NewService(pythonBridge)

	// Configurar el procesador de imágenes
	processorConfig := &image_processor.Config{
		Contrast:               cfg.Imaging.Contrast,
		SharpenSigma:           cfg.Imaging.SharpenSigma,
		EdgeThreshold:          cfg.Imaging.EdgeThreshold,
		GLCMOffsetDistance:     1,
		ForegroundThreshold:    128,
		EdgeDetectionThreshold: 50.0,
		TempDir:                cfg.Imaging.TempDir,
	}

	processor := image_processor.NewImageProcessor(processorConfig, pythonService)

	return &ImageProcessingService{
		Processor:     processor,
		PythonService: pythonService,
		Logger:        logger,
	}
}

// Process procesa una imagen utilizando el procesador interno
func (s *ImageProcessingService) Process(img image.Image) (image.Image, error) {
	return s.Processor.Process(img)
}

// ExtractFeatures extrae características de una imagen utilizando el procesador interno
func (s *ImageProcessingService) ExtractFeatures(img image.Image, path string) (map[string]float64, map[string]interface{}, error) {
	return s.Processor.ExtractFeatures(img, path)
}

// PythonFeaturesStatus verifica el estado de las características de Python
func (s *ImageProcessingService) PythonFeaturesStatus() (bool, string) {
	return s.Processor.PythonFeaturesStatus()
}

// Close cierra los recursos utilizados por el servicio
func (s *ImageProcessingService) Close() error {
	// Aquí se pueden cerrar recursos si es necesario
	return nil
}
