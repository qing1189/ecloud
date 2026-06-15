.PHONY: help build up down restart logs clean backup restore

# 默认目标
help:
	@echo "eCloud Web 管理界面 - Makefile 命令"
	@echo ""
	@echo "开发命令："
	@echo "  make build          - 构建 Docker 镜像"
	@echo "  make up             - 启动服务"
	@echo "  make down           - 停止服务"
	@echo "  make restart        - 重启服务"
	@echo "  make logs           - 查看日志"
	@echo "  make shell          - 进入容器 shell"
	@echo ""
	@echo "数据管理："
	@echo "  make backup         - 备份数据"
	@echo "  make restore        - 恢复数据（需指定文件）"
	@echo "  make clean          - 清理数据和容器"
	@echo ""
	@echo "快速部署："
	@echo "  make deploy         - 一键部署（构建+启动）"
	@echo "  make redeploy       - 重新部署（停止+清理+构建+启动）"
	@echo ""
	@echo "本地开发："
	@echo "  make dev-backend    - 本地运行后端"
	@echo "  make dev-frontend   - 本地运行前端"
	@echo ""

# 构建镜像
build:
	@echo "正在构建 Docker 镜像..."
	docker-compose build

# 启动服务
up:
	@echo "正在启动服务..."
	docker-compose up -d
	@echo "服务已启动！访问: http://localhost:8088"

# 停止服务
down:
	@echo "正在停止服务..."
	docker-compose down

# 重启服务
restart:
	@echo "正在重启服务..."
	docker-compose restart

# 查看日志
logs:
	docker-compose logs -f

# 进入容器
shell:
	docker-compose exec ecloud-web sh

# 查看管理员密码
password:
	@docker-compose logs ecloud-web | grep -A 2 "管理员密码已生成" || echo "未找到密码（可能已初始化）"

# 备份数据
backup:
	@echo "正在备份数据..."
	@mkdir -p backups
	@tar -czf backups/ecloud-backup-$$(date +%Y%m%d-%H%M%S).tar.gz data/store/
	@echo "备份完成: backups/ecloud-backup-$$(date +%Y%m%d-%H%M%S).tar.gz"

# 恢复数据（使用: make restore FILE=backup-file.tar.gz）
restore:
	@if [ -z "$(FILE)" ]; then \
		echo "错误: 请指定备份文件"; \
		echo "用法: make restore FILE=backups/ecloud-backup-20260615-120000.tar.gz"; \
		exit 1; \
	fi
	@echo "正在恢复数据..."
	@docker-compose stop
	@tar -xzf $(FILE)
	@docker-compose start
	@echo "数据恢复完成"

# 清理数据和容器
clean:
	@echo "警告: 此操作将删除所有数据和容器！"
	@read -p "确认继续？(yes/no): " confirm && [ "$$confirm" = "yes" ] || exit 1
	@echo "正在清理..."
	@docker-compose down -v
	@rm -rf data/store/*
	@echo "清理完成"

# 一键部署
deploy:
	@echo "开始一键部署..."
	@./deploy.sh

# 重新部署
redeploy:
	@echo "开始重新部署..."
	@docker-compose down
	@docker-compose build --no-cache
	@docker-compose up -d
	@echo "重新部署完成"

# 本地开发 - 后端
dev-backend:
	@echo "启动后端开发服务..."
	go run main.go server

# 本地开发 - 前端
dev-frontend:
	@echo "启动前端开发服务..."
	cd web && npm run dev

# 健康检查
health:
	@curl -s http://localhost:8088/health && echo "✓ 服务正常" || echo "✗ 服务异常"

# 查看服务状态
status:
	@docker-compose ps

# 更新镜像
update:
	@echo "正在更新镜像..."
	@make backup
	@git pull
	@docker-compose build --no-cache
	@docker-compose up -d
	@echo "更新完成"

# 初始化环境
init:
	@echo "初始化环境..."
	@mkdir -p data/store backups
	@[ -f .env ] || cp .env.example .env
	@echo "环境初始化完成"
