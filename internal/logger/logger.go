package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger = zap.NewNop().Sugar()

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl.Sugar()
	return nil
}
