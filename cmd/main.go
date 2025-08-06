package main

import (
	"log"
	"net/http"

	"github.com/uragamarco/proyecto-balistica/internal/api"
	"github.com/uragamarco/proyecto-balistica/internal/config"
	"github.com/uragamarco/proyecto-balistica/internal/services/chroma"
	"github.com/uragamarco/proyecto-balistica/internal/services/image_processor"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("./configs/default.yml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize services
	imgProcCfg := &image_processor.Config{
		Contrast:      cfg.Imaging.Contrast,
		SharpenSigma:  cfg.Imaging.SharpenSigma,
		SharpenAmount: cfg.Imaging.SharpenAmount,
		EdgeThreshold: cfg.Imaging.EdgeThreshold,
	}

	imgProcessor := image_processor.NewImageProcessor(imgProcCfg)

	chromaCfg := &chroma.Config{
		ColorThreshold: cfg.Chroma.ColorThreshold,
		SampleSize:     cfg.Chroma.SampleSize,
	}

	chromaSvc := chroma.NewService(chromaCfg)

	// Initialize API handlers
	handlers := api.NewHandlers(imgProcessor, chromaSvc)

	// Create and start server
	server := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: api.NewRouter(handlers),
	}

	log.Printf("Starting server on %s", cfg.Server.Address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
