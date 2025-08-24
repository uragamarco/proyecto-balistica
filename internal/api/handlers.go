package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/uragamarco/proyecto-balistica/internal/models"
	"github.com/uragamarco/proyecto-balistica/internal/services/chroma"
	"github.com/uragamarco/proyecto-balistica/internal/services/classification"
	"github.com/uragamarco/proyecto-balistica/internal/services/comparison"
	"github.com/uragamarco/proyecto-balistica/internal/storage"
	"go.uber.org/zap"
)

// Interfaces para testing
type ImageProcessorInterface interface {
	Process(img image.Image) (image.Image, error)
	ExtractFeatures(img image.Image, tempPath string) (map[string]float64, map[string]interface{}, error)
	PythonFeaturesStatus() (bool, string)
}

type ChromaServiceInterface interface {
	Analyze(img image.Image) (*chroma.ChromaAnalysis, error)
}

// APIError representa un error de la API con código HTTP y mensaje
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error implementa la interfaz error
func (e APIError) Error() string {
	return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
}

// NewAPIError crea un nuevo error de API
func NewAPIError(code int, message string, err error) APIError {
	details := ""
	if err != nil {
		details = err.Error()
	}
	return APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

type Handlers struct {
	imageProcessor        ImageProcessorInterface
	chromaService         ChromaServiceInterface
	advancedComparison    *comparison.AdvancedComparison
	storageService        *storage.StorageService
	classificationService *classification.ClassificationService
	Logger                *zap.Logger
}

func NewHandlers(logger *zap.Logger, ip ImageProcessorInterface, cs ChromaServiceInterface, ss *storage.StorageService, classificationSvc *classification.ClassificationService) *Handlers {
	return &Handlers{
		imageProcessor:        ip,
		chromaService:         cs,
		advancedComparison:    comparison.NewAdvancedComparison(logger),
		storageService:        ss,
		classificationService: classificationSvc,
		Logger:                logger,
	}
}

// respondWithError envía una respuesta de error en formato JSON
func (h *Handlers) respondWithError(w http.ResponseWriter, apiErr APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Code)
	
	// Registrar el error en los logs
	h.Logger.Error("API Error",
		zap.Int("code", apiErr.Code),
		zap.String("message", apiErr.Message),
		zap.String("details", apiErr.Details))
	
	// Enviar respuesta JSON
	if err := json.NewEncoder(w).Encode(apiErr); err != nil {
		h.Logger.Error("Error al serializar respuesta de error", zap.Error(err))
	}
}

// respondWithJSON envía una respuesta exitosa en formato JSON
func (h *Handlers) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			h.Logger.Error("Error al serializar respuesta JSON", zap.Error(err))
			return
		}
	}
}

// combineClassificationIndicators combina indicadores de clasificación de arma y calibre
func (h *Handlers) combineClassificationIndicators(weaponIndicators, caliberIndicators map[string]float64) map[string]float64 {
	combined := make(map[string]float64)

	// Copiar indicadores de arma con prefijo
	for key, value := range weaponIndicators {
		combined["weapon_"+key] = value
	}

	// Copiar indicadores de calibre con prefijo
	for key, value := range caliberIndicators {
		combined["caliber_"+key] = value
	}

	return combined
}

func convertColorData(cd []chroma.ColorData) []models.ColorData {
	result := make([]models.ColorData, len(cd))
	for i, c := range cd {
		r, g, b, _ := c.Color.RGBA()
		result[i] = models.ColorData{
			Color: models.RGB{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
			},
			Frequency: c.Frequency,
		}
	}
	return result
}

// calculateAverageVariance calcula el promedio de las varianzas de color
func calculateAverageVariance(variances map[string]float64) float64 {
	if len(variances) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range variances {
		sum += v
	}

	return sum / float64(len(variances))
}

