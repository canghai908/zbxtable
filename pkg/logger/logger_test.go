package logger

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestGetLogLevelByRunMode(t *testing.T) {
	tests := []struct {
		runmode  string
		expected int
	}{
		{"dev", 5},     // Debug
		{"test", 4},    // Info
		{"prod", 3},    // Warn
		{"unknown", 4}, // Default to Info
		{"", 4},        // Default to Info
	}

	for _, tt := range tests {
		t.Run(tt.runmode, func(t *testing.T) {
			result := getLogLevelByRunMode(tt.runmode)
			if result != tt.expected {
				t.Errorf("getLogLevelByRunMode(%q) = %d, want %d", tt.runmode, result, tt.expected)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		level    int
		expected logrus.Level
	}{
		{0, logrus.PanicLevel},
		{1, logrus.FatalLevel},
		{2, logrus.ErrorLevel},
		{3, logrus.WarnLevel},
		{4, logrus.InfoLevel},
		{5, logrus.DebugLevel},
		{6, logrus.TraceLevel},
		{99, logrus.InfoLevel}, // Default
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.level)), func(t *testing.T) {
			result := parseLogLevel(tt.level)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%d) = %v, want %v", tt.level, result, tt.expected)
			}
		})
	}
}

func TestInitLoggerWithRunMode(t *testing.T) {
	tests := []struct {
		runmode       string
		expectedLevel logrus.Level
	}{
		{"dev", logrus.DebugLevel},
		{"test", logrus.InfoLevel},
		{"prod", logrus.WarnLevel},
	}

	for _, tt := range tests {
		t.Run(tt.runmode, func(t *testing.T) {
			// 使用临时目录进行测试
			err := InitLoggerWithRunMode(tt.runmode, "", 7, 10000, 100, true)
			if err != nil {
				t.Fatalf("InitLoggerWithRunMode failed: %v", err)
			}

			if Log == nil {
				t.Fatal("Log instance is nil")
			}

			if Log.GetLevel() != tt.expectedLevel {
				t.Errorf("Log level = %v, want %v", Log.GetLevel(), tt.expectedLevel)
			}
		})
	}
}
