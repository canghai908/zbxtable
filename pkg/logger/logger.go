package logger

import (
	"fmt"
	"os"
	"path/filepath"
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
		level = logrus.InfoLevel
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

	// 如果未配置日志路径，使用默认路径：./log/yyyy-MM-dd.log
	if logPath == "" {
		logPath = filepath.Join(".", "log", fmt.Sprintf("%s.log", time.Now().Format("2006-01-02")))
	}

	// 如果启用按天分割，修改日志文件名为日期格式
	if daily {
		logDir := filepath.Dir(logPath)
		logPath = filepath.Join(logDir, fmt.Sprintf("%s.log", time.Now().Format("2006-01-02")))
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
