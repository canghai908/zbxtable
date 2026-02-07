package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLogPathHandling(t *testing.T) {
	tests := []struct {
		name        string
		logPath     string
		daily       bool
		expectedDir string
		description string
	}{
		{
			name:        "目录路径-log",
			logPath:     "log",
			daily:       true,
			expectedDir: "log",
			description: "配置 log_path = log 时，日志应该在 log 目录下",
		},
		{
			name:        "目录路径-./log",
			logPath:     "./log",
			daily:       true,
			expectedDir: "log",
			description: "配置 log_path = ./log 时，日志应该在 log 目录下",
		},
		{
			name:        "目录路径-logs",
			logPath:     "logs",
			daily:       true,
			expectedDir: "logs",
			description: "配置 log_path = logs 时，日志应该在 logs 目录下",
		},
		{
			name:        "完整文件路径",
			logPath:     "log/app.log",
			daily:       true,
			expectedDir: "log",
			description: "配置完整文件路径时，日志应该在指定目录下",
		},
		{
			name:        "空路径",
			logPath:     "",
			daily:       true,
			expectedDir: "log",
			description: "未配置路径时，使用默认 log 目录",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟 InitLogger 中的路径处理逻辑
			logPath := tt.logPath

			if logPath == "" {
				logPath = filepath.Join(".", "log", time.Now().Format("2006-01-02")+".log")
			} else {
				if filepath.Ext(logPath) == "" {
					// 这是一个目录路径
					if tt.daily {
						logPath = filepath.Join(logPath, time.Now().Format("2006-01-02")+".log")
					} else {
						logPath = filepath.Join(logPath, "app.log")
					}
				} else {
					// 这是一个完整的文件路径
					if tt.daily {
						logDir := filepath.Dir(logPath)
						logPath = filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")
					}
				}
			}

			// 获取日志目录
			logDir := filepath.Dir(logPath)
			// 清理路径以便比较
			logDir = filepath.Clean(logDir)
			expectedDir := filepath.Clean(tt.expectedDir)

			if logDir != expectedDir {
				t.Errorf("%s\n期望日志目录: %s\n实际日志目录: %s\n完整路径: %s",
					tt.description, expectedDir, logDir, logPath)
			} else {
				t.Logf("✓ %s\n  日志目录: %s\n  完整路径: %s",
					tt.description, logDir, logPath)
			}
		})
	}
}

func TestInitLoggerWithDifferentPaths(t *testing.T) {
	// 创建临时测试目录
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	tests := []struct {
		name        string
		logPath     string
		expectedDir string
	}{
		{
			name:        "相对目录路径-log",
			logPath:     "log",
			expectedDir: "log",
		},
		{
			name:        "相对目录路径-logs",
			logPath:     "logs",
			expectedDir: "logs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 初始化日志
			err := InitLogger(tt.logPath, 4, 7, 10000, 100, true)
			if err != nil {
				t.Fatalf("InitLogger 失败: %v", err)
			}

			// 检查目录是否创建
			expectedDir := filepath.Join(tempDir, tt.expectedDir)
			if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
				t.Errorf("期望的日志目录未创建: %s", expectedDir)
			} else {
				t.Logf("✓ 日志目录创建成功: %s", expectedDir)
			}

			// 写入一条日志测试
			Log.Info("测试日志")

			// 检查日志文件是否存在
			logFile := filepath.Join(expectedDir, time.Now().Format("2006-01-02")+".log")
			if _, err := os.Stat(logFile); os.IsNotExist(err) {
				t.Errorf("日志文件未创建: %s", logFile)
			} else {
				t.Logf("✓ 日志文件创建成功: %s", logFile)
			}
		})
	}
}

