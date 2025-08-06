package main

import (
	"log"
	"net/http"

	"github.com/uragamarco/proyecto-balistica/internal/api"
	"github.com/uragamarco/proyecto-balistica/internal/config"
	"github.com/uragamarco/proyecto-balistica/internal/services/chroma"
)

func main() {
	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}

	// Inicializar ChromaDB
	err = chroma.Init(cfg.ChromaURL, cfg.CollectionName)
	if err != nil {
		log.Fatalf("Error inicializando ChromaDB: %v", err)
	}

	// Configurar rutas
	mux := http.NewServeMux()
	mux.HandleFunc("/api/upload", api.UploadHandler)
	mux.Handle("/", http.FileServer(http.Dir("./static")) 

	// Iniciar servidor
	log.Printf("Servidor iniciado en %s", cfg.ServerAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, mux))
}