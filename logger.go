package rcon

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
}

type DefaultLogger struct{}

func (l *DefaultLogger) Info(...interface{})  {}
func (l *DefaultLogger) Error(...interface{}) {}
func (l *DefaultLogger) Debug(...interface{}) {}
