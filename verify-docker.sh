#!/bin/bash

# Docker 配置验证脚本

set -e

echo "=========================================="
echo "eCloud Docker 配置验证"
echo "=========================================="
echo ""

# 检查 1: Dockerfile.web 存在
echo "[1/10] 检查 Dockerfile.web..."
if [ -f "Dockerfile.web" ]; then
    echo "✓ Dockerfile.web 存在"
else
    echo "✗ Dockerfile.web 不存在"
    exit 1
fi

# 检查 2: docker-compose.yml 存在
echo "[2/10] 检查 docker-compose.yml..."
if [ -f "docker-compose.yml" ]; then
    echo "✓ docker-compose.yml 存在"
else
    echo "✗ docker-compose.yml 不存在"
    exit 1
fi

# 检查 3: .env.example 存在
echo "[3/10] 检查 .env.example..."
if [ -f ".env.example" ]; then
    echo "✓ .env.example 存在"
else
    echo "✗ .env.example 不存在"
    exit 1
fi

# 检查 4: .dockerignore 存在
echo "[4/10] 检查 .dockerignore..."
if [ -f ".dockerignore" ]; then
    echo "✓ .dockerignore 存在"
else
    echo "✗ .dockerignore 不存在"
    exit 1
fi

# 检查 5: deploy.sh 存在且可执行
echo "[5/10] 检查 deploy.sh..."
if [ -f "deploy.sh" ] && [ -x "deploy.sh" ]; then
    echo "✓ deploy.sh 存在且可执行"
else
    echo "✗ deploy.sh 不存在或不可执行"
    exit 1
fi

# 检查 6: Makefile 存在
echo "[6/10] 检查 Makefile..."
if [ -f "Makefile" ]; then
    echo "✓ Makefile 存在"
else
    echo "✗ Makefile 不存在"
    exit 1
fi

# 检查 7: 前端目录和文件
echo "[7/10] 检查前端项目..."
if [ -d "web" ] && [ -f "web/package.json" ] && [ -f "web/vite.config.js" ]; then
    echo "✓ 前端项目完整"
else
    echo "✗ 前端项目不完整"
    exit 1
fi

# 检查 8: Go 项目文件
echo "[8/10] 检查 Go 项目..."
if [ -f "go.mod" ] && [ -f "go.sum" ] && [ -f "main.go" ]; then
    echo "✓ Go 项目完整"
else
    echo "✗ Go 项目不完整"
    exit 1
fi

# 检查 9: cmd/server.go 存在
echo "[9/10] 检查 server.go..."
if [ -f "cmd/server.go" ]; then
    echo "✓ cmd/server.go 存在"
else
    echo "✗ cmd/server.go 不存在"
    exit 1
fi

# 检查 10: 文档文件
echo "[10/10] 检查文档..."
docs=("DOCKER_DEPLOY.md" "WEB_USAGE.md" "TEST_GUIDE.md")
all_docs_exist=true
for doc in "${docs[@]}"; do
    if [ ! -f "$doc" ]; then
        echo "✗ $doc 不存在"
        all_docs_exist=false
    fi
done

if [ "$all_docs_exist" = true ]; then
    echo "✓ 所有文档完整"
else
    exit 1
fi

echo ""
echo "=========================================="
echo "✓ 所有配置验证通过！"
echo "=========================================="
echo ""
echo "可以执行以下命令开始部署："
echo "  ./deploy.sh          # 一键部署"
echo "  docker-compose up -d # 使用 docker-compose"
echo "  make deploy          # 使用 Makefile"
echo ""