func (h *Handlers) ProcessImage(w http.ResponseWriter, r *http.Request) {
	// Validar método HTTP
	if r.Method != http.MethodPost {
		h.respondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "Método no permitido", nil))
		return
	}

	// Parse multipart form (20MB limit)
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		h.respondWithError(w, NewAPIError(http.StatusBadRequest, "Error al procesar el formulario", err))
		return
	}

	// Obtener archivo de imagen
	file, handler, err := r.FormFile("image")
	if err != nil {
		h.respondWithError(w, NewAPIError(http.StatusBadRequest, "Error al obtener la imagen", err))
		return
	}
	defer file.Close()

	// Validar tipo de archivo
	contentType := handler.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/tiff" {
		h.respondWithError(w, NewAPIError(
			http.StatusBadRequest, 
			"Formato de imagen no soportado", 
			fmt.Errorf("tipo de contenido no válido: %s. Use JPEG, PNG o TIFF", contentType)))
		return
	}

	// Registrar información sobre la imagen recibida
	h.Logger.Info("Procesando imagen", 
		zap.String("filename", handler.Filename),
		zap.String("content-type", contentType),
		zap.Int64("size", handler.Size))

	// Leer contenido de imagen en memoria
	var imgBytes bytes.Buffer
	if _, copyErr := io.Copy(&imgBytes, file); copyErr != nil {
		h.respondWithError(w, NewAPIError(http.StatusInternalServerError, "Error al leer la imagen", err))
		return
	}

	// Decodificar imagen
	img, imgFormat, err := image.Decode(bytes.NewReader(imgBytes.Bytes()))
	if err != nil {
		h.respondWithError(w, NewAPIError(http.StatusInternalServerError, "Error al decodificar la imagen", err))
		return
	}
	
	h.Logger.Debug("Imagen decodificada correctamente", 
		zap.String("format", imgFormat))

	// Crear archivo temporal para procesamiento
	tempDir := "/tmp/balistica"
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	// Asegurar que el directorio temporal existe
	if mkdirErr := os.MkdirAll(tempDir, 0755); mkdirErr != nil {
		h.respondWithError(w, NewAPIError(http.StatusInternalServerError, "Error al crear directorio temporal", mkdirErr))
		return
	}

	tempFile, err := os.CreateTemp(tempDir, "balistica_*.png")
	if err != nil {
		h.respondWithError(w, NewAPIError(http.StatusInternalServerError, "Error al crear archivo temporal", err))
		return
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	h.Logger.Debug("Archivo temporal creado", zap.String("path", tempFile.Name()))

	// Escribir imagen en archivo temporal
	if _, writeErr := tempFile.Write(imgBytes.Bytes()); writeErr != nil {
		h.respondWithError(w, NewAPIError(http.StatusInternalServerError, "Error al escribir archivo temporal", writeErr))
		return
	}

	// Procesar imagen
	processStart := time.Now()
	processedImg, err := h.imageProcessor.Process(img)
	if err != nil {
		h.respondWithError(w, NewAPIError(http.StatusInternalServerError, "Error al procesar la imagen", err))
		return
	}
	h.Logger.Debug("Imagen procesada", zap.Duration("duration", time.Since(processStart)))

	// Extraer características
	featuresStart := time.Now()
	features, metadata, err := h.imageProcessor.ExtractFeatures(processedImg, tempFile.Name())
	if err != nil {
		h.respondWithError(w, NewAPIError(http.StatusInternalServerError, "Error al extraer características", err))
		return
	}
	
	// Validar que las características no contengan valores infinitos o NaN
	for key, value := range features {
		if math.IsInf(value, 0) || math.IsNaN(value) {
			features[key] = 0.0 // Reemplazar con valor por defecto
			h.Logger.Warn("Feature contiene valor infinito o NaN, reemplazando con 0.0", 
				zap.String("feature", key), 
				zap.Float64("original_value", value))
		}
	}
	
	h.Logger.Debug("Características extraídas", zap.Duration("duration", time.Since(featuresStart)))

	// Analizar croma
	chromaStart := time.Now()
	chromaAnalysis, err := h.chromaService.Analyze(processedImg)
	if err != nil {
		h.respondWithError(w, NewAPIError(http.StatusInternalServerError, "Error al analizar el croma", err))
		return
	}
	h.Logger.Debug("Análisis de croma completado", zap.Duration("duration", time.Since(chromaStart)))

	// Generar hash de imagen
	hash := models.GenerateImageHashFromBytes(imgBytes.Bytes())

	// Verificar estado de características Python
	pyEnabled, pyStatus := h.imageProcessor.PythonFeaturesStatus()
	if pyStatus != "" {
		h.Logger.Info("Estado de características Python", zap.String("status", pyStatus))
	}

	// Calcular confianza del análisis
	confidence := calculateAnalysisConfidence(features)
	
	// Validar que la confianza sea un valor finito
	if math.IsInf(confidence, 0) || math.IsNaN(confidence) {
		confidence = 0.0 // Valor por defecto si es infinito o NaN
		h.Logger.Warn("Confidence calculada es infinita o NaN, usando valor por defecto", zap.Float64("original_confidence", confidence))
	}

	// Calcular varianza de color y validar
	colorVariance := calculateAverageVariance(chromaAnalysis.ColorVariance)
	if math.IsInf(colorVariance, 0) || math.IsNaN(colorVariance) {
		colorVariance = 0.0 // Valor por defecto si es infinito o NaN
		h.Logger.Warn("ColorVariance calculada es infinita o NaN, usando valor por defecto", zap.Float64("original_variance", colorVariance))
	}

	// Realizar clasificación automática si el servicio está disponible
	var classification *models.ClassificationResult
	if h.classificationService != nil {
		classificationStart := time.Now()
		
		// Generar ID único para el análisis
		analysisID := fmt.Sprintf("analysis_%d", time.Now().UnixNano())
		
		// Realizar clasificación
		classificationResult, err := h.classificationService.ClassifyBallistic(r.Context(), analysisID, features)
		if err != nil {
			h.Logger.Warn("Error en clasificación automática", zap.Error(err))
			// No fallar el análisis completo por error en clasificación
		} else {
			// Convertir resultado de clasificación al modelo de respuesta
			classification = &models.ClassificationResult{
				WeaponType:   classificationResult.WeaponType.WeaponType,
				Caliber:      classificationResult.Caliber.Caliber,
				Confidence:   (classificationResult.WeaponType.Confidence + classificationResult.Caliber.Confidence) / 2,
				Indicators:   h.combineClassificationIndicators(classificationResult.WeaponType.Indicators, classificationResult.Caliber.Indicators),
				Evidence:     append(classificationResult.WeaponType.Evidence, classificationResult.Caliber.Evidence...),
				OverallScore: classificationResult.OverallScore,
			}
			
			h.Logger.Info("Clasificación automática completada",
				zap.String("weapon_type", classification.WeaponType),
				zap.String("caliber", classification.Caliber),
				zap.Float64("confidence", classification.Confidence),
				zap.Duration("duration", time.Since(classificationStart)))
		}
	}

	// Preparar respuesta
	response := models.BallisticAnalysis{
		Features: features,
		ChromaData: models.ChromaAnalysis{
			DominantColors: convertColorData(chromaAnalysis.DominantColors),
			ColorVariance:  colorVariance,
		},
		Classification: classification, // Incluir clasificación automática
		Metadata: models.AnalysisMetadata{
			Timestamp:          time.Now().UTC().Format(time.RFC3339),
			ImageHash:          hash,
			ProcessorVersion:   "1.4.0", // Versión actualizada
			PythonFeaturesUsed: pyEnabled,
			Confidence:         confidence,
		},
	}
	
	// Agregar metadatos adicionales si están disponibles
	if filename, ok := metadata["filename"].(string); ok {
		response.Metadata.Filename = filename
	} else {
		response.Metadata.Filename = handler.Filename // Usar el nombre del archivo subido
	}
	
	if contentType, ok := metadata["content_type"].(string); ok {
		response.Metadata.ContentType = contentType
	} else {
		response.Metadata.ContentType = contentType // Usar el tipo de contenido detectado
	}
	
	if fileSize, ok := metadata["file_size"].(int64); ok {
		response.Metadata.FileSize = fileSize
	} else {
		response.Metadata.FileSize = handler.Size // Usar el tamaño del archivo subido
	}

	// Registrar información sobre el análisis completado
	h.Logger.Info("Análisis completado",
		zap.String("image_hash", hash),
		zap.Float64("confidence", confidence),
		zap.Bool("python_features", pyEnabled))

	// Responder con JSON
	h.respondWithJSON(w, http.StatusOK, response)
}

