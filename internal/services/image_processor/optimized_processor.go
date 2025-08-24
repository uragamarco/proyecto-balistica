package image_processor

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/uragamarco/proyecto-balistica/internal/services/python_features"
	"go.uber.org/zap"
)

// OptimizedImageProcessor versión optimizada del procesador de imágenes
type OptimizedImageProcessor struct {
	config         *Config
	pythonFeatures *python_features.Service
	featureCache   *FeatureCache
	workerPool     *WorkerPool
}

// FeatureCache sistema de cache para características calculadas
type FeatureCache struct {
	mu    sync.RWMutex
	cache map[string]CachedFeatures
	ttl   time.Duration
}

type CachedFeatures struct {
	features  map[string]float64
	advanced  map[string]interface{}
	timestamp time.Time
}

// WorkerPool pool de workers para procesamiento paralelo
type WorkerPool struct {
	workers   int
	jobs      chan FeatureJob
	results   chan FeatureResult
	wg        sync.WaitGroup
}

type FeatureJob struct {
	img    image.Image
	region image.Rectangle
	jobID  int
}

type FeatureResult struct {
	jobID    int
	features map[string]float64
	err      error
}

// NewOptimizedImageProcessor crea una nueva instancia optimizada
func NewOptimizedImageProcessor(cfg *Config, pyService *python_features.Service) *OptimizedImageProcessor {
	cache := &FeatureCache{
		cache: make(map[string]CachedFeatures),
		ttl:   5 * time.Minute, // Cache por 5 minutos
	}

	workerPool := &WorkerPool{
		workers: runtime.NumCPU(),
		jobs:    make(chan FeatureJob, 100),
		results: make(chan FeatureResult, 100),
	}

	// Inicializar workers
	for i := 0; i < workerPool.workers; i++ {
		go workerPool.worker()
	}

	return &OptimizedImageProcessor{
		config:         cfg,
		pythonFeatures: pyService,
		featureCache:   cache,
		workerPool:     workerPool,
	}
}

// ExtractFeaturesOptimized extracción optimizada de características
func (oip *OptimizedImageProcessor) ExtractFeaturesOptimized(img image.Image, originalPath string) (map[string]float64, map[string]interface{}, error) {
	start := time.Now()
	defer func() {
		if oip.config.Logger != nil {
			oip.config.Logger.Info("Feature extraction completed",
				zap.Duration("duration", time.Since(start)))
		}
	}()

	// Verificar cache primero
	cacheKey := oip.generateCacheKey(originalPath, img.Bounds())
	if cached, found := oip.featureCache.get(cacheKey); found {
		return cached.features, cached.advanced, nil
	}

	// Procesamiento paralelo de características locales
	localFeatures, err := oip.extractLocalFeaturesParallel(img)
	if err != nil {
		return nil, nil, fmt.Errorf("error extracting local features: %w", err)
	}

	// Características avanzadas con Python (si está disponible)
	var advancedFeatures map[string]interface{}
	if oip.pythonFeatures != nil {
		advancedFeatures, err = oip.extractAdvancedFeaturesOptimized(img, originalPath)
		if err != nil {
			oip.config.Logger.Warn("Advanced features extraction failed", zap.Error(err))
			advancedFeatures = make(map[string]interface{})
		}
	} else {
		advancedFeatures = make(map[string]interface{})
	}

	// Guardar en cache
	oip.featureCache.set(cacheKey, CachedFeatures{
		features:  localFeatures,
		advanced:  advancedFeatures,
		timestamp: time.Now(),
	})

	return localFeatures, advancedFeatures, nil
}

// extractLocalFeaturesParallel extracción paralela de características locales
func (oip *OptimizedImageProcessor) extractLocalFeaturesParallel(img image.Image) (map[string]float64, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Dividir imagen en regiones para procesamiento paralelo
	numRegions := oip.workerPool.workers
	regionHeight := height / numRegions

	// Enviar trabajos a workers
	for i := 0; i < numRegions; i++ {
		startY := i * regionHeight
		endY := startY + regionHeight
		if i == numRegions-1 {
			endY = height // Última región toma el resto
		}

		region := image.Rect(0, startY, width, endY)
		oip.workerPool.jobs <- FeatureJob{
			img:    img,
			region: region,
			jobID:  i,
		}
	}

	// Recopilar resultados
	regionFeatures := make([]map[string]float64, numRegions)
	for i := 0; i < numRegions; i++ {
		result := <-oip.workerPool.results
		if result.err != nil {
			return nil, result.err
		}
		regionFeatures[result.jobID] = result.features
	}

	// Combinar características de todas las regiones
	return oip.combineRegionFeatures(regionFeatures, img), nil
}

// worker función del worker para procesamiento paralelo
func (wp *WorkerPool) worker() {
	for job := range wp.jobs {
		features := make(map[string]float64)
		
		// Calcular GLCM para la región
		glcmFeatures := calculateGLCMFeaturesOptimized(job.img, job.region)
		for i, feature := range []string{"glcm_contrast", "glcm_energy", "glcm_homogeneity"} {
			if i < len(glcmFeatures) {
				features[feature] = glcmFeatures[i]
			}
		}

		// Calcular características de forma para la región
		shapeFeatures := calculateShapeFeaturesOptimized(job.img, job.region)
		for i, feature := range []string{"circularity", "aspect_ratio"} {
			if i < len(shapeFeatures) {
				features[feature] = shapeFeatures[i]
			}
		}

		wp.results <- FeatureResult{
			jobID:    job.jobID,
			features: features,
			err:      nil,
		}
	}
}

