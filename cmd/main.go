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

// AppInterface define la interfaz común para ambas aplicaciones
type AppInterface interface {
	Run() error
	Shutdown(ctx context.Context) error
}

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

	// Determinar si usar cache basado en configuración
	cacheEnabled := cfg.Cache != nil && cfg.Cache.Enabled

	if cacheEnabled {
		logger.Info("Iniciando aplicación balística con cache",
			zap.String("versión", cfg.App.Version),
			zap.String("entorno", cfg.App.Environment),
			zap.Bool("cache_enabled", true))
	} else {
		logger.Info("Iniciando aplicación balística",
			zap.String("versión", cfg.App.Version),
			zap.String("entorno", cfg.App.Environment),
			zap.Bool("cache_enabled", false))
	}

	// Crear aplicación apropiada basada en configuración de cache
	var balisticaApp AppInterface
	if cacheEnabled {
		appWithCache, err := app.NewAppWithCache(cfg, logger)
		if err != nil {
			logger.Fatal("Error inicializando aplicación con cache", zap.Error(err))
		}
		balisticaApp = appWithCache
	} else {
		appStandard, err := app.NewApp(cfg, logger)
		if err != nil {
			logger.Fatal("Error inicializando aplicación", zap.Error(err))
		}
		balisticaApp = appStandard
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

	// Log de información de ejecución específica según el tipo de aplicación
	if cacheEnabled {
		logger.Info("Aplicación con cache en ejecución",
			zap.String("puerto", cfg.Server.Port),
			zap.String("cache_dir", cfg.Cache.Directory),
			zap.Duration("memory_ttl", cfg.Cache.MemoryTTL),
			zap.Duration("disk_ttl", cfg.Cache.DiskTTL),
			zap.Int("max_memory_mb", cfg.Cache.MaxMemoryMB))
	} else {
		logger.Info("Aplicación en ejecución", zap.String("puerto", cfg.Server.Port))
	}

	// Esperar señal de apagado
	sig := <-sigChan
	logger.Info("Recibida señal de apagado", zap.String("señal", sig.String()))

	// Mostrar estadísticas de cache si está habilitado
	if cacheEnabled {
		if appWithCache, ok := balisticaApp.(*app.AppWithCache); ok {
			stats := appWithCache.GetCacheStats()
			logger.Info("Estadísticas de cache antes del apagado", zap.Any("stats", stats))
		}
	}

	// Apagado controlado
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := balisticaApp.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error en apagado controlado", zap.Error(err))
	}

	if cacheEnabled {
		logger.Info("Aplicación con cache detenida correctamente")
	} else {
		logger.Info("Aplicación detenida correctamente")
	}
}
