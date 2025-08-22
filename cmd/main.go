package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/uragamarco/proyecto-balistica/internal/app"
	"github.com/uragamarco/proyecto-balistica/internal/config"

	"go.uber.org/zap"
)

func main() {
	// Cargar configuración principal
	cfg, err := config.Load("configs/default.yml")
	if err != nil {
		panic("Error cargando configuración: " + err.Error())
	}

	// Inicializar logger con Zap
	logger, err := config.NewLogger(cfg.Logging.Level, cfg.Logging.Output)
	if err != nil {
		panic("Error inicializando logger: " + err.Error())
	}
	defer func() {
		// Asegurar que todos los logs se vacíen antes de salir
		_ = logger.Sync()
	}()

	// Recuperar panics y registrarlos
	defer func() {
		if r := recover(); r != nil {
			logger.Fatal("Panic recuperado",
				zap.Any("razón", r),
				zap.Stack("stack"))
		}
	}()

	logger.Info("Iniciando aplicación balística",
		zap.String("versión", cfg.App.Version),
		zap.String("entorno", cfg.App.Environment))

	// Crear aplicación con inyección de logger
	balisticaApp, err := app.NewApp(cfg, logger)
	if err != nil {
		logger.Fatal("Error inicializando aplicación", zap.Error(err))
	}

	// Canal para manejar señales de sistema
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ejecutar aplicación en goroutine
	go func() {
		if err := balisticaApp.Run(); err != nil {
			logger.Fatal("Error ejecutando aplicación", zap.Error(err))
		}
	}()

	logger.Info("Aplicación en ejecución", zap.String("puerto", cfg.Server.Port))

	// Esperar señal de apagado
	sig := <-sigChan
	logger.Info("Recibida señal de apagado", zap.String("señal", sig.String()))

	// Apagado controlado
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := balisticaApp.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error en apagado controlado", zap.Error(err))
	}

	logger.Info("Aplicación detenida correctamente")
}
