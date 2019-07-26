package logger

import (
	"fmt"
	"sync/atomic"

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
	panic("RegisterSink not implemented")
}

//SessionName not used currently
func (lagerZapLogger *LagerZapLogger) SessionName() string {
	panic("SessionName not implemented")
}

//Session not used currently
func (lagerZapLogger *LagerZapLogger) Session(task string, data ...lager.Data) lager.Logger {
	sid := atomic.AddUint32(&lagerZapLogger.nextSession, 1)

	var sessionIDstr string

	if l.sessionID != "" {
		sessionIDstr = fmt.Sprintf("%s.%d", l.sessionID, sid)
	} else {
		sessionIDstr = fmt.Sprintf("%d", sid)
	}

	return &logger{
		component: l.component,
		task:      fmt.Sprintf("%s.%s", l.task, task),
		sinks:     l.sinks,
		sessionID: sessionIDstr,
		data:      l.baseData(data...),
	}
}

//WithData is not used currently
func (lagerZapLogger *LagerZapLogger) WithData(data lager.Data) lager.Logger {
	panic("WithData not implemented")
}

//Debug has verbose message
func (lagerZapLogger *LagerZapLogger) Debug(action string, data ...lager.Data) {
	lagerZapLogger.sugaredLogger.Debugw(action, data)
}

//Info is default log level
func (lagerZapLogger *LagerZapLogger) Info(action string, data ...lager.Data) {
	lagerZapLogger.sugaredLogger.Infow(action, data)
}

//Error is for logging errors
func (lagerZapLogger *LagerZapLogger) Error(action string, err error, data ...lager.Data) {
	lagerZapLogger.sugaredLogger.Errorw(action, data)
}

//Fatal is for logging fatal messages. The system shutdowns after logging the message
func (lagerZapLogger *LagerZapLogger) Fatal(action string, err error, data ...lager.Data) {
	lagerZapLogger.sugaredLogger.Fatalw(action, data)
}
