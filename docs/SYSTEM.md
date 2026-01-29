# ZbxTable 系统说明文档

## 目录

- [系统概述](#系统概述)
- [系统架构](#系统架构)
- [功能特性](#功能特性)
- [技术栈](#技术栈)
- [安装部署](#安装部署)
- [配置说明](#配置说明)
- [使用指南](#使用指南)
- [API 文档](#api-文档)
- [常见问题](#常见问题)

## 系统概述

ZbxTable 是一个基于 Go 语言开发的 Zabbix 报表系统，旨在为 Zabbix 监控系统提供强大的数据分析、报表生成和可视化功能。

### 主要特点

- 🎨 **自定义拓扑图**：支持自定义绘制网络拓扑图，直观展示设备关系
- 📊 **设备分类展示**：按照主机组、标签等维度对设备进行分类展示和导出
- 📈 **告警分析**：导出指定时间段内的 Zabbix 告警信息，支持 Excel 格式
- 🔍 **告警统计**：分析特定时间段的告警数据，生成告警 Top 10 等统计报表
- 📅 **定时报表**：支持配置定时任务，自动生成和发送报表
- 📧 **多渠道通知**：支持邮件、企业微信等多种告警通知方式
- 🌐 **多租户支持**：支持多个 Zabbix 实例管理
- 🔐 **权限管理**：完善的用户权限管理体系

## 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                         用户层                               │
│                    (Web Browser)                            │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                       前端层 (Vue.js)                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ 拓扑管理 │  │ 报表管理 │  │ 告警分析 │  │ 系统管理 │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    后端层 (Go + Gin)                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ API 路由 │  │ 业务逻辑 │  │ 数据处理 │  │ 任务调度 │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└─────────────────────────────────────────────────────────────┘
                            │
                ┌───────────┴───────────┐
                ▼                       ▼
┌──────────────────────┐    ┌──────────────────────┐
│   数据层 (MySQL)      │    │  Zabbix API          │
│  ┌────────────────┐  │    │  ┌────────────────┐  │
│  │ 用户数据       │  │    │  │ 监控数据       │  │
│  │ 报表配置       │  │    │  │ 告警信息       │  │
│  │ 拓扑数据       │  │    │  │ 主机信息       │  │
│  └────────────────┘  │    │  └────────────────┘  │
└──────────────────────┘    └──────────────────────┘
```

## 功能特性

### 1. 拓扑管理

- **自定义拓扑图**：使用可视化编辑器绘制网络拓扑
- **设备关联**：将 Zabbix 主机与拓扑节点关联
- **实时状态**：显示设备实时监控状态
- **告警展示**：在拓扑图上直观显示告警信息

### 2. 报表管理

- **主机报表**：生成主机性能报表（CPU、内存、磁盘等）
- **告警报表**：导出告警历史记录
- **自定义报表**：支持自定义报表模板
- **多格式导出**：支持 Excel、PDF 等多种格式

### 3. 告警分析

- **告警统计**：按时间、主机、严重程度统计告警
- **Top N 分析**：告警主机 Top 10、告警项 Top 10
- **趋势分析**：告警趋势图表展示
- **告警详情**：查看告警详细信息和处理记录

### 4. 定时任务

- **定时报表**：配置定时生成报表任务
- **自动发送**：支持邮件、企业微信自动发送
- **任务日志**：记录任务执行历史和结果
- **灵活调度**：支持 Cron 表达式配置执行时间

### 5. 系统管理

- **用户管理**：用户账号的创建、编辑、删除
- **权限管理**：基于角色的权限控制
- **Zabbix 实例管理**：支持多个 Zabbix 实例
- **系统配置**：邮件服务器、企业微信等配置

## 技术栈

### 后端

- **语言**：Go 1.23+
- **Web 框架**：Gin
- **ORM**：GORM
- **数据库**：MySQL 8.0+ / PostgreSQL
- **日志**：Logrus + Lumberjack（日志轮转）
- **任务调度**：Cron
- **Zabbix API**：zabbix-go

### 前端

- **框架**：Vue.js 2.x
- **UI 组件**：Ant Design Vue
- **图表**：ECharts
- **拓扑编辑**：自定义 Canvas 组件
- **HTTP 客户端**：Axios

### 部署

- **容器化**：Docker + Docker Compose
- **反向代理**：Nginx（可选）
- **进程管理**：Systemd / Supervisor

## 安装部署

### 方式一：Docker 部署（推荐）

详见 [Docker 部署指南](../deployments/docker/README.md)

### 方式二：二进制部署

1. **下载二进制文件**

```bash
# 从 GitHub Releases 下载最新版本
wget https://github.com/canghai908/zbxtable/releases/latest/download/zbxtable-linux-amd64.tar.gz
tar -xzf zbxtable-linux-amd64.tar.gz
cd zbxtable
```

2. **配置数据库**

```bash
# 创建数据库
mysql -uroot -p -e "CREATE DATABASE zbxtable DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -uroot -p -e "CREATE USER 'zbxtable'@'localhost' IDENTIFIED BY 'zbxtablepwd123';"
mysql -uroot -p -e "GRANT ALL PRIVILEGES ON zbxtable.* TO 'zbxtable'@'localhost';"
```

3. **修改配置文件**

```bash
cp config/app.conf.example config/app.conf
vim config/app.conf
```

4. **启动服务**

```bash
# 直接启动
./zbxtable web

# 或使用 systemd
sudo cp zbxtable.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable zbxtable
sudo systemctl start zbxtable
```

### 方式三：源码编译

详见 [README.md](../README.md) 的编译部分

## 配置说明

### 应用配置 (config/app.conf)

```ini
; 应用配置
appname   = zbxtable          ; 应用名称
httpport  = 8085              ; HTTP 端口
runmode   = prod              ; 运行模式：dev/prod
timeout   = 12                ; 超时时间（秒）
token     = your_secret_token ; API Token

; 日志配置
log_level = 6                 ; 日志级别：0-6 (0=Panic, 1=Info, 2=Error, 3=Warn, 4=Info, 5=Debug, 6=Trace)
log_path  =                   ; 日志路径，留空则默认为 ./log/yyyy-MM-dd.log
maxlines  = 1000              ; 最大行数
maxsize   = 100               ; 单个日志文件最大大小（MB）
maxdays   = 10                ; 日志保留天数
daily     = true              ; 是否按天分割日志

; 数据库配置
dbtype    = mysql             ; 数据库类型：mysql/postgresql
dbhost    = localhost         ; 数据库主机
dbuser    = zbxtable          ; 数据库用户
dbpass    = zbxtablepwd123    ; 数据库密码
dbname    = zbxtable          ; 数据库名
dbport    = 3306              ; 数据库端口
```

### 日志配置说明

ZbxTable 使用 Logrus + Lumberjack 实现日志管理，具有以下特性：

- **按日期命名**：日志文件自动按日期命名，格式为 `yyyy-MM-dd.log`
- **自动分割**：当日志文件达到 `maxsize` 设置的大小时自动分割
- **自动压缩**：旧日志文件自动压缩为 `.gz` 格式
- **自动清理**：超过 `maxdays` 天数的日志自动删除
- **默认路径**：如果 `log_path` 为空，默认使用 `./log/yyyy-MM-dd.log`

日志级别说明：

| 级别 | 说明 | 适用场景 |
|------|------|----------|
| 0 | Panic | 严重错误，程序无法继续运行 |
| 1 | Info | 一般信息，生产环境推荐 |
| 2 | Error | 错误信息 |
| 3 | Warn | 警告信息 |
| 4 | Info | 信息级别 |
| 5 | Debug | 调试信息 |
| 6 | Trace | 跟踪信息，开发环境使用 |

## 使用指南

### 首次使用

1. **访问安装向导**

首次启动后，访问 `http://localhost:8085/install`，按照向导完成安装。

2. **配置 Zabbix 连接**

在安装向导中填写 Zabbix API 地址、用户名和密码。

3. **创建管理员账号**

设置系统管理员账号和密码。

4. **完成安装**

点击完成安装，系统会自动初始化数据库。

### 基本操作

#### 1. 添加 Zabbix 实例

系统管理 → Zabbix 实例管理 → 添加实例

填写实例信息：
- 实例名称
- API 地址
- 用户名/密码 或 API Token

#### 2. 创建拓扑图

拓扑管理 → 新建拓扑

- 使用拖拽方式添加节点
- 关联 Zabbix 主机
- 添加连接线
- 保存拓扑

#### 3. 生成报表

报表管理 → 新建报表

- 选择报表类型（主机报表/告警报表）
- 选择时间范围
- 选择主机或主机组
- 生成并导出报表

#### 4. 配置定时任务

任务管理 → 新建任务

- 设置任务名称
- 选择报表模板
- 配置 Cron 表达式
- 设置通知方式
- 启用任务

## API 文档

详见 [API 文档](./api/README.md)

### 认证

所有 API 请求需要在 Header 中携带 Token：

```
Authorization: Bearer your_token_here
```

### 常用接口

#### 获取主机列表

```http
GET /api/v1/hosts
```

#### 获取告警列表

```http
GET /api/v1/alarms?start_time=2024-01-01&end_time=2024-01-31
```

#### 生成报表

```http
POST /api/v1/reports
Content-Type: application/json

{
  "type": "host",
  "host_ids": [10084, 10085],
  "start_time": "2024-01-01 00:00:00",
  "end_time": "2024-01-31 23:59:59"
}
```

## 常见问题

### 1. 无法连接 Zabbix API

**问题**：提示 "connect Zabbix API failed"

**解决方案**：
- 检查 Zabbix API 地址是否正确
- 确认 Zabbix 用户名和密码正确
- 检查网络连接是否正常
- 如果使用 HTTPS，确认证书有效

### 2. 数据库连接失败

**问题**：提示 "database connection failed"

**解决方案**：
- 检查数据库配置是否正确
- 确认数据库服务是否启动
- 检查数据库用户权限
- 测试数据库连接：`mysql -h host -u user -p`

### 3. 日志文件过大

**问题**：日志文件占用大量磁盘空间

**解决方案**：
- 调整 `maxdays` 参数，减少日志保留天数
- 调整 `maxsize` 参数，减小单个日志文件大小
- 降低 `log_level`，减少日志输出
- 定期清理旧日志文件

### 4. 报表生成失败

**问题**：生成报表时提示错误

**解决方案**：
- 检查时间范围是否合理
- 确认选择的主机存在且有数据
- 查看日志文件获取详细错误信息
- 检查磁盘空间是否充足

### 5. 定时任务不执行

**问题**：配置的定时任务没有按时执行

**解决方案**：
- 检查 Cron 表达式是否正确
- 确认任务已启用
- 查看任务日志了解执行情况
- 检查系统时间是否正确

### 6. 前端页面无法访问

**问题**：浏览器无法打开页面

**解决方案**：
- 检查服务是否正常启动
- 确认端口没有被占用
- 检查防火墙设置
- 查看浏览器控制台错误信息

## 性能优化

### 1. 数据库优化

- 定期清理历史数据
- 为常用查询字段添加索引
- 调整 MySQL 配置参数
- 使用数据库连接池

### 2. 应用优化

- 启用缓存机制
- 调整日志级别
- 优化查询语句
- 使用异步任务处理耗时操作

### 3. 系统优化

- 增加服务器内存
- 使用 SSD 硬盘
- 配置 Nginx 反向代理
- 启用 Gzip 压缩

## 安全建议

1. **修改默认密码**：首次安装后立即修改管理员密码
2. **使用 HTTPS**：生产环境建议配置 HTTPS
3. **限制访问**：使用防火墙限制访问来源
4. **定期备份**：定期备份数据库和配置文件
5. **更新系统**：及时更新到最新版本
6. **审计日志**：定期检查系统日志

## 更新日志

详见 [CHANGELOG.md](../CHANGELOG.md)

## 贡献指南

欢迎提交 Issue 和 Pull Request！

详见 [CONTRIBUTING.md](../CONTRIBUTING.md)

## 许可证

Apache-2.0 License

## 联系我们

- GitHub: https://github.com/canghai908/zbxtable
- 官网: https://zbxtable.com
- 邮箱: support@zbxtable.com

## 致谢

感谢所有为 ZbxTable 做出贡献的开发者！

