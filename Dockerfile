# 多阶段构建 Dockerfile for ZbxTable
# 阶段1: 构建前端
FROM node:16-alpine AS frontend-builder

WORKDIR /app/frontend

# 复制前端源码
COPY zbxtable-web/package*.json ./
RUN npm install --registry=https://registry.npmmirror.com

COPY zbxtable-web/ ./
RUN npm run build

# 阶段2: 构建后端
FROM golang:1.23-alpine AS backend-builder

# 安装必要的构建工具
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# 复制 go mod 文件
COPY zbxtable/go.mod zbxtable/go.sum ./
RUN go mod download

# 复制后端源码
COPY zbxtable/ ./

# 从前端构建阶段复制构建产物
COPY --from=frontend-builder /app/frontend/web ./web

# 构建后端应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o zbxtable main.go

# 阶段3: 运行时镜像
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata wget && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=backend-builder /app/zbxtable .
COPY --from=backend-builder /app/web ./web

# 创建必要的目录
RUN mkdir -p /app/log /app/download /app/template /app/config

# 设置权限
RUN chmod +x /app/zbxtable

# 暴露端口
EXPOSE 8088

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8088/ || exit 1

# 启动应用
CMD ["/app/zbxtable", "web"]

