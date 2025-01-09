package env

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ZapLogger *zap.Logger

// type LeveledZap struct {
// 	*zap.Logger
// }
// func (z *LeveledZap) Error(msg string, keysAndValues ...interface{}) {
// 	l.WithFields(keysAndValues).Error(msg)
// }

// func (z *LeveledZap) Info(msg string, keysAndValues ...interface{}) {
// 	l.WithFields(keysAndValues).Info(msg)
// }
// func (z *LeveledZap) Debug(msg string, keysAndValues ...interface{}) {
// 	l.WithFields(keysAndValues).Debug(msg)
// }

// func (z *LeveledZap) Warn(msg string, keysAndValues ...interface{}) {
// 	l.WithFields(keysAndValues).Warn(msg)
// }

func createLogger(this_service string) *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	// level := strings.ToLower(os.Getenv("DEBUG_LEVEL"))
	// WARNING: THIS MUST RUN AFTER THE ENV IS PARSED/SET UP
	s, _ := Env.GetUserService(this_service)
	level := s.GetParamString("debug_level")

	var zap_level zapcore.Level
	switch level {
	case "debug":
		zap_level = zap.DebugLevel
	case "info":
		zap_level = zap.InfoLevel
	case "warn":
		zap_level = zap.WarnLevel
	case "error":
		zap_level = zap.ErrorLevel
	default:
		zap_level = zap.InfoLevel
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap_level),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
		InitialFields: map[string]interface{}{
			"pid": os.Getpid(),
		},
	}

	logger, err := config.Build()
	if err != nil {
		log.Fatal("cannot build zap logger from config")
	}
	return zap.Must(logger, nil)
}

func SetupLogging(this_service string) {
	ZapLogger = createLogger(this_service)
	zap.ReplaceGlobals(zap.Must(ZapLogger, nil))
}
