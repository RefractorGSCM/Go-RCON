package presets

import "log"

type DebugLogger struct{}

func (dl *DebugLogger) Info(args ...interface{}) {
	newArgs := []interface{}{"[INFO] "}
	newArgs = append(newArgs, args...)
	log.Print(newArgs...)
}

func (dl *DebugLogger) Error(args ...interface{}) {
	newArgs := []interface{}{"[ERROR] "}
	newArgs = append(newArgs, args...)
	log.Print(newArgs...)
}

func (dl *DebugLogger) Debug(args ...interface{}) {
	newArgs := []interface{}{"[DEBUG] "}
	newArgs = append(newArgs, args...)
	log.Print(newArgs...)
}
