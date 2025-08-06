package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Address string
		Timeout struct {
			Read  time.Duration
			Write time.Duration
			Idle  time.Duration
		}
	}
	Database struct {
		ChromaURL string
	}
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error leyendo configuración: %w", err)
	}

	var cfg Config

	// Configuración del servidor
	cfg.Server.Address = viper.GetString("SERVER_ADDRESS")
	cfg.Server.Timeout.Read = viper.GetDuration("SERVER_TIMEOUT_READ")
	cfg.Server.Timeout.Write = viper.GetDuration("SERVER_TIMEOUT_WRITE")
	cfg.Server.Timeout.Idle = viper.GetDuration("SERVER_TIMEOUT_IDLE")

	// Configuración de ChromaDB
	cfg.Database.ChromaURL = viper.GetString("CHROMA_URL")

	return &cfg, nil
}
