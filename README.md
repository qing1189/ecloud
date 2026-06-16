# eCloud 云电脑监控系统

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![Vue Version](https://img.shields.io/badge/vue-3.x-green.svg)](https://vuejs.org)

自动化监控和管理移动云电脑实例，支持多用户权限管理的 Web 管理平台。

---

## ✨ 主要特性

- 🔐 **多用户权限管理**：支持管理员和普通用户两种角色
- 🖥️ **云账号管理**：支持公众版和政企版账号
- 📊 **实时监控**：自动检测机器状态，异常自动开机
- 📝 **操作日志**：完整记录所有操作，支持审计
- 🛡️ **安全可靠**：bcrypt 密码加密，JWT Token 认证
- 🌐 **现代化界面**：Vue 3 + Element Plus 响应式设计
- 🐳 **容器化部署**：Docker Compose 一键部署

---

## 🚀 快速开始

### 使用 Docker Compose（推荐）

```bash
# 1. 克隆仓库
git clone <repository-url>
cd ecloud

# 2. 修改环境变量（可选，建议修改默认密码）
cp .env.example .env
nano .env

# 3. 启动服务
docker-compose up -d

# 4. 访问管理界面
http://localhost:8088
```

**默认登录账号**：
- 用户名：`admin`
- 密码：`admin123456`（在 `.env` 文件中配置）

⚠️ **生产环境请务必修改密码！**

详细部署说明请查看：[快速开始指南](QUICKSTART.md)

---

## 📋 功能概览

### 管理员功能
- ✅ 用户管理（创建、编辑、删除用户）
- ✅ 查看所有云账号和监控状态
- ✅ 查看所有操作日志
- ✅ 重置用户密码

### 普通用户功能
- ✅ 管理自己的云账号
- ✅ 配置监控任务
- ✅ 查看自己的操作日志
- ✅ 修改自己的密码

---

## 🔐 权限模型

| 角色 | 权限说明 |
|------|---------|
| **admin** | 管理所有用户、查看所有云账号、管理所有监控任务、查看所有日志 |
| **user** | 仅管理自己的云账号、查看自己的日志、修改自己的密码 |

---

## 🏗️ 技术栈

### 后端
- **语言**：Go 1.21+
- **框架**：标准库 net/http
- **数据库**：SQLite 3
- **认证**：JWT (golang-jwt/jwt)
- **密码加密**：bcrypt

### 前端
- **框架**：Vue 3 (Composition API)
- **UI 库**：Element Plus
- **状态管理**：Pinia
- **路由**：Vue Router
- **构建工具**：Vite

---

## 📦 项目结构

```
ecloud/
├── cmd/                    # 命令行入口
│   ├── root.go
│   ├── server.go          # Web 服务器启动
│   └── trust.go           # 设备信任命令
├── pkg/
│   ├── api/               # API 路由和处理器
│   │   ├── handlers/      # 请求处理器
│   │   ├── middleware.go  # 认证中间件
│   │   └── router.go      # 路由配置
│   ├── service/           # 业务逻辑层
│   │   ├── user/          # 用户管理
│   │   ├── auth/          # 认证服务
│   │   ├── account/       # 账号管理
│   │   ├── monitor/       # 监控服务
│   │   └── logger/        # 日志服务
│   └── store/             # 数据存储层
├── web/                   # Vue 前端项目
│   ├── src/
│   │   ├── views/         # 页面组件
│   │   ├── stores/        # Pinia 状态管理
│   │   ├── api/           # API 接口定义
│   │   └── router/        # 路由配置
│   └── vite.config.js
├── docker-compose.yml     # Docker Compose 配置
├── Dockerfile.web         # Web 服务 Docker 镜像
├── .env.example           # 环境变量示例
└── README.md
```

---

## 📖 文档

- [快速开始指南](QUICKSTART.md) - 5 分钟快速部署
- [多用户系统改造报告](MULTI_USER_MIGRATION.md) - 完整的技术文档
- [API 文档](#) - API 接口说明（待补充）

---

## 🔒 安全建议

1. **生产环境在 `.env` 文件中设置强密码**（至少12位，包含大小写字母、数字、特殊字符）
2. 使用 Nginx 反向代理并配置 HTTPS
3. 配置防火墙规则限制访问 IP
4. 定期备份数据库文件（`ecloud.db`）
5. 不要将 `.env` 文件提交到版本控制系统

---

## 🛠️ 开发指南

### 本地开发

#### 后端开发

```bash
# 安装依赖
go mod download

# 运行后端
go run . server

# 编译
go build -o ecloud-server .
```

#### 前端开发

```bash
cd web

# 安装依赖
npm install

# 开发模式
npm run dev

# 构建生产版本
npm run build
```

### 环境变量

复制 `.env.example` 为 `.env` 并根据需要修改：

```env
# 管理员账号配置
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123456

# 服务配置
WEB_PORT=8088
TZ=Asia/Shanghai
LOG_LEVEL=info
```

---

## 📝 更新日志

### v2.0.0 - 多用户版本 (2026-06-16)
- ✨ 新增多用户权限管理系统
- 🔐 密码加密升级为 bcrypt
- 🔑 JWT 认证升级为标准库实现
- 🎨 前端界面全面改版
- 📊 新增用户管理页面
- 🛡️ 增加基于角色的访问控制（RBAC）

### v1.0.0 - 初始版本
- 云账号管理
- 实时监控功能
- 操作日志记录
- 单管理员模式

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

## 📄 许可证

[MIT License](LICENSE)

---

## 📞 联系方式

如有问题，请提交 Issue 或联系开发团队。

---

**开发团队** | **最后更新**：2026-06-16
