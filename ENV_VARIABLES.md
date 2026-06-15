# 环境变量配置说明

## 工作原理

### Docker Compose 自动加载 .env

Docker Compose **会自动加载项目根目录下的 `.env` 文件**，无需在 `docker-compose.yml` 中显式配置 `env_file`。

这是 Docker Compose 的默认行为，文档参考：
https://docs.docker.com/compose/environment-variables/set-environment-variables/

---

## 配置方式

### 方式一：使用 .env 文件（推荐）

```bash
# 1. 从示例复制（deploy.sh 自动执行）
cp .env.example .env

# 2. 编辑配置
vim .env

# 3. 启动服务
docker-compose up -d
```

**优点：**
- ✅ Docker Compose 自动加载
- ✅ 配置集中管理
- ✅ 方便修改
- ✅ 可以不提交到 Git（已在 .gitignore）

### 方式二：使用默认值

```bash
# 不创建 .env 文件，直接启动
docker-compose up -d
```

**说明：**
- docker-compose.yml 中使用了默认值：`${WEB_PORT:-8088}`
- 如果 `.env` 不存在，会使用默认值：
  - `WEB_PORT` → 8088
  - `TZ` → Asia/Shanghai
  - `LOG_LEVEL` → info

### 方式三：环境变量覆盖

```bash
# 通过环境变量覆盖
WEB_PORT=9000 docker-compose up -d

# 或
export WEB_PORT=9000
docker-compose up -d
```

---

## 配置文件

### .env.example（模板）

```bash
# Web 服务端口
WEB_PORT=8088

# 时区设置
TZ=Asia/Shanghai

# 日志级别 (debug, info, warn, error)
LOG_LEVEL=info
```

### .env（实际使用）

```bash
# 从 .env.example 复制后修改
cp .env.example .env

# 修改端口示例
WEB_PORT=9000
```

---

## docker-compose.yml 中的使用

```yaml
services:
  ecloud-web:
    ports:
      # 格式：${变量名:-默认值}
      - "${WEB_PORT:-8088}:8088"
    
    environment:
      # 传递给容器内部
      - TZ=${TZ:-Asia/Shanghai}
      - LOG_LEVEL=${LOG_LEVEL:-info}
```

**说明：**
- `${WEB_PORT:-8088}` - 如果 `WEB_PORT` 未定义，使用 `8088`
- Docker Compose 会从 `.env` 读取变量值
- 通过 `environment` 传递给容器内部

---

## 变量优先级

从高到低：

1. **Shell 环境变量** - `export WEB_PORT=9000`
2. **命令行参数** - `WEB_PORT=9000 docker-compose up`
3. **.env 文件** - 项目根目录的 `.env`
4. **docker-compose.yml 中的默认值** - `${WEB_PORT:-8088}`

---

## 常见问题

### Q1: 为什么不在 docker-compose.yml 中显式配置 `env_file: .env`？

**A:** 因为 Docker Compose 会自动加载 `.env` 文件，显式配置反而会在文件不存在时报错。

**自动加载的优点：**
- ✅ .env 不存在时，使用默认值，不会报错
- ✅ .env 存在时，自动加载变量
- ✅ 更灵活，更符合 Docker Compose 最佳实践

### Q2: 如何验证环境变量是否生效？

```bash
# 查看 Docker Compose 解析后的配置
docker-compose config

# 检查容器内的环境变量
docker exec ecloud-web env | grep -E "TZ|LOG_LEVEL"

# 查看端口映射
docker-compose ps
```

### Q3: 修改 .env 后需要重启吗？

**是的**，需要重新创建容器：

```bash
# 停止服务
docker-compose down

# 重新启动（会读取新的 .env）
docker-compose up -d
```

### Q4: .env 文件应该提交到 Git 吗？

**不应该**！.env 包含本地配置，不应提交。

**.gitignore 中已配置：**
```
config.yml
.env
```

**应该提交的是 .env.example（模板文件）**

---

## 部署流程

### 自动部署（推荐）

```bash
# deploy.sh 会自动处理 .env
./deploy.sh
```

**脚本会自动：**
1. 检查 `.env` 是否存在
2. 如果不存在，从 `.env.example` 复制
3. 提示用户可以修改配置
4. 启动服务

### 手动部署

```bash
# 1. 创建 .env
cp .env.example .env

# 2. 修改配置（可选）
vim .env

# 3. 创建数据目录
mkdir -p data/store

# 4. 启动服务
docker-compose up -d
```

---

## 最佳实践

### 1. 使用 .env 文件
✅ **推荐：** 创建 `.env` 文件，集中管理配置

```bash
cp .env.example .env
vim .env
```

### 2. 不提交敏感信息
✅ **推荐：** .env 加入 .gitignore

```gitignore
.env
config.yml
store/*.json
```

### 3. 提供 .env.example
✅ **推荐：** 提交 `.env.example` 作为模板

```bash
# 包含默认值和注释
WEB_PORT=8088  # Web 服务端口
TZ=Asia/Shanghai  # 时区
```

### 4. 使用默认值
✅ **推荐：** docker-compose.yml 中提供默认值

```yaml
ports:
  - "${WEB_PORT:-8088}:8088"  # 默认 8088
```

---

## 总结

✅ **Docker Compose 自动加载 .env 文件**
- 无需在 docker-compose.yml 中配置 `env_file`
- .env 不存在时，使用默认值
- .env 存在时，自动加载变量

✅ **deploy.sh 自动处理**
- 首次部署自动从 .env.example 复制
- 简化用户操作

✅ **配置灵活**
- 支持 .env 文件
- 支持环境变量
- 支持默认值

---

**参考文档：**
- Docker Compose 环境变量：https://docs.docker.com/compose/environment-variables/
- .env 文件格式：https://docs.docker.com/compose/environment-variables/env-file/
