package middleware

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uragamarco/proyecto-balistica/internal/services/cache"
	"go.uber.org/zap"
)

// CacheMiddleware middleware para cache de respuestas HTTP
type CacheMiddleware struct {
	cacheService *cache.CacheService
	logger       *zap.Logger
	config       *CacheMiddlewareConfig
}

// CacheMiddlewareConfig configuración del middleware de cache
type CacheMiddlewareConfig struct {
	DefaultTTL     time.Duration // TTL por defecto para respuestas
	CacheableStatus []int        // Códigos de estado que se pueden cachear
	CacheableRoutes []string     // Rutas que se pueden cachear
	IgnoreHeaders   []string     // Headers a ignorar en la clave de cache
	MaxBodySize     int64        // Tamaño máximo del body para cachear
	Enabled         bool         // Habilitar/deshabilitar middleware
}

// CachedResponse respuesta cacheada
type CachedResponse struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       []byte              `json:"body"`
	Timestamp  time.Time           `json:"timestamp"`
	TTL        time.Duration       `json:"ttl"`
}

// responseWriter wrapper para capturar la respuesta
type responseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

// NewCacheMiddleware crea una nueva instancia del middleware de cache
func NewCacheMiddleware(cacheService *cache.CacheService, logger *zap.Logger, config *CacheMiddlewareConfig) *CacheMiddleware {
	if config == nil {
		config = &CacheMiddlewareConfig{
			DefaultTTL:      5 * time.Minute,
			CacheableStatus: []int{200, 201, 202},
			CacheableRoutes: []string{"/api/analyze", "/api/compare", "/api/classify"},
			IgnoreHeaders:   []string{"Authorization", "Cookie", "Set-Cookie"},
			MaxBodySize:     10 * 1024 * 1024, // 10MB
			Enabled:         true,
		}
	}

	return &CacheMiddleware{
		cacheService: cacheService,
		logger:       logger,
		config:       config,
	}
}

// Handler middleware handler para Gin
func (cm *CacheMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cm.config.Enabled || !cm.shouldCache(c) {
			c.Next()
			return
		}

		// Generar clave de cache
		cacheKey := cm.generateCacheKey(c)

		// Intentar obtener respuesta del cache
		if cachedResp, found := cm.getCachedResponse(cacheKey); found {
			cm.serveCachedResponse(c, cachedResp)
			return
		}

		// Crear wrapper para capturar la respuesta
		wrapper := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer([]byte{}),
			statusCode:     200,
		}
		c.Writer = wrapper

		// Procesar request
		c.Next()

		// Cachear respuesta si es apropiado
		if cm.shouldCacheResponse(wrapper.statusCode, wrapper.body.Len()) {
			cm.cacheResponse(cacheKey, wrapper)
		}
	}
}

// shouldCache determina si la request debe ser cacheada
func (cm *CacheMiddleware) shouldCache(c *gin.Context) bool {
	// Solo cachear GET requests
	if c.Request.Method != "GET" && c.Request.Method != "POST" {
		return false
	}

	// Verificar si la ruta es cacheable
	path := c.Request.URL.Path
	for _, route := range cm.config.CacheableRoutes {
		if strings.HasPrefix(path, route) {
			return true
		}
	}

	return false
}

// shouldCacheResponse determina si la respuesta debe ser cacheada
func (cm *CacheMiddleware) shouldCacheResponse(statusCode int, bodySize int) bool {
	// Verificar código de estado
	for _, code := range cm.config.CacheableStatus {
		if statusCode == code {
			break
		}
	}

	// Verificar tamaño del body
	if int64(bodySize) > cm.config.MaxBodySize {
		return false
	}

	return true
}

// generateCacheKey genera una clave única para la request
func (cm *CacheMiddleware) generateCacheKey(c *gin.Context) string {
	hash := md5.New()

	// Incluir método y path
	hash.Write([]byte(c.Request.Method))
	hash.Write([]byte(c.Request.URL.Path))

	// Incluir query parameters
	if c.Request.URL.RawQuery != "" {
		hash.Write([]byte(c.Request.URL.RawQuery))
	}

	// Incluir headers relevantes
	for name, values := range c.Request.Header {
		if !cm.shouldIgnoreHeader(name) {
			for _, value := range values {
				hash.Write([]byte(name + ":" + value))
			}
		}
	}

	// Para POST requests, incluir body si es pequeño
	if c.Request.Method == "POST" && c.Request.ContentLength < 1024 {
		if body, err := io.ReadAll(c.Request.Body); err == nil {
			hash.Write(body)
			// Restaurar body para el handler
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}
	}

	return fmt.Sprintf("http_cache_%x", hash.Sum(nil))
}

