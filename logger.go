package main

import (
	"code.cloudfoundry.org/lager"
	"go.uber.org/zap"
)

// Ensure LagerZapLogger adheres to the Logger interface.
var _ lager.Logger = &LagerZapLogger{}

// LagerZapLogger is implementing the Lager interface. The OSBAPI expects us to use the lager logger,
// but we wanted to use the Zap logger for its fast, leveled, and structured logging.
// The zap methods are wrapped in the Lager method calls and is merely mapping them.
type LagerZapLogger struct {
	logger *zap.SugaredLogger
}

// NewLagerZapLogger constructs and returns a new LagerZapLogger with a pointer field of type SugaredLogger.
func NewLagerZapLogger(zap *zap.SugaredLogger) *LagerZapLogger {
	return &LagerZapLogger{
		logger: zap,
	}
}

// RegisterSink represents a write destination for a Logger. It provides a thread-safe interface for writing logs.
func (lagerZapLogger *LagerZapLogger) RegisterSink(sink lager.Sink) {}

// SessionName returns the name of the session. This is normally added when initializing a new logger but zap doesn't require nor need it.
// but the OSBAPI does. Currently it's only returning an empty string.
func (lagerZapLogger *LagerZapLogger) SessionName() string {
	return ""
}

// Session sets the session of the logger and returns a new logger with a nested session. We are currently
// returning the same logger back.
func (lagerZapLogger *LagerZapLogger) Session(task string, data ...lager.Data) lager.Logger {
	return lagerZapLogger
}

// WithData creates a new child with the parent fields and returns a logger with newly added data. We are currently
// returning the same logger back.
func (lagerZapLogger *LagerZapLogger) WithData(data lager.Data) lager.Logger {
	return lagerZapLogger
}

// Debug logs a message at debug level. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (lagerZapLogger *LagerZapLogger) Debug(action string, data ...lager.Data) {
	lagerZapLogger.logger.Debugw(action, createFields(data)...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (lagerZapLogger *LagerZapLogger) Info(action string, data ...lager.Data) {
	lagerZapLogger.logger.Infow(action, createFields(data)...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (lagerZapLogger *LagerZapLogger) Error(action string, err error, data ...lager.Data) {
	lagerZapLogger.logger.Errorw(action, createFields(data)...)
}

// Fatal is for logging fatal messages.
// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (lagerZapLogger *LagerZapLogger) Fatal(action string, err error, data ...lager.Data) {
	lagerZapLogger.logger.Fatalw(action, createFields(data)...)
}

// createFields converts the structured log data that the lager library uses to
// the format that zap expects. lager uses a list of maps while zap expects a flat list
// of alternating keys and values.
func createFields(data []lager.Data) []interface{} {
	var fields []interface{}

	// Copying items from data and appending them to fields
	for _, item := range data {
		for k, v := range item {
			fields = append(fields, k, v)
		}
	}
	return fields
}
