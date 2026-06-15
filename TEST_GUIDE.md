# eCloud Web 管理界面 - 测试指南

## 🎯 当前完成状态

### ✅ 已完成（后端）

1. **存储层** (`pkg/store/`)
   - JSON 文件线程安全读写
   - 账号、认证、日志数据管理

2. **认证服务** (`pkg/service/auth/`)
   - 密码哈希验证（临时简化版，待升级 bcrypt）
   - JWT Token 生成验证（临时简化版，待升级真正的 JWT）
   - 首次启动生成随机密码

3. **账号管理服务** (`pkg/service/account/`)
   - 账号增删改查
   - 密码 Base64 加密存储
   - 安全脱敏输出

4. **监控管理器** (`pkg/service/monitor/`)
   - 多账号并发监控
   - 动态启停任务
   - 热加载配置
   - 支持公众版和政企版

5. **操作日志服务** (`pkg/service/logger/`)
   - 异步事件记录
   - 分页查询
   - 按账号/类型筛选

6. **API 接口** (`pkg/api/`)
   - RESTful API 设计
   - JWT 认证中间件
   - CORS 支持
   - 统一响应格式

7. **Web 服务命令** (`cmd/server.go`)
   - 启动 Web 服务器
   - 自动加载账号任务
   - 优雅关闭

### ⏳ 待完成

- **前端项目**（Vue 3 + Element Plus）
- **Docker 镜像**（包含前后端）
- **完整文档**

---

## 🚀 快速测试

### 1. 安装依赖

由于权限限制，请手动运行以下命令安装依赖：

```bash
go get github.com/gin-gonic/gin@v1.10.0
go get github.com/golang-jwt/jwt/v5@v5.2.1
go get golang.org/x/crypto@v0.25.0
go mod tidy
```

### 2. 编译项目

```bash
go build -o ecloud .
```

### 3. 启动 Web 服务

```bash
./ecloud server
```

**预期输出：**
```
=================================================
  ecloud computer auto boot

  Author: Samler
=================================================
=================================================
首次启动检测到，管理员密码已生成:

    密码: Xy9Kp2mN4Q

请妥善保存密码，可通过 API 修改
=================================================
[Info]  2026-06-15 12:00:00 [定时任务] 初始化中
=================================================
Web 服务已启动
监听地址: http://0.0.0.0:8088
API 文档: http://0.0.0.0:8088/api/
=================================================
```

### 4. 测试 API

#### 健康检查
```bash
curl http://localhost:8088/health
# 输出: OK
```

#### 登录获取 Token
```bash
curl -X POST http://localhost:8088/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"password": "Xy9Kp2mN4Q"}'
```

**响应示例：**
```json
{
  "success": true,
  "data": {
    "token": "admin:2026-06-16T12:00:00+08:00:hash_xyz"
  }
}
```

#### 添加公众版账号
```bash
curl -X POST http://localhost:8088/api/accounts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "name": "我的公众版账号",
    "type": "public",
    "username": "13800138000",
    "password": "your_password",
    "monitor_config": {
      "enabled": true,
      "interval": 60,
      "machines": []
    }
  }'
```

#### 查看所有账号
```bash
curl -X GET http://localhost:8088/api/accounts \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

#### 查看监控状态
```bash
curl -X GET http://localhost:8088/api/monitor/status \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

#### 查看操作日志
```bash
curl -X GET "http://localhost:8088/api/logs?page=1&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

#### 修改密码
```bash
curl -X POST http://localhost:8088/api/auth/change-password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "old_password": "Xy9Kp2mN4Q",
    "new_password": "NewPassword123"
  }'
