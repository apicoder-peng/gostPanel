#!/usr/bin/env bash

# ========================================================
#  Gost 节点一键安装脚本
#  系统要求: Ubuntu, Debian, CentOS, Alpine
#  作者: code-gopher
# ========================================================

set -eo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PLAIN='\033[0m'

# 配置信息
BASE_PATH="/etc/gost-node"
BIN_PATH="/usr/local/bin/gost-node"
CONF_FILE="${BASE_PATH}/config.yaml"
GOST_VERSION="3.2.6"
GH_PROXY=${GH_PROXY:-""}

# 日志输出
info() { echo -e "${GREEN}[信息]${PLAIN} $1"; }
warn() { echo -e "${YELLOW}[警告]${PLAIN} $1"; }
error() { echo -e "${RED}[错误]${PLAIN} $1"; exit 1; }

# 检查 Root 权限
check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "请使用 root 用户运行此脚本"
    fi
    info "权限检查通过"
}

# 检测系统架构
get_arch() {
    local arch=$(uname -m)
    case "$arch" in
        x86_64) echo "amd64" ;;
        aarch64) echo "arm64" ;;
        armv7l) echo "armv7" ;;
        *) error "暂不支持该架构: $arch" ;;
    esac
}

# 检查端口占用
check_port() {
    local port=$1
    if command -v ss >/dev/null 2>&1; then
        if ss -tpln | grep -q ":$port "; then
            error "端口 $port 已被占用，请更换端口后重试。"
        fi
    elif command -v netstat >/dev/null 2>&1; then
        if netstat -tpln | grep -q ":$port "; then
            error "端口 $port 已被占用，请更换端口后重试。"
        fi
    fi
}

# 卸载功能
uninstall() {
    info "正在卸载 Gost 节点..."
    if command -v systemctl >/dev/null 2>&1 && [[ -f /etc/systemd/system/gost-node.service ]]; then
        systemctl stop gost-node || true
        systemctl disable gost-node || true
        rm -f /etc/systemd/system/gost-node.service
        systemctl daemon-reload
    elif [[ -f /etc/init.d/gost-node ]]; then
        rc-service gost-node stop || true
        rc-update del gost-node default || true
        rm -f /etc/init.d/gost-node
    fi
    rm -rf "$BASE_PATH"
    rm -f "$BIN_PATH"
    info "Gost 节点已成功移除。"
}

# 安装二进制文件
install_bin() {
    local arch=$(get_arch)
    local url="${GH_PROXY}https://github.com/go-gost/gost/releases/download/v${GOST_VERSION}/gost_${GOST_VERSION}_linux_${arch}.tar.gz"
    
    info "正在下载 Gost v${GOST_VERSION} (${arch})..."
    info "下载地址: $url"
    
    local tmp_dir=$(mktemp -d)
    
    # 下载并解压
    if wget -q --show-progress -O- "$url" | tar -zx -C "$tmp_dir"; then
        info "下载完成"
    else
        error "下载或解压失败，请检查：\n  1. 网络连接是否正常\n  2. GitHub 是否可访问\n  3. 架构 ${arch} 是否支持"
    fi
    
    # 检查文件是否存在
    if [[ ! -f "${tmp_dir}/gost" ]]; then
        error "解压后未找到 gost 文件"
    fi
    
    mv "${tmp_dir}/gost" "$BIN_PATH"
    chmod +x "$BIN_PATH"
    rm -rf "$tmp_dir"
    info "二进制文件已安装到 $BIN_PATH"
}