// calculateGLCMFeaturesOptimized versión optimizada del cálculo GLCM
func calculateGLCMFeaturesOptimized(img image.Image, region image.Rectangle) []float64 {
	glcm := make(map[[2]uint8]int)
	offset := 1 // Distancia fija para optimización

	// Calcular GLCM solo en la región especificada
	for y := region.Min.Y; y < region.Max.Y-offset; y++ {
		for x := region.Min.X; x < region.Max.X-offset; x++ {
			gray1 := getGrayValue(img.At(x, y))
			gray2 := getGrayValue(img.At(x+offset, y))
			glcm[[2]uint8{gray1, gray2}]++
		}
	}

	if len(glcm) == 0 {
		return []float64{0, 0, 0}
	}

	return []float64{
		calculateContrast(glcm),
		calculateEnergy(glcm),
		calculateHomogeneity(glcm),
	}
}

// calculateShapeFeaturesOptimized versión optimizada del cálculo de forma
func calculateShapeFeaturesOptimized(img image.Image, region image.Rectangle) []float64 {
	var area, perimeter float64
	var minX, maxX, minY, maxY int = region.Max.X, region.Min.X, region.Max.Y, region.Min.Y

	// Calcular área, perímetro y bounding box en una sola pasada
	for y := region.Min.Y; y < region.Max.Y; y++ {
		for x := region.Min.X; x < region.Max.X; x++ {
			if isForeground(img.At(x, y)) {
				area++
				// Actualizar bounding box
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
				// Verificar si es pixel de borde
				if isEdgePixelOptimized(img, x, y, region) {
					perimeter++
				}
			}
		}
	}

	if area == 0 || perimeter == 0 {
		return []float64{0, 0}
	}

	// Calcular métricas
	circularity := (4 * math.Pi * area) / (perimeter * perimeter)
	width := float64(maxX - minX)
	height := float64(maxY - minY)
	aspectRatio := width / math.Max(height, 1)

	return []float64{circularity, aspectRatio}
}

// isEdgePixelOptimized versión optimizada de detección de bordes
func isEdgePixelOptimized(img image.Image, x, y int, region image.Rectangle) bool {
	// Verificar solo vecinos necesarios dentro de la región
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if nx >= region.Min.X && nx < region.Max.X && ny >= region.Min.Y && ny < region.Max.Y {
				if !isForeground(img.At(nx, ny)) {
					return true
				}
			}
		}
	}
	return false
}

// combineRegionFeatures combina características de múltiples regiones
func (oip *OptimizedImageProcessor) combineRegionFeatures(regionFeatures []map[string]float64, img image.Image) map[string]float64 {
	combined := make(map[string]float64)
	numRegions := float64(len(regionFeatures))

	// Promediar características numéricas
	for _, features := range regionFeatures {
		for key, value := range features {
			combined[key] += value / numRegions
		}
	}

	// Agregar características globales de la imagen completa
	bounds := img.Bounds()
	combined["image_width"] = float64(bounds.Dx())
	combined["image_height"] = float64(bounds.Dy())
	combined["image_area"] = float64(bounds.Dx() * bounds.Dy())

	return combined
}

// extractAdvancedFeaturesOptimized extracción optimizada de características avanzadas
func (oip *OptimizedImageProcessor) extractAdvancedFeaturesOptimized(img image.Image, originalPath string) (map[string]interface{}, error) {
	// Redimensionar imagen si es muy grande para optimizar procesamiento Python
	bounds := img.Bounds()
	maxDimension := 1024
	if bounds.Dx() > maxDimension || bounds.Dy() > maxDimension {
		img = imaging.Fit(img, maxDimension, maxDimension, imaging.Lanczos)
	}

	// Crear archivo temporal optimizado
	tempFile := filepath.Join(oip.config.TempDir, fmt.Sprintf("opt_temp_%d.jpg", time.Now().UnixNano()))
	defer os.Remove(tempFile)

	if err := imaging.Save(img, tempFile); err != nil {
		return nil, fmt.Errorf("error saving temp file: %w", err)
	}

	response, err := oip.pythonFeatures.ExtractFeatures(tempFile)
	if err != nil {
		return nil, err
	}
	
	// Convertir PythonResponse a map[string]interface{}
	result := make(map[string]interface{})
	if response != nil {
		result["hu_moments"] = response.HuMoments
		result["firing_pin_marks"] = response.FiringPinMarks
		result["striation_patterns"] = response.StriationPatterns
		result["contour_area"] = response.ContourArea
		result["contour_len"] = response.ContourLen
		result["lbp_uniformity"] = response.LBPUniformity
	}
	
	return result, nil
}

// Métodos de cache
func (fc *FeatureCache) get(key string) (CachedFeatures, bool) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	cached, exists := fc.cache[key]
	if !exists {
		return CachedFeatures{}, false
	}

	// Verificar TTL
	if time.Since(cached.timestamp) > fc.ttl {
		delete(fc.cache, key)
		return CachedFeatures{}, false
	}

	return cached, true
}

func (fc *FeatureCache) set(key string, features CachedFeatures) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.cache[key] = features
}

func (oip *OptimizedImageProcessor) generateCacheKey(path string, bounds image.Rectangle) string {
	return fmt.Sprintf("%s_%dx%d", filepath.Base(path), bounds.Dx(), bounds.Dy())
}

// Cleanup limpia recursos
func (oip *OptimizedImageProcessor) Cleanup() {
	close(oip.workerPool.jobs)
	oip.workerPool.wg.Wait()
}