```

---

## 📋 完整 API 列表

### 认证相关

| 方法 | 路径 | 描述 | 需要认证 |
|------|------|------|---------|
| POST | `/api/auth/login` | 登录获取 Token | ❌ |
| GET | `/api/auth/verify` | 验证 Token | ✅ |
| POST | `/api/auth/change-password` | 修改密码 | ✅ |

### 账号管理

| 方法 | 路径 | 描述 | 需要认证 |
|------|------|------|---------|
| GET | `/api/accounts` | 列出所有账号 | ✅ |
| POST | `/api/accounts` | 添加账号 | ✅ |
| GET | `/api/accounts/:id` | 获取账号详情 | ✅ |
| PUT | `/api/accounts/:id` | 更新账号 | ✅ |
| DELETE | `/api/accounts/:id` | 删除账号 | ✅ |
| POST | `/api/accounts/:id/toggle` | 启用/停用监控 | ✅ |

### 监控状态

| 方法 | 路径 | 描述 | 需要认证 |
|------|------|------|---------|
| GET | `/api/monitor/status` | 所有任务状态 | ✅ |
| GET | `/api/monitor/status/:id` | 单个任务状态 | ✅ |
| POST | `/api/monitor/reload` | 重新加载所有任务 | ✅ |

### 操作日志

| 方法 | 路径 | 描述 | 需要认证 |
|------|------|------|---------|
| GET | `/api/logs?page=1&limit=50&account_id=xxx&type=boot` | 日志列表 | ✅ |

---

## 🔧 数据文件位置

所有数据存储在 `store/` 目录下：

```
store/
├── accounts.json    # 账号配置
├── auth.json        # 管理员密码
└── logs.json        # 操作日志
```

**账号配置示例：**
```json
{
  "accounts": [
    {
      "id": "acc_1718456789",
      "name": "我的公众版账号",
      "type": "public",
      "username": "13800138000",
      "password": "MTIzNDU2",
      "monitor_config": {
        "enabled": true,
        "interval": 60,
        "machines": []
      },
      "created_at": "2026-06-15T12:00:00Z",
      "updated_at": "2026-06-15T12:00:00Z"
    }
  ],
  "version": 1
}
```

---

## ⚠️ 已知限制（临时实现）

由于依赖安装受限，以下功能使用简化版实现：

1. **密码哈希** - 当前使用简单前缀，待升级为 bcrypt
2. **JWT Token** - 当前使用简单字符串，待升级为真正的 JWT
3. **密码加密** - 当前使用 Base64 编码，待升级为 AES-256-GCM

**安装依赖后需要升级的代码：**
- `pkg/service/auth/manager.go` - hashPassword/checkPassword 函数
- `pkg/service/auth/jwt.go` - GenerateToken/ValidateToken 函数

---

## 🎉 功能验证

### 1. 多账号并发监控

添加 3 个账号，查看监控状态：

```bash
# 添加账号 1
curl -X POST http://localhost:8088/api/accounts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN" \
  -d '{"name": "账号1", "type": "public", "username": "user1", "password": "pass1", "monitor_config": {"enabled": true, "interval": 60, "machines": []}}'

# 添加账号 2
curl -X POST http://localhost:8088/api/accounts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN" \
  -d '{"name": "账号2", "type": "public", "username": "user2", "password": "pass2", "monitor_config": {"enabled": true, "interval": 120, "machines": []}}'

# 查看状态
curl http://localhost:8088/api/monitor/status -H "Authorization: Bearer TOKEN"
```

### 2. 热加载配置

修改账号监控间隔无需重启：

```bash
# 更新账号配置
curl -X PUT http://localhost:8088/api/accounts/acc_1718456789 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN" \
  -d '{"monitor_config": {"enabled": true, "interval": 180, "machines": []}}'

# 监控任务会自动重新加载
```

### 3. 查看操作日志

```bash
curl "http://localhost:8088/api/logs?page=1&limit=20" \
  -H "Authorization: Bearer TOKEN"
```

---

## 📝 下一步计划

1. **安装 Go 依赖**，升级认证和加密实现
2. **开发前端界面**（Vue 3 + Element Plus）
3. **创建 Docker 镜像**
4. **编写完整文档**

---

## 🐛 故障排查

### 服务无法启动
- 检查端口 8088 是否被占用：`lsof -i:8088`
- 查看日志输出中的错误信息

### 认证失败
- 确认使用首次启动生成的密码
- 检查 `store/auth.json` 文件是否存在

### 账号无法启动监控
- 检查账号类型和凭证是否正确
- 公众版需要设备信任（先运行 `./ecloud trust`）
- 查看 `store/logs.json` 中的错误日志

---

**文档生成时间**：2026-06-15  
**当前版本**：v0.2-alpha（Web 服务后端）
