package otel

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	// meterName HTTP 指标使用的 Meter 名称。
	meterName = "zbxtable/http"
)

// GinMetricsMiddleware 返回一个 Gin 中间件，为每个 HTTP 请求记录业务指标：
//   - http.server.request.duration：请求耗时直方图（秒），其 count 即请求数；
//   - http.server.active_requests：当前并发处理中的请求数。
//
// 指标按 方法 / 路由模板 / 状态码 维度聚合，上报到 Zabbix 8 APM 的 Metrics。
// 使用路由模板（c.FullPath()）而非真实路径，避免路径参数导致的维度爆炸。
// 未启用指标时，全局 MeterProvider 为 Noop，记录开销可忽略。
func GinMetricsMiddleware() gin.HandlerFunc {
	meter := otel.GetMeterProvider().Meter(meterName)

	duration, err := meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithUnit("s"),
		metric.WithDescription("HTTP 服务端请求处理耗时"),
	)
	if err != nil {
		duration = nil
	}

	active, err := meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithUnit("{request}"),
		metric.WithDescription("当前正在处理的 HTTP 请求数"),
	)
	if err != nil {
		active = nil
	}

	return func(c *gin.Context) {
		start := time.Now()

		method := c.Request.Method
		if active != nil {
			inflightAttrs := metric.WithAttributes(
				attribute.String("http.request.method", method),
			)
			active.Add(c.Request.Context(), 1, inflightAttrs)
			defer active.Add(c.Request.Context(), -1, inflightAttrs)
		}

		c.Next()

		if duration != nil {
			route := c.FullPath()
			if route == "" {
				// 未匹配到任何路由（如 404），统一归并以避免维度爆炸。
				route = "unmatched"
			}
			attrs := metric.WithAttributes(
				attribute.String("http.request.method", method),
				attribute.String("http.route", route),
				attribute.Int("http.response.status_code", c.Writer.Status()),
			)
			duration.Record(c.Request.Context(), time.Since(start).Seconds(), attrs)
		}
	}
}
