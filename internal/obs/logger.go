package obs

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a configured zap logger based on the provided config.
func NewLogger(cfg Config) (*zap.Logger, error) {
	var zapCfg zap.Config
	var lvl zapcore.Level
	if cfg.Env == "prod" {
		zapCfg = zap.NewProductionConfig()
		zapCfg.Sampling = &zap.SamplingConfig{Initial: 100, Thereafter: 100}

	} else {
		zapCfg = zap.NewDevelopmentConfig()
	}

	if cfg.LogJSON {
		zapCfg.Encoding = "json"
	} else {
		zapCfg.Encoding = "console"
	}

	if err := lvl.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		lvl = zapcore.InfoLevel
	}

	zapCfg.Level = zap.NewAtomicLevelAt(lvl)

	zapCfg.EncoderConfig.TimeKey = "ts"
	zapCfg.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder

	zapCfg.EncoderConfig.LevelKey = "level"
	zapCfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	zapCfg.EncoderConfig.MessageKey = "msg"
	zapCfg.EncoderConfig.NameKey = "logger"

	zapCfg.EncoderConfig.CallerKey = "caller"
	zapCfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	l, err := zapCfg.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	l.With(
		zap.String("service.name", cfg.ServiceName),
		zap.String("service.version", cfg.ServiceVersion),
		zap.String("deployment.environment", cfg.Env),
	)

	return l, nil
}