// shouldIgnoreHeader determina si un header debe ser ignorado
func (cm *CacheMiddleware) shouldIgnoreHeader(headerName string) bool {
	for _, ignored := range cm.config.IgnoreHeaders {
		if strings.EqualFold(headerName, ignored) {
			return true
		}
	}
	return false
}

// getCachedResponse obtiene una respuesta del cache
func (cm *CacheMiddleware) getCachedResponse(cacheKey string) (*CachedResponse, bool) {
	data, found := cm.cacheService.Get(cacheKey)
	if !found {
		return nil, false
	}

	cachedResp, ok := data.(*CachedResponse)
	if !ok {
		cm.logger.Warn("Invalid cached response type", zap.String("key", cacheKey))
		return nil, false
	}

	// Verificar TTL
	if time.Since(cachedResp.Timestamp) > cachedResp.TTL {
		cm.cacheService.Delete(cacheKey)
		return nil, false
	}

	cm.logger.Debug("Cache hit", zap.String("key", cacheKey))
	return cachedResp, true
}

// serveCachedResponse sirve una respuesta desde el cache
func (cm *CacheMiddleware) serveCachedResponse(c *gin.Context, cachedResp *CachedResponse) {
	// Establecer headers
	for name, values := range cachedResp.Headers {
		for _, value := range values {
			c.Header(name, value)
		}
	}

	// Agregar header de cache
	c.Header("X-Cache", "HIT")
	c.Header("X-Cache-Timestamp", cachedResp.Timestamp.Format(time.RFC3339))

	// Escribir respuesta
	c.Data(cachedResp.StatusCode, c.GetHeader("Content-Type"), cachedResp.Body)
	c.Abort()
}

// cacheResponse almacena una respuesta en el cache
func (cm *CacheMiddleware) cacheResponse(cacheKey string, wrapper *responseWriter) {
	// Capturar headers (excluyendo los que no deben cachearse)
	headers := make(map[string][]string)
	for name, values := range wrapper.Header() {
		if !cm.shouldIgnoreHeader(name) {
			headers[name] = values
		}
	}

	cachedResp := &CachedResponse{
		StatusCode: wrapper.statusCode,
		Headers:    headers,
		Body:       wrapper.body.Bytes(),
		Timestamp:  time.Now(),
		TTL:        cm.config.DefaultTTL,
	}

	if err := cm.cacheService.Set(cacheKey, cachedResp); err != nil {
		cm.logger.Error("Failed to cache response", 
			zap.String("key", cacheKey), 
			zap.Error(err))
	} else {
		cm.logger.Debug("Response cached", 
			zap.String("key", cacheKey),
			zap.Int("size", len(cachedResp.Body)))
	}
}

// GetStats obtiene estadísticas del cache
func (cm *CacheMiddleware) GetStats() map[string]interface{} {
	stats := cm.cacheService.GetStats()
	return map[string]interface{}{
		"cache_stats": stats,
		"middleware_config": map[string]interface{}{
			"enabled":          cm.config.Enabled,
			"default_ttl":      cm.config.DefaultTTL.String(),
			"cacheable_routes": cm.config.CacheableRoutes,
			"max_body_size":    cm.config.MaxBodySize,
		},
	}
}

// ClearCache limpia todo el cache
func (cm *CacheMiddleware) ClearCache() error {
	return cm.cacheService.Clear()
}

// Implementación de responseWriter
func (rw *responseWriter) Write(data []byte) (int, error) {
	rw.body.Write(data)
	return rw.ResponseWriter.Write(data)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) WriteString(s string) (int, error) {
	rw.body.WriteString(s)
	return rw.ResponseWriter.WriteString(s)
}

// CacheStatsHandler handler para obtener estadísticas del cache
func (cm *CacheMiddleware) CacheStatsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := cm.GetStats()
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   stats,
		})
	}
}

// CacheClearHandler handler para limpiar el cache
func (cm *CacheMiddleware) CacheClearHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := cm.ClearCache(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Cache cleared successfully",
		})
	}
}

// FeatureCacheKey genera una clave específica para características de imagen
func (cm *CacheMiddleware) FeatureCacheKey(imagePath string, imageSize int64, processingParams map[string]interface{}) string {
	return cm.cacheService.GenerateKey("features", imagePath, imageSize, processingParams)
}

// ComparisonCacheKey genera una clave específica para comparaciones
func (cm *CacheMiddleware) ComparisonCacheKey(image1Path, image2Path string, algorithm string) string {
	return cm.cacheService.GenerateKey("comparison", image1Path, image2Path, algorithm)
}

// ClassificationCacheKey genera una clave específica para clasificaciones
func (cm *CacheMiddleware) ClassificationCacheKey(imagePath string, model string, threshold float64) string {
	return cm.cacheService.GenerateKey("classification", imagePath, model, threshold)
}