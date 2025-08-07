package prodjlink

import (
	"github.com/inconshreveable/log15"

	"github.com/autonomouskoi/akcore/svc/log"
)

var _ log15.Logger = pdjLogger{}

type pdjLogger struct {
	log.Logger
}

func (l pdjLogger) Crit(msg string, ctx ...interface{}) {
	l.Error(msg, ctx...)
}

func (l pdjLogger) GetHandler() log15.Handler {
	panic("not implemented")
}

func (l pdjLogger) SetHandler(_ log15.Handler) {
	panic("not implemented")
}

func (l pdjLogger) New(ctx ...interface{}) log15.Logger {
	return pdjLogger{
		Logger: l,
	}
}

func (l pdjLogger) Debug(msg string, ctx ...interface{}) {
	if msg == "Sending packet" {
		return
	}
	l.Logger.Debug(msg, ctx...)
}
