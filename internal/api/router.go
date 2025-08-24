package api

import (
	"net/http"
	"strings"

	"github.com/uragamarco/proyecto-balistica/internal/handlers"
)

func NewRouter(h *Handlers) *http.ServeMux {
	mux := http.NewServeMux()

	// Middleware CORS
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			// Manejar preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}

	// Servir archivos estáticos desde la carpeta web
	fs := http.FileServer(http.Dir("web/"))
	mux.Handle("/", fs)

	// Existing endpoints con CORS
	mux.Handle("/api/process", corsHandler(http.HandlerFunc(h.ProcessImage)))
	mux.Handle("/api/compare", corsHandler(http.HandlerFunc(h.CompareSamples)))

	// Storage endpoints for analyses con CORS
	mux.Handle("/api/analyses", corsHandler(http.HandlerFunc(h.GetAnalysesHandler)))
	mux.Handle("/api/analyses/", corsHandler(http.HandlerFunc(h.GetAnalysisHandler)))
	mux.Handle("/api/analyses/search", corsHandler(http.HandlerFunc(h.SearchAnalysesHandler)))
	mux.Handle("/api/analyses/delete/", corsHandler(http.HandlerFunc(h.DeleteAnalysisHandler)))

	// Storage endpoints for comparisons con CORS
	mux.Handle("/api/comparisons/", corsHandler(http.HandlerFunc(h.GetComparisonHandler)))
	mux.Handle("/api/comparisons/sample/", corsHandler(http.HandlerFunc(h.GetComparisonsBySampleHandler)))
	mux.Handle("/api/comparisons/similarity", corsHandler(http.HandlerFunc(h.GetSimilarComparisonsHandler)))
	mux.Handle("/api/comparisons/date-range", corsHandler(http.HandlerFunc(h.GetComparisonsByDateRangeHandler)))

	// Advanced search endpoint con CORS
	mux.Handle("/api/search/advanced", corsHandler(http.HandlerFunc(h.AdvancedSearchHandler)))

	// Crear handler de clasificación
	classificationHandler := handlers.NewClassificationHandler(h.classificationService, h.Logger)

	// Classification endpoints con CORS
	mux.Handle("/api/classification/classify", corsHandler(http.HandlerFunc(classificationHandler.ClassifyBallistic)))
	mux.Handle("/api/classification/history", corsHandler(http.HandlerFunc(classificationHandler.GetClassificationHistory)))
	mux.Handle("/api/classification/analysis/", corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/classification/analysis/") {
			classificationHandler.GetClassificationByAnalysisID(w, r)
		}
	})))
	mux.Handle("/api/classification/search/weapon/", corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/classification/search/weapon/") {
			classificationHandler.SearchByWeaponType(w, r)
		}
	})))
	mux.Handle("/api/classification/search/caliber/", corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/classification/search/caliber/") {
			classificationHandler.SearchByCaliber(w, r)
		}
	})))

	// Health endpoint con CORS
	mux.Handle("/api/health", corsHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"ballistic-analysis-api"}`))
	})))

	return mux
}
