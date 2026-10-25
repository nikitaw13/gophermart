package logger

import (
	"go.uber.org/zap"
)

// NewLogger creates a production zap logger at the given level.
func NewLogger(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zapLogger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return zapLogger, nil
}
