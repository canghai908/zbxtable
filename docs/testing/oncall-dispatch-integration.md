# 分发规则与值班排班实际联调

使用显式指定的 PostgreSQL 测试库。每个测试创建独立 schema，真实写入规则、排班、代班、告警、接收人快照和通知日志，结束后自动删除 schema。进程内队列和配置缓存也隔离并恢复。不会写入开发或生产库，也不会使用真实邮件或企业微信账号。

## 生成的规则与数据

测试用户：primary（当班）、next-shift（下一班）、cover（代班）、fallback（兜底）、global（全局默认）、缺少通知地址用户、禁用用户；用户组 ops 包含 primary、cover、fallback。

普通规则示例：`主机包含 prod- AND (告警内容包含 CPU OR 告警内容包含 Memory)`。普通规则关联 `schedule_ids`，实例默认规则通知 fallback，全局默认规则通知 global。旧格式规则仍直接通知固定用户。测试同时生成用户组、固定人员与排班混合的规则，以及按渠道生效的维护屏蔽规则。

模拟告警携带主机、实例、指标 ID、指标值、告警内容、等级、事件 ID。通过实际 Webhook 数据适配器 MsAdd 写入告警，并由 GenAlert 读取数据库中的规则与排班进行分发。

| 场景 | 期望 |
| --- | --- |
| 旧条件格式 + 固定用户 | cover 收到 |
| 嵌套 AND/OR 命中 CPU | primary 收到 |
| OR 的 Memory 分支命中 | primary 收到 |
| AND 主机不符 | 实例默认 fallback 收到 |
| OR 所有分支不符 | 实例默认 fallback 收到 |
| 其他实例 | 全局默认 global 收到 |
| 禁用普通规则 | 实例默认 fallback 收到 |
| 多规则、用户、用户组与排班同时命中 | primary/cover/fallback 各一次 |
| 固定模式多人同时值班 | primary 与 next-shift 收到 |
| 日轮换已交班 | next-shift 收到 |
| 周轮换已交班 | next-shift 收到 |
| 临时代班生效 | cover 收到 |
| 临时代班已过期 | primary 收到 |
| 临时代班尚未开始 | primary 收到 |
| 不在覆盖星期内 | fallback 收到 |
| 排班尚未开始 | fallback 收到 |
| 排班禁用 + 兜底用户组 | ops 可达成员收到 |
| 当班人缺少渠道地址 | fallback 收到 |
| 当班人已禁用 | fallback 收到 |
| 当班与兜底均不可达 | 不入队，状态为无可用接收人 |
| 仅屏蔽邮件 | 企业微信仍入队 |
| 屏蔽全部渠道 | 不入队，状态为屏蔽 |
| 屏蔽规则已过期 | primary 收到 |

每个场景断言实际队列中的接收人、告警事件及指标字段、通知状态。排班分发额外核对数据库接收人快照来源。原有 PostgreSQL 测试还覆盖同事件重试沿用首次快照、排班变更后新事件重新解析、删除引用保护、代班冲突和审计。

## 实际 HTTP 与通知投递

- 通过管理接口创建排班/规则后，使用真实本地 HTTP 服务接收 Webhook，核对 token 拒绝、告警入库、入队接收人、实例名称、快照、告警查询及规则预览一致性。管理接口测试使用测试管理员身份注入，不验证登录页面或正式会话登录流程。
- 邮件使用真实本地 SMTP TCP 端口和生产发送函数，核对 CRAM-MD5、信封地址、MIME 解码后的正文、故障/恢复模板和通知日志；同时测试 SMTP 550 拒绝及日志存储失败。
- 群机器人使用真实本地 HTTP 接收端；测试传输层仅将企业微信机器人路径重定向至本地，拒绝其他目标。核对请求正文和事件日志；另测 HTTP 200 业务拒绝、非法 JSON、缺少 errcode、超限响应、非 200、网络异常、响应读取失败及日志存储失败。

联调修复：恢复邮件按原始状态选择恢复模板，展示格式化使用局部副本，避免改变队列事件；邮件发送错误与日志错误不再被返回值吞掉；机器人必须收到有效 `errcode: 0` 才记为成功，失败不会再误记成功，错误文本不包含请求 URL 中的机器人 Key。

## 复现

需要 Go 项目要求的版本和可创建 schema 的 PostgreSQL 测试账号；无需新增 Go 依赖。

```sh
export ZBXTABLE_ONCALL_TEST_DSN='postgres://test_user@127.0.0.1:15439/testdb?sslmode=disable'
go test ./... -count=1
go test -race ./internal/model ./internal/handler -count=1
go test ./internal/model ./internal/handler -run TestOnCall -count=3
go vet ./...
go build ./...
git ls-files '*.go' | xargs gofmt -l
git ls-files --others --exclude-standard '*.go' | xargs gofmt -l
```

不设置测试库环境变量时，数据库集成测试明确跳过，其他单元测试照常运行。此次验证不代表真实外部 SMTP/企业微信账号配置、网络或第三方服务已通过验证；通知实际抵达本地接收端。

相关测试：`internal/model/oncall_scenarios_postgres_test.go`、`oncall_delivery_postgres_test.go`、`oncall_robot_postgres_test.go`、`oncall_postgres_test.go`、`internal/handler/oncall_postgres_test.go`。

## 本次结果（2026-11-20）

- PostgreSQL 集成开启后，`go test ./... -count=1` 全部通过（数据库测试实际执行）。
- `go test -race ./internal/model ./internal/handler -count=1` 通过。
- `go vet ./...`、`go build ./...`、已跟踪及新增 Go 文件的 gofmt 检查通过。
- 前端 `node --test tests/*.test.js` 33 项通过；本次未修改前端代码。
- 确认数据库中测试 schema 自动清理，无遗留测试规则和排班。
