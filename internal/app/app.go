package app

import (
	"context"
	"net/http"
	"time"

	"github.com/uragamarco/proyecto-balistica/internal/api"
	"github.com/uragamarco/proyecto-balistica/internal/config"
	"github.com/uragamarco/proyecto-balistica/internal/services"
	"github.com/uragamarco/proyecto-balistica/internal/services/ballistic_detector"
	"github.com/uragamarco/proyecto-balistica/internal/services/chroma"
	"github.com/uragamarco/proyecto-balistica/internal/services/classification"
	"github.com/uragamarco/proyecto-balistica/internal/storage"
	"github.com/uragamarco/proyecto-balistica/pkg/integration"
	"go.uber.org/zap"
)

// App representa la aplicación principal
type App struct {
	Config               *config.Config
	Logger               *zap.Logger
	HTTPServer           *http.Server
	PythonSvc            *integration.PythonService
	ImageService         *services.ImageProcessingService
	StorageService       *storage.StorageService
	BallisticDetector    *ballistic_detector.BallisticDetector
	ClassificationService *classification.ClassificationService
}

// NewApp crea una nueva instancia de la aplicación
func NewApp(cfg *config.Config, logger *zap.Logger) (*App, error) {
	// Inicializar servicio Python con logger
	pythonSvc := integration.NewPythonService(logger, cfg.Python.Timeout)

	// Inicializar servicio de procesamiento de imágenes con logger
	imageService := services.NewImageProcessingService(logger, cfg)

	// Inicializar servicio de chroma
	chromaService := chroma.NewService(&chroma.Config{
		ColorThreshold: cfg.Chroma.ColorThreshold,
		SampleSize:     cfg.Chroma.SampleSize,
	})

	// Inicializar servicio de almacenamiento
	storageService, err := storage.NewStorageService("ballistics.db", logger)
	if err != nil {
		return nil, err
	}

	// Inicializar detector balístico
	ballisticDetector := ballistic_detector.NewBallisticDetector(logger)

	// Inicializar servicio de clasificación
	classificationService := classification.NewClassificationService(
		ballisticDetector,
		storageService,
		logger,
	)

	// Crear handlers con inyección de dependencias
	handlers := api.NewHandlers(logger, imageService.Processor, chromaService, storageService, classificationService)

	// Configurar router
	router := api.NewRouter(handlers)

	// Crear servidor HTTP con timeout configurable
	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return &App{
		Config:               cfg,
		Logger:               logger,
		HTTPServer:           httpServer,
		PythonSvc:            pythonSvc,
		ImageService:         imageService,
		StorageService:       storageService,
		BallisticDetector:    ballisticDetector,
		ClassificationService: classificationService,
	}, nil
}

// Run inicia la aplicación
func (a *App) Run() error {
	a.Logger.Info("Iniciando servidor HTTP",
		zap.String("puerto", a.Config.Server.Port),
		zap.String("ambiente", a.Config.App.Environment))

	if err := a.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown detiene la aplicación de manera controlada
func (a *App) Shutdown(ctx context.Context) error {
	a.Logger.Info("Apagando aplicación...")

	// Apagar servidor HTTP primero
	if err := a.HTTPServer.Shutdown(ctx); err != nil {
		a.Logger.Error("Error al apagar servidor HTTP", zap.Error(err))
		return err
	}

	// Cerrar otros servicios
	if err := a.PythonSvc.Close(); err != nil {
		a.Logger.Warn("Error al cerrar servicio Python", zap.Error(err))
	}

	// Cerrar servicio de almacenamiento
	if err := a.StorageService.Close(); err != nil {
		a.Logger.Warn("Error al cerrar servicio de almacenamiento", zap.Error(err))
	}

	// No es necesario cerrar el servicio de imágenes ya que no implementa Close()

	a.Logger.Info("Aplicación detenida correctamente")
	return nil
}
