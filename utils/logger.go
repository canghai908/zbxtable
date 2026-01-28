package utils

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Log 全局日志实例
	Log *logrus.Logger
)

// InitLogger 初始化日志系统
func InitLogger(logPath string, logLevel int, maxDays, maxLines, maxSize int, daily bool) error {
	Log = logrus.New()

	// 设置日志格式
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 设置日志级别
	level := logrus.InfoLevel
	switch logLevel {
	case 0:
		level = logrus.PanicLevel
	case 1:
		level = logrus.InfoLevel // 修改：level 1 应该是 InfoLevel，而不是 FatalLevel
	case 2:
		level = logrus.ErrorLevel
	case 3:
		level = logrus.WarnLevel
	case 4:
		level = logrus.InfoLevel
	case 5:
		level = logrus.DebugLevel
	case 6:
		level = logrus.TraceLevel
	}
	Log.SetLevel(level)

	// 如果配置了日志文件路径，则输出到文件
	if logPath != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(logPath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		// 配置日志轮转
		writer := &lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    maxSize,    // MB
			MaxBackups: maxDays,    // 保留天数
			MaxAge:     maxDays,    // 保留天数
			Compress:   true,       // 压缩旧日志
			LocalTime:  true,
		}

		// 如果启用按天分割，需要自定义文件名
		if daily {
			// 这里简化处理，实际可以按日期动态生成文件名
			// lumberjack 本身不支持按天分割，但可以通过 MaxAge 和 MaxBackups 实现类似效果
		}

		Log.SetOutput(writer)
	} else {
		// 默认输出到标准输出
		Log.SetOutput(os.Stdout)
	}

	return nil
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
