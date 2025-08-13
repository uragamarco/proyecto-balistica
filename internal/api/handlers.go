package api

import (
	"bytes"
	"encoding/json"
	"image"
	"io"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/uragamarco/proyecto-balistica/internal/models"
	"github.com/uragamarco/proyecto-balistica/internal/services/chroma"
	imgproc "github.com/uragamarco/proyecto-balistica/internal/services/image_processor"
)

type Handlers struct {
	imageProcessor *imgproc.ImageProcessor
	chromaService  *chroma.Service
}

func NewHandlers(ip *imgproc.ImageProcessor, cs *chroma.Service) *Handlers {
	return &Handlers{
		imageProcessor: ip,
		chromaService:  cs,
	}
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

func (h *Handlers) ProcessImage(w http.ResponseWriter, r *http.Request) {
	// Validar método HTTP
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Obtener archivo de imagen
	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving image: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validar tipo de archivo
	contentType := handler.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		http.Error(w, "Formato de imagen no soportado. Use JPEG o PNG", http.StatusBadRequest)
		return
	}

	// Leer contenido de imagen en memoria
	var imgBytes bytes.Buffer
	if _, err := io.Copy(&imgBytes, file); err != nil {
		http.Error(w, "Error reading image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Decodificar imagen
	img, _, err := image.Decode(bytes.NewReader(imgBytes.Bytes()))
	if err != nil {
		http.Error(w, "Error decoding image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Crear archivo temporal para procesamiento
	tempDir := h.imageProcessor.Config.TempDir
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	tempFile, err := os.CreateTemp(tempDir, "balistica_*.png")
	if err != nil {
		http.Error(w, "Error creating temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	// Escribir imagen en archivo temporal
	if _, err := tempFile.Write(imgBytes.Bytes()); err != nil {
		http.Error(w, "Error writing temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Procesar imagen
	processedImg, err := h.imageProcessor.Process(img)
	if err != nil {
		http.Error(w, "Error processing image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extraer características
	features, err := h.imageProcessor.ExtractFeatures(processedImg, tempFile.Name())
	if err != nil {
		http.Error(w, "Error extracting features: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Analizar croma
	chromaAnalysis, err := h.chromaService.Analyze(processedImg)
	if err != nil {
		http.Error(w, "Error analyzing chroma: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generar hash de imagen
	hash := models.GenerateImageHashFromBytes(imgBytes.Bytes())

	// Verificar estado de características Python
	pyEnabled, pyStatus := h.imageProcessor.PythonFeaturesStatus()
	if pyStatus != "" {
		h.imageProcessor.Logger.Printf("Python features status: %s", pyStatus)
	}

	// Calcular confianza del análisis
	confidence := calculateAnalysisConfidence(features)

	// Preparar respuesta
	response := models.BallisticAnalysis{
		Features: features,
		ChromaData: models.ChromaAnalysis{
			DominantColors: convertColorData(chromaAnalysis.DominantColors),
			ColorVariance:  chromaAnalysis.ColorVariance,
		},
		Metadata: models.AnalysisMetadata{
			Timestamp:          time.Now().UTC().Format(time.RFC3339),
			ImageHash:          hash,
			ProcessorVersion:   "1.4.0", // Versión actualizada
			PythonFeaturesUsed: pyEnabled,
			Confidence:         confidence,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
	}
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
	if hu, ok := features["hu_moment_1"]; ok {
		qualityScore += math.Abs(hu) * 0.3
		keyFeatures++
	}
	if area, ok := features["contour_area"]; ok && area > 0 {
		qualityScore += math.Log(area) * 0.2
		keyFeatures++
	}
	if str, ok := features["striation_density"]; ok {
		qualityScore += str * 0.2
		keyFeatures++
	}

	// Si no hay características clave, devolver confianza baja
	if keyFeatures == 0 {
		return 0.0
	}

	// Normalizar puntaje de calidad
	qualityScore = qualityScore / float64(keyFeatures)

	// Factor de completitud (asumiendo 15 características esperadas)
	completeness := float64(featureCount) / 15.0
	completeness = math.Min(completeness, 1.0) // No más de 1

	// Combinar factores
	confidence := (qualityScore * 0.7) + (completeness * 0.3)

	// Limitar a rango [0, 1]
	return math.Max(0, math.Min(1, confidence))
}

func (h *Handlers) CompareSamples(w http.ResponseWriter, r *http.Request) {
	// Validar método HTTP
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var comparisonRequest struct {
		Sample1   map[string]float64 `json:"sample1"`
		Sample2   map[string]float64 `json:"sample2"`
		Weights   map[string]float64 `json:"weights,omitempty"`
		Threshold float64            `json:"threshold,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&comparisonRequest); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validar muestras de entrada
	if len(comparisonRequest.Sample1) == 0 || len(comparisonRequest.Sample2) == 0 {
		http.Error(w, "Muestras vacías", http.StatusBadRequest)
		return
	}

	// Implementar lógica de comparación
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
	confidence := similarityScore * 0.95 // Factor de ajuste

	// Preparar respuesta extendida
	response := models.ComparisonResult{
		Similarity:      similarityScore,
		Match:           match,
		Confidence:      confidence,
		FeatureWeights:  comparisonRequest.Weights,
		DiffPerFeature:  calculateFeatureDiffs(comparisonRequest.Sample1, comparisonRequest.Sample2),
		AreasOfInterest: identifyCriticalDifferences(comparisonRequest.Sample1, comparisonRequest.Sample2),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
	}
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
		return 0.0
	}

	var totalWeight, weightedSum float64
	validFeatures := 0

	for feature, value1 := range f1 {
		value2, exists := f2[feature]
		if !exists {
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
		}

		diff := math.Abs(value1 - value2)
		normalizedDiff := diff / maxVal

		// Acumular similitud ponderada
		weightedSum += weight * (1 - normalizedDiff)
		totalWeight += weight
		validFeatures++
	}

	if validFeatures == 0 || totalWeight == 0 {
		return 0.0
	}

	return weightedSum / totalWeight
}
