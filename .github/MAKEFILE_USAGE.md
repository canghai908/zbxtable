# ✅ 工作流更新：使用 Makefile 编译

## 📝 更新说明

已更新 GitHub Actions 工作流，使用项目的 Makefile 进行编译和测试，而不是直接使用 Go 命令。

---

## 🔧 主要变更

### Build 工作流 (build.yml)

**之前：**
```yaml
- name: 下载依赖
  run: |
    go mod download
    go mod verify

- name: 代码静态检查
  run: go vet ./...

- name: 运行测试
  run: go test -v -race -coverprofile=coverage.out ./...

- name: 编译检查
  run: |
    GOOS=linux GOARCH=amd64 go build -o /dev/null main.go
    GOOS=windows GOARCH=amd64 go build -o /dev/null main.go
    GOOS=darwin GOARCH=arm64 go build -o /dev/null main.go
```

**现在：**
```yaml
- name: 安装依赖
  run: make install

- name: 代码静态检查
  run: make vet

- name: 运行测试
  run: make test

- name: 生成测试覆盖率
  run: make coverage

- name: 编译应用
  run: make build

- name: 验证编译产物
  run: |
    if [ -f build/zbxtable ]; then
      echo "✅ 编译成功"
      ls -lh build/zbxtable
    else
      echo "❌ 编译失败"
      exit 1
    fi
```

### Release 工作流 (release.yml)

**之前：**
```yaml
- name: 编译多平台二进制文件
  run: |
    # 手动设置版本信息和 LDFLAGS
    VERSION=${{ steps.version.outputs.version }}
    GIT_HASH=$(git rev-parse --short HEAD)
    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    LDFLAGS="-X main.version=${VERSION} ..."
    
    # 手动编译每个平台
    GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" ...
    GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" ...
    # ... 更多平台

- name: 创建发布包
  run: |
    # 手动创建发布包
    for platform in "${platforms[@]}"; do
      # 复制文件、打包等
    done
```

**现在：**
```yaml
- name: 编译所有平台并打包
  run: |
    echo "📦 开始编译所有平台..."
    make build-all
    
    echo "📦 创建发布包..."
    make package
    
    echo "✅ 编译和打包完成！"
    ls -lh dist/releases/
```

---

## ✨ 优势

### 1. 代码复用
- ✅ 本地开发和 CI/CD 使用相同的构建命令
- ✅ 避免重复的构建逻辑
- ✅ 更容易维护

### 2. 版本信息自动化
- ✅ Makefile 自动从 git 获取版本信息
- ✅ 自动设置 LDFLAGS
- ✅ 无需在工作流中手动配置

### 3. 简化工作流
- ✅ 工作流文件更简洁
- ✅ 更容易理解和修改
- ✅ 减少出错可能

### 4. 统一构建流程
- ✅ 本地：`make build`
- ✅ CI：`make build`
- ✅ 发布：`make build-all && make package`

---

## 📋 Makefile 命令说明

### 开发常用命令

```bash
# 安装依赖
make install

# 格式化代码
make fmt

# 静态检查
make vet

# 运行测试
make test

# 生成测试覆盖率
make coverage

# 编译当前平台
make build

# 运行应用
make run
```

### CI/CD 使用命令

```bash
# Build 工作流使用
make install    # 安装依赖
make vet        # 静态检查
make test       # 运行测试
make coverage   # 生成覆盖率
make build      # 编译应用

# Release 工作流使用
make build-all  # 编译所有平台
make package    # 创建发布包
```

### 其他有用命令

```bash
# 清理构建产物
make clean

# 代码质量检查（格式化+静态检查+lint+测试）
make quality

# 查看版本信息
make version

# 查看帮助
make help
```

---

## 🔍 Makefile 关键配置

### 版本信息自动获取

```makefile
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.gitHash=$(GIT_HASH) -X main.buildTime=$(BUILD_TIME) -w -s"
```

### 多平台编译

```makefile
PLATFORMS := linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64

build-all:
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-$${platform%/*}-$${platform#*/} main.go; \
	done
```

### 打包发布

```makefile
package: build-all
	@for platform in $(PLATFORMS); do \
		# 创建发布包目录
		# 复制二进制文件、配置文件、文档
		# 创建 tar.gz 压缩包
	done
	# 生成 SHA256 校验和
```

