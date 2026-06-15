#!/bin/bash

# eCloud Web 服务 - 快速启动脚本
# 用途：一键部署 Docker 服务

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印信息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 打印欢迎信息
print_welcome() {
    echo ""
    echo "================================================="
    echo "  eCloud 云电脑监控 - Docker 快速启动"
    echo "  Version: 1.0.0"
    echo "================================================="
    echo ""
}

# 检查 Docker 是否安装
check_docker() {
    print_info "检查 Docker 环境..."

    if ! command -v docker &> /dev/null; then
        print_error "Docker 未安装，请先安装 Docker"
        echo "安装文档：https://docs.docker.com/engine/install/"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose 未安装，请先安装 Docker Compose"
        echo "安装文档：https://docs.docker.com/compose/install/"
        exit 1
    fi

    print_success "Docker 环境检查通过"
}

# 创建必要的目录
create_directories() {
    print_info "创建数据目录..."

    mkdir -p data/store
    mkdir -p backups

    print_success "目录创建完成"
}

# 初始化环境变量
init_env() {
    print_info "初始化环境变量..."

    if [ ! -f .env ]; then
        if [ -f .env.example ]; then
            cp .env.example .env
            print_success "已从 .env.example 复制环境变量"
        else
            print_warning ".env.example 文件不存在，使用默认配置"
        fi
    else
        print_info ".env 文件已存在，跳过初始化"
    fi
}

# 构建镜像
build_image() {
    print_info "构建 Docker 镜像..."
    echo "这可能需要几分钟时间，请耐心等待..."

    if docker-compose build; then
        print_success "镜像构建成功"
    else
        print_error "镜像构建失败"
        exit 1
    fi
}

# 启动服务
start_service() {
    print_info "启动服务..."

    if docker-compose up -d; then
        print_success "服务启动成功"
    else
        print_error "服务启动失败"
        exit 1
    fi
}

# 等待服务就绪
wait_for_service() {
    print_info "等待服务就绪..."

    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:8088/health > /dev/null 2>&1; then
            print_success "服务已就绪"
            return 0
        fi

        echo -n "."
        sleep 1
        attempt=$((attempt + 1))
    done

    echo ""
    print_warning "服务启动超时，但可能仍在启动中"
    print_info "请手动检查：docker-compose logs -f"
}

# 显示管理员密码
show_password() {
    print_info "正在获取管理员密码..."
    sleep 2

    local password=$(docker-compose logs ecloud-web 2>/dev/null | grep -A 2 "管理员密码已生成" | grep "密码:" | awk '{print $2}')

    if [ -n "$password" ]; then
        echo ""
        echo "================================================="
        echo -e "${GREEN}首次启动检测到，管理员密码已生成：${NC}"
        echo ""
        echo -e "    ${YELLOW}密码: $password${NC}"
        echo ""
        echo "请妥善保存密码，可在 Web 界面修改"
        echo "================================================="
    else
        print_info "未检测到新密码（可能已初始化过）"
    fi
}

# 显示访问信息
show_access_info() {
    local port=$(grep WEB_PORT .env 2>/dev/null | cut -d '=' -f2 || echo "8088")

    echo ""
    echo "================================================="
    echo -e "${GREEN}服务已成功启动！${NC}"
    echo "================================================="
    echo ""
    echo -e "访问地址: ${BLUE}http://localhost:$port${NC}"
    echo ""
    echo "常用命令："
    echo "  查看日志：docker-compose logs -f"
    echo "  停止服务：docker-compose stop"
    echo "  重启服务：docker-compose restart"
    echo "  进入容器：docker-compose exec ecloud-web sh"
    echo ""
    echo "文档："
    echo "  Docker 部署：DOCKER_DEPLOY.md"
    echo "  使用指南：WEB_USAGE.md"
    echo "  API 测试：TEST_GUIDE.md"
    echo ""
    echo "================================================="
}

# 主函数
main() {
    print_welcome

    check_docker
    create_directories
    init_env
    build_image
    start_service
    wait_for_service
    show_password
    show_access_info

    echo ""
    print_success "部署完成！"
    echo ""
}

# 执行主函数
main
