package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

const (
	envDev  = "dev"
	envProd = "prod"
)

func SetupLogger(env string) *logrus.Logger {
	var log *logrus.Logger

	switch env {
	case envDev:
		log = &logrus.Logger{
			Out:       os.Stdout,
			Formatter: &logrus.TextFormatter{FullTimestamp: true},
			Level:     logrus.DebugLevel,
		}
	case envProd:
		log = &logrus.Logger{
			Out:       os.Stdout,
			Formatter: &logrus.JSONFormatter{},
			Level:     logrus.InfoLevel,
		}
	}
	return log
}
