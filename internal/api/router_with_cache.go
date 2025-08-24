package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uragamarco/proyecto-balistica/internal/handlers"
	"github.com/uragamarco/proyecto-balistica/internal/middleware"
	"github.com/uragamarco/proyecto-balistica/internal/services/cache"
	"go.uber.org/zap"
)

// RouterWithCache router mejorado con sistema de cache
type RouterWithCache struct {
	handlers        *Handlers
	cacheService    *cache.CacheService
	cacheMiddleware *middleware.CacheMiddleware
	logger          *zap.Logger
	engine          *gin.Engine
}

// NewRouterWithCache crea un nuevo router con cache integrado
func NewRouterWithCache(handlers *Handlers, logger *zap.Logger, cacheDir string) (*RouterWithCache, error) {
	// Configurar cache service
	cacheConfig := &cache.CacheConfig{
		MemoryTTL:    5 * time.Minute,
		DiskTTL:      30 * time.Minute,
		MaxMemoryMB:  100, // 100MB de cache en memoria
		CacheDir:     cacheDir,
		Enabled:      true,
		Compress:     true,
	}

	cacheService, err := cache.NewCacheService(cacheConfig, logger)
	if err != nil {
		return nil, err
	}

	// Configurar middleware de cache
	cacheMiddlewareConfig := &middleware.CacheMiddlewareConfig{
		DefaultTTL:      5 * time.Minute,
		CacheableStatus: []int{200, 201, 202},
		CacheableRoutes: []string{"/api/process", "/api/compare", "/api/classify", "/api/analyses"},
		IgnoreHeaders:   []string{"Authorization", "Cookie", "Set-Cookie", "X-Request-ID"},
		MaxBodySize:     50 * 1024 * 1024, // 50MB
		Enabled:         true,
	}

	cacheMiddleware := middleware.NewCacheMiddleware(cacheService, logger, cacheMiddlewareConfig)

	// Configurar Gin engine
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	// Middleware global
	engine.Use(gin.Recovery())
	engine.Use(corsMiddleware())
	engine.Use(loggingMiddleware(logger))

	router := &RouterWithCache{
		handlers:        handlers,
		cacheService:    cacheService,
		cacheMiddleware: cacheMiddleware,
		logger:          logger,
		engine:          engine,
	}

	router.setupRoutes()
	return router, nil
}

// setupRoutes configura todas las rutas con cache apropiado
func (r *RouterWithCache) setupRoutes() {
	// Servir archivos estáticos
	r.engine.Static("/static", "./web")
	r.engine.StaticFile("/", "./web/index.html")
	r.engine.StaticFile("/favicon.ico", "./web/favicon.ico")

	// Grupo API con cache middleware
	api := r.engine.Group("/api")
	api.Use(r.cacheMiddleware.Handler())

	// Rutas principales de análisis (con cache)
	api.POST("/process", r.wrapHandler(r.handlers.ProcessImage))
	api.POST("/compare", r.wrapHandler(r.handlers.CompareSamples))

	// Rutas de análisis almacenados (con cache)
	api.GET("/analyses", r.wrapHandler(r.handlers.GetAnalysesHandler))
	api.GET("/analyses/:id", r.wrapHandler(r.handlers.GetAnalysisHandler))
	api.GET("/analyses/search", r.wrapHandler(r.handlers.SearchAnalysesHandler))
	api.DELETE("/analyses/:id", r.wrapHandler(r.handlers.DeleteAnalysisHandler))

	// Rutas de comparaciones (con cache)
	api.GET("/comparisons/:id", r.wrapHandler(r.handlers.GetComparisonHandler))
	api.GET("/comparisons/sample/:sampleId", r.wrapHandler(r.handlers.GetComparisonsBySampleHandler))
	api.GET("/comparisons/similarity", r.wrapHandler(r.handlers.GetSimilarComparisonsHandler))
	api.GET("/comparisons/date-range", r.wrapHandler(r.handlers.GetComparisonsByDateRangeHandler))

	// Rutas de búsqueda avanzada (con cache)
	api.POST("/search/advanced", r.wrapHandler(r.handlers.AdvancedSearchHandler))

	// Rutas de clasificación (con cache)
	classificationHandler := handlers.NewClassificationHandler(r.handlers.classificationService, r.logger)
	api.POST("/classification/classify", r.wrapHandler(classificationHandler.ClassifyBallistic))
	api.GET("/classification/history", r.wrapHandler(classificationHandler.GetClassificationHistory))
	api.GET("/classification/analysis/:analysisId", r.wrapHandler(classificationHandler.GetClassificationByAnalysisID))
	api.GET("/classification/search/weapon/:weaponType", r.wrapHandler(classificationHandler.SearchByWeaponType))
	api.GET("/classification/search/caliber/:caliber", r.wrapHandler(classificationHandler.SearchByCaliber))

	// Rutas de gestión de cache (sin cache)
	cacheGroup := r.engine.Group("/api/cache")
	cacheGroup.GET("/stats", r.cacheMiddleware.CacheStatsHandler())
	cacheGroup.POST("/clear", r.cacheMiddleware.CacheClearHandler())
	cacheGroup.GET("/health", r.cacheHealthHandler())

	// Rutas de salud y métricas (sin cache)
	health := r.engine.Group("/health")
	health.GET("/", r.healthHandler())
	health.GET("/cache", r.cacheHealthHandler())
	health.GET("/detailed", r.detailedHealthHandler())
}

