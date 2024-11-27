package libs

import (
	"fmt"
	"os"
	"time"

	"github.com/nekomeowww/xo"
	"github.com/nekomeowww/xo/logger"
	"github.com/nekomeowww/xo/logger/loki"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/lingticio/llmg/internal/configs"
	"github.com/lingticio/llmg/internal/meta"
)

func NewLogger() func(config *configs.Config) (*logger.Logger, error) {
	return func(config *configs.Config) (*logger.Logger, error) {
		logLevel, err := logger.ReadLogLevelFromEnv()
		if err != nil {
			logLevel = zapcore.InfoLevel
		}

		var isFatalLevel bool
		if logLevel == zapcore.FatalLevel {
			isFatalLevel = true
			logLevel = zapcore.InfoLevel
		}

		logFormat, readFormatError := logger.ReadLogFormatFromEnv()

		logger, err := logger.NewLogger(
			logger.WithLevel(logLevel),
			logger.WithAppName(config.Meta.App),
			logger.WithNamespace(config.Meta.Namespace),
			logger.WithLogFilePath(xo.RelativePathBasedOnPwdOf("./logs/"+config.Meta.App)),
			logger.WithFormat(logFormat),
			logger.WithLokiRemoteConfig(lo.Ternary(os.Getenv("LOG_LOKI_REMOTE_URL") != "", &loki.Config{
				Url:          os.Getenv("LOG_LOKI_REMOTE_URL"),
				BatchMaxSize: 2000,             //nolint:mnd
				BatchMaxWait: 10 * time.Second, //nolint:mnd
			}, nil)),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create logger: %w", err)
		}
		if isFatalLevel {
			logger.Error("fatal log level is unacceptable, fallbacks to info level")
		}
		if readFormatError != nil {
			logger.Error("failed to read log format from env, fallbacks to json")
		}

		logger = logger.WithAndSkip(
			1,
			zap.String("commit", meta.LastCommit),
			zap.String("version", meta.Version),
			zap.String("env", meta.Env),
		)

		return logger, nil
	}
}