---

## 🎯 工作流程对比

### Build 工作流

**之前：**
```
检出代码 → 设置环境 → 构建前端 → 复制前端
  ↓
go mod download → go mod verify
  ↓
gofmt 检查 → go vet → go test
  ↓
手动编译多个平台验证
```

**现在：**
```
检出代码 → 设置环境 → 构建前端 → 复制前端
  ↓
make install
  ↓
gofmt 检查 → make vet → make test → make coverage
  ↓
make build → 验证编译产物
```

### Release 工作流

**之前：**
```
检出代码 → 设置环境 → 构建前端 → 复制前端
  ↓
手动设置版本变量和 LDFLAGS
  ↓
手动编译 5 个平台（重复代码）
  ↓
手动创建发布包（复杂的 shell 脚本）
  ↓
生成校验和 → 创建 Release
```

**现在：**
```
检出代码 → 设置环境 → 构建前端 → 复制前端
  ↓
make build-all（自动编译所有平台）
  ↓
make package（自动创建发布包）
  ↓
创建 Release
```

---

## 📂 构建产物

### make build
```
build/
└── zbxtable          # 当前平台的可执行文件
```

### make build-all
```
dist/
├── zbxtable-linux-amd64
├── zbxtable-linux-arm64
├── zbxtable-windows-amd64.exe
├── zbxtable-darwin-amd64
└── zbxtable-darwin-arm64
```

### make package
```
dist/
├── releases/
│   ├── zbxtable-v1.0.0-linux-amd64.tar.gz
│   ├── zbxtable-v1.0.0-linux-arm64.tar.gz
│   ├── zbxtable-v1.0.0-windows-amd64.tar.gz
│   ├── zbxtable-v1.0.0-darwin-amd64.tar.gz
│   ├── zbxtable-v1.0.0-darwin-arm64.tar.gz
│   └── SHA256SUMS
└── zbxtable-v1.0.0-*/
    ├── zbxtable (或 zbxtable.exe)
    ├── config/
    ├── README.md
    ├── README.zh-CN.md
    └── LICENSE
```

---

## 🧪 本地测试

### 测试 Build 流程

```bash
cd /Users/canghai/dev/code/zbxtable/zbxtable

# 1. 安装依赖
make install

# 2. 格式化代码
make fmt

# 3. 静态检查
make vet

# 4. 运行测试
make test

# 5. 生成覆盖率
make coverage

# 6. 编译
make build

# 7. 验证
ls -lh build/zbxtable
```

### 测试 Release 流程

```bash
cd /Users/canghai/dev/code/zbxtable/zbxtable

# 1. 编译所有平台
make build-all

# 2. 创建发布包
make package

# 3. 验证
ls -lh dist/releases/
```

---

## ✅ 更新的文件

1. ✅ `zbxtable/.github/workflows/build.yml`
2. ✅ `zbxtable/.github/workflows/release.yml`

---

## 🎯 下一步

### 1. 提交更新

```bash
cd /Users/canghai/dev/code/zbxtable/zbxtable
git add .github/
git commit -m "refactor: 使用 Makefile 进行编译和打包"
git push
```

### 2. 验证工作流

推送后查看 Actions 页面，确认工作流正常运行。

### 3. 本地测试（推荐）

```bash
# 快速测试
make build

# 完整测试
make quality

# 测试发布流程
make build-all && make package
```

---

## 💡 最佳实践

### 1. 本地开发

```bash
# 日常开发
make fmt        # 格式化代码
make vet        # 静态检查
make test       # 运行测试
make build      # 编译
make run        # 运行

# 提交前
make quality    # 完整的代码质量检查
```

### 2. 发布前

```bash
# 本地测试完整发布流程
make clean
make build-all
make package

# 检查发布包
ls -lh dist/releases/
tar -tzf dist/releases/zbxtable-*.tar.gz | head -20
```

### 3. 清理

```bash
# 清理构建产物
make clean

# 清理并重新构建
make clean && make build
```

---

## 📚 相关文档

- **Makefile**: `zbxtable/Makefile`
- **工作流**: `zbxtable/.github/workflows/`
- **更新说明**: `zbxtable/.github/UPDATE_NOTES.md`

---

**更新时间：** 2026-02-07  
**状态：** ✅ 已完成

