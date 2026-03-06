# ZbxTable Docker 部署指南

本文档介绍如何在 `zbxtable` 项目中使用 Docker / Docker Compose 快速部署，并通过 `deployments/docker/docker.sh` 脚本进行日常运维。

## 目录

- [前置要求](#前置要求)
- [目录结构](#目录结构)
- [快速开始（推荐）](#快速开始推荐)
- [docker.sh 脚本命令](#dockersh-脚本命令)
- [手动 Docker Compose 命令](#手动-docker-compose-命令)
- [配置说明](#配置说明)
- [备份与恢复](#备份与恢复)
- [故障排查](#故障排查)

## 前置要求

- Docker 20.10+
- Docker Compose 2.0+（或兼容 `docker-compose` 命令）
- 至少 2GB 可用内存
- 至少 10GB 可用磁盘空间

## 目录结构

`zbxtable` 项目下与 Docker 相关的核心文件：

- `docker-compose.yml`
- `Dockerfile`
- `config/`（容器启动后会自动生成 `app.conf`）
- `deployments/docker/docker.sh`
- `deployments/docker/mysql/init.sql`

## 快速开始（推荐）

1. 进入项目目录

```bash
cd /Users/canghai/dev/code/zbxtable/zbxtable
```

2. 赋予脚本执行权限（首次）

```bash
chmod +x deployments/docker/docker.sh
```

3. 启动服务

```bash
./deployments/docker/docker.sh up
```

> 脚本会自动创建运行目录：`config`、`log`、`download`、`template`。
>
> 首次启动时，程序会自动生成 `config/app.conf`。

4. 查看日志

```bash
./deployments/docker/docker.sh logs zbxtable
```

5. 访问应用

- 地址：`http://localhost:8088`
- 首次访问会进入安装向导

## docker.sh 脚本命令

在项目根目录执行：

```bash
./deployments/docker/docker.sh <命令> [参数]
```

支持命令：

- `up`：启动服务（mysql + zbxtable，自动创建运行目录）
- `down`：停止并移除容器
- `restart`：重启服务
- `ps`：查看服务状态
- `logs [service]`：查看日志，支持指定服务名（如 `zbxtable` / `mysql`）
- `build`：重新构建镜像
- `rebuild`：强制重建并启动（自动创建运行目录 + `down + build --no-cache + up -d`）
- `pull`：拉取基础镜像
- `backup [文件路径]`：备份数据库（默认输出 `backup_时间.sql`）
- `restore <文件路径>`：从 SQL 文件恢复数据库
- `help`：查看帮助

## 手动 Docker Compose 命令

如需不经过脚本，可直接使用：

```bash
docker compose -f docker-compose.yml up -d
docker compose -f docker-compose.yml logs -f zbxtable
docker compose -f docker-compose.yml down
```

> 若你的环境仅支持 `docker-compose`，将 `docker compose` 替换为 `docker-compose` 即可。

## 配置说明

### 1) 端口

- 应用端口：`8088`
- 数据库端口：`3306`

### 2) 挂载目录

`docker-compose.yml` 默认挂载：

- `./config:/app/config`
- `./log:/app/log`
- `./download:/app/download`
- `./template:/app/template`

### 3) 配置文件

- `config/app.conf` 由程序启动后自动生成，无需手动创建。
- 如需连接 Compose 内置 MySQL，请确保自动生成后的配置中 `dbhost=mysql`、`dbport=3306`。

### 4) 时区

默认使用：`Asia/Shanghai`

## 备份与恢复

### 备份数据库

```bash
./deployments/docker/docker.sh backup
```

或指定路径：

```bash
./deployments/docker/docker.sh backup ./backup/zbxtable_$(date +%F).sql
```

### 恢复数据库

```bash
./deployments/docker/docker.sh restore ./backup/zbxtable_2026-03-06.sql
```

## 故障排查

### 1) 容器启动失败

```bash
./deployments/docker/docker.sh ps
./deployments/docker/docker.sh logs zbxtable
./deployments/docker/docker.sh logs mysql
```

### 2) 数据库连接失败

请检查：

- `config/app.conf` 中 `dbhost` 是否为 `mysql`
- `dbuser/dbpass/dbname` 是否与 `docker-compose.yml` 一致
- MySQL 容器是否正常运行

### 3) 配置文件说明

`config/app.conf` 由程序自动生成；如果未生成，请先查看应用日志定位原因：

```bash
./deployments/docker/docker.sh logs zbxtable
```

### 4) 端口冲突

如果 8088 或 3306 已被占用，请修改 `docker-compose.yml` 的端口映射。

---

如有问题请提交：

- GitHub Issues: https://github.com/canghai908/zbxtable/issues
