package log

import (
	"io"

	"github.com/labstack/gommon/log"
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

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args)
}

func SetLevel(level string) {
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
		log.Warnf("because the provided log level (%s) was invalid, log level was set to info", level)
	}
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
