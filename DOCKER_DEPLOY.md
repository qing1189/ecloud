# Docker 部署指南

## 📦 部署方式

提供三种部署方式：

1. **Docker Compose（推荐）** - 一键部署，适合生产环境
2. **Docker 命令** - 手动部署，适合测试环境
3. **Docker + Nginx** - HTTPS 反向代理，适合对外服务

---

## 🚀 方式一：Docker Compose 部署（推荐）

### 1. 准备环境

```bash
# 克隆项目（或复制文件）
cd ecloud

# 创建数据目录
mkdir -p data/store
```

### 2. 配置环境变量

```bash
# 复制环境变量示例
cp .env.example .env

# 编辑环境变量（可选）
vim .env
```

**默认配置：**
- 端口：8088
- 时区：Asia/Shanghai
- 日志级别：info

### 3. 构建并启动

```bash
# 构建并启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f ecloud-web

# 查看首次生成的管理员密码
docker-compose logs ecloud-web | grep "密码"
```

**预期输出：**
```
ecloud-web | [Info] 首次启动检测到，管理员密码已生成:
ecloud-web | 
ecloud-web |     密码: Xy9Kp2mN4Q
ecloud-web | 
ecloud-web | [Info] Web 服务已启动
ecloud-web | [Info] 访问地址: http://0.0.0.0:8088
```

### 4. 访问界面

打开浏览器访问：**http://localhost:8088**

使用生成的密码登录。

### 5. 常用命令

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose stop

# 重启服务
docker-compose restart

# 查看日志
docker-compose logs -f

# 进入容器
docker-compose exec ecloud-web sh

# 停止并删除容器
docker-compose down

# 停止并删除容器和数据卷
docker-compose down -v
```

---

## 🔨 方式二：Docker 命令部署

### 1. 构建镜像

```bash
# 使用 Dockerfile.web 构建
docker build -f Dockerfile.web -t ecloud-web:latest .
```

**构建过程：**
- 阶段 1：构建前端（Node.js）
- 阶段 2：构建后端（Go）
- 阶段 3：创建运行镜像（Alpine）

### 2. 创建数据目录

```bash
mkdir -p $(pwd)/data/store
```

### 3. 启动容器

```bash
docker run -d \
  --name ecloud-web \
  --restart unless-stopped \
  -p 8088:8088 \
  -v $(pwd)/data/store:/app/store \
  -e TZ=Asia/Shanghai \
  ecloud-web:latest
```

### 4. 查看日志

```bash
# 查看所有日志
docker logs ecloud-web

# 实时查看日志
docker logs -f ecloud-web

# 查看管理员密码
docker logs ecloud-web | grep "密码"
```

### 5. 常用命令

```bash
# 启动容器
docker start ecloud-web

# 停止容器
docker stop ecloud-web

# 重启容器
docker restart ecloud-web

# 进入容器
docker exec -it ecloud-web sh

# 删除容器
docker rm -f ecloud-web
```

---

## 🔐 方式三：Docker + Nginx（HTTPS）

适合需要对外提供服务的场景。

### 1. 准备 SSL 证书

```bash
# 创建 SSL 证书目录
mkdir -p ssl

# 将证书文件放入 ssl/ 目录
# - ssl/cert.pem (证书文件)
# - ssl/key.pem (私钥文件)
```

**获取免费 SSL 证书：**
- Let's Encrypt：https://letsencrypt.org/
- Certbot：https://certbot.eff.org/

### 2. 修改 Nginx 配置

```bash
# 复制配置示例
cp nginx.conf.example nginx.conf

# 编辑配置文件
vim nginx.conf

# 修改以下内容：
# 1. server_name：改为你的域名
# 2. ssl_certificate：证书路径
# 3. ssl_certificate_key：私钥路径
```

### 3. 修改 docker-compose.yml

取消注释 Nginx 部分：

```yaml
nginx:
  image: nginx:alpine
  container_name: ecloud-nginx
  restart: unless-stopped
  ports:
    - "80:80"
    - "443:443"
  volumes:
    - ./nginx.conf:/etc/nginx/nginx.conf:ro
    - ./ssl:/etc/nginx/ssl:ro
  depends_on:
    - ecloud-web
  networks:
    - ecloud-network
```

### 4. 启动服务

```bash
docker-compose up -d
```

### 5. 访问界面

- HTTP：http://your-domain.com → 自动跳转到 HTTPS
- HTTPS：https://your-domain.com

---

## 📊 健康检查

所有部署方式都包含健康检查：

```bash
# 检查容器健康状态
docker ps

