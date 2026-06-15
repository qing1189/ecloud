# Dockerfile 镜像源配置说明

## 默认配置（海外环境）

`Dockerfile.web` 默认使用官方镜像源，适用于海外服务器。

```dockerfile
# 前端构建 - 使用官方 Node 镜像
FROM node:20-alpine AS frontend-builder

# 后端构建 - 使用官方 Go 镜像
FROM golang:1.23-alpine AS backend-builder

# 下载依赖 - 使用 Go 官方代理
RUN go mod download
```

---

## 中国大陆环境

如果你的服务器在中国大陆，可以使用国内镜像加速：

### 方式一：使用 Dockerfile.web.cn（推荐）

我们提供了一个专门的中国版 Dockerfile：

```bash
# 使用中国版 Dockerfile 构建
docker build -f Dockerfile.web.cn -t ecloud-web .

# 或修改 docker-compose.yml
services:
  ecloud-web:
    build:
      context: .
      dockerfile: Dockerfile.web.cn  # 使用中国版
```

### 方式二：修改 Dockerfile.web

如果你想直接修改主 Dockerfile：

```dockerfile
# 1. 添加 Go 代理（第 31 行附近）
RUN go env -w GOPROXY=https://goproxy.cn,direct && \
    go mod download

# 2. 添加 Alpine 镜像源（第 47 行附近）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk --no-cache add ca-certificates tzdata wget
```

---

## 镜像源对比

### 官方源（海外环境）

| 组件 | 镜像源 | 速度 |
|------|--------|------|
| Node.js | registry.npmjs.org | 快 🟢 |
| Go Modules | proxy.golang.org | 快 🟢 |
| Alpine Packages | dl-cdn.alpinelinux.org | 快 🟢 |

### 国内镜像（中国环境）

| 组件 | 镜像源 | 速度 |
|------|--------|------|
| Node.js | registry.npmmirror.com | 快 🟢 |
| Go Modules | goproxy.cn | 快 🟢 |
| Alpine Packages | mirrors.aliyun.com | 快 🟢 |

---

## 推荐配置

### 海外服务器（默认）

✅ **使用 `Dockerfile.web`**
- 官方镜像源
- 无需配置代理
- 构建速度快

```bash
docker-compose up -d --build
```

### 中国大陆服务器

✅ **使用 `Dockerfile.web.cn`**
- 国内镜像加速
- 构建速度更快
- 避免网络超时

```bash
# 修改 docker-compose.yml
dockerfile: Dockerfile.web.cn

# 或直接构建
docker build -f Dockerfile.web.cn -t ecloud-web .
```

---

## 常见问题

### Q1: 如何判断我需要哪个版本？

**A:** 根据服务器位置：

- 🌍 **海外服务器** → 使用 `Dockerfile.web`（默认）
- 🇨🇳 **中国服务器** → 使用 `Dockerfile.web.cn`

**测试方法：**
```bash
# 测试官方源速度
time curl -I https://proxy.golang.org

# 如果超时或很慢，使用中国版
```

### Q2: 构建失败，提示网络超时？

**A:** 可能需要使用国内镜像：

```bash
# 1. 使用中国版 Dockerfile
cp Dockerfile.web.cn.example Dockerfile.web.cn

# 2. 修改 docker-compose.yml
services:
  ecloud-web:
    build:
      dockerfile: Dockerfile.web.cn

# 3. 重新构建
docker-compose build --no-cache
```

### Q3: 如何切换镜像源？

**A:** 修改 `docker-compose.yml`：

```yaml
services:
  ecloud-web:
    build:
      context: .
      dockerfile: Dockerfile.web      # 海外版
      # dockerfile: Dockerfile.web.cn # 中国版
```

---

## 构建时间对比

### 海外服务器使用官方源

```
前端构建: 2-3 分钟
后端构建: 1-2 分钟
总时间: 3-5 分钟 ✅
```

### 中国服务器使用官方源

```
前端构建: 8-15 分钟 ❌（npm 慢）
后端构建: 5-10 分钟 ❌（go mod 慢）
总时间: 15-25 分钟 ❌（可能超时）
```

### 中国服务器使用国内镜像

```
前端构建: 2-3 分钟 ✅
后端构建: 1-2 分钟 ✅
总时间: 3-5 分钟 ✅
```

---

## 最佳实践

### 1. 根据环境选择

✅ **海外** → `Dockerfile.web`（默认）
✅ **中国** → `Dockerfile.web.cn`

### 2. 构建失败时

如果构建超时或失败：

```bash
# 1. 尝试使用国内镜像
cp Dockerfile.web.cn.example Dockerfile.web.cn

# 2. 清理缓存重新构建
docker-compose build --no-cache
```

### 3. 网络优化

```bash
# 使用构建参数指定镜像源
docker build \
  --build-arg GOPROXY=https://goproxy.cn,direct \
  -f Dockerfile.web \
  -t ecloud-web .
```

---

## 总结

- ✅ **默认使用官方源**（适合海外环境）
- ✅ **提供中国版 Dockerfile**（适合中国环境）
- ✅ **灵活切换**（通过 docker-compose.yml 配置）

**当前配置：** 海外环境（官方源）✅
