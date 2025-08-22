package config

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(level string, output string) (*zap.Logger, error) {
	// Convertir nivel de string a zapcore.Level
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, err
	}

	// Configuración de producción
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.MessageKey = "mensaje"
	config.EncoderConfig.LevelKey = "nivel"

	// Configurar salida
	if output == "file" {
		config.OutputPaths = []string{"logs/app.log"}
		config.ErrorOutputPaths = []string{"logs/errors.log"}
	} else {
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}
	}

	return config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.ErrorLevel),
	)
}