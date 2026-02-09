[English](./README.md) | 简体中文

# ZbxTable

[![Build and Test](https://github.com/canghai908/zbxtable/actions/workflows/build.yml/badge.svg)](https://github.com/canghai908/zbxtable/actions/workflows/build.yml)
[![Release](https://github.com/canghai908/zbxtable/actions/workflows/release.yml/badge.svg)](https://github.com/canghai908/zbxtable/actions/workflows/release.yml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.23-blue)](https://golang.org/)

ZbxTable 是一个使用 Go 语言开发的 Zabbix 智能运维平台，提供强大的监控数据展示、告警管理和 AI 智能分析功能。

## ✨ 3.0 版本亮点

**ZbxTable 3.0** 是一个里程碑式的版本更新，带来了多项重大功能增强：

- 🤖 **AI 智能分析** - 集成 AI 大模型（Ollama/Deepseek），智能分析告警原因并提供解决方案
- 🌐 **多租户架构** - 支持同时管理多个 Zabbix 实例，实例级别的数据隔离和权限控制
- 📢 **增强的告警管理** - 全新的告警分发规则引擎，支持告警屏蔽、状态追踪和事件日志
- 💾 **数据库优化** - 完整支持 MySQL 和 PostgreSQL，使用 GORM 统一数据访问层
- 🔐 **安全性增强** - 敏感信息加密存储，系统自动生成加密密钥
- ⚡ **缓存系统重构** - 使用 go-cache 替代 Redis，简化部署架构
- 📝 **日志系统升级** - 按日期自动分割，自动压缩历史日志

## 功能特性

### 核心功能
- 🎨 **拓扑管理** - 自定义绘制拓扑图，实时状态显示
- 📊 **设备管理** - 设备分类展示和导出（Linux、Windows、网络设备、存储等）
- 📈 **报表系统** - 导出指定时间段内的 Zabbix 告警和指标数据到 Excel
- 🔍 **告警分析** - 分析特定时间段的告警数据，生成告警 Top 10 等统计
- 📅 **定时任务** - 定时报表生成和自动发送

### 告警管理
- 📧 **多渠道通知** - 支持邮件、企业微信、Webhook 等多种通知方式
- 🎯 **告警分发** - 灵活的告警规则引擎，支持多维度匹配条件
- 🔕 **告警屏蔽** - 支持时间段、主机、告警级别等多维度屏蔽规则
- 📊 **告警追踪** - 完整的告警状态追踪和事件日志记录
- 🤖 **AI 分析** - 智能分析告警原因，提供解决方案建议

### 系统管理
- 🌐 **多实例管理** - 同时管理多个 Zabbix 实例，统一运维平台
- 🔐 **权限控制** - 基于角色的访问控制（RBAC），细粒度权限管理
- ⚙️ **配置管理** - 数据库驱动的配置系统，支持运行时更新
- 📱 **菜单管理** - 动态菜单系统，基于角色的菜单权限控制
- 🔒 **安全加密** - 敏感信息自动加密存储

## 系统架构

![1](/zbxtable.png)

## 组件说明

**ZbxTable**: 后端使用 [Gin 框架](https://github.com/gin-gonic/gin) 和 Go 语言开发。

**ZbxTable-Web**: 前端使用 [Vue](https://github.com/vuejs/vue) 开发。

## 🛠️ 技术栈

### 后端技术
- **语言**: Go 1.23+
- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 5.7+ / PostgreSQL 12+
- **AI 集成**: Ollama / Deepseek API

### 前端技术
- **框架**: Vue.js 2.x
- **UI 组件**: Ant Design Vue
- **图表**: ECharts
- **拓扑图**: G6

### 告警接收
- **Webhook 方式**: ZbxTable 提供 Webhook API 接口，直接接收 Zabbix 告警
- **接口地址**: `http://your-zbxtable-server:8088/v1/receive`
- **认证方式**: Token 认证
- **支持格式**: JSON

在 Zabbix 中配置 Webhook 媒介类型，将告警发送到 ZbxTable 的 Webhook 接口即可实现告警接收和分发。详细配置请参考[官方文档](https://zbxtable.com)。

## 在线演示

[https://demo.zbxtable.com](https://demo.zbxtable.com)

**默认账号：** `admin / Zbxtable`

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

## ⚠️ 重要提示

**ZbxTable 3.0 与 2.x 版本数据库结构不兼容，建议全新安装。** 如果您正在使用 2.x 版本，请备份重要数据后重新安装 3.0 版本。

## 快速开始

### 方式一：一键安装脚本（推荐）

适用于 CentOS、Ubuntu、Debian 等 Linux 系统，自动完成所有配置。

```bash
# 在线安装
curl -fsSL https://dl.cactifans.com/zbxtable/install.sh | sudo bash

# 或者下载后安装
wget https://dl.cactifans.com/zbxtable/install.sh
sudo bash install.sh
```

**安装脚本功能：**
- ✅ 自动检测操作系统类型
- ✅ 自动下载最新版本
- ✅ 创建系统用户和目录
- ✅ 配置 systemd 服务
- ✅ 配置防火墙规则
- ✅ 自动启动服务

**安装后访问：**
- 首次安装：`http://your-server-ip:8088`
- 默认账号：`admin / Zbxtable`（请立即修改密码）

### 方式二：二进制部署

#### 1. 下载二进制文件

从 GitHub Releases 或镜像站下载：

```bash
# 从 GitHub 下载（以 v3.0.0 为例）
wget https://github.com/canghai908/zbxtable/releases/download/v3.0.0/zbxtable-v3.0.0-linux-amd64.tar.gz

# 或从镜像站下载
wget https://dl.cactifans.com/zbxtable/zbxtable-latest-linux-amd64.tar.gz

# 解压
tar -xzf zbxtable-v3.0.0-linux-amd64.tar.gz
cd zbxtable-v3.0.0-linux-amd64
```

#### 2. 创建数据库

**MySQL：**

```sql
CREATE DATABASE zbxtable CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'zbxtable'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON zbxtable.* TO 'zbxtable'@'localhost';
FLUSH PRIVILEGES;
```

**PostgreSQL：**

```sql
CREATE DATABASE zbxtable;
CREATE USER zbxtable WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE zbxtable TO zbxtable;
```

#### 3. 启动服务

**创建系统用户：**

```bash
# 创建 zbxtable 系统用户
sudo useradd -r -s /sbin/nologin zbxtable

# 创建安装目录
sudo mkdir -p /usr/local/zbxtable

# 复制二进制文件
sudo cp zbxtable /usr/local/zbxtable/

# 设置权限
sudo chown -R zbxtable:zbxtable /usr/local/zbxtable
sudo chmod +x /usr/local/zbxtable/zbxtable
```

**配置 systemd 服务：**

```bash
# 创建 systemd 服务文件
sudo tee /etc/systemd/system/zbxtable.service > /dev/null << EOF
[Unit]
Description=ZbxTable - Zabbix Monitoring Tool
After=network.target

[Service]
Type=simple
User=zbxtable
Group=zbxtable
WorkingDirectory=/usr/local/zbxtable
ExecStart=/usr/local/zbxtable/zbxtable web
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

# 重载 systemd
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start zbxtable

# 设置开机自启
sudo systemctl enable zbxtable

# 查看服务状态
sudo systemctl status zbxtable
```

#### 4. 访问安装向导

访问 `http://your-server-ip:8088`，按照安装向导配置数据库连接并完成安装。

## 📚 文档

- [3.0 版本发布说明](./RELEASE_3.0.md) - 详细的版本更新内容
- [系统说明文档](./docs/SYSTEM.md) - 完整的系统文档
- [API 文档](./docs/api/README.md) - API 接口文档
- [官方网站](https://zbxtable.com) - 更多文档和教程

## 📊 系统要求

### 最低配置

- **CPU**：2 核
- **内存**：4 GB
- **磁盘**：20 GB
- **操作系统**：CentOS 7/8、Ubuntu 18.04+、Debian 9+、RHEL 7/8

### 推荐配置

- **CPU**：4 核
- **内存**：8 GB
- **磁盘**：50 GB SSD
- **数据库**：MySQL 5.7+ 或 PostgreSQL 12+

### 网络要求

- 能够访问 Zabbix Server API
- 如使用邮件通知，需访问 SMTP 服务器
- 如使用企业微信，需访问企业微信 API
- 如使用 Deepseek API，需访问互联网

## 🔄 从 2.1 升级到 3.0

**⚠️ 重要：ZbxTable 3.0 与 2.x 版本数据库结构完全不兼容，无法直接升级。**

建议操作步骤：

1. **备份 2.x 版本的重要数据**（拓扑图、自定义配置等）
2. **全新安装 3.0 版本**（使用一键安装脚本或二进制部署）
3. **重新配置系统**（Zabbix 实例、告警规则等）

如果必须保留历史数据，请联系技术支持获取数据迁移方案。

## 🔒 安全建议

1. **修改默认密码** - 首次登录后立即修改 admin 密码
2. **使用 HTTPS** - 配置 Nginx 反向代理，启用 SSL 证书
3. **防火墙配置** - 仅开放必要端口，限制访问来源 IP
4. **定期备份** - 每天备份数据库和配置文件
5. **权限管理** - 使用最小权限原则，定期审查用户权限

## 📦 代码仓库

**ZbxTable**: [https://github.com/canghai908/zbxtable](https://github.com/canghai908/zbxtable)

**ZbxTable-Web**: [https://github.com/canghai908/zbxtable-web](https://github.com/canghai908/zbxtable-web)

## 👥 开发团队

**后端开发**: [canghai908](https://github.com/canghai908)

**前端开发**: [ahyiru](https://github.com/ahyiru)

## 🤝 贡献

欢迎贡献代码！请随时提交 Pull Request。

## 📞 支持

- **GitHub Issues**: [https://github.com/canghai908/zbxtable/issues](https://github.com/canghai908/zbxtable/issues)
- **官方文档**: [https://zbxtable.com](https://zbxtable.com)
- **在线演示**: [https://demo.zbxtable.com](https://demo.zbxtable.com)

## 📱 关注我们

扫码关注微信公众号，获取最新动态、技术文章和使用技巧：

<div align="center">
  <img src="docs/wechat.png" alt="微信公众号" width="400"/>
  <p>微信扫一扫关注公众号</p>
</div>

## 📄 许可证

ZbxTable 使用 Apache-2.0 许可证。详见 [LICENSE](LICENSE) 文件。

---

**ZbxTable 3.0 - 让 Zabbix 监控更智能、更强大！** 🚀
