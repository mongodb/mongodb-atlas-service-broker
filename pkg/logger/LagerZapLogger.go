package logger

import (
	"code.cloudfoundry.org/lager"
	"go.uber.org/zap"
)

// Ensure LagerZapLogger adheres to the Logger interface.
var _ lager.Logger = &LagerZapLogger{}

//LagerZapLogger STRUCT
type LagerZapLogger struct {
	SugaredLogger *zap.SugaredLogger
}

//NewLagerZapLogger constructor
func NewLagerZapLogger(zap *zap.SugaredLogger) *LagerZapLogger {
	return &LagerZapLogger{
		SugaredLogger: zap,
	}
}

//RegisterSink is not used currently
func (lagerZapLogger *LagerZapLogger) RegisterSink(sink lager.Sink) {
	//
}

//SessionName not used currently
func (lagerZapLogger *LagerZapLogger) SessionName() string {
	return ""
}

//Session not used currently
func (lagerZapLogger *LagerZapLogger) Session(task string, data ...lager.Data) lager.Logger {
	return nil
}

//WithData is not used currently
func (lagerZapLogger *LagerZapLogger) WithData(data lager.Data) lager.Logger {
	return nil
}

//Debug has verbose message
func (lagerZapLogger *LagerZapLogger) Debug(action string, data ...lager.Data) {
	lagerZapLogger.SugaredLogger.Debug(action, nil)
}

//Info is default log level
func (lagerZapLogger *LagerZapLogger) Info(action string, data ...lager.Data) {
	lagerZapLogger.SugaredLogger.Info(action, nil)
}

//Error is for logging errors
func (lagerZapLogger *LagerZapLogger) Error(action string, err error, data ...lager.Data) {
	lagerZapLogger.SugaredLogger.Error(action, nil)
}

//Fatal is for logging fatal messages. The system shutdowns after logging the message
func (lagerZapLogger *LagerZapLogger) Fatal(action string, err error, data ...lager.Data) {
	lagerZapLogger.SugaredLogger.Fatal(action, nil)
}