# 手动测试健康检查
curl http://localhost:8088/health
```

**健康检查参数：**
- 检查间隔：30 秒
- 超时时间：3 秒
- 启动延迟：10 秒
- 重试次数：3 次

---

## 💾 数据备份

### 备份数据

```bash
# 创建备份目录
mkdir -p backups

# 备份存储目录
tar -czf backups/ecloud-backup-$(date +%Y%m%d-%H%M%S).tar.gz data/store/

# 或使用 docker cp
docker cp ecloud-web:/app/store ./backups/store-$(date +%Y%m%d-%H%M%S)
```

### 恢复数据

```bash
# 停止服务
docker-compose stop

# 恢复备份
tar -xzf backups/ecloud-backup-20260615-120000.tar.gz

# 或使用 docker cp
docker cp ./backups/store-20260615-120000 ecloud-web:/app/store

# 启动服务
docker-compose start
```

### 自动备份脚本

创建 `backup.sh`：

```bash
#!/bin/bash
BACKUP_DIR="./backups"
RETENTION_DAYS=30

# 创建备份
mkdir -p $BACKUP_DIR
tar -czf $BACKUP_DIR/ecloud-backup-$(date +%Y%m%d-%H%M%S).tar.gz data/store/

# 删除旧备份
find $BACKUP_DIR -name "ecloud-backup-*.tar.gz" -mtime +$RETENTION_DAYS -delete

echo "备份完成：$(date)"
```

设置定时任务（crontab）：

```bash
# 每天凌晨 2 点备份
0 2 * * * /path/to/ecloud/backup.sh >> /var/log/ecloud-backup.log 2>&1
```

---

## 🔧 故障排查

### 容器无法启动

```bash
# 查看容器日志
docker-compose logs ecloud-web

# 检查端口占用
lsof -i:8088

# 检查数据目录权限
ls -la data/store
```

### 无法访问 Web 界面

```bash
# 检查容器状态
docker-compose ps

# 检查健康状态
docker inspect ecloud-web | grep -A 10 Health

# 测试端口连通性
curl -v http://localhost:8088/health
```

### 数据丢失

```bash
# 确认数据卷挂载
docker inspect ecloud-web | grep -A 10 Mounts

# 检查数据文件
ls -la data/store/
```

### 镜像构建失败

```bash
# 清理构建缓存
docker builder prune

# 重新构建（不使用缓存）
docker-compose build --no-cache
```

---

## 🔄 更新升级

### 更新镜像

```bash
# 1. 备份数据（重要！）
tar -czf backups/ecloud-backup-$(date +%Y%m%d).tar.gz data/store/

# 2. 停止服务
docker-compose stop

# 3. 拉取最新代码
git pull

# 4. 重新构建镜像
docker-compose build --no-cache

# 5. 启动服务
docker-compose up -d

# 6. 查看日志确认启动成功
docker-compose logs -f
```

### 回滚版本

```bash
# 停止服务
docker-compose stop

# 恢复数据
tar -xzf backups/ecloud-backup-20260615.tar.gz

# 使用旧镜像启动
docker-compose up -d
```

---

## 🎯 生产环境建议

### 1. 安全配置

```bash
# 修改默认端口（.env）
WEB_PORT=8888

# 配置防火墙
ufw allow 8888/tcp

# 仅允许特定 IP
iptables -A INPUT -p tcp --dport 8888 -s 192.168.1.0/24 -j ACCEPT
iptables -A INPUT -p tcp --dport 8888 -j DROP
```

### 2. 资源限制

修改 `docker-compose.yml`：

```yaml
services:
  ecloud-web:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
```

### 3. 日志管理

```yaml
services:
  ecloud-web:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

### 4. 监控集成

使用 Prometheus + Grafana：

```yaml
# 添加 Prometheus 监控
services:
  prometheus:
    image: prom/prometheus
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"

  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    depends_on:
      - prometheus
```

---

## 📝 环境变量说明

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| WEB_PORT | 8088 | Web 服务端口 |
| TZ | Asia/Shanghai | 时区 |
| LOG_LEVEL | info | 日志级别 |
| DATA_DIR | ./data/store | 数据目录 |
| BACKUP_DIR | ./backups | 备份目录 |

---

## 🎉 快速开始

最简单的部署方式：

```bash
# 1. 创建目录
mkdir -p data/store

# 2. 复制环境变量
cp .env.example .env

# 3. 启动服务
docker-compose up -d

# 4. 查看密码
docker-compose logs ecloud-web | grep "密码"

# 5. 访问界面
# 浏览器打开：http://localhost:8088
```

就这么简单！🚀

---

**文档更新时间：** 2026-06-15  
**Docker 版本要求：** >= 20.10  
**Docker Compose 版本：** >= 2.0
