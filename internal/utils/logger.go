package utils

import (
	"github.com/sirupsen/logrus"
)

type LogFields map[string]interface{}

type Logger interface {
	WithFields(fields LogFields) Logger
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type logger struct {
	entry *logrus.Entry
}

func NewLogger() Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.999Z07:00",
	})
	l.SetLevel(logrus.DebugLevel)

	return &logger{
		entry: logrus.NewEntry(l),
	}
}

func (l *logger) WithFields(fields LogFields) Logger {
	return &logger{
		entry: l.entry.WithFields(logrus.Fields(fields)),
	}
}

func (l *logger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

func (l *logger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

func (l *logger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

func (l *logger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}
