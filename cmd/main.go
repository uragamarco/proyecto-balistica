package main

import (
	"log"
	"net/http"
	"os"

	"github.com/uragamarco/proyecto-balistica/internal/api"
	"github.com/uragamarco/proyecto-balistica/internal/app"
	"github.com/uragamarco/proyecto-balistica/internal/config"
	"github.com/uragamarco/proyecto-balistica/internal/services/chroma"
	"github.com/uragamarco/proyecto-balistica/internal/services/image_processor"
	"github.com/uragamarco/proyecto-balistica/internal/services/python_features" // Nuevo paquete
	"github.com/uragamarco/proyecto-balistica/pkg/integration"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Cargar configuración
	cfg, err := config.LoadConfig("configs/default.yml")
	if err != nil {
		panic("Error loading config: " + err.Error())
	}

	// Inicializar logger
	logger := initLogger(cfg)
	defer func() {
		_ = logger.Sync() // Asegurar que todos los logs se escriban
	}()

	// Crear aplicación
	application := app.NewApp(cfg, logger)

	// Ejecutar aplicación
	if err := application.Run(); err != nil {
		logger.Error("Application failed", zap.Error(err))
		os.Exit(1)
	}

	// Initialize Python integration
	pyBridge := integration.NewRPCExtractor()
	if err := pyBridge.HealthCheck(); err != nil {
		log.Printf("WARNING: Python features disabled - %v", err)
		pyBridge = nil // Permite operación sin Python
	}

	// Initialize services
	imgProcCfg := &image_processor.Config{
		Contrast:      cfg.Imaging.Contrast,
		SharpenSigma:  cfg.Imaging.SharpenSigma,
		EdgeThreshold: cfg.Imaging.EdgeThreshold,
		PythonBridge:  pyBridge, // Inyectamos el bridge
		TempDir:       cfg.Imaging.TempDir,
		Logger:        log.New(os.Stdout, "IMG_PROC: ", log.LstdFlags),
	}

	imgProcessor := image_processor.NewImageProcessor(imgProcCfg)

	chromaCfg := &chroma.Config{
		ColorThreshold: cfg.Chroma.ColorThreshold,
		SampleSize:     cfg.Chroma.SampleSize,
	}

	chromaSvc := chroma.NewService(chromaCfg)

	// Initialize feature extraction service
	featureSvc := python_features.NewService(pyBridge) // Opcional si Python está disponible

	// Initialize API handlers
	handlers := api.NewHandlers(imgProcessor, chromaSvc, featureSvc)

	// Create and start server
	server := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: api.NewRouter(handlers),
	}

	log.Printf("Starting server on %s", cfg.Server.Address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

func initLogger(cfg *config.Config) *zap.Logger {
	var logger *zap.Logger
	var err error

	if cfg.Environment == "production" {
		// Configuración de producción: JSON format, más rápido
		config := zap.NewProductionConfig()
		config.OutputPaths = []string{"stdout", cfg.Logging.File}
		logger, err = config.Build()
	} else {
		// Configuración de desarrollo: más legible
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, err = config.Build()
	}

	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// Reemplazar logger global
	zap.ReplaceGlobals(logger)
	return logger
}