// calculateAnalysisConfidence calcula la confianza del análisis
func calculateAnalysisConfidence(features map[string]float64) float64 {
	if len(features) == 0 {
		return 0.0
	}

	// Variables para cálculo de confianza
	var (
		featureCount = len(features)
		qualityScore = 0.0
		keyFeatures  = 0
	)

	// Ponderar características clave
	if hu, ok := features["hu_moment_1"]; ok && !math.IsInf(hu, 0) && !math.IsNaN(hu) {
		qualityScore += math.Abs(hu) * 0.3
		keyFeatures++
	}
	if area, ok := features["contour_area"]; ok && area > 0 && !math.IsInf(area, 0) && !math.IsNaN(area) {
		logArea := math.Log(area)
		if !math.IsInf(logArea, 0) && !math.IsNaN(logArea) {
			qualityScore += logArea * 0.2
			keyFeatures++
		}
	}
	if str, ok := features["striation_density"]; ok && !math.IsInf(str, 0) && !math.IsNaN(str) {
		qualityScore += str * 0.2
		keyFeatures++
	}

	// Si no hay características clave, devolver confianza baja
	if keyFeatures == 0 {
		return 0.0
	}

	// Normalizar puntaje de calidad
	qualityScore = qualityScore / float64(keyFeatures)

	// Verificar que qualityScore sea finito
	if math.IsInf(qualityScore, 0) || math.IsNaN(qualityScore) {
		qualityScore = 0.0
	}

	// Factor de completitud (asumiendo 15 características esperadas)
	completeness := float64(featureCount) / 15.0
	completeness = math.Min(completeness, 1.0) // No más de 1

	// Combinar factores
	confidence := (qualityScore * 0.7) + (completeness * 0.3)

	// Verificar que confidence sea finito y limitar a rango [0, 1]
	if math.IsInf(confidence, 0) || math.IsNaN(confidence) {
		confidence = 0.0
	}
	return math.Max(0, math.Min(1, confidence))
}

