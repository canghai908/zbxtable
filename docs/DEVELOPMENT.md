# 本地开发与编译指南

本文档介绍如何在本地搭建 ZbxTable 前后端开发环境、进行联调测试，以及从源码编译单二进制可执行文件、跨平台发布包和 Docker 镜像。

## 目录

- [1. 环境要求](#1-环境要求)
- [2. 仓库与目录结构](#2-仓库与目录结构)
- [3. 本地开发与前后端联调](#3-本地开发与前后端联调)
- [4. 源码编译与打包发布](#4-源码编译与打包发布)
- [5. 代码规范检查与测试](#5-代码规范检查与测试)
- [6. 常见问题排查](#6-常见问题排查)

---

## 1. 环境要求

| 组件 | 版本要求 | 说明 |
| :--- | :--- | :--- |
| **Go** | `1.26+` | 以 `zbxtable/go.mod` 声明版本为准 |
| **Node.js** | `16.x ~ 20.x` | Vue CLI 4 构建依赖；Node.js `17+` 需启用 `--openssl-legacy-provider`（`yarn build` 已内置） |
| **Yarn** | `1.22+`（Classic） | 推荐使用 `yarn` 管理前端依赖（也可使用 `npm`） |
| **数据库** | MySQL `5.7+ / 8.0+` 或 PostgreSQL `12+` | 本地开发任选其一即可 |
| **构建工具** | Git、GNU Make、Bash | 执行 `Makefile` 目标与打包脚本 |

---

## 2. 仓库与目录结构

ZbxTable 采用前后端双仓库并列结构，建议在同一工作区目录下并排克隆两个仓库（`3.0` 分支）：

```bash
mkdir -p zbxtable-workspace && cd zbxtable-workspace

# 克隆后端仓库（3.0 分支）
git clone -b 3.0 https://github.com/canghai908/zbxtable.git zbxtable

# 克隆前端仓库（3.0 分支）
git clone -b 3.0 https://github.com/canghai908/zbxtable-web.git zbxtable-web
```

> **注意**：`zbxtable/scripts/build-linux-amd64.sh` 与 `zbxtable/Dockerfile` 默认按上述同级目录（`../zbxtable-web`）查找前端工程。

### 核心目录说明

```text
zbxtable-workspace/
├── zbxtable/                    # Go 后端仓库
│   ├── main.go                  # 程序入口与版本信息注入变量
│   ├── cmd/web.go               # CLI web 子命令、启动流程与配置读取
│   ├── api/v1/                  # Gin 路由注册与中间件挂载
│   │   └── web/                 # go:embed 前端静态资源挂载包（api/v1/web/web/ 为构建时生成）
│   ├── internal/                # 核心业务逻辑（handler / middleware / model）
│   ├── pkg/                     # 公共组件（logger / assets / otel / utils 等）
│   ├── config/                  # 本地运行配置（app.conf，已加入 .gitignore）
│   ├── scripts/                 # 跨平台完整构建脚本
│   ├── deployments/docker/      # Docker Compose 部署与运维脚本
│   └── Makefile                 # 常用开发、测试、编译与打包命令
│
└── zbxtable-web/                # Vue 2 前端仓库
    ├── src/
    │   ├── pages/               # 业务页面
    │   ├── components/          # 公共组件
    │   ├── services/            # 后端 API 请求封装
    │   ├── router/              # 路由与动态菜单配置
    │   └── store/               # Vuex 状态管理
    ├── .env.development         # 本地开发后端代理地址配置
    ├── vue.config.js            # Webpack 构建、开发代理与版本号注入配置
    └── web/                     # 前端生产构建输出目录（yarn build 生成）
```

---

## 3. 本地开发与前后端联调

日常开发推荐采用**前后端分离热启动**模式：后端独立监听 `:8088` 提供 API 与 WebSocket，前端通过 `yarn serve` 监听 `:8080` 并将接口请求代理到后端。

### 3.1 准备 `go:embed` 静态目录占位（首次克隆必看）

后端 `zbxtable/api/v1/web/embed.go` 使用了 `//go:embed all:web`，而 `api/v1/web/web/` 目录被 `.gitignore` 忽略。首次克隆后端仓库后，若尚未编译并同步前端产物，直接运行 `go run`、`go build` 或 `go test` 会提示 `pattern all:web: no matching files found`。

首次本地开发后端前，可先创建一个最小占位文件（或参考 [4.2 节](#42-编译包含前端页面的单二进制本机平台) 先构建一次前端）：

```bash
cd zbxtable
mkdir -p api/v1/web/web
[ -f api/v1/web/web/index.html ] || echo "<!doctype html><title>ZbxTable Dev</title>" > api/v1/web/web/index.html
```

### 3.2 准备数据库

在本地创建空数据库（表结构会在安装完成或后端启动时通过 GORM 自动迁移）：

**MySQL：**

```sql
CREATE DATABASE zbxtable CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

**PostgreSQL：**

```sql
CREATE DATABASE zbxtable;
```

### 3.3 配置后端（两种方式任选其一）

#### 方式 A：通过安装向导自动生成（推荐首次搭建）

无需手动编写配置文件。当 `./config/app.conf` 不存在时，后端启动会自动进入**安装向导模式**。等前后端启动后访问 `http://127.0.0.1:8080/install`（或已嵌入前端时访问 `http://127.0.0.1:8088/install`），按页面提示填写数据库信息即可自动生成 `config/app.conf` 并初始化默认管理员账号（`admin / Zbxtable`）。

#### 方式 B：手动创建 `config/app.conf`

在 `zbxtable/config/app.conf` 中创建配置文件（切勿提交到 Git）：

```ini
httpport = 8088
runmode  = dev
timeout  = 12

log_level = 3
log_path  = log
maxlines  = 1000
maxsize   = 0
maxdays   = 10
daily     = true

; database: mysql 或 postgresql
dbtype = mysql
dbhost = 127.0.0.1
dbport = 3306
dbuser = root
dbpass = your_password
dbname = zbxtable
```

> **配置优先级说明**：
> 1. **端口覆盖**：环境变量 `ZBXTABLE_HTTP_PORT` 优先级高于配置文件中的 `httpport`。
> 2. **`.env` 覆盖规则**：若 `zbxtable/.env` 文件存在，`GetConfKey` 会**完全从 `.env` 读取配置，不再回退读取 `config/app.conf`**（但启动安装状态检查仍依赖 `config/app.conf` 中的 `dbtype`/`dbhost`/`dbname` 非空）。本地开发建议统一使用 `config/app.conf`。

### 3.4 启动后端服务

```bash
cd zbxtable

# 下载并整理 Go 模块依赖
make install

# 方式 1：编译并启动（输出到 build/zbxtable 并执行 web 子命令）
make run

# 方式 2：直接使用 go run 启动
go run main.go web

# 方式 3：使用 air 热重载开发（未安装 air 时会自动执行 go install）
make dev
```

后端默认监听 `http://127.0.0.1:8088`。

### 3.5 启动前端开发服务器

打开新终端窗口进入 `zbxtable-web`：

```bash
cd zbxtable-web

# 安装依赖
yarn install

# 启动本地热更新开发服务器
yarn serve
```

> **Node.js 17+ / 18 / 20 提示**：如果本地 Node.js 版本 >= 17，执行 `yarn serve` 报 `ERR_OSSL_EVP_UNSUPPORTED`（`digital envelope routines::unsupported`），请携带环境变量启动：
>
> ```bash
> NODE_OPTIONS=--openssl-legacy-provider yarn serve
> ```

启动成功后访问 `http://127.0.0.1:8080`：
- `vue.config.js` 已配置开发代理，自动将 `/v1`、`/download`、`/public`、`/upload`、`/install` 请求及 `/ws` WebSocket 连接转发至 `.env.development` 中的 `VUE_APP_API_BASE_URL`（默认 `http://127.0.0.1:8088`）。
- 如修改了后端监听端口，同步调整 `zbxtable-web/.env.development` 中的 `VUE_APP_API_BASE_URL` 并重启 `yarn serve` 即可。

---

## 4. 源码编译与打包发布

ZbxTable 生产部署采用**单二进制内嵌资源**架构：
1. 前端构建产物（`zbxtable-web/web/`）复制到 `zbxtable/api/v1/web/web/`，编译时通过 `//go:embed` 打包进 Go 二进制；
2. 报表模板与脚本资源（`zbxtable/pkg/assets/files/`）同样内嵌于二进制，进程启动时自动释放至运行目录 `assets/`。

### 4.1 单独编译前端

```bash
cd zbxtable-web
yarn install
yarn build
```

- `package.json` 的 `build` 脚本已内置 `NODE_OPTIONS=--openssl-legacy-provider`。
- 构建产物输出到 `zbxtable-web/web/`（含 `index.html`、`static/`、`css/` 及 Gzip 预压缩资源）。
- **前端版本号解析优先级**（定义于 `vue.config.js`）：
  1. 环境变量 `VERSION`（如 `VERSION=v3.0.8 yarn build`）
  2. 当前提交精确匹配的 Git Tag（`git describe --tags --exact-match`）
  3. 最近 Tag + 提交偏移（`git describe --tags --always`）
  4. `package.json` 中的 `version` 字段

### 4.2 编译包含前端页面的单二进制（本机平台）

在完成前端构建后，将静态资源同步到后端嵌入目录并执行编译：

```bash
# 1. 构建前端
cd zbxtable-web
yarn install && yarn build

# 2. 同步前端产物到后端 go:embed 目录
cd ../zbxtable
rm -rf api/v1/web/web
mkdir -p api/v1/web/web
cp -R ../zbxtable-web/web/. api/v1/web/web/

# 3. 编译当前系统架构二进制
make build
```

编译产物位于 `zbxtable/build/zbxtable`，可通过以下命令验证注入的版本信息并运行：

```bash
./build/zbxtable -v
./build/zbxtable web
```

> `Makefile` 编译时会自动通过 `-ldflags` 注入 `main.version`（当前精确 Git Tag 或 `dev`）、`main.gitHash`（短提交哈希）和 `main.buildTime`（UTC 构建时间）。

### 4.3 交叉编译 Linux 二进制与发布归档包

#### 一键构建 Linux AMD64（前端构建 + 资源同步 + 后端交叉编译）

仓库内置了完整构建脚本 `scripts/build-linux-amd64.sh`，可在 macOS / Linux 开发机上一键生成内嵌最新前端的 Linux `amd64` 可执行文件：

```bash
cd zbxtable
make build-linux-amd64
# 输出产物：zbxtable/dist/zbxtable-linux-amd64
```

#### 构建多平台二进制与 `.tar.gz` 发布包

若需同时编译 `linux/amd64` 与 `linux/arm64`（需先按 [4.2 节](#42-编译包含前端页面的单二进制本机平台) 同步好 `api/v1/web/web/`）：

```bash
cd zbxtable

# 仅交叉编译所有目标平台二进制（输出至 dist/zbxtable-linux-amd64、dist/zbxtable-linux-arm64）
make build-all

# 编译并打包为发布归档（输出至 dist/releases/*.tar.gz）
make package
```

### 4.4 本地构建 Docker 镜像

`zbxtable/Dockerfile` 采用多阶段构建（依次构建 `zbxtable-web` 前端、`zbxtable` 后端及 Alpine 运行时镜像），构建上下文需指向包含两个仓库的父级工作区目录：

```bash
cd zbxtable

# 使用 Makefile 构建镜像（内部执行 docker build -f Dockerfile ..）
make docker-build

# 或使用 Docker 运维脚本启动完整 MySQL + ZbxTable 容器环境
./deployments/docker/docker.sh up
```

更多容器化部署与备份恢复命令请参考 [Docker 部署指南](../deployments/docker/README.md)。

### 4.5 Makefile 常用命令速查

在 `zbxtable/` 目录下执行：

| 命令 | 说明 |
| :--- | :--- |
| `make help` | 查看所有支持的 Make 目标说明 |
| `make install` | 执行 `go mod download` 与 `go mod tidy` |
| `make build` | 编译当前平台二进制到 `build/zbxtable` |
| `make run` | 编译并启动 `./build/zbxtable web` |
| `make dev` | 使用 `air` 启动热重载开发模式 |
| `make build-linux-amd64` | 自动编译前端、同步静态资源并交叉编译 `dist/zbxtable-linux-amd64` |
| `make build-all` | 交叉编译 `linux/amd64` 和 `linux/arm64` 二进制到 `dist/` |
| `make package` | 生成多平台 `.tar.gz` 发布压缩包到 `dist/releases/` |
| `make fmt` | 使用 `gofmt -w .` 格式化 Go 代码 |
| `make vet` | 运行 `go vet ./...` 静态检查 |
| `make lint` | 运行 `golangci-lint run ./...`（未安装时自动安装） |
| `make test` | 运行单元测试（`go test -v -race ./...`） |
| `make coverage` | 生成测试覆盖率报告 `coverage.out` 与 `coverage.html` |
| `make quality` | 依次执行 `fmt`、`vet`、`lint`、`test` 全量质量检查 |
| `make clean` | 清理 `build/`、`dist/` 及测试覆盖率产物 |

---

## 5. 代码规范检查与测试

提交代码或发起 Pull Request 前，请确保对应仓库通过以下检查：

### 5.1 后端检查与测试

```bash
cd zbxtable

# 格式化与静态分析
make fmt
make vet

# 运行全部单元测试（含竞态检测）
make test

# 运行指定包或单个测试函数
go test -v ./internal/model -run TestDispatchRuleWorkflow
```

- **时区敏感测试**：告警屏蔽与值班排班相关规则建议在不同 `TZ` 下验证（与 CI 一致）：
  ```bash
  TZ=UTC go test ./internal/model -run 'Test(MuteRuleConditionTimeChannel|DispatchRuleWorkflow|DispatchOvernightAndBoundaries)$' -count=1
  ```
- **数据库集成测试（可选）**：
  - 通用数据库探针测试：设置 `ZBXTABLE_TEST_DBTYPE`、`ZBXTABLE_TEST_DBHOST`、`ZBXTABLE_TEST_DBPORT`、`ZBXTABLE_TEST_DBUSER`、`ZBXTABLE_TEST_DBPASS`、`ZBXTABLE_TEST_DBNAME` 后运行 `go test -v ./internal/model -run TestConfigValueTextProbe`（未设置时自动跳过）。
  - 值班与告警分发 PostgreSQL 集成测试：设置 `ZBXTABLE_ONCALL_TEST_DSN`，详见 [值班分发集成测试说明](./testing/oncall-dispatch-integration.md)。

### 5.2 前端检查与构建验证

```bash
cd zbxtable-web

# ESLint 代码规范检查
yarn lint

# 生产构建校验
yarn build
```

---

## 6. 常见问题排查

### Q1：执行 `go build` 或 `go test` 报错 `api/v1/web/embed.go:9:12: pattern all:web: no matching files found`
**原因**：`api/v1/web/web/` 目录不存在或为空（该目录不纳入 Git 版本控制）。  
**解决**：
- 仅调试后端或跑单测：执行 `mkdir -p api/v1/web/web && echo "dev" > api/v1/web/web/index.html`；
- 需要完整前端页面：按 [4.2 节](#42-编译包含前端页面的单二进制本机平台) 构建 `zbxtable-web` 并复制 `web/` 到 `api/v1/web/web/`。

### Q2：Node.js 18 / 20 下运行 `yarn serve` 报错 `error:0308010C:digital envelope routines::unsupported`
**原因**：Webpack 4 使用的旧版哈希算法在 Node.js 17+ 的 OpenSSL 3 中默认禁用。  
**解决**：使用 `NODE_OPTIONS=--openssl-legacy-provider yarn serve` 启动开发服务器（`yarn build` 已在 `package.json` 中内置该参数）。

### Q3：修改了前端代码，刷新页面没有变化
**原因**：可能直接访问了后端端口 `http://127.0.0.1:8088`（该端口返回的是上次编译进后端二进制的静态文件）。  
**解决**：本地联调时请访问前端开发服务器 `http://127.0.0.1:8080`；若要验证 `:8088` 单二进制效果，需重新执行 `yarn build`、同步 `api/v1/web/web/` 并重新 `make build`。

### Q4：修改 `config/app.conf` 后配置未生效
**原因**：检查 `zbxtable/` 根目录下是否存在 `.env` 文件。一旦存在 `.env`，`GetConfKey` 将只读取 `.env` 中的键值，不会回退读取 `config/app.conf`。  
**解决**：删除多余的 `.env` 文件，或将所需配置项完整迁移至 `.env` 后重启后端进程。

### Q5：本地 `8088` 端口被其他进程占用
**原因**：后端默认监听 `8088`。  
**解决**：启动后端时指定 `ZBXTABLE_HTTP_PORT=8089 go run main.go web`（或修改 `config/app.conf` 中的 `httpport`），同时将 `zbxtable-web/.env.development` 中的 `VUE_APP_API_BASE_URL` 改为 `http://127.0.0.1:8089` 并重启 `yarn serve`。