// wrapHandler convierte http.HandlerFunc a gin.HandlerFunc
func (r *RouterWithCache) wrapHandler(handler http.HandlerFunc) gin.HandlerFunc {
	return gin.WrapF(handler)
}

// corsMiddleware middleware CORS para Gin
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Cache, X-Cache-Timestamp")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// loggingMiddleware middleware de logging para Gin
func loggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
		)
		return ""
	})
}

// healthHandler handler de salud general
func (r *RouterWithCache) healthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "ballistics-analysis",
			"version":   "1.0.0",
		})
	}
}

// cacheHealthHandler handler de salud del cache
func (r *RouterWithCache) cacheHealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := r.cacheService.GetStats()
		
		health := "healthy"
		if stats.HitRatio < 0.1 {
			health = "warning" // Baja tasa de aciertos
		}
		if stats.MemoryUsageMB > 90 {
			health = "critical" // Uso de memoria alto
		}

		c.JSON(http.StatusOK, gin.H{
			"status":      health,
			"cache_stats": stats,
			"timestamp":   time.Now().Format(time.RFC3339),
		})
	}
}

// detailedHealthHandler handler de salud detallada
func (r *RouterWithCache) detailedHealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := r.cacheService.GetStats()
		middlewareStats := r.cacheMiddleware.GetStats()

		c.JSON(http.StatusOK, gin.H{
			"status":           "healthy",
			"timestamp":        time.Now().Format(time.RFC3339),
			"service":          "ballistics-analysis",
			"version":          "1.0.0",
			"cache_service":    stats,
			"cache_middleware": middlewareStats,
			"uptime":           time.Since(time.Now()).String(),
		})
	}
}

// GetEngine retorna el engine de Gin para uso externo
func (r *RouterWithCache) GetEngine() *gin.Engine {
	return r.engine
}

// GetCacheService retorna el servicio de cache
func (r *RouterWithCache) GetCacheService() *cache.CacheService {
	return r.cacheService
}

// GetCacheMiddleware retorna el middleware de cache
func (r *RouterWithCache) GetCacheMiddleware() *middleware.CacheMiddleware {
	return r.cacheMiddleware
}

// CacheFeatures cachea características de imagen específicas
func (r *RouterWithCache) CacheFeatures(imagePath string, imageSize int64, features map[string]float64, advanced map[string]interface{}) error {
	cacheKey := r.cacheMiddleware.FeatureCacheKey(imagePath, imageSize, map[string]interface{}{
		"processor": "optimized",
		"version":   "1.0",
	})

	data := map[string]interface{}{
		"features": features,
		"advanced": advanced,
		"metadata": map[string]interface{}{
			"image_path": imagePath,
			"image_size": imageSize,
			"timestamp":  time.Now(),
		},
	}

	return r.cacheService.Set(cacheKey, data)
}

// GetCachedFeatures obtiene características cacheadas
func (r *RouterWithCache) GetCachedFeatures(imagePath string, imageSize int64) (map[string]float64, map[string]interface{}, bool) {
	cacheKey := r.cacheMiddleware.FeatureCacheKey(imagePath, imageSize, map[string]interface{}{
		"processor": "optimized",
		"version":   "1.0",
	})

	data, found := r.cacheService.Get(cacheKey)
	if !found {
		return nil, nil, false
	}

	cachedData, ok := data.(map[string]interface{})
	if !ok {
		return nil, nil, false
	}

	features, _ := cachedData["features"].(map[string]float64)
	advanced, _ := cachedData["advanced"].(map[string]interface{})

	return features, advanced, true
}

// CacheComparison cachea resultado de comparación
func (r *RouterWithCache) CacheComparison(image1Path, image2Path string, algorithm string, result interface{}) error {
	cacheKey := r.cacheMiddleware.ComparisonCacheKey(image1Path, image2Path, algorithm)
	return r.cacheService.Set(cacheKey, result)
}

// GetCachedComparison obtiene resultado de comparación cacheado
func (r *RouterWithCache) GetCachedComparison(image1Path, image2Path string, algorithm string) (interface{}, bool) {
	cacheKey := r.cacheMiddleware.ComparisonCacheKey(image1Path, image2Path, algorithm)
	return r.cacheService.Get(cacheKey)
}

// CacheClassification cachea resultado de clasificación
func (r *RouterWithCache) CacheClassification(imagePath string, model string, threshold float64, result interface{}) error {
	cacheKey := r.cacheMiddleware.ClassificationCacheKey(imagePath, model, threshold)
	return r.cacheService.Set(cacheKey, result)
}

// GetCachedClassification obtiene resultado de clasificación cacheado
func (r *RouterWithCache) GetCachedClassification(imagePath string, model string, threshold float64) (interface{}, bool) {
	cacheKey := r.cacheMiddleware.ClassificationCacheKey(imagePath, model, threshold)
	return r.cacheService.Get(cacheKey)
}

// Shutdown cierra el router y limpia recursos
func (r *RouterWithCache) Shutdown() error {
	r.logger.Info("Shutting down router with cache")
	return r.cacheService.Clear()
}