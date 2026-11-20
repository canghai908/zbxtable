// Package otel 封装 OpenTelemetry 的初始化逻辑，用于将链路（Trace）数据
// 通过 OTLP 协议上报到 Zabbix 8 的 APM（OpenTelemetry 接收端），
// 以便在 Zabbix 中查看服务的分布式追踪与性能数据。
//
// 使用方式：在服务启动时调用 Init，获得 shutdown 函数并在退出时调用，
// 用于刷新缓冲区中尚未发送的 Span。
package otel

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// enabled 标记 OpenTelemetry 是否已成功启用，供其它包（如数据库层）判断
// 是否需要挂载追踪插件，避免在未启用时产生额外开销。
var enabled atomic.Bool

// Enabled 返回 OpenTelemetry 追踪是否已启用。
func Enabled() bool {
	return enabled.Load()
}

// Config OpenTelemetry 初始化配置。
type Config struct {
	// Enabled 是否启用追踪上报。
	Enabled bool
	// Endpoint OTLP 接收端地址，例如 Zabbix server 的地址 "127.0.0.1:4317"
	// （gRPC）或 "127.0.0.1:4318"（HTTP）。不要带协议前缀。
	Endpoint string
	// Protocol 传输协议，可选 "grpc"（默认）或 "http"。
	Protocol string
	// Insecure 为 true 时使用明文连接（不启用 TLS）。
	Insecure bool
	// ServiceName 服务名，在 Zabbix APM 中用于标识该服务。
	ServiceName string
	// ServiceVersion 服务版本。
	ServiceVersion string
	// Environment 部署环境标识，例如 prod/dev。
	Environment string
	// SampleRatio 采样率，范围 [0,1]。1 表示全采样，0 表示不采样。
	SampleRatio float64
	// LogsEnabled 是否同时将应用日志通过 OTLP 上报到 Zabbix 8 APM 的 Logs。
	LogsEnabled bool
	// MetricsEnabled 是否同时将指标（运行时/HTTP）通过 OTLP 上报到 Zabbix 8 APM 的 Metrics。
	MetricsEnabled bool
	// MetricInterval 指标导出周期，默认 15s。
	MetricInterval time.Duration
	// Headers 附加到 OTLP 请求上的头部，例如用于鉴权。
	Headers map[string]string
}

// buildResource 构建上报用的资源信息（服务名、版本、部署环境）。
func buildResource(cfg Config) *resource.Resource {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(cfg.ServiceName),
	}
	if cfg.ServiceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersion(cfg.ServiceVersion))
	}
	if cfg.Environment != "" {
		attrs = append(attrs, semconv.DeploymentEnvironment(cfg.Environment))
	}
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, attrs...),
	)
	if err != nil {
		// 合并失败通常是 schema 不一致，退化为仅使用自定义属性。
		return resource.NewWithAttributes(semconv.SchemaURL, attrs...)
	}
	return res
}

// Init 根据配置初始化全局 TracerProvider 与文本传播器。
// 返回的 shutdown 用于在进程退出时优雅关闭并刷新未上报的 Span。
// 当 cfg.Enabled 为 false 时，返回一个空操作的 shutdown，不做任何上报。
func Init(ctx context.Context, cfg Config) (shutdown func(context.Context) error, err error) {
	noop := func(context.Context) error { return nil }

	if !cfg.Enabled {
		return noop, nil
	}
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return noop, fmt.Errorf("otel: endpoint 未配置，追踪未启用")
	}

	if cfg.ServiceName == "" {
		cfg.ServiceName = "ZbxTable"
	}
	if cfg.SampleRatio <= 0 {
		cfg.SampleRatio = 1.0
	}
	if cfg.SampleRatio > 1 {
		cfg.SampleRatio = 1.0
	}

	// 构建 OTLP Trace 导出器（支持 gRPC / HTTP 两种协议）。
	exporter, err := newExporter(ctx, cfg)
	if err != nil {
		return noop, fmt.Errorf("otel: 创建 OTLP 导出器失败: %w", err)
	}

	// 资源信息：服务名、版本、部署环境等，Zabbix APM 以此聚合。
	res := buildResource(cfg)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))),
	)

	otel.SetTracerProvider(tp)
	// 配置 W3C TraceContext + Baggage 传播，便于跨服务串联调用链。
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	enabled.Store(true)

	// 收集所有需要在退出时关闭的组件。
	shutdowns := []func(context.Context) error{
		func(ctx context.Context) error {
			enabled.Store(false)
			return tp.Shutdown(ctx)
		},
	}

	// 可选：初始化日志上报（OTLP Logs -> Zabbix 8 APM）。
	if cfg.LogsEnabled {
		logShutdown, logErr := setupLogs(ctx, cfg)
		if logErr != nil {
			// 日志上报失败不影响追踪：记录错误并继续。
			err = logErr
		} else {
			shutdowns = append(shutdowns, logShutdown)
		}
	}

	// 可选：初始化指标上报（OTLP Metrics -> Zabbix 8 APM）。
	if cfg.MetricsEnabled {
		metricShutdown, metricErr := setupMetrics(ctx, cfg)
		// 即使 runtime 采集启动失败，MeterProvider 仍可用，shutdown 非空。
		if metricShutdown != nil {
			shutdowns = append(shutdowns, metricShutdown)
		}
		if metricErr != nil {
			err = metricErr
		}
	}

	shutdown = func(ctx context.Context) error {
		var firstErr error
		// 逆序关闭。
		for i := len(shutdowns) - 1; i >= 0; i-- {
			if e := shutdowns[i](ctx); e != nil && firstErr == nil {
				firstErr = e
			}
		}
		return firstErr
	}
	return shutdown, err
}

// newExporter 根据协议创建对应的 OTLP Trace 导出器。
func newExporter(ctx context.Context, cfg Config) (*otlptrace.Exporter, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Protocol)) {
	case "http", "https", "otlphttp":
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(cfg.Endpoint),
		}
		if cfg.Insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		if len(cfg.Headers) > 0 {
			opts = append(opts, otlptracehttp.WithHeaders(cfg.Headers))
		}
		return otlptracehttp.New(ctx, opts...)
	default: // grpc
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
		}
		if cfg.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		if len(cfg.Headers) > 0 {
			opts = append(opts, otlptracegrpc.WithHeaders(cfg.Headers))
		}
		return otlptracegrpc.New(ctx, opts...)
	}
}
