package logutil

import "testing"

func TestName(t *testing.T) {
	InitGlobalLogger(InfoLevel)
	Debug("debug")
	Info("info")
	Warn("warn")
	Error("error")
}
