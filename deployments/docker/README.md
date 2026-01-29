# ZbxTable Docker 部署指南

本文档介绍如何使用 Docker 和 Docker Compose 部署 ZbxTable。

## 目录

- [前置要求](#前置要求)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [常用命令](#常用命令)
- [故障排查](#故障排查)

## 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少 2GB 可用内存
- 至少 10GB 可用磁盘空间

## 快速开始

### 方式一：使用 Docker Compose（推荐）

1. **克隆项目**

```bash
git clone https://github.com/canghai908/zbxtable.git
cd zbxtable/zbxtable
```

2. **准备配置文件**

```bash
# 复制配置文件模板
cp config/app.conf.example config/app.conf

# 编辑配置文件，修改数据库连接信息
vim config/app.conf
```

配置示例：

```ini
; zbxtable
appname   = zbxtable
httpport  = 8085
runmode   = prod
timeout   = 12
token     = ced46eae0ffa411f8e7bbec90cc6e68d
log_level = 6
; 日志路径，留空则默认为 ./log/yyyy-MM-dd.log
log_path  = 
maxlines  = 1000
maxsize   = 100
maxdays   = 10
daily     = true

; database
dbtype    = mysql
dbhost    = mysql
dbuser    = zbxtable
dbpass    = zbxtablepwd123
dbname    = zbxtable
dbport    = 3306
```

3. **启动服务**

```bash
# 启动所有服务（MySQL + ZbxTable）
docker-compose up -d

# 查看日志
docker-compose logs -f zbxtable
```

4. **访问应用**

打开浏览器访问：`http://localhost:8085`

首次访问会进入安装向导，按照提示完成安装。

### 方式二：仅使用 Docker

1. **构建镜像**

```bash
# 在项目根目录下构建
docker build -t zbxtable:latest -f zbxtable/Dockerfile .
```

2. **运行容器**

```bash
docker run -d \
  --name zbxtable \
  -p 8085:8085 \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/log:/app/log \
  -v $(pwd)/download:/app/download \
  -v $(pwd)/template:/app/template \
  -e TZ=Asia/Shanghai \
  zbxtable:latest
```

## 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| TZ | 时区设置 | Asia/Shanghai |

### 数据卷

| 容器路径 | 说明 | 建议挂载 |
|----------|------|----------|
| /app/config | 配置文件目录 | 是 |
| /app/log | 日志文件目录 | 是 |
| /app/download | 下载文件目录 | 是 |
| /app/template | 模板文件目录 | 可选 |

### 端口

| 端口 | 说明 |
|------|------|
| 8085 | Web 服务端口 |

## 常用命令

### Docker Compose 命令

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose stop

# 重启服务
docker-compose restart

# 查看日志
docker-compose logs -f zbxtable

# 查看服务状态
docker-compose ps

# 停止并删除容器
docker-compose down

# 停止并删除容器及数据卷
docker-compose down -v
```

### Docker 命令

```bash
# 查看容器日志
docker logs -f zbxtable

# 进入容器
docker exec -it zbxtable sh

# 重启容器
docker restart zbxtable

# 停止容器
docker stop zbxtable

# 删除容器
docker rm zbxtable

# 查看容器状态
docker ps -a | grep zbxtable
```

## 数据库初始化

如果使用 Docker Compose，MySQL 会自动初始化。如果使用外部数据库，需要手动创建数据库：

```sql
CREATE DATABASE IF NOT EXISTS zbxtable DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'zbxtable'@'%' IDENTIFIED BY 'zbxtablepwd123';
GRANT ALL PRIVILEGES ON zbxtable.* TO 'zbxtable'@'%';
FLUSH PRIVILEGES;
```

## 升级

### 使用 Docker Compose 升级

```bash
# 拉取最新代码
git pull

# 重新构建镜像
docker-compose build

# 重启服务
docker-compose up -d

# 查看日志确认启动成功
docker-compose logs -f zbxtable
```

### 使用 Docker 升级

```bash
# 停止旧容器
docker stop zbxtable
docker rm zbxtable

# 重新构建镜像
docker build -t zbxtable:latest -f zbxtable/Dockerfile .

# 启动新容器
docker run -d \
  --name zbxtable \
  -p 8085:8085 \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/log:/app/log \
  -v $(pwd)/download:/app/download \
  zbxtable:latest
```

## 故障排查

### 1. 容器无法启动

```bash
# 查看容器日志
docker logs zbxtable

# 检查配置文件
cat config/app.conf

# 检查端口占用
netstat -tunlp | grep 8085
```

### 2. 无法连接数据库

```bash
# 检查数据库容器状态
docker-compose ps mysql

# 测试数据库连接
docker exec -it zbxtable-mysql mysql -uzbxtable -pzbxtablepwd123 -e "SELECT 1"

# 检查网络连接
docker network inspect zbxtable_zbxtable-network
```

### 3. 日志文件过大

日志文件会自动按天分割并压缩，默认保留 10 天。可以在 `config/app.conf` 中调整：

```ini
maxdays   = 7    ; 保留7天
maxsize   = 100  ; 单个文件最大100MB
```

### 4. 查看健康检查状态

```bash
# 查看容器健康状态
docker inspect --format='{{.State.Health.Status}}' zbxtable

# 查看健康检查日志
docker inspect --format='{{range .State.Health.Log}}{{.Output}}{{end}}' zbxtable
```

## 性能优化

### 1. 调整 MySQL 配置

编辑 `docker-compose.yml`，在 mysql 服务的 command 部分添加：

```yaml
command:
  - --character-set-server=utf8mb4
  - --collation-server=utf8mb4_unicode_ci
  - --max_connections=1000
  - --innodb_buffer_pool_size=1G
```

### 2. 调整应用日志级别

在生产环境建议将日志级别设置为 1（Info）或 2（Error）：

```ini
log_level = 1
```

## 安全建议

1. **修改默认密码**：修改 `docker-compose.yml` 中的数据库密码
2. **使用 HTTPS**：建议在前端配置 Nginx 反向代理并启用 HTTPS
3. **限制端口访问**：使用防火墙限制 8085 端口的访问
4. **定期备份**：定期备份数据库和配置文件

## 备份与恢复

### 备份

```bash
# 备份数据库
docker exec zbxtable-mysql mysqldump -uzbxtable -pzbxtablepwd123 zbxtable > backup.sql

# 备份配置文件
tar -czf config_backup.tar.gz config/

# 备份数据卷
docker run --rm -v zbxtable_mysql_data:/data -v $(pwd):/backup alpine tar -czf /backup/mysql_data_backup.tar.gz /data
```

### 恢复

```bash
# 恢复数据库
docker exec -i zbxtable-mysql mysql -uzbxtable -pzbxtablepwd123 zbxtable < backup.sql

# 恢复配置文件
tar -xzf config_backup.tar.gz
```

## 支持

如有问题，请访问：

- GitHub Issues: https://github.com/canghai908/zbxtable/issues
- 官方文档: https://zbxtable.com

## 许可证

Apache-2.0 License

