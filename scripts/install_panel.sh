#!/bin/bash

# Gost Panel Binary Installation Script
# Author: code-gopher
# Repository: https://github.com/code-gopher/gostPanel

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}错误: 请以 root 用户运行此脚本${NC}"
  exit 1
fi

# Configuration
APP_NAME="gost-panel"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/gost-panel"
SERVICE_FILE="/etc/systemd/system/gost-panel.service"
REPO="code-gopher/gostPanel"
GH_PROXY=${GH_PROXY:-""}

# Detect Architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)  PLATFORM="linux-amd64" ;;
    aarch64|arm64) PLATFORM="linux-arm64" ;;
    *) echo -e "${RED}不支持的架构: $ARCH${NC}"; exit 1 ;;
esac

# Functions
function info() { echo -e "${GREEN}[INFO]${NC} $1"; }
function warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
function error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

function check_memory() {
    total_mem=$(free -m | grep Mem | awk '{print $2}')
    if [ "$total_mem" -le 1000 ]; then
        error "系统内存不足 ($total_mem MB)。二进制部署模式要求内存必须大于 1GB。"
    fi
    info "内存检查通过: $total_mem MB"
}

function install_panel() {
    check_memory
    info "正在获取最新版本信息..."
    LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$LATEST_TAG" ]; then
        LATEST_TAG=$(curl -s "${GH_PROXY}https://github.com/$REPO/releases/latest" | grep -oP 'v[0-9]+\.[0-9]+\.[0-9]+' | head -n 1)
    fi
    
    if [ -z "$LATEST_TAG" ]; then
        error "无法获取最新版本号，请检查网络或手动指定。"
    fi
    info "最新版本: $LATEST_TAG"

    TEMP_DIR=$(mktemp -d)
    TAR_FILE="$TEMP_DIR/gost-panel.tar.gz"
    DOWNLOAD_URL="${GH_PROXY}https://github.com/$REPO/releases/download/$LATEST_TAG/gost-panel-$PLATFORM.tar.gz"
    
    info "正在下载发布包: $DOWNLOAD_URL"
    curl -L "$DOWNLOAD_URL" -o "$TAR_FILE"
    
    info "正在解压..."
    tar -xzf "$TAR_FILE" -C "$TEMP_DIR"
    
    # 获取二进制文件名 (发布包里可能是 gost-panel-linux-amd64)
    EXTRACTED_BIN=$(find "$TEMP_DIR" -name "gost-panel-linux-*" -type f)
    if [ -z "$EXTRACTED_BIN" ]; then
        error "解压后未找到二进制文件"
    fi

    chmod +x "$EXTRACTED_BIN"
    mv "$EXTRACTED_BIN" "$INSTALL_DIR/$APP_NAME"

    # 处理配置目录
    if [ ! -d "$CONFIG_DIR/config" ]; then
        mkdir -p "$CONFIG_DIR/config"
    fi

    if [ -f "$TEMP_DIR/config/config.yaml" ]; then
        if [ ! -f "$CONFIG_DIR/config/config.yaml" ]; then
            cp "$TEMP_DIR/config/config.yaml" "$CONFIG_DIR/config/config.yaml"
            info "已从发布包初始化配置文件: $CONFIG_DIR/config/config.yaml"
        else
            warn "配置文件已存在，跳过覆盖"
        fi
    fi

    # 清理临时目录
    rm -rf "$TEMP_DIR"

    # Create systemd service
    info "创建系统服务..."
    cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Gost Panel Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$CONFIG_DIR
ExecStart=$INSTALL_DIR/$APP_NAME -c $CONFIG_DIR/config/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable "$APP_NAME"
    systemctl start "$APP_NAME"

    info "Gost Panel 安装完成并已启动！"
    systemctl status gost-panel
}

function uninstall_panel() {
    info "正在卸载 Gost Panel..."
    systemctl stop "$APP_NAME" || true
    systemctl disable "$APP_NAME" || true
    rm -f "$SERVICE_FILE"
    systemctl daemon-reload
    rm -f "$INSTALL_DIR/$APP_NAME"
    warn "二进制文件已移除。配置文件目录 $CONFIG_DIR (包含数据库) 已保留。"
    info "卸载完成。"
}

# Main
ACTION=${1:-install}

case $ACTION in
    install)
        install_panel
        ;;
    uninstall)
        uninstall_panel
        ;;
    *)
        echo "用法: $0 [install|uninstall]"
        exit 1
        ;;
esac
