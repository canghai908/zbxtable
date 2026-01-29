# ZbxTable 快速开始指南

本指南将帮助您快速部署和使用 ZbxTable。

## 目录

- [环境要求](#环境要求)
- [Docker 快速部署](#docker-快速部署)
- [二进制快速部署](#二进制快速部署)
- [首次配置](#首次配置)
- [基本使用](#基本使用)
- [下一步](#下一步)

## 环境要求

### Docker 部署

- Docker 20.10+
- Docker Compose 2.0+
- 2GB+ 内存
- 10GB+ 磁盘空间

### 二进制部署

- Linux/Windows/macOS 操作系统
- MySQL 8.0+ 或 PostgreSQL 12+
- 1GB+ 内存
- 5GB+ 磁盘空间

## Docker 快速部署

### 1. 克隆项目

```bash
git clone https://github.com/canghai908/zbxtable.git
cd zbxtable/zbxtable
```

### 2. 准备配置文件

```bash
# 复制配置文件模板
cp config/app.conf.example config/app.conf

# 编辑配置文件（可选，使用默认配置也可以）
vim config/app.conf
```

**重要配置项**：

```ini
# 数据库配置（Docker 部署使用默认值即可）
dbtype    = mysql
dbhost    = mysql        # Docker 服务名
dbuser    = zbxtable
dbpass    = zbxtablepwd123
dbname    = zbxtable
dbport    = 3306

# 日志配置（使用默认值即可）
log_level = 6
log_path  =              # 留空，自动使用 ./log/yyyy-MM-dd.log
maxsize   = 100
maxdays   = 10
daily     = true
```

### 3. 启动服务

```bash
# 启动所有服务（MySQL + ZbxTable）
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f zbxtable
```

### 4. 访问应用

打开浏览器访问：`http://localhost:8085`

首次访问会自动跳转到安装向导页面。

## 二进制快速部署

### 1. 下载程序

```bash
# 下载最新版本
wget https://github.com/canghai908/zbxtable/releases/latest/download/zbxtable-linux-amd64.tar.gz

# 解压
tar -xzf zbxtable-linux-amd64.tar.gz
cd zbxtable
```

### 2. 准备数据库

```bash
# 登录 MySQL
mysql -uroot -p

# 创建数据库和用户
CREATE DATABASE zbxtable DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'zbxtable'@'localhost' IDENTIFIED BY 'zbxtablepwd123';
GRANT ALL PRIVILEGES ON zbxtable.* TO 'zbxtable'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

### 3. 配置应用

```bash
# 复制配置文件
cp config/app.conf.example config/app.conf

# 编辑配置文件
vim config/app.conf
```

修改数据库配置：

```ini
dbtype    = mysql
dbhost    = localhost
dbuser    = zbxtable
dbpass    = zbxtablepwd123
dbname    = zbxtable
dbport    = 3306
```

### 4. 启动服务

```bash
# 直接启动（前台运行）
./zbxtable web

# 或使用 nohup 后台运行
nohup ./zbxtable web > zbxtable.log 2>&1 &

# 或使用 systemd（推荐）
sudo cp zbxtable.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable zbxtable
sudo systemctl start zbxtable
sudo systemctl status zbxtable
```

### 5. 访问应用

打开浏览器访问：`http://localhost:8085`

## 首次配置

### 1. 安装向导

首次访问会进入安装向导，按照以下步骤完成安装：

#### 步骤 1：环境检查

系统会自动检查环境配置，确保所有项目都通过。

#### 步骤 2：数据库配置

- 数据库类型：MySQL 或 PostgreSQL
- 数据库地址：localhost 或 IP 地址
- 数据库端口：3306（MySQL）或 5432（PostgreSQL）
- 数据库名称：zbxtable
- 数据库用户：zbxtable
- 数据库密码：您设置的密码

点击"测试连接"确保数据库连接正常。

#### 步骤 3：Zabbix 配置

- Zabbix Web 地址：如 `http://zabbix.example.com`
- Zabbix 用户名：Admin（或其他有权限的用户）
- Zabbix 密码：您的 Zabbix 密码
- 或使用 API Token（Zabbix 5.4+）

点击"测试连接"确保 Zabbix API 连接正常。

#### 步骤 4：管理员账号

- 用户名：admin（建议）
- 密码：设置一个强密码
- 确认密码：再次输入密码
- 邮箱：管理员邮箱

#### 步骤 5：完成安装

点击"完成安装"，系统会自动初始化数据库并创建管理员账号。

### 2. 登录系统

安装完成后，使用刚才创建的管理员账号登录系统。

## 基本使用

### 1. 添加 Zabbix 实例（可选）

如果您有多个 Zabbix 实例，可以添加更多实例：

1. 进入"系统管理" → "Zabbix 实例管理"
2. 点击"添加实例"
3. 填写实例信息：
   - 实例名称：如"生产环境"
   - API 地址：Zabbix Web 地址
   - 用户名/密码 或 API Token
4. 点击"保存"

### 2. 查看主机列表

1. 进入"主机管理" → "主机列表"
2. 系统会自动从 Zabbix 同步主机信息
3. 可以按主机组、状态等条件筛选

### 3. 生成报表

#### 主机性能报表

1. 进入"报表管理" → "主机报表"
2. 选择时间范围（如最近 7 天）
3. 选择主机或主机组
4. 选择监控项（CPU、内存、磁盘等）
5. 点击"生成报表"
6. 下载 Excel 文件

#### 告警报表

1. 进入"报表管理" → "告警报表"
2. 选择时间范围
3. 选择主机或主机组（可选）
4. 选择告警级别（可选）
5. 点击"生成报表"
6. 下载 Excel 文件

### 4. 查看告警统计

1. 进入"告警分析" → "告警统计"
2. 选择时间范围
3. 查看：
   - 告警趋势图
   - 告警主机 Top 10
   - 告警项 Top 10
   - 按严重程度统计

### 5. 创建拓扑图

1. 进入"拓扑管理" → "新建拓扑"
2. 输入拓扑名称
3. 使用工具栏添加节点：
   - 拖拽添加设备节点
   - 关联 Zabbix 主机
   - 添加连接线
   - 添加文本标签
4. 保存拓扑
5. 查看拓扑时可以看到设备实时状态

### 6. 配置定时任务

1. 进入"任务管理" → "新建任务"
2. 填写任务信息：
   - 任务名称：如"每日性能报表"
   - 报表类型：主机报表/告警报表
   - 时间范围：最近 1 天
   - 执行时间：Cron 表达式，如 `0 9 * * *`（每天 9:00）
   - 通知方式：邮件/企业微信
   - 接收人：邮箱地址或企业微信用户
3. 启用任务
4. 查看任务日志了解执行情况

## 下一步

### 配置邮件通知

1. 进入"系统管理" → "系统配置" → "邮件配置"
2. 填写 SMTP 服务器信息：
   - SMTP 服务器：如 `smtp.gmail.com`
   - SMTP 端口：465（SSL）或 25
   - 发件人邮箱
   - 发件人密码
3. 点击"测试邮件"确保配置正确
4. 保存配置

### 配置企业微信通知

1. 进入"系统管理" → "系统配置" → "企业微信配置"
2. 填写企业微信信息：
   - 企业 ID
   - 应用 ID
   - 应用 Secret
3. 点击"测试消息"确保配置正确
4. 保存配置

### 用户管理

1. 进入"系统管理" → "用户管理"
2. 添加新用户：
   - 用户名
   - 密码
   - 邮箱
   - 角色（管理员/普通用户）
3. 设置用户权限

### 查看日志

日志文件位于 `./log/` 目录下，按日期命名：

```bash
# 查看今天的日志
tail -f log/2024-01-29.log

# 查看所有日志文件
ls -lh log/

# 旧日志会自动压缩
ls -lh log/*.gz
```

## 常见问题

### 1. 无法访问页面

- 检查服务是否启动：`docker-compose ps` 或 `systemctl status zbxtable`
- 检查端口是否被占用：`netstat -tunlp | grep 8085`
- 检查防火墙设置

### 2. 数据库连接失败

- 检查数据库服务是否启动
- 检查数据库配置是否正确
- 测试数据库连接：`mysql -h host -u user -p`

### 3. Zabbix API 连接失败

- 检查 Zabbix Web 地址是否正确
- 检查用户名和密码是否正确
- 确认 Zabbix 用户有足够的权限
- 检查网络连接

### 4. 报表生成失败

- 检查时间范围是否合理
- 确认选择的主机有监控数据
- 查看日志文件了解详细错误
- 检查磁盘空间是否充足

## 获取帮助

- 📖 [完整文档](./SYSTEM.md)
- 🐳 [Docker 部署指南](../deployments/docker/README.md)
- 🔧 [API 文档](./api/README.md)
- 🌐 [官方网站](https://zbxtable.com)
- 💬 [GitHub Issues](https://github.com/canghai908/zbxtable/issues)

## 下一步学习

1. 阅读[系统说明文档](./SYSTEM.md)了解更多功能
2. 查看[API 文档](./api/README.md)学习如何使用 API
3. 参考[Docker 部署指南](../deployments/docker/README.md)了解生产环境部署
4. 访问[官方网站](https://zbxtable.com)获取更多教程

祝您使用愉快！

