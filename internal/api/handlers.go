package api

import (
	"encoding/json"
	"image"
	"math"
	"net/http"

	"github.com/uragamarco/proyecto-balistica/internal/models"
	"github.com/uragamarco/proyecto-balistica/internal/services/chroma"
	"github.com/uragamarco/proyecto-balistica/internal/services/image_processor"
)

type Handlers struct {
	imageProcessor *image_processor.ImageProcessor
	chromaService  *chroma.Service
}

func NewHandlers(ip *image_processor.ImageProcessor, cs *chroma.Service) *Handlers {
	return &Handlers{
		imageProcessor: ip,
		chromaService:  cs,
	}
}

func convertColorData(cd []chroma.ColorData) []models.ColorData {
	result := make([]models.ColorData, len(cd))
	for i, c := range cd {
		result[i] = models.ColorData{
			Color: models.RGB{
				R: c.Color.R,
				G: c.Color.G,
				B: c.Color.B,
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
