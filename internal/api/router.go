package api

import "net/http"

func NewRouter(handlers *Handlers) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/process", handlers.ProcessImage)
	mux.HandleFunc("/api/compare", handlers.CompareSamples)

	return mux
}
