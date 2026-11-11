package otel

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// metricsEnabled 标记指标上报是否已启用。
var metricsEnabled atomic.Bool

// MetricsEnabled 返回指标上报是否已启用。
func MetricsEnabled() bool {
	return metricsEnabled.Load()
}

// setupMetrics 初始化 OpenTelemetry 指标 MeterProvider（OTLP gRPC/HTTP 导出器），
// 注册为全局 MeterProvider，并启动 Go 运行时指标采集（内存/GC/协程数等）。
// 返回用于优雅关闭的 shutdown 函数。
func setupMetrics(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	exporter, err := newMetricExporter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("otel: 创建 OTLP 指标导出器失败: %w", err)
	}

	interval := cfg.MetricInterval
	if interval <= 0 {
		interval = 15 * time.Second
	}

	res := buildResource(cfg)
	reader := sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(interval))
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	metricsEnabled.Store(true)

	// 启动 Go 运行时指标采集（基于全局 MeterProvider）。
	if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(interval)); err != nil {
		// 运行时指标采集失败不影响其它指标上报，仅停止运行时采集。
		return func(ctx context.Context) error {
			metricsEnabled.Store(false)
			return mp.Shutdown(ctx)
		}, fmt.Errorf("otel: 启动运行时指标采集失败: %w", err)
	}

	return func(ctx context.Context) error {
		metricsEnabled.Store(false)
		return mp.Shutdown(ctx)
	}, nil
}

// newMetricExporter 根据协议创建 OTLP 指标导出器。
func newMetricExporter(ctx context.Context, cfg Config) (sdkmetric.Exporter, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Protocol)) {
	case "http", "https", "otlphttp":
		opts := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpoint(cfg.Endpoint)}
		if cfg.Insecure {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}
		if len(cfg.Headers) > 0 {
			opts = append(opts, otlpmetrichttp.WithHeaders(cfg.Headers))
		}
		return otlpmetrichttp.New(ctx, opts...)
	default: // grpc
		opts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(cfg.Endpoint)}
		if cfg.Insecure {
			opts = append(opts, otlpmetricgrpc.WithInsecure())
		}
		if len(cfg.Headers) > 0 {
			opts = append(opts, otlpmetricgrpc.WithHeaders(cfg.Headers))
		}
		return otlpmetricgrpc.New(ctx, opts...)
	}
}