# 生成配置文件
configure() {
    local api_port="${1:-39000}"
    local tcp_port="${2:-39001}"
    local user="${3:-}"
    local pass="${4:-}"
    
    info "正在生成配置文件..."

    # 如果账号为空，生成随机账号
    if [[ -z "$user" ]]; then
        info "生成随机用户名..."
        user=$(head -c 32 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c 12)
    fi
    
    # 如果密码为空，生成随机密码
    if [[ -z "$pass" ]]; then
        info "生成随机密码..."
        pass=$(head -c 32 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c 12)
    fi
    
    mkdir -p "$BASE_PATH"
    cat > "$CONF_FILE" <<EOF
services:
- name: relay-tcp-tunnel
  addr: ":$tcp_port"
  handler:
    type: relay
  listener:
    type: tcp
api:
  addr: ":$api_port"
  pathPrefix: /api
  auth:
    username: $user
    password: $pass
EOF
    info "配置文件已保存到 $CONF_FILE"
    
    # 导出变量用于最后显示
    GLOBAL_API_PORT=$api_port
    GLOBAL_TCP_PORT=$tcp_port
    GLOBAL_USER=$user
    GLOBAL_PASS=$pass
}

# 设置系统服务
setup_service() {
    if command -v systemctl >/dev/null 2>&1; then
        info "正在配置 systemd 服务..."
        cat > /etc/systemd/system/gost-node.service <<EOF
[Unit]
Description=Gost Panel Node Service
After=network.target

[Service]
Type=simple
ExecStart=$BIN_PATH -C $CONF_FILE
Restart=always
User=root

[Install]
WantedBy=multi-user.target
EOF
        systemctl daemon-reload
        systemctl enable gost-node
        systemctl restart gost-node
    elif [[ -f /sbin/openrc-run ]]; then
        info "正在配置 OpenRC 服务..."
        cat > /etc/init.d/gost-node <<EOF
#!/sbin/openrc-run
description="Gost Panel Node Service"
command="$BIN_PATH"
command_args="-C $CONF_FILE"
command_background=true
pidfile="/run/gost-node.pid"
depend() {
    need net
}
EOF
        chmod +x /etc/init.d/gost-node
        rc-update add gost-node default
        rc-service gost-node start
    else
        warn "未识别的服务管理器。请手动启动: $BIN_PATH -C $CONF_FILE"
    fi
}

# 显示安装成功信息
show_info() {
    local ip=$(curl -s -4 https://api.ipify.org 2>/dev/null || curl -s -4 https://ifconfig.me 2>/dev/null || echo "您的公网IP")
    echo -e "\n${GREEN}================================================${PLAIN}"
    echo -e "${GREEN}       Gost 节点安装成功！${PLAIN}"
    echo -e "------------------------------------------------"
    echo -e "  API 地址   : ${BLUE}http://${ip}:${GLOBAL_API_PORT}/api${PLAIN}"
    echo -e "  用户名     : ${BLUE}${GLOBAL_USER}${PLAIN}"
    echo -e "  密码       : ${BLUE}${GLOBAL_PASS}${PLAIN}"
    echo -e "  TCP 转发端口: ${BLUE}${GLOBAL_TCP_PORT}${PLAIN}"
    echo -e "------------------------------------------------"
    echo -e "${YELLOW}  请使用以上信息在面板中添加节点。${PLAIN}"
    echo -e "${YELLOW}  TCP 转发端口用于单跳或多跳隧道转发。${PLAIN}"
    echo -e "${GREEN}================================================${PLAIN}\n"
}

# 主程序
main() {
    echo -e "${GREEN}========================================${PLAIN}"
    echo -e "${GREEN}  Gost 节点安装脚本 v${GOST_VERSION}${PLAIN}"
    echo -e "${GREEN}========================================${PLAIN}\n"
    
    check_root
    
    # 检查第一个参数是否为 uninstall
    if [[ "${1:-}" == "uninstall" ]]; then
        uninstall
        exit 0
    fi
    
    # 解析安装参数
    local api_port="${1:-39000}"
    local tcp_port="${2:-39001}"
    local user="${3:-}"
    local pass="${4:-}"
    
    info "开始安装 Gost 节点..."
    info "配置参数: API端口=$api_port, TCP端口=$tcp_port, 用户=$user"
    
    # 检查端口占用
    check_port "$api_port"
    check_port "$tcp_port"
    
    install_bin
    configure "$api_port" "$tcp_port" "$user" "$pass"
    setup_service
    show_info
}

main "$@"
