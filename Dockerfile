# 前端构建阶段
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# 后端构建阶段
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app

# 复制依赖文件
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 复制后端源码
COPY backend/ .

# 从前端构建阶段复制 dist 到后端 embedding 路径
COPY --from=frontend-builder /app/backend/internal/router/dist ./internal/router/dist

# 编译（关闭 CGO，使用纯 Go SQLite 驱动）
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" -o gost-panel cmd/server/main.go

# 运行阶段
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

# 复制程序和配置
COPY --from=backend-builder /app/gost-panel .
COPY --from=backend-builder /app/config ./config

# 创建目录
RUN mkdir -p /app/data /app/logs

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

CMD ["./gost-panel"]
