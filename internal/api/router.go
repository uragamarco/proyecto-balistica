package api

import (
	"net/http"
	"strings"

	"github.com/uragamarco/proyecto-balistica/internal/handlers"
)

func NewRouter(h *Handlers) *http.ServeMux {
	mux := http.NewServeMux()

	// Existing endpoints
	mux.HandleFunc("/api/process", h.ProcessImage)
	mux.HandleFunc("/api/compare", h.CompareSamples)

	// Storage endpoints for analyses
	mux.HandleFunc("/api/analyses", h.GetAnalysesHandler)
	mux.HandleFunc("/api/analyses/", h.GetAnalysisHandler)
	mux.HandleFunc("/api/analyses/search", h.SearchAnalysesHandler)
	mux.HandleFunc("/api/analyses/delete/", h.DeleteAnalysisHandler)

	// Storage endpoints for comparisons
	mux.HandleFunc("/api/comparisons/", h.GetComparisonHandler)
	mux.HandleFunc("/api/comparisons/sample/", h.GetComparisonsBySampleHandler)
	mux.HandleFunc("/api/comparisons/similarity", h.GetSimilarComparisonsHandler)
	mux.HandleFunc("/api/comparisons/date-range", h.GetComparisonsByDateRangeHandler)

	// Advanced search endpoint
	mux.HandleFunc("/api/search/advanced", h.AdvancedSearchHandler)

	// Crear handler de clasificación
	classificationHandler := handlers.NewClassificationHandler(h.classificationService, h.Logger)

	// Classification endpoints
	mux.HandleFunc("/api/classification/classify", classificationHandler.ClassifyBallistic)
	mux.HandleFunc("/api/classification/history", classificationHandler.GetClassificationHistory)
	mux.HandleFunc("/api/classification/analysis/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/classification/analysis/") {
			classificationHandler.GetClassificationByAnalysisID(w, r)
		}
	})
	mux.HandleFunc("/api/classification/search/weapon/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/classification/search/weapon/") {
			classificationHandler.SearchByWeaponType(w, r)
		}
	})
	mux.HandleFunc("/api/classification/search/caliber/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/classification/search/caliber/") {
			classificationHandler.SearchByCaliber(w, r)
		}
	})

	// Health endpoint
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"ballistic-analysis-api"}`))
	})

	return mux
}
