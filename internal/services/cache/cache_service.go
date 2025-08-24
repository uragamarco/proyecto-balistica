package cache

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CacheService servicio de cache persistente para características balísticas
type CacheService struct {
	memoryCache  *MemoryCache
	diskCache    *DiskCache
	logger       *zap.Logger
	config       *CacheConfig
}

// CacheConfig configuración del sistema de cache
type CacheConfig struct {
	MemoryTTL    time.Duration // TTL para cache en memoria
	DiskTTL      time.Duration // TTL para cache en disco
	MaxMemoryMB  int           // Límite de memoria en MB
	CacheDir     string        // Directorio para cache en disco
	Enabled      bool          // Habilitar/deshabilitar cache
	Compress     bool          // Comprimir datos en disco
}

// MemoryCache cache en memoria con LRU
type MemoryCache struct {
	mu       sync.RWMutex
	data     map[string]*CacheEntry
	lruList  []string
	maxSize  int
	currSize int
}

// DiskCache cache persistente en disco
type DiskCache struct {
	mu       sync.RWMutex
	cacheDir string
	compress bool
	logger   *zap.Logger
}

// CacheEntry entrada del cache con metadatos
type CacheEntry struct {
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       time.Duration `json:"ttl"`
	Size      int         `json:"size"`
	Hits      int         `json:"hits"`
	Key       string      `json:"key"`
}

// CacheStats estadísticas del cache
type CacheStats struct {
	MemoryHits   int64   `json:"memory_hits"`
	DiskHits     int64   `json:"disk_hits"`
	Misses       int64   `json:"misses"`
	HitRatio     float64 `json:"hit_ratio"`
	MemoryUsageMB float64 `json:"memory_usage_mb"`
	DiskUsageMB   float64 `json:"disk_usage_mb"`
	EntryCount    int     `json:"entry_count"`
}

// NewCacheService crea una nueva instancia del servicio de cache
func NewCacheService(config *CacheConfig, logger *zap.Logger) (*CacheService, error) {
	if !config.Enabled {
		return &CacheService{
			config: config,
			logger: logger,
		}, nil
	}

	// Crear directorio de cache si no existe
	if err := os.MkdirAll(config.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	memoryCache := &MemoryCache{
		data:    make(map[string]*CacheEntry),
		lruList: make([]string, 0),
		maxSize: config.MaxMemoryMB * 1024 * 1024, // Convertir MB a bytes
	}

	diskCache := &DiskCache{
		cacheDir: config.CacheDir,
		compress: config.Compress,
		logger:   logger,
	}

	service := &CacheService{
		memoryCache: memoryCache,
		diskCache:   diskCache,
		logger:      logger,
		config:      config,
	}

	// Limpiar cache expirado al inicializar
	go service.cleanupExpiredEntries()

	return service, nil
}

// Get obtiene un valor del cache (memoria -> disco -> miss)
func (cs *CacheService) Get(key string) (interface{}, bool) {
	if !cs.config.Enabled {
		return nil, false
	}

	// 1. Buscar en memoria primero
	if data, found := cs.memoryCache.get(key); found {
		cs.logger.Debug("Cache hit in memory", zap.String("key", key))
		return data, true
	}

	// 2. Buscar en disco
	if data, found := cs.diskCache.get(key, cs.config.DiskTTL); found {
		cs.logger.Debug("Cache hit in disk", zap.String("key", key))
		
		// Promover a memoria para acceso rápido futuro
		cs.memoryCache.set(key, data, cs.config.MemoryTTL)
		return data.Data, true
	}

	cs.logger.Debug("Cache miss", zap.String("key", key))
	return nil, false
}

// Set almacena un valor en el cache (memoria + disco)
func (cs *CacheService) Set(key string, data interface{}) error {
	if !cs.config.Enabled {
		return nil
	}

	// Almacenar en memoria
	cs.memoryCache.set(key, data, cs.config.MemoryTTL)

	// Almacenar en disco de forma asíncrona
	go func() {
		if err := cs.diskCache.set(key, data, cs.config.DiskTTL); err != nil {
			cs.logger.Error("Failed to store in disk cache", 
				zap.String("key", key), 
				zap.Error(err))
		}
	}()

	cs.logger.Debug("Stored in cache", zap.String("key", key))
	return nil
}

// Delete elimina una entrada del cache
func (cs *CacheService) Delete(key string) error {
	if !cs.config.Enabled {
		return nil
	}

	cs.memoryCache.delete(key)
	cs.diskCache.delete(key)

	cs.logger.Debug("Deleted from cache", zap.String("key", key))
	return nil
}

// GetStats obtiene estadísticas del cache
func (cs *CacheService) GetStats() CacheStats {
	if !cs.config.Enabled {
		return CacheStats{}
	}

	memoryStats := cs.memoryCache.getStats()
	diskStats := cs.diskCache.getStats()

	totalHits := memoryStats.MemoryHits + diskStats.DiskHits
	totalRequests := totalHits + memoryStats.Misses

	hitRatio := 0.0
	if totalRequests > 0 {
		hitRatio = float64(totalHits) / float64(totalRequests)
	}

	return CacheStats{
		MemoryHits:    memoryStats.MemoryHits,
		DiskHits:      diskStats.DiskHits,
		Misses:        memoryStats.Misses,
		HitRatio:      hitRatio,
		MemoryUsageMB: float64(memoryStats.MemoryUsageMB),
		DiskUsageMB:   diskStats.DiskUsageMB,
		EntryCount:    memoryStats.EntryCount + diskStats.EntryCount,
	}
}

// Clear limpia todo el cache
func (cs *CacheService) Clear() error {
	if !cs.config.Enabled {
		return nil
	}

	cs.memoryCache.clear()
	return cs.diskCache.clear()
}

// GenerateKey genera una clave de cache basada en parámetros
func (cs *CacheService) GenerateKey(prefix string, params ...interface{}) string {
	hash := md5.New()
	for _, param := range params {
		switch v := param.(type) {
		case string:
			hash.Write([]byte(v))
		case []byte:
			hash.Write(v)
		default:
			data, _ := json.Marshal(v)
			hash.Write(data)
		}
	}
	return fmt.Sprintf("%s_%x", prefix, hash.Sum(nil))
}

// Métodos de MemoryCache
func (mc *MemoryCache) get(key string) (interface{}, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	entry, exists := mc.data[key]
	if !exists {
		return nil, false
	}

	// Verificar TTL
	if time.Since(entry.Timestamp) > entry.TTL {
		delete(mc.data, key)
		mc.removeLRU(key)
		return nil, false
	}

	entry.Hits++
	mc.updateLRU(key)
	return entry.Data, true
}

func (mc *MemoryCache) set(key string, data interface{}, ttl time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Calcular tamaño aproximado
	dataBytes, _ := json.Marshal(data)
	size := len(dataBytes)

	// Verificar si necesitamos espacio
	for mc.currSize+size > mc.maxSize && len(mc.lruList) > 0 {
		mc.evictLRU()
	}

	entry := &CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		TTL:       ttl,
		Size:      size,
		Hits:      0,
		Key:       key,
	}

	mc.data[key] = entry
	mc.currSize += size
	mc.updateLRU(key)
}

