package logutil

import "testing"

func TestName(t *testing.T) {
	err := InitGlobLogger()
	if err != nil {
		t.Fatal(err)
	}
	for {
		go func() {
			Info("info")
		}()
	}
}

func TestName2(t *testing.T) {
	err := InitGlobLogger(LogConfig{
		LogLevel:          LevelInfo,
		LogFormat:         Json,
		LogPath:           "./logs",
		LogFileName:       "test.log",
		LogFileMaxSize:    200,
		LogFileMaxBackups: 7,
		LogMaxAge:         14,
		LogCompress:       true,
	})
	if err != nil {
		t.Fatal(err)
	}
	Debug("debug")
	Info("info")
	Warn("warn")
}
