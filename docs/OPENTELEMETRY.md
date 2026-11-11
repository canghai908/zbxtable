# OpenTelemetry 链路追踪（对接 Zabbix 8 APM）

ZbxTable 内置 OpenTelemetry（OTel）埋点，可将 HTTP 请求链路、数据库 SQL
调用、应用日志与运行时指标以 OTLP 协议上报到 **Zabbix 8 的 APM** 接收端，
在 Zabbix 中查看服务的分布式追踪（Traces）、日志（Logs）、指标（Metrics）、
调用耗时与异常。

该功能默认**关闭**，不影响现有部署；开启后对性能影响极小。

## 埋点范围

- **HTTP 服务端**：每个进入的 HTTP 请求都会生成一个 Span（基于
  `otelgin` 中间件），记录路由、方法、状态码与耗时。
- **数据库**：启用追踪后自动为每条 GORM 执行的 SQL 生成子 Span（基于
  `gorm.io/plugin/opentelemetry`），便于定位慢查询。
- **日志**：启用后通过 `otellogrus` 桥接，将 logrus 日志在**继续写本地
  文件的同时**以 OTLP 上报到 Zabbix APM 的 Logs 视图。
- **指标**：启用后同时采集两类指标，通过周期性 Reader 以 OTLP 上报到
  Zabbix APM 的 Metrics 视图：
  - **运行时指标**：内存、GC、协程数等（基于
    `contrib/instrumentation/runtime`）；
  - **HTTP 业务指标**：`http.server.request.duration`（请求耗时直方图，其
    count 即请求数）与 `http.server.active_requests`（并发请求数），按
    方法 / 路由模板 / 状态码维度聚合。路由使用模板（如 `/v1/host/:hostid`）
    而非真实路径，避免维度爆炸。
- **链路传播**：使用 W3C TraceContext + Baggage，可与上下游服务串联。

## 配置项

**推荐在页面配置**：登录后进入「系统配置 → APM 监控」分区设置并保存，配置存储
在数据库中。也可在 `config/app.conf`（或 `.env`）中设置同名键作为回退（数据库
中对应值为空时才生效）。

> ⚠️ 无论从页面还是配置文件修改，**保存后都需要重启服务才能生效**（埋点在
> 进程启动时一次性初始化）。

以下为各配置键：

| 配置项 | 说明 | 默认值 |
| --- | --- | --- |
| `otel_enable` | 是否启用追踪上报（`true`/`false`），数据库默认关闭 | `false` |
| `otel_endpoint` | OTLP 接收端地址，**不带协议前缀**。gRPC 一般为 `host:4317`，HTTP 为 `host:4318` | `127.0.0.1:4317` |
| `otel_protocol` | 传输协议：`grpc`（默认）或 `http` | `grpc` |
| `otel_insecure` | 是否使用明文连接（不启用 TLS）：`true`/`false` | `true` |
| `otel_service_name` | 服务名，在 Zabbix APM 中标识该服务 | `ZbxTable` |
| `otel_sample_ratio` | 采样率，范围 `[0,1]`，`1` 为全采样 | `1.0` |
| `otel_logs_enable` | 是否同时上报应用日志到 Zabbix APM Logs（`true`/`false`） | 跟随 `otel_enable` |
| `otel_metrics_enable` | 是否同时上报指标到 Zabbix APM Metrics（`true`/`false`） | 跟随 `otel_enable` |
| `otel_metrics_interval` | 指标导出周期（秒） | `15` |

> 服务版本（`service.version`）与部署环境（`deployment.environment`，取自
> `runmode`）会自动作为资源属性上报，无需单独配置。

### 示例

```ini
; OpenTelemetry APM
otel_enable       = true
otel_endpoint     = zabbix-server:4317
otel_protocol     = grpc
otel_insecure     = true
otel_service_name = ZbxTable
otel_sample_ratio = 1.0
otel_logs_enable  = true
otel_metrics_enable = true
otel_metrics_interval = 15
```

## Zabbix 8 侧配置

1. 确认 Zabbix 8 已启用并开放 OpenTelemetry/OTLP 接收端口
   （gRPC 默认 `4317`，HTTP 默认 `4318`）。
2. 将 `otel_endpoint` 指向该接收端地址。
3. 重启 ZbxTable 服务后，在 Zabbix 的 APM / 追踪视图中按
   `otel_service_name` 查找服务并查看链路数据。

## 说明

- 配置项缺失或 `otel_enable=false` 时，追踪完全关闭，使用全局 Noop
  Provider，几乎无额外开销。
- 初始化失败（例如 `endpoint` 未配置）只会记录日志并禁用追踪，**不会**
  影响服务正常启动与运行。
- 进程退出时会在 5 秒内优雅刷新缓冲区中尚未上报的 Span、日志与指标。
- 日志上报不会替换本地文件日志，二者同时生效。
- 若要让 APM 中的日志自动带上 `trace_id`/`span_id`（点击一条 Trace 即可
  关联查看对应日志），需在记录日志时携带请求 `context`
  （如 `logger.Log.WithContext(c.Request.Context()).Info(...)`）。该关联为
  可选增强，当前基础接入已可将日志上报到 Logs 视图。