func (mc *MemoryCache) delete(key string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if entry, exists := mc.data[key]; exists {
		mc.currSize -= entry.Size
		delete(mc.data, key)
		mc.removeLRU(key)
	}
}

func (mc *MemoryCache) clear() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.data = make(map[string]*CacheEntry)
	mc.lruList = make([]string, 0)
	mc.currSize = 0
}

func (mc *MemoryCache) updateLRU(key string) {
	// Mover al frente de la lista LRU
	mc.removeLRU(key)
	mc.lruList = append([]string{key}, mc.lruList...)
}

func (mc *MemoryCache) removeLRU(key string) {
	for i, k := range mc.lruList {
		if k == key {
			mc.lruList = append(mc.lruList[:i], mc.lruList[i+1:]...)
			break
		}
	}
}

func (mc *MemoryCache) evictLRU() {
	if len(mc.lruList) == 0 {
		return
	}

	key := mc.lruList[len(mc.lruList)-1]
	if entry, exists := mc.data[key]; exists {
		mc.currSize -= entry.Size
		delete(mc.data, key)
	}
	mc.lruList = mc.lruList[:len(mc.lruList)-1]
}

func (mc *MemoryCache) getStats() CacheStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var totalHits int64
	for _, entry := range mc.data {
		totalHits += int64(entry.Hits)
	}

	return CacheStats{
		MemoryHits:    totalHits,
		MemoryUsageMB: float64(mc.currSize) / (1024 * 1024),
		EntryCount:    len(mc.data),
	}
}

// Métodos de DiskCache
func (dc *DiskCache) get(key string, ttl time.Duration) (*CacheEntry, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	filePath := dc.getFilePath(key)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, false
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	// Verificar TTL
	if time.Since(entry.Timestamp) > ttl {
		os.Remove(filePath)
		return nil, false
	}

	return &entry, true
}

func (dc *DiskCache) set(key string, data interface{}, ttl time.Duration) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	entry := &CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		TTL:       ttl,
		Key:       key,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	filePath := dc.getFilePath(key)
	return os.WriteFile(filePath, jsonData, 0644)
}

func (dc *DiskCache) delete(key string) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	filePath := dc.getFilePath(key)
	os.Remove(filePath)
}

func (dc *DiskCache) clear() error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	return os.RemoveAll(dc.cacheDir)
}

func (dc *DiskCache) getFilePath(key string) string {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(key)))
	return filepath.Join(dc.cacheDir, hash+".cache")
}

func (dc *DiskCache) getStats() CacheStats {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	var totalSize int64
	var count int

	filepath.Walk(dc.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && filepath.Ext(path) == ".cache" {
			totalSize += info.Size()
			count++
		}
		return nil
	})

	return CacheStats{
		DiskUsageMB: float64(totalSize) / (1024 * 1024),
		EntryCount:  count,
	}
}

// cleanupExpiredEntries limpia entradas expiradas periódicamente
func (cs *CacheService) cleanupExpiredEntries() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cs.logger.Debug("Running cache cleanup")
		
		// Limpiar memoria
		cs.memoryCache.mu.Lock()
		for key, entry := range cs.memoryCache.data {
			if time.Since(entry.Timestamp) > entry.TTL {
				cs.memoryCache.currSize -= entry.Size
				delete(cs.memoryCache.data, key)
				cs.memoryCache.removeLRU(key)
			}
		}
		cs.memoryCache.mu.Unlock()

		// Limpiar disco
		filepath.Walk(cs.diskCache.cacheDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && filepath.Ext(path) == ".cache" {
				if time.Since(info.ModTime()) > cs.config.DiskTTL {
					os.Remove(path)
				}
			}
			return nil
		})
	}
}