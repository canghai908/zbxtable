package otel

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/bridges/otellogrus"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// logProvider 保存已初始化的日志 Provider，供挂载 logrus Hook 使用。
var logProvider atomic.Pointer[sdklog.LoggerProvider]

// logsEnabled 标记日志上报是否已启用。
var logsEnabled atomic.Bool

// LogsEnabled 返回日志上报是否已启用。
func LogsEnabled() bool {
	return logsEnabled.Load()
}

// setupLogs 初始化 OpenTelemetry 日志 Provider（OTLP gRPC/HTTP 导出器），
// 并注册为全局 LoggerProvider。返回用于优雅关闭的 shutdown 函数。
func setupLogs(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	exporter, err := newLogExporter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("otel: 创建 OTLP 日志导出器失败: %w", err)
	}

	res := buildResource(cfg)
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)

	global.SetLoggerProvider(lp)
	logProvider.Store(lp)
	logsEnabled.Store(true)

	return func(ctx context.Context) error {
		logsEnabled.Store(false)
		return lp.Shutdown(ctx)
	}, nil
}

// newLogExporter 根据协议创建 OTLP 日志导出器。
func newLogExporter(ctx context.Context, cfg Config) (sdklog.Exporter, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Protocol)) {
	case "http", "https", "otlphttp":
		opts := []otlploghttp.Option{otlploghttp.WithEndpoint(cfg.Endpoint)}
		if cfg.Insecure {
			opts = append(opts, otlploghttp.WithInsecure())
		}
		if len(cfg.Headers) > 0 {
			opts = append(opts, otlploghttp.WithHeaders(cfg.Headers))
		}
		return otlploghttp.New(ctx, opts...)
	default: // grpc
		opts := []otlploggrpc.Option{otlploggrpc.WithEndpoint(cfg.Endpoint)}
		if cfg.Insecure {
			opts = append(opts, otlploggrpc.WithInsecure())
		}
		if len(cfg.Headers) > 0 {
			opts = append(opts, otlploggrpc.WithHeaders(cfg.Headers))
		}
		return otlploggrpc.New(ctx, opts...)
	}
}

// AttachLogrusHook 将 OpenTelemetry 日志桥接 Hook 挂载到给定的 logrus 实例，
// 使日志在继续写本地文件的同时，通过 OTLP 上报到 Zabbix 8 APM 的 Logs。
// 若日志上报未启用或 logger 为空，则不做任何操作。
func AttachLogrusHook(l *logrus.Logger) {
	if l == nil || !logsEnabled.Load() {
		return
	}
	lp := logProvider.Load()
	if lp == nil {
		return
	}
	hook := otellogrus.NewHook("ZbxTable",
		otellogrus.WithLoggerProvider(lp),
		// 上报除 Trace 级别以外的所有级别，避免过于冗长。
		otellogrus.WithLevels([]logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
			logrus.InfoLevel,
			logrus.DebugLevel,
		}),
	)
	l.AddHook(hook)
}
