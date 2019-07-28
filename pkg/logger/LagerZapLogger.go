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
	component     string
	task          string
	sessionID     string
	nextSession   uint32
}

//NewLagerZapLogger constructor
func NewLagerZapLogger(zap *zap.SugaredLogger, component string) *LagerZapLogger {
	return &LagerZapLogger{
		sugaredLogger: zap,
		component:     component,
		task:          component,
	}
}

//GetSugaredLogger returns the pointer of the sugaredLogger in the LagerZapLogger class
func (lagerZapLogger *LagerZapLogger) GetSugaredLogger() *zap.SugaredLogger {
	return lagerZapLogger.sugaredLogger
}

//RegisterSink is not used currently
func (lagerZapLogger *LagerZapLogger) RegisterSink(sink lager.Sink) {
	panic("RegisterSink(sink lager.Sink) not implemented")
}

//SessionName not used currently
func (lagerZapLogger *LagerZapLogger) SessionName() string {
	return lagerZapLogger.task
}

//Session of the logger
func (lagerZapLogger *LagerZapLogger) Session(task string, data ...lager.Data) lager.Logger {

	sid := atomic.AddUint32(&lagerZapLogger.nextSession, 1)

	var sessionIDstr string

	if lagerZapLogger.sessionID != "" {
		sessionIDstr = fmt.Sprintf("%s.%d", lagerZapLogger.sessionID, sid)
	} else {
		sessionIDstr = fmt.Sprintf("%d", sid)
	}

	return &LagerZapLogger{
		component: lagerZapLogger.component,
		task:      fmt.Sprintf("%s.%s", lagerZapLogger.task, task),
		sessionID: sessionIDstr,
	}
	//return lager.NewLogger("my-app")
}

//WithData creates a new child with the parent fields
func (lagerZapLogger *LagerZapLogger) WithData(data lager.Data) lager.Logger {
	return lagerZapLogger.With(data)
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

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func (lagerZapLogger *LagerZapLogger) With(data ...lager.Data) lager.Logger {
	return &LagerZapLogger{
		component: lagerZapLogger.component,
		task:      lagerZapLogger.task,
		sessionID: lagerZapLogger.sessionID,
	}
}
