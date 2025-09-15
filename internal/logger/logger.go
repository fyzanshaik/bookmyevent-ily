package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Entry
}

func New(level string) *Logger {
	logger := logrus.New()

	logLevel, err := logrus.ParseLevel(strings.ToLower(level))
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	if isProduction() {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	return &Logger{Entry: logrus.NewEntry(logger)}
}

func (l *Logger) WithService(serviceName string) *Logger {
	return &Logger{Entry: l.Entry.WithField("service", serviceName)}
}

func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{Entry: l.Entry.WithField("request_id", requestID)}
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	return &Logger{Entry: l.Entry.WithFields(logrus.Fields(fields))}
}

func isProduction() bool {
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	return env == "production" || env == "prod"
}