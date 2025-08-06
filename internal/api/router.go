package api

import (
	"net/http"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Configuración de endpoints
	mux.HandleFunc("/api/upload", uploadHandler)
	mux.HandleFunc("/api/login", loginHandler)

	// Puedes agregar middlewares aquí si lo necesitas
	return mux
}
