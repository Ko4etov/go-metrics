package logger

import (
	"go.uber.org/zap"
)

var Logger zap.SugaredLogger

func Initialize(level string) error {

	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()

	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Logger = *zl.Sugar()

	return nil
}
