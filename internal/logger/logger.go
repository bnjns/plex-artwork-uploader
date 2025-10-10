package logger

import (
	"github.com/sirupsen/logrus"
)

func New(verbosity int) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		DisableLevelTruncation: true,
		PadLevelText:           true,
	})

	switch verbosity {
	case 2:
		logger.SetLevel(logrus.TraceLevel)
	case 1:
		logger.SetLevel(logrus.DebugLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	return logger
}
