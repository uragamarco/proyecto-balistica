package app

import (
	"context"
	"net/http"
	"path/filepath"
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

// AppWithCache representa la aplicación principal con sistema de cache integrado
type AppWithCache struct {
	Config                *config.Config
	Logger                *zap.Logger
	HTTPServer            *http.Server
	PythonSvc             *integration.PythonService
	ImageService          *services.ImageProcessingService
	StorageService        *storage.StorageService
	BallisticDetector     *ballistic_detector.BallisticDetector
	ClassificationService *classification.ClassificationService
	RouterWithCache       *api.RouterWithCache
}

// NewAppWithCache crea una nueva instancia de la aplicación con cache
func NewAppWithCache(cfg *config.Config, logger *zap.Logger) (*AppWithCache, error) {
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

	// Configurar directorio de cache
	cacheDir := filepath.Join(".", "cache")
	if cfg.Cache != nil && cfg.Cache.Directory != "" {
		cacheDir = cfg.Cache.Directory
	}

	// Configurar router con cache
	routerWithCache, err := api.NewRouterWithCache(handlers, logger, cacheDir)
	if err != nil {
		return nil, err
	}

	// Crear servidor HTTP con timeout configurable
	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      routerWithCache.GetEngine(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &AppWithCache{
		Config:                cfg,
		Logger:                logger,
		HTTPServer:            httpServer,
		PythonSvc:             pythonSvc,
		ImageService:          imageService,
		StorageService:        storageService,
		BallisticDetector:     ballisticDetector,
		ClassificationService: classificationService,
		RouterWithCache:       routerWithCache,
	}, nil
}

// Run inicia la aplicación
func (a *AppWithCache) Run() error {
	a.Logger.Info("Iniciando servidor HTTP con cache",
		zap.String("puerto", a.Config.Server.Port),
		zap.String("ambiente", a.Config.App.Environment),
		zap.String("cache_enabled", "true"))

	// Mostrar estadísticas de cache al inicio
	go a.logCacheStats()

	if err := a.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// logCacheStats registra estadísticas de cache periódicamente
func (a *AppWithCache) logCacheStats() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := a.RouterWithCache.GetCacheService().GetStats()
			a.Logger.Info("Estadísticas de cache",
				zap.Int64("memory_hits", stats.MemoryHits),
				zap.Int64("disk_hits", stats.DiskHits),
				zap.Int64("misses", stats.Misses),
				zap.Float64("hit_ratio", stats.HitRatio),
				zap.Int("entry_count", stats.EntryCount),
				zap.Float64("memory_usage_mb", stats.MemoryUsageMB),
				zap.Float64("disk_usage_mb", stats.DiskUsageMB))
		}
	}
}

// Shutdown detiene la aplicación de manera controlada
func (a *AppWithCache) Shutdown(ctx context.Context) error {
	a.Logger.Info("Apagando aplicación con cache...")

	// Mostrar estadísticas finales de cache
	stats := a.RouterWithCache.GetCacheService().GetStats()
	a.Logger.Info("Estadísticas finales de cache",
		zap.Int64("memory_hits", stats.MemoryHits),
		zap.Int64("disk_hits", stats.DiskHits),
		zap.Int64("total_misses", stats.Misses),
		zap.Float64("final_hit_ratio", stats.HitRatio),
		zap.Int("entry_count", stats.EntryCount))

	// Apagar servidor HTTP primero
	if err := a.HTTPServer.Shutdown(ctx); err != nil {
		a.Logger.Error("Error al apagar servidor HTTP", zap.Error(err))
		return err
	}

	// Cerrar router con cache
	if err := a.RouterWithCache.Shutdown(); err != nil {
		a.Logger.Warn("Error al cerrar router con cache", zap.Error(err))
	}

	// Cerrar otros servicios
	if err := a.PythonSvc.Close(); err != nil {
		a.Logger.Warn("Error al cerrar servicio Python", zap.Error(err))
	}

	// Cerrar servicio de almacenamiento
	if err := a.StorageService.Close(); err != nil {
		a.Logger.Warn("Error al cerrar servicio de almacenamiento", zap.Error(err))
	}

	a.Logger.Info("Aplicación con cache detenida correctamente")
	return nil
}

// GetCacheService retorna el servicio de cache para uso externo
func (a *AppWithCache) GetCacheService() interface{} {
	return a.RouterWithCache.GetCacheService()
}

// GetCacheStats retorna estadísticas actuales del cache
func (a *AppWithCache) GetCacheStats() interface{} {
	return a.RouterWithCache.GetCacheService().GetStats()
}

// ClearCache limpia todo el cache
func (a *AppWithCache) ClearCache() error {
	a.Logger.Info("Limpiando cache manualmente")
	return a.RouterWithCache.GetCacheService().Clear()
}

// CacheFeatures cachea características de imagen
func (a *AppWithCache) CacheFeatures(imagePath string, imageSize int64, features map[string]float64, advanced map[string]interface{}) error {
	return a.RouterWithCache.CacheFeatures(imagePath, imageSize, features, advanced)
}

// GetCachedFeatures obtiene características cacheadas
func (a *AppWithCache) GetCachedFeatures(imagePath string, imageSize int64) (map[string]float64, map[string]interface{}, bool) {
	return a.RouterWithCache.GetCachedFeatures(imagePath, imageSize)
}

// CacheComparison cachea resultado de comparación
func (a *AppWithCache) CacheComparison(image1Path, image2Path string, algorithm string, result interface{}) error {
	return a.RouterWithCache.CacheComparison(image1Path, image2Path, algorithm, result)
}

// GetCachedComparison obtiene resultado de comparación cacheado
func (a *AppWithCache) GetCachedComparison(image1Path, image2Path string, algorithm string) (interface{}, bool) {
	return a.RouterWithCache.GetCachedComparison(image1Path, image2Path, algorithm)
}

// CacheClassification cachea resultado de clasificación
func (a *AppWithCache) CacheClassification(imagePath string, model string, threshold float64, result interface{}) error {
	return a.RouterWithCache.CacheClassification(imagePath, model, threshold, result)
}

// GetCachedClassification obtiene resultado de clasificación cacheado
func (a *AppWithCache) GetCachedClassification(imagePath string, model string, threshold float64) (interface{}, bool) {
	return a.RouterWithCache.GetCachedClassification(imagePath, model, threshold)
}