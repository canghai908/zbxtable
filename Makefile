# Makefile for ZbxTable

# 变量定义
APP_NAME := zbxtable
VERSION := $(shell git describe --tags --exact-match 2>/dev/null || echo "dev")
GIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.gitHash=$(GIT_HASH) -X main.buildTime=$(BUILD_TIME) -w -s"

# Go 相关
GOCMD := go
GOBUILD := CGO_ENABLED=0 $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# 输出目录
BUILD_DIR := build
DIST_DIR := dist

# 平台
PLATFORMS := linux/amd64 linux/arm64

.PHONY: all build clean test coverage lint docker docker-build docker-up docker-down help

# 默认目标
all: clean build

# 帮助信息
help:
	@echo "ZbxTable Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build          - 编译应用"
	@echo "  make build-all      - 编译所有平台版本"
	@echo "  make clean          - 清理编译产物"
	@echo "  make test           - 运行测试"
	@echo "  make coverage       - 生成测试覆盖率报告"
	@echo "  make lint           - 代码检查"
	@echo "  make docker-build   - 构建 Docker 镜像"
	@echo "  make docker-up      - 启动 Docker 服务"
	@echo "  make docker-down    - 停止 Docker 服务"
	@echo "  make install        - 安装依赖"
	@echo "  make run            - 运行应用"
	@echo "  make fmt            - 格式化代码"
	@echo "  make vet            - 代码静态检查"
	@echo ""

# 编译
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) main.go
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

# 编译所有平台
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		CGO_ENABLED=0 GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GOCMD) build $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-$${platform%/*}-$${platform#*/} main.go; \
		echo "Built $(DIST_DIR)/$(APP_NAME)-$${platform%/*}-$${platform#*/}"; \
	done
	@echo "All builds complete!"

# 清理
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

# 安装依赖
install:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies installed!"

# 运行
run: build
	@echo "Running $(APP_NAME)..."
	@$(BUILD_DIR)/$(APP_NAME) web

# 测试
test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

# 测试覆盖率
coverage:
	@echo "Generating coverage report..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 代码检查
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

# 格式化代码
fmt:
	@echo "Formatting code..."
	@gofmt -w .
	@echo "Code formatted!"

# 静态检查
vet:
	@echo "Running go vet..."
	@$(GOCMD) vet ./...
	@echo "Vet complete!"

# Docker 构建
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) -f Dockerfile ..
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "Docker image built: $(APP_NAME):$(VERSION)"

# Docker 启动
docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d
	@echo "Docker services started!"
	@echo "Access the application at http://localhost:8085"

# Docker 停止
docker-down:
	@echo "Stopping Docker services..."
	docker-compose down
	@echo "Docker services stopped!"

# Docker 日志
docker-logs:
	docker-compose logs -f zbxtable

# Docker 重启
docker-restart:
	@echo "Restarting Docker services..."
	docker-compose restart
	@echo "Docker services restarted!"

# 开发模式（热重载）
dev:
	@echo "Starting development mode..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

# 生成 mock
mock:
	@echo "Generating mocks..."
	@which mockgen > /dev/null || (echo "Installing mockgen..." && go install github.com/golang/mock/mockgen@latest)
	@find . -name "*_mock.go" -delete
	@go generate ./...
	@echo "Mocks generated!"

# 数据库迁移
migrate-up:
	@echo "Running database migrations..."
	@# 这里可以添加数据库迁移命令
	@echo "Migrations complete!"

migrate-down:
	@echo "Rolling back database migrations..."
	@# 这里可以添加数据库回滚命令
	@echo "Rollback complete!"

# 版本信息
version:
	@echo "Version: $(VERSION)"
	@echo "Git Hash: $(GIT_HASH)"
	@echo "Build Time: $(BUILD_TIME)"

# 打包发布
package: build-all
	@echo "Creating release packages..."
	@mkdir -p $(DIST_DIR)/releases
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		pkg_name="$(APP_NAME)-$(VERSION)-$$os-$$arch"; \
		mkdir -p $(DIST_DIR)/$$pkg_name; \
		cp $(DIST_DIR)/$(APP_NAME)-$$os-$$arch $(DIST_DIR)/$$pkg_name/$(APP_NAME); \
		chmod +x $(DIST_DIR)/$$pkg_name/$(APP_NAME); \
		cp -r config $(DIST_DIR)/$$pkg_name/; \
		cp README.md $(DIST_DIR)/$$pkg_name/; \
		cp LICENSE $(DIST_DIR)/$$pkg_name/ 2>/dev/null || true; \
		cd $(DIST_DIR) && tar -czf releases/$$pkg_name.tar.gz $$pkg_name; \
		cd ..; \
		echo "Created $(DIST_DIR)/releases/$$pkg_name.tar.gz"; \
	done
	@echo "Release packages created in $(DIST_DIR)/releases/"

# 安装到系统
install-bin: build
	@echo "Installing $(APP_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(APP_NAME)
	@echo "$(APP_NAME) installed successfully!"

# 卸载
uninstall-bin:
	@echo "Uninstalling $(APP_NAME)..."
	@sudo rm -f /usr/local/bin/$(APP_NAME)
	@echo "$(APP_NAME) uninstalled!"

# 检查代码质量
quality: fmt vet lint test
	@echo "Code quality check complete!"

