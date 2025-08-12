package api

import (
	"encoding/json"
	"image"
	"math"
	"net/http"

	"github.com/uragamarco/proyecto-balistica/internal/models"
	"github.com/uragamarco/proyecto-balistica/internal/services/chroma"
	imgproc "github.com/uragamarco/proyecto-balistica/internal/services/image_processor"
)

type Handlers struct {
	imageProcessor *imgproc.ImageProcessor
	chromaService  *chroma.Service
}

// Constructor corregido
func NewHandlers(ip *imgproc.ImageProcessor, cs *chroma.Service) *Handlers {
	return &Handlers{
		imageProcessor: ip,
		chromaService:  cs,
	}
}

// Función convertColorData corregida
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
	}package api

import (
	"encoding/json"
	"image"
	"math"
	"net/http"
	"path/filepath"
	"strings"

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

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Error decoding image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Process image
	processedImg, err := h.imageProcessor.Process(img)
	if err != nil {
		http.Error(w, "Error processing image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract features using the original file path for Python processing
	originalPath := filepath.Join("/tmp", handler.Filename) // Temporary path simulation
	features, err := h.imageProcessor.ExtractFeatures(processedImg, originalPath)
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

	// Prepare response
	response := models.BallisticAnalysis{
		Features: features, // Now a map[string]float64
		ChromaData: models.ChromaAnalysis{
			DominantColors: convertColorData(chromaAnalysis.DominantColors),
			ColorVariance:  chromaAnalysis.ColorVariance,
		},
		// ProcessedImage: processedImg, // Removed as it's not JSON serializable
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) CompareSamples(w http.ResponseWriter, r *http.Request) {
	var comparisonRequest struct {
		Sample1 map[string]float64 `json:"sample1"`
		Sample2 map[string]float64 `json:"sample2"`
		Weights map[string]float64 `json:"weights,omitempty"` // New: Feature weights
	}

	err := json.NewDecoder(r.Body).Decode(&comparisonRequest)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Implement comparison logic
	similarityScore := h.compareFeatures(comparisonRequest.Sample1, comparisonRequest.Sample2, comparisonRequest.Weights)

	response := map[string]float64{
		"similarity": similarityScore,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Updated to work with feature maps and weights
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

		// Calculate feature weight (default to 1.0 if not specified)
		weight := 1.0
		if w, ok := weights[feature]; ok {
			weight = w
		}

		// Calculate normalized difference (0-1 range)
		diff := math.Abs(value1 - value2)
		normalizedDiff := diff / (1 + math.Max(value1, value2)) // Adaptive normalization
		
		// Accumulate weighted similarity
		weightedSum += weight * (1 - normalizedDiff)
		totalWeight += weight
		validFeatures++
	}

	if validFeatures == 0 || totalWeight == 0 {
		return 0.0
	}

	return weightedSum / totalWeight
}
	return result
}

func (h *Handlers) ProcessImage(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Get image file
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving image", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Error decoding image", http.StatusInternalServerError)
		return
	}

	// Process image
	processedImg, err := h.imageProcessor.Process(img)
	if err != nil {
		http.Error(w, "Error processing image", http.StatusInternalServerError)
		return
	}

	// Extract features
	features, err := h.imageProcessor.ExtractFeatures(processedImg)
	if err != nil {
		http.Error(w, "Error extracting features", http.StatusInternalServerError)
		return
	}

	// Analyze chroma
	chromaAnalysis, err := h.chromaService.Analyze(processedImg)
	if err != nil {
		http.Error(w, "Error analyzing chroma", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := models.BallisticAnalysis{
		Features: features,
		ChromaData: models.ChromaAnalysis{
			DominantColors: convertColorData(chromaAnalysis.DominantColors),
			ColorVariance:  chromaAnalysis.ColorVariance,
		},
		ProcessedImage: processedImg,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) CompareSamples(w http.ResponseWriter, r *http.Request) {
	var comparisonRequest struct {
		Sample1 []float64 `json:"sample1"`
		Sample2 []float64 `json:"sample2"`
	}

	err := json.NewDecoder(r.Body).Decode(&comparisonRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Implement comparison logic
	similarityScore := h.compareFeatures(comparisonRequest.Sample1, comparisonRequest.Sample2)

	response := map[string]float64{
		"similarity": similarityScore,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) compareFeatures(f1, f2 []float64) float64 {
	if len(f1) != len(f2) || len(f1) == 0 {
		return 0.0
	}

	var sum float64
	for i := range f1 {
		diff := f1[i] - f2[i]
		sum += diff * diff
	}

	distance := math.Sqrt(sum)
	maxPossible := math.Sqrt(float64(len(f1)) * 255)
	similarity := 1 - (distance / maxPossible)

	return math.Max(0, math.Min(1, similarity)) // Clamp between 0 and 1
}
