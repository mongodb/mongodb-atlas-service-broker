package logger

import (
	"code.cloudfoundry.org/lager"
	"go.uber.org/zap"
)

// Ensure LagerZapLogger adheres to the Logger interface.
var _ lager.Logger = &LagerZapLogger{}

//LagerZapLogger STRUCT
type LagerZapLogger struct {
	sugaredLogger *zap.SugaredLogger
}

//NewLagerZapLogger constructor
func NewLagerZapLogger(zap *zap.SugaredLogger) *LagerZapLogger {
	return &LagerZapLogger{
		sugaredLogger: zap,
	}
}

//GetSugaredLogger returns the pointer of the sugaredLogger in the LagerZapLogger class
func (lagerZapLogger *LagerZapLogger) GetSugaredLogger() *zap.SugaredLogger {
	return lagerZapLogger.sugaredLogger
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
	lagerZapLogger.sugaredLogger.Debugw(action, nil)
}

//Info is default log level
func (lagerZapLogger *LagerZapLogger) Info(action string, data ...lager.Data) {
	lagerZapLogger.sugaredLogger.Infow(action, nil)
}

//Error is for logging errors
func (lagerZapLogger *LagerZapLogger) Error(action string, err error, data ...lager.Data) {
	lagerZapLogger.sugaredLogger.Errorw(action, nil)
}

//Fatal is for logging fatal messages. The system shutdowns after logging the message
func (lagerZapLogger *LagerZapLogger) Fatal(action string, err error, data ...lager.Data) {
	lagerZapLogger.sugaredLogger.Fatalw(action, nil)
}
