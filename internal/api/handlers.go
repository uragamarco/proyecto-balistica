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
	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get image file
	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving image: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file content into memory for hash generation
	var imgBytes bytes.Buffer
	tee := io.TeeReader(file, &imgBytes)

	// Decode image from the TeeReader
	img, _, err := image.Decode(tee)
	if err != nil {
		http.Error(w, "Error decoding image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create temporary file for Python processing
	tempDir := h.imageProcessor.Config.TempDir
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	tempFile, err := os.CreateTemp(tempDir, "balistica_*.png")
	if err != nil {
		http.Error(w, "Error creating temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write image to temp file
	if _, err := tempFile.Write(imgBytes.Bytes()); err != nil {
		http.Error(w, "Error writing temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Process image
	processedImg, err := h.imageProcessor.Process(img)
	if err != nil {
		http.Error(w, "Error processing image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract features using the temporary file path
	features, err := h.imageProcessor.ExtractFeatures(processedImg, tempFile.Name())
	if err != nil {
		http.Error(w, "Error extracting features: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Analyze chroma
	chromaAnalysis, err := h.chromaService.Analyze(processedImg)
	if err != nil {
		http.Error(w, "Error analyzing chroma: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate image hash
	hash := models.GenerateImageHashFromBytes(imgBytes.Bytes())

	// Check Python status
	pyEnabled, pyStatus := h.imageProcessor.PythonFeaturesStatus()
	if pyStatus != "" {
		// Log Python status for debugging
		h.imageProcessor.Logger.Printf("Python features status: %s", pyStatus)
	}

	// Prepare response
	response := models.BallisticAnalysis{
		Features: features,
		ChromaData: models.ChromaAnalysis{
			DominantColors: convertColorData(chromaAnalysis.DominantColors),
			ColorVariance:  chromaAnalysis.ColorVariance,
		},
		Metadata: models.AnalysisMetadata{
			Timestamp:          time.Now().UTC().Format(time.RFC3339),
			ImageHash:          hash,
			ProcessorVersion:   "1.3.0",
			PythonFeaturesUsed: pyEnabled,
			Confidence:         calculateAnalysisConfidence(features), // Nueva función
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
	}
}

// calculateAnalysisConfidence estima confianza basada en características
func calculateAnalysisConfidence(features map[string]float64) float64 {
	// Variables para cálculo de confianza
	var (
		featureCount = len(features)
		qualityScore = 0.0
	)

	// Ponderar características clave
	if hu, ok := features["hu_moment_1"]; ok {
		qualityScore += math.Abs(hu) * 0.3
	}
	if area, ok := features["contour_area"]; ok && area > 0 {
		qualityScore += math.Log(area) * 0.2
	}
	if str, ok := features["striation_density"]; ok {
		qualityScore += str * 0.2
	}

	// Factor de completitud
	completeness := float64(featureCount) / 15.0 // Asumiendo 15 características esperadas

	// Combinar factores
	confidence := (qualityScore * 0.7) + (completeness * 0.3)

	// Limitar a rango 0-1
	return math.Max(0, math.Min(1, confidence))
}

func (h *Handlers) CompareSamples(w http.ResponseWriter, r *http.Request) {
	var comparisonRequest struct {
		Sample1   map[string]float64 `json:"sample1"`
		Sample2   map[string]float64 `json:"sample2"`
		Weights   map[string]float64 `json:"weights,omitempty"`
		Threshold float64            `json:"threshold,omitempty"` // Nuevo: umbral personalizado
	}

	err := json.NewDecoder(r.Body).Decode(&comparisonRequest)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Implement comparison logic
	similarityScore := h.compareFeatures(
		comparisonRequest.Sample1,
		comparisonRequest.Sample2,
		comparisonRequest.Weights,
	)

	// Determinar coincidencia
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
	json.NewEncoder(w).Encode(response)
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

// identifyCriticalDifferences identifica áreas con mayores diferencias
func identifyCriticalDifferences(s1, s2 map[string]float64) []string {
	var critical []string
	threshold := 0.2 // Diferencia significativa

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

		// Calcular peso (valor por defecto 1.0)
		weight := 1.0
		if w, ok := weights[feature]; ok {
			weight = w
		}

		// Calcular diferencia normalizada
		maxVal := math.Max(math.Abs(value1), math.Abs(value2))
		if maxVal == 0 {
			maxVal = 1 // Evitar división por cero
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
