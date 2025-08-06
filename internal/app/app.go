package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/uragamarco/proyecto-balistica/internal/api"
	"github.com/uragamarco/proyecto-balistica/internal/config"
	"github.com/uragamarco/proyecto-balistica/internal/services/chroma"
	"github.com/uragamarco/proyecto-balistica/internal/services/image_processor"
)

type Application struct {
	server *http.Server
	logger zerolog.Logger
	config *config.Config
}

func New(cfg *config.Config, imgProcessor *image_processor.ImageProcessor, chromaSvc *chroma.Service) (*Application, error) {
	app := &Application{
		config: cfg,
		logger: zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}

	// Inicializar handlers
	handlers := api.NewHandlers(imgProcessor, chromaSvc)

	// Configurar el enrutador con los handlers
	router := api.NewRouter(handlers)

	app.server = &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  cfg.Server.Timeout.Read,
		WriteTimeout: cfg.Server.Timeout.Write,
		IdleTimeout:  cfg.Server.Timeout.Idle,
	}

	return app, nil
}

func (a *Application) Run() error {
	a.logger.Info().Msgf("Servidor iniciado en %s", a.server.Addr)

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal().Err(err).Msg("Error al iniciar el servidor")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	a.logger.Info().Msg("Apagando servidor...")
	return a.server.Shutdown(ctx)
}
