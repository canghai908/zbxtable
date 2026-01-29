[English](./README.md) | 简体中文

# ZbxTable

[![Build Status](https://drone.cactifans.org/api/badges/canghai908/zbxtable/status.svg?ref=refs/heads/2.1)](https://drone.cactifans.org/canghai908/zbxtable)

ZbxTable 是一个使用 Go 语言开发的 Zabbix 报表系统。

## 功能特性

- 🎨 自定义绘制拓扑图
- 📊 设备分类展示和导出
- 📈 导出指定时间段内的 Zabbix 告警信息到 Excel
- 🔍 分析特定时间段的告警数据，生成告警 Top 10 等统计
- 📅 定时报表生成和自动发送
- 📧 多渠道通知（邮件、企业微信）
- 🌐 多租户支持，管理多个 Zabbix 实例
- 🔐 基于角色的权限控制

## 系统架构

![1](/zbxtable.png)

## 组件说明

**ZbxTable**: 后端使用 [Gin 框架](https://github.com/gin-gonic/gin) 和 Go 语言开发。

**ZbxTable-Web**: 前端使用 [Vue](https://github.com/vuejs/vue) 开发。

**MS-Agent**: 安装在 Zabbix Server 上，用于接收 Zabbix Server 产生的告警并发送到 ZbxTable。

## 在线演示

[https://demo.zbxtable.com](https://demo.zbxtable.com)

## 兼容性

| Zabbix 版本    | 兼容性 |
|:---------------|:------:|
| 7.4.x          |   ✅   |
| 7.2.x          |   ✅   |
| 7.0.x LTS      |   ✅   |
| 6.4.x          |   ✅   |
| 6.2.x          |   ✅   |
| 6.0.x LTS      |   ✅   |
| 5.4.x          |   ✅   |
| 5.2.x          |   ✅   |
| 5.0.x LTS      |   ✅   |
| 4.4.x          |   ✅   |
| 4.2.x          |   ✅   |
| 4.0.x LTS      |   ✅   |
| 3.4.x          | 未测试 |
| 3.2.x          | 未测试 |
| 3.0.x LTS      | 未测试 |

## 快速开始

### Docker 部署（推荐）

```bash
# 克隆仓库
git clone https://github.com/canghai908/zbxtable.git
cd zbxtable/zbxtable

# 复制配置文件
cp config/app.conf.example config/app.conf

# 编辑配置（修改数据库设置）
vim config/app.conf

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f zbxtable
```

访问 `http://localhost:8085` 即可使用应用。

详细的 Docker 部署说明请参考 [Docker 部署指南](./deployments/docker/README.md)。

### 二进制部署

```bash
# 下载最新版本
wget https://github.com/canghai908/zbxtable/releases/latest/download/zbxtable-linux-amd64.tar.gz
tar -xzf zbxtable-linux-amd64.tar.gz
cd zbxtable

# 配置数据库
cp config/app.conf.example config/app.conf
vim config/app.conf

# 启动服务
./zbxtable web
```

## 文档

- [系统说明文档](./docs/SYSTEM.md) - 完整的系统文档
- [Docker 部署指南](./deployments/docker/README.md) - Docker 部署说明
- [API 文档](./docs/api/README.md) - API 接口文档
- [官方网站](https://zbxtable.com) - 更多文档和教程

## 配置说明

### 日志配置

ZbxTable 使用 Logrus + Lumberjack 进行日志管理，具有以下特性：

- **按日期命名**：日志文件自动按日期命名（yyyy-MM-dd.log）
- **自动轮转**：日志达到配置大小时自动轮转
- **自动压缩**：旧日志文件自动压缩为 .gz 格式
- **自动清理**：超过保留期限的日志自动删除
- **默认路径**：如果 `log_path` 为空，默认使用 `./log/yyyy-MM-dd.log`

在 `config/app.conf` 中的配置示例：

```ini
log_level = 6      # 日志级别：0-6 (0=Panic, 1=Info, 2=Error, 3=Warn, 4=Info, 5=Debug, 6=Trace)
log_path  =        # 日志路径，留空则默认为 ./log/yyyy-MM-dd.log
maxsize   = 100    # 单个日志文件最大大小（MB）
maxdays   = 10     # 日志保留天数
daily     = true   # 启用按天分割日志
```

## 代码仓库

**ZbxTable**: [https://github.com/canghai908/zbxtable](https://github.com/canghai908/zbxtable)

**ZbxTable-Web**: [https://github.com/canghai908/zbxtable-web](https://github.com/canghai908/zbxtable-web)

**MS-Agent**: [https://github.com/canghai908/ms-agent](https://github.com/canghai908/ms-agent)

## 源码编译

要求：Go >= 1.23

```bash
# 克隆仓库
mkdir -p $GOPATH/src/github.com/canghai908
cd $GOPATH/src/github.com/canghai908
git clone https://github.com/canghai908/zbxtable.git
cd zbxtable/zbxtable

# 下载前端资源
wget -q -c https://dl.cactifans.com/stable/zbxtable/web-latest.tar.gz && tar xf web-latest.tar.gz

# 编译
go build -o zbxtable main.go

# 或使用控制脚本
./control build
./control pack
```

## Docker 构建

```bash
# 从源码构建
docker build -t zbxtable:latest -f zbxtable/Dockerfile .

# 或使用 docker-compose
docker-compose build
```

## 开发团队

**后端开发**: [canghai908](https://github.com/canghai908)

**前端开发**: [ahyiru](https://github.com/ahyiru)

## 贡献

欢迎贡献代码！请随时提交 Pull Request。

## 支持

- GitHub Issues: [https://github.com/canghai908/zbxtable/issues](https://github.com/canghai908/zbxtable/issues)
- 文档: [https://zbxtable.com](https://zbxtable.com)

## 许可证

ZbxTable 使用 Apache-2.0 许可证。详见 [LICENSE](LICENSE) 文件。
