package logger

import (
	"testing"
)

func TestInitializeLogger(t *testing.T) {
	testcases := [][2]string{
		{"debug", "test.log"},
		{"info", "stderr"},
		{"warn", ""},
		{"", ""},
		{"nil", "nil"},
	}
	for _, tc := range testcases {
		var logLevel *string
		var logFile *string
		logLevel = &tc[0]
		logFile = &tc[1]
		if tc[0] == "nil" {
			logLevel = nil
		}
		if tc[1] == "nil" {
			logFile = nil
		}

		InitializeLogger(logLevel, logFile)

		if log == nil {
			t.Fatal("Logger should be initialized")
		}

		// if log.Core().Enabled(zap.DebugLevel) == false {
		// 	t.Fatal("Logger should have debug level enabled")
		// }
	}
}