func (h *Handlers) CompareSamples(w http.ResponseWriter, r *http.Request) {
	// Validar método HTTP
	if r.Method != http.MethodPost {
		h.respondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "Método no permitido", nil))
		return
	}

	var comparisonRequest struct {
		Sample1   map[string]float64 `json:"sample1"`
		Sample2   map[string]float64 `json:"sample2"`
		Weights   map[string]float64 `json:"weights,omitempty"`
		Threshold float64            `json:"threshold,omitempty"`
		UseAdvanced bool             `json:"use_advanced,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&comparisonRequest); err != nil {
		h.respondWithError(w, NewAPIError(http.StatusBadRequest, "Cuerpo de solicitud inválido", err))
		return
	}

	// Validar muestras de entrada
	if len(comparisonRequest.Sample1) == 0 || len(comparisonRequest.Sample2) == 0 {
		h.respondWithError(w, NewAPIError(http.StatusBadRequest, "Muestras vacías", errors.New("ambas muestras deben contener datos")))
		return
	}

	h.Logger.Debug("Solicitud de comparación recibida", 
		zap.Int("sample1_features", len(comparisonRequest.Sample1)),
		zap.Int("sample2_features", len(comparisonRequest.Sample2)),
		zap.Bool("use_advanced", comparisonRequest.UseAdvanced))

	// Usar comparación avanzada si se solicita
	if comparisonRequest.UseAdvanced {
		h.handleAdvancedComparison(w, comparisonRequest)
		return
	}

	// Implementar lógica de comparación básica
	similarityScore := h.compareFeatures(
		comparisonRequest.Sample1,
		comparisonRequest.Sample2,
		comparisonRequest.Weights,
	)

	// Determinar umbral de coincidencia
	matchThreshold := 0.85
	if comparisonRequest.Threshold > 0 {
		matchThreshold = comparisonRequest.Threshold
	}

	match := similarityScore >= matchThreshold
	confidence := h.calculateBasicConfidence(similarityScore, comparisonRequest.Sample1, comparisonRequest.Sample2)

	// Preparar respuesta extendida
	response := models.ComparisonResult{
		Similarity:      similarityScore,
		Match:           match,
		Confidence:      confidence,
		FeatureWeights:  comparisonRequest.Weights,
		DiffPerFeature:  calculateFeatureDiffs(comparisonRequest.Sample1, comparisonRequest.Sample2),
		AreasOfInterest: identifyCriticalDifferences(comparisonRequest.Sample1, comparisonRequest.Sample2),
	}

	// Registrar información sobre la comparación
	h.Logger.Info("Comparación completada",
		zap.Float64("similarity", similarityScore),
		zap.Bool("match", match),
		zap.Float64("confidence", confidence),
		zap.Float64("threshold", matchThreshold))

	// Responder con JSON
	h.respondWithJSON(w, http.StatusOK, response)
}

// calculateFeatureDiffs calcula diferencias por característica
func calculateFeatureDiffs(s1, s2 map[string]float64) map[string]float64 {
	diffs := make(map[string]float64)
	for k, v1 := range s1 {
		if v2, ok := s2[k]; ok {
			diffs[k] = math.Abs(v1 - v2)
		}
	}
	return diffs
}

// handleAdvancedComparison maneja comparaciones usando el servicio avanzado
func (h *Handlers) handleAdvancedComparison(w http.ResponseWriter, comparisonRequest struct {
	Sample1     map[string]float64 `json:"sample1"`
	Sample2     map[string]float64 `json:"sample2"`
	Weights     map[string]float64 `json:"weights,omitempty"`
	Threshold   float64            `json:"threshold,omitempty"`
	UseAdvanced bool               `json:"use_advanced,omitempty"`
}) {
	// Convertir weights a FeatureWeights o usar valores por defecto
	weights := h.convertToFeatureWeights(comparisonRequest.Weights)
	
	// Usar el servicio de comparación avanzada
	result := h.advancedComparison.CompareAdvanced(
		comparisonRequest.Sample1,
		comparisonRequest.Sample2,
		weights,
	)

	if result == nil {
		h.respondWithError(w, NewAPIError(http.StatusInternalServerError, "Error en comparación avanzada", nil))
		return
	}

	h.Logger.Info("Comparación avanzada completada",
		zap.Float64("similarity", result.Similarity),
		zap.Bool("match", result.Match),
		zap.Float64("confidence", result.Confidence),
		zap.Float64("ballistic_score", result.BallisticIndicators.OverallBallisticScore))

	h.respondWithJSON(w, http.StatusOK, result)
}

// calculateBasicConfidence calcula la confianza para comparaciones básicas
func (h *Handlers) calculateBasicConfidence(similarity float64, sample1, sample2 map[string]float64) float64 {
	// Calcular confianza basada en múltiples factores
	baseConfidence := similarity
	
	// Factor de ajuste basado en número de características
	featureCount := float64(len(sample1))
	if len(sample2) < len(sample1) {
		featureCount = float64(len(sample2))
	}
	
	// Penalizar si hay pocas características
	featurePenalty := 1.0
	if featureCount < 5 {
		featurePenalty = 0.8
	} else if featureCount < 10 {
		featurePenalty = 0.9
	}
	
	// Calcular varianza de las diferencias para estabilidad
	diffs := calculateFeatureDiffs(sample1, sample2)
	var variance float64
	if len(diffs) > 0 {
		var sum, mean float64
		for _, diff := range diffs {
			sum += diff
		}
		mean = sum / float64(len(diffs))
		
		for _, diff := range diffs {
			variance += math.Pow(diff-mean, 2)
		}
		variance /= float64(len(diffs))
	}
	
	// Penalizar alta varianza (inconsistencia)
	variancePenalty := 1.0
	if variance > 0.1 {
		variancePenalty = 0.9
	} else if variance > 0.05 {
		variancePenalty = 0.95
	}
	
	confidence := baseConfidence * featurePenalty * variancePenalty
	
	// Asegurar que la confianza esté en el rango [0, 1]
	if confidence > 1.0 {
		confidence = 1.0
	} else if confidence < 0.0 {
		confidence = 0.0
	}
	
	return confidence
}

// convertToFeatureWeights convierte un map[string]float64 a FeatureWeights
func (h *Handlers) convertToFeatureWeights(weights map[string]float64) comparison.FeatureWeights {
	// Usar valores por defecto si no se proporcionan weights
	if len(weights) == 0 {
		return comparison.GetDefaultWeights()
	}
	
	// Crear FeatureWeights con valores por defecto
	fw := comparison.GetDefaultWeights()
	
	// Sobrescribir con valores proporcionados
	if val, exists := weights["striation_features"]; exists {
		fw.StriationFeatures = val
	}
	if val, exists := weights["firing_pin_features"]; exists {
		fw.FiringPinFeatures = val
	}
	if val, exists := weights["breech_face_features"]; exists {
		fw.BreechFaceFeatures = val
	}
	if val, exists := weights["ejector_features"]; exists {
		fw.EjectorFeatures = val
	}
	if val, exists := weights["extractor_features"]; exists {
		fw.ExtractorFeatures = val
	}
	if val, exists := weights["geometric_features"]; exists {
		fw.GeometricFeatures = val
	}
	if val, exists := weights["texture_features"]; exists {
		fw.TextureFeatures = val
	}
	if val, exists := weights["color_features"]; exists {
		fw.ColorFeatures = val
	}
	if val, exists := weights["shape_features"]; exists {
		fw.ShapeFeatures = val
	}
	if val, exists := weights["contour_features"]; exists {
		fw.ContourFeatures = val
	}
	
	return fw
}

// identifyCriticalDifferences identifica diferencias críticas
func identifyCriticalDifferences(s1, s2 map[string]float64) []string {
	var critical []string
	const threshold = 0.2 // Diferencia significativa

	for feature, v1 := range s1 {
		if v2, ok := s2[feature]; ok {
			diff := math.Abs(v1 - v2)
			if diff > threshold {
				critical = append(critical, feature)
			}
		}
	}
	return critical
}

func (h *Handlers) compareFeatures(f1, f2, weights map[string]float64) float64 {
	if len(f1) == 0 || len(f2) == 0 {
		h.Logger.Warn("Comparación con características vacías", 
			zap.Int("f1_length", len(f1)), 
			zap.Int("f2_length", len(f2)))
		return 0.0
	}

	var totalWeight, weightedSum float64
	validFeatures := 0
	skippedFeatures := 0

	for feature, value1 := range f1 {
		value2, exists := f2[feature]
		if !exists {
			skippedFeatures++
			continue
		}

		// Obtener peso (valor por defecto 1.0)
		weight := 1.0
		if w, ok := weights[feature]; ok {
			weight = w
		}

		// Calcular diferencia normalizada
		maxVal := math.Max(math.Abs(value1), math.Abs(value2))
		if maxVal < 1e-9 { // Evitar división por cero
			maxVal = 1
			h.Logger.Debug("Valor máximo muy pequeño, usando 1 para evitar división por cero", 
				zap.String("feature", feature))
		}

		diff := math.Abs(value1 - value2)
		normalizedDiff := diff / maxVal

		// Acumular similitud ponderada
		weightedSum += weight * (1 - normalizedDiff)
		totalWeight += weight
		validFeatures++
	}

	// Registrar estadísticas de comparación
	h.Logger.Debug("Estadísticas de comparación", 
		zap.Int("total_features_f1", len(f1)),
		zap.Int("total_features_f2", len(f2)),
		zap.Int("valid_features", validFeatures),
		zap.Int("skipped_features", skippedFeatures))

	if validFeatures == 0 || totalWeight == 0 {
		h.Logger.Warn("No se encontraron características válidas para comparar", 
			zap.Int("valid_features", validFeatures), 
			zap.Float64("total_weight", totalWeight))
		return 0.0
	}

	similarity := weightedSum / totalWeight
	h.Logger.Debug("Similitud calculada", zap.Float64("similarity", similarity))
	return similarity
}

// Storage handler methods

// GetAnalysesHandler recupera todos los análisis con paginación
func (h *Handlers) GetAnalysesHandler(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	analyses, err := h.storageService.GetAllAnalysis(limit, offset)
	if err != nil {
		h.Logger.Error("Error al recuperar análisis", zap.Error(err))
		http.Error(w, "Error al recuperar análisis", http.StatusInternalServerError)
		return
	}

	totalCount, err := h.storageService.GetAnalysisCount()
	if err != nil {
		h.Logger.Warn("Error al obtener conteo total", zap.Error(err))
		totalCount = 0
	}

	response := map[string]interface{}{
		"data":     analyses,
		"total":    totalCount,
		"limit":    limit,
		"offset":   offset,
		"has_more": len(analyses) == limit,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.Logger.Error("Error al codificar respuesta", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// GetAnalysisHandler recupera un análisis por ID
func (h *Handlers) GetAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/analyses/")
	id := strings.TrimSuffix(path, "/")

	if id == "" {
		http.Error(w, "ID de análisis requerido", http.StatusBadRequest)
		return
	}

	analysis, err := h.storageService.GetAnalysis(id)
	if err != nil {
		h.Logger.Error("Error al recuperar análisis", zap.String("id", id), zap.Error(err))
		http.Error(w, "Análisis no encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(analysis); err != nil {
		h.Logger.Error("Error al codificar respuesta", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// SearchAnalysesHandler busca análisis por ruta de imagen
func (h *Handlers) SearchAnalysesHandler(w http.ResponseWriter, r *http.Request) {
	imagePath := r.URL.Query().Get("image_path")

	if imagePath == "" {
		http.Error(w, "Parámetro image_path requerido", http.StatusBadRequest)
		return
	}

	analyses, err := h.storageService.SearchAnalysisByImagePath(imagePath)
	if err != nil {
		h.Logger.Error("Error al buscar análisis", zap.String("image_path", imagePath), zap.Error(err))
		http.Error(w, "Error al buscar análisis", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(analyses); err != nil {
		h.Logger.Error("Error al codificar respuesta", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// DeleteAnalysisHandler elimina un análisis por ID
func (h *Handlers) DeleteAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/analyses/delete/")
	id := strings.TrimSuffix(path, "/")

	if id == "" {
		http.Error(w, "ID de análisis requerido", http.StatusBadRequest)
		return
	}

	err := h.storageService.DeleteAnalysis(id)
	if err != nil {
		h.Logger.Error("Error al eliminar análisis", zap.String("id", id), zap.Error(err))
		http.Error(w, "Error al eliminar análisis", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"message": "Análisis eliminado exitosamente"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.Logger.Error("Error al codificar respuesta", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// GetComparisonHandler recupera una comparación por ID
func (h *Handlers) GetComparisonHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/comparisons/")
	id := strings.TrimSuffix(path, "/")

	if id == "" {
		http.Error(w, "ID de comparación requerido", http.StatusBadRequest)
		return
	}

	comparison, err := h.storageService.GetComparison(id)
	if err != nil {
		h.Logger.Error("Error al recuperar comparación", zap.String("id", id), zap.Error(err))
		http.Error(w, "Comparación no encontrada", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(comparison); err != nil {
		h.Logger.Error("Error al codificar respuesta", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// GetComparisonsBySampleHandler recupera comparaciones por muestra
func (h *Handlers) GetComparisonsBySampleHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/comparisons/sample/")
	sampleID := strings.TrimSuffix(path, "/")

	if sampleID == "" {
		http.Error(w, "ID de muestra requerido", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	comparisons, err := h.storageService.GetComparisonsBySample(sampleID, limit, offset)
	if err != nil {
		h.Logger.Error("Error al recuperar comparaciones", zap.String("sample_id", sampleID), zap.Error(err))
		http.Error(w, "Error al recuperar comparaciones", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(comparisons); err != nil {
		h.Logger.Error("Error al codificar respuesta", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// GetSimilarComparisonsHandler recupera comparaciones similares
func (h *Handlers) GetSimilarComparisonsHandler(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	comparisons, err := h.storageService.GetMatchingComparisons(limit, offset)
	if err != nil {
		h.Logger.Error("Error al recuperar comparaciones similares", zap.Error(err))
		http.Error(w, "Error al recuperar comparaciones", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(comparisons); err != nil {
		h.Logger.Error("Error al codificar respuesta", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// GetComparisonsByDateRangeHandler recupera comparaciones por rango de fechas
func (h *Handlers) GetComparisonsByDateRangeHandler(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		http.Error(w, "Parámetros start_date y end_date requeridos (formato: 2006-01-02)", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Formato de start_date inválido (usar: 2006-01-02)", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Formato de end_date inválido (usar: 2006-01-02)", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		if l, parseErr := strconv.Atoi(limitStr); parseErr == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, parseErr := strconv.Atoi(offsetStr); parseErr == nil && o >= 0 {
			offset = o
		}
	}

	comparisons, err := h.storageService.GetComparisonsByDateRange(startDate, endDate, limit, offset)
	if err != nil {
		h.Logger.Error("Error al recuperar comparaciones por fecha", 
			zap.String("start_date", startDateStr), 
			zap.String("end_date", endDateStr), 
			zap.Error(err))
		http.Error(w, "Error al recuperar comparaciones", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(comparisons); err != nil {
		h.Logger.Error("Error al codificar respuesta", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

// AdvancedSearchHandler realiza búsqueda avanzada usando similitud coseno
func (h *Handlers) AdvancedSearchHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Features  map[string]float64 `json:"features"`
		Threshold float64            `json:"threshold"`
		Limit     int               `json:"limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Formato de solicitud inválido", http.StatusBadRequest)
		return
	}

	if len(request.Features) == 0 {
		http.Error(w, "Características requeridas para búsqueda", http.StatusBadRequest)
		return
	}

	if request.Threshold <= 0 {
		request.Threshold = 0.8
	}

	if request.Limit <= 0 {
		request.Limit = 10
	}

	results, err := h.storageService.SearchSimilarAnalysis(request.Features, request.Threshold, request.Limit)
	if err != nil {
		h.Logger.Error("Error en búsqueda avanzada", zap.Error(err))
		http.Error(w, "Error en búsqueda avanzada", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		h.Logger.Error("Error al codificar respuesta", zap.Error(err))
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}
