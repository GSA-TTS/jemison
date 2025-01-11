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

func createLogger(thisService string) *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	// level := strings.ToLower(os.Getenv("DEBUG_LEVEL"))
	// WARNING: THIS MUST RUN AFTER THE ENV IS PARSED/SET UP
	s, _ := Env.GetUserService(thisService)
	level := s.GetParamString("debug_level")

	var zapLevel zapcore.Level

	switch level {
	case "debug":
		zapLevel = zap.DebugLevel
	case "info":
		zapLevel = zap.InfoLevel
	case "warn":
		zapLevel = zap.WarnLevel
	case "error":
		zapLevel = zap.ErrorLevel
	default:
		zapLevel = zap.InfoLevel
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
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

func SetupLogging(thisService string) {
	ZapLogger = createLogger(thisService)
	zap.ReplaceGlobals(zap.Must(ZapLogger, nil))
}
