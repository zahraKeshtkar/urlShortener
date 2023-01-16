package log

import (
	"io"

	"github.com/sirupsen/logrus"
)

var defaultLogger *logrus.Logger

func InitLogger() {
	defaultLogger = logrus.New()
}

func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

func Trace(args ...interface{}) {
	defaultLogger.Trace(args...)
}

func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args)
}

func SetLevel(level logrus.Level) {
	defaultLogger.SetLevel(level)
}

func WithFields(keysAndValues logrus.Fields) *logrus.Entry {
	return defaultLogger.WithFields(keysAndValues)
}

func SetOutput(w io.Writer) {
	defaultLogger.SetOutput(w)
}

func SetFormat(formatter logrus.Formatter) {
	defaultLogger.SetFormatter(formatter)
}
