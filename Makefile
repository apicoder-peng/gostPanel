# 版本信息
VERSION ?= $(shell git describe --tags --always)
LDFLAGS := -s -w -X 'gost-panel/internal/config.Version=$(VERSION)'

.PHONY: all build build-frontend build-backend clean dev help release

# 默认目标
all: build

# 帮助信息
help:
	@echo "Available commands:"
	@echo "  make build          - Build both frontend and backend"
	@echo "  make build-frontend - Build frontend only"
	@echo "  make build-backend  - Build backend only"
	@echo "  make dev            - Run in development mode (hot reload)"
	@echo "  make run            - Build frontend and run backend"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make release        - Build multi-platform release"
	@echo ""
	@echo "💡 Tip for Windows users: Run these commands in Git Bash or WSL for compatibility."

# 完整构建
build: build-frontend build-backend
	@echo "Build complete! Binary: backend/gost-panel.exe"

# 构建前端
build-frontend:
	@echo "Building frontend..."
	cd frontend && npm install && npm run build
	@echo "Frontend build complete"
	
# 构建后端（包含嵌入的前端）
build-backend:
	@echo "Building backend..."
	cd backend && go build -ldflags="$(LDFLAGS)" -o gost-panel.exe cmd/server/main.go
	@echo "Backend build complete"

# 运行（构建前端并启动后端）
run: build-frontend
	@echo "Starting backend..."
	cd backend && go run cmd/server/main.go

# 清理构建产物
clean:
	@echo "Cleaning artifacts..."
	rm -f backend/gost-panel.exe
	rm -f backend/main.exe
	rm -rf backend/internal/router/dist
	rm -rf frontend/dist
	@echo "Clean complete"

# 构建多平台发布版本
release: build-frontend
	@echo "Building multi-platform release..."
	cd backend && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o gost-panel-windows-amd64.exe cmd/server/main.go
	cd backend && CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o gost-panel-linux-amd64 cmd/server/main.go
	cd backend && CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o gost-panel-linux-arm64 cmd/server/main.go
	cd backend && CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o gost-panel-darwin-amd64 cmd/server/main.go
	cd backend && CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o gost-panel-darwin-arm64 cmd/server/main.go
	@echo "Multi-platform build complete"
