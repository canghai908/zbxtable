package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Log 全局日志实例
	Log *logrus.Logger
)

// InitLogger 初始化日志系统
// logPath: 日志文件路径，如果为空则默认为 ./log/yyyy-MM-dd.log
// logLevel: 日志级别 (0-6)
// maxDays: 日志保留天数
// maxLines: 最大行数（暂未使用）
// maxSize: 单个日志文件最大大小（MB），0表示不限制
// daily: 是否按天分割日志文件
func InitLogger(logPath string, logLevel int, maxDays, maxLines, maxSize int, daily bool) error {
	Log = logrus.New()

	// 启用调用者信息报告（显示文件名和行号）
	Log.SetReportCaller(true)

	// 设置日志格式
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			// 自定义调用者信息格式：只显示文件名（不含完整路径）和行号
			filename := filepath.Base(f.File)
			return "", fmt.Sprintf("[%s:%d]", filename, f.Line)
		},
	})

	// 设置日志级别
	level := parseLogLevel(logLevel)
	Log.SetLevel(level)

	// 如果未配置日志路径，使用默认路径：./log/yyyy-MM-dd.log
	if logPath == "" {
		logPath = filepath.Join(".", "log", fmt.Sprintf("%s.log", time.Now().Format("2006-01-02")))
	} else {
		// 如果配置的是目录路径（不包含文件扩展名），则在该目录下创建日志文件
		// 判断是否为目录：没有扩展名或者以 / 结尾
		if filepath.Ext(logPath) == "" {
			// 这是一个目录路径，需要在该目录下创建日志文件
			if daily {
				logPath = filepath.Join(logPath, fmt.Sprintf("%s.log", time.Now().Format("2006-01-02")))
			} else {
				logPath = filepath.Join(logPath, "app.log")
			}
		} else {
			// 这是一个完整的文件路径
			if daily {
				// 如果启用按天分割，修改日志文件名为日期格式
				logDir := filepath.Dir(logPath)
				logPath = filepath.Join(logDir, fmt.Sprintf("%s.log", time.Now().Format("2006-01-02")))
			}
		}
	}

	// 确保日志目录存在
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// 配置日志轮转
	writer := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    maxSize, // MB，0表示不限制
		MaxBackups: maxDays, // 保留的旧日志文件数量
		MaxAge:     maxDays, // 保留天数
		Compress:   true,    // 压缩旧日志
		LocalTime:  true,    // 使用本地时间
	}

	Log.SetOutput(writer)

	return nil
}

// InitLoggerWithRunMode 根据 runmode 初始化日志系统
// runmode: 运行模式 (dev/test/prod)
// logPath: 日志文件路径，如果为空则默认为 ./log/yyyy-MM-dd.log
// maxDays: 日志保留天数
// maxLines: 最大行数（暂未使用）
// maxSize: 单个日志文件最大大小（MB），0表示不限制
// daily: 是否按天分割日志文件
func InitLoggerWithRunMode(runmode, logPath string, maxDays, maxLines, maxSize int, daily bool) error {
	// 根据 runmode 自动设置日志级别
	logLevel := getLogLevelByRunMode(runmode)
	return InitLogger(logPath, logLevel, maxDays, maxLines, maxSize, daily)
}

// getLogLevelByRunMode 根据运行模式返回对应的日志级别
// dev: Debug级别 (5) - 开发环境，记录详细的调试信息
// test: Info级别 (4) - 测试环境，记录一般信息
// prod: Warn级别 (3) - 生产环境，只记录警告和错误
func getLogLevelByRunMode(runmode string) int {
	switch runmode {
	case "dev":
		return 5 // Debug级别
	case "test":
		return 4 // Info级别
	case "prod":
		return 3 // Warn级别
	default:
		return 4 // 默认Info级别
	}
}

// parseLogLevel 解析日志级别数字为 logrus.Level
func parseLogLevel(logLevel int) logrus.Level {
	switch logLevel {
	case 0:
		return logrus.PanicLevel
	case 1:
		return logrus.FatalLevel
	case 2:
		return logrus.ErrorLevel
	case 3:
		return logrus.WarnLevel
	case 4:
		return logrus.InfoLevel
	case 5:
		return logrus.DebugLevel
	case 6:
		return logrus.TraceLevel
	default:
		return logrus.InfoLevel
	}
}

// 兼容 beego/logs 的常用方法
func Info(args ...interface{}) {
	Log.Info(args...)
}

func Error(args ...interface{}) {
	Log.Error(args...)
}

func Warn(args ...interface{}) {
	Log.Warn(args...)
}

func Debug(args ...interface{}) {
	Log.Debug(args...)
}

func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

func Errorf(format string, args ...interface{}) {
	Log.Errorf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	Log.Warnf(format, args...)
}

func Debugf(format string, args ...interface{}) {
	Log.Debugf(format, args...)
}
