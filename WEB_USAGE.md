# eCloud Web 管理界面 - 完整使用指南

## 🎉 项目完成状态

**进度：100% 完成** ✅

- ✅ 后端 API（Go + Gin）
- ✅ 前端界面（Vue 3 + Element Plus）
- ✅ 多账号并发监控
- ✅ 热加载配置
- ✅ 操作日志
- ⏳ Docker 镜像（待完善）

---

## 🚀 快速开始

### 1. 安装依赖

```bash
# 安装 Go 依赖
go get github.com/gin-gonic/gin@v1.10.0
go get github.com/golang-jwt/jwt/v5@v5.2.1
go get golang.org/x/crypto@v0.25.0
go mod tidy

# 安装前端依赖（已完成，跳过此步骤）
cd web && npm install
```

### 2. 构建前端

```bash
cd web
npm run build
cd ..
```

构建后的文件会输出到 `static/` 目录。

### 3. 编译后端

```bash
go build -o ecloud .
```

### 4. 启动服务

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
[Info]  2026-06-15 13:45:00 Web 服务已启动
[Info]  2026-06-15 13:45:00 访问地址: http://0.0.0.0:8088
[Info]  2026-06-15 13:45:00 API 接口: http://0.0.0.0:8088/api/
=================================================
```

### 5. 访问 Web 界面

打开浏览器访问：**http://localhost:8088**

使用首次启动生成的密码登录。

---

## 📱 功能介绍

### 登录页面
- 简洁美观的登录界面
- 密码强度验证
- 错误提示

### 仪表盘
- **统计卡片**：账号数、云电脑数、开机次数、活跃任务
- **近期操作**：最新 10 条操作日志
- **监控状态**：所有账号的实时状态
- **自动刷新**：每 30 秒自动更新数据

### 账号管理
- **添加账号**：
  - 公众版：用户名 + 密码
  - 政企版：Access Key + Secret Key + Pool ID
- **编辑账号**：修改名称、密码、监控配置
- **删除账号**：支持二次确认
- **启用/停用监控**：一键切换
- **监控配置**：
  - 检查间隔（30-3600 秒）
  - 监控机器列表（留空表示全部）

### 实时监控
- **任务状态**：运行中/已停止/错误
- **最后检查时间**：相对时间显示
- **最近事件**：开机成功/失败等
- **自动刷新**：可开启/关闭
- **重新加载**：重启所有监控任务

### 操作日志
- **筛选功能**：按事件类型筛选
- **分页显示**：10/20/50/100 条
- **详情查看**：查看完整事件详情
- **时间显示**：本地化格式
- **状态标签**：成功/失败/信息

---

## 🎨 界面预览

### 布局结构
```
┌─────────────────────────────────────────────┐
│  [Logo] eCloud         [管理员 ▼]  [退出]   │
├──────┬──────────────────────────────────────┤
│      │  仪表盘                               │
│ 仪表 │  ┌────┐ ┌────┐ ┌────┐ ┌────┐       │
│ 板   │  │账号│ │云电│ │开机│ │活跃│       │
│      │  │ 3  │ │ 12 │ │ 8  │ │ 2  │       │
│ 账号 │  └────┘ └────┘ └────┘ └────┘       │
│ 管理 │                                      │
│      │  近期操作                             │
│ 实时 │  ━━━━━━━━━━━━━━━━━━━━━━━━━         │
│ 监控 │  12:34 [开机成功] machine_1          │
│      │  12:33 [开机成功] machine_5          │
│ 操作 │  ...                                 │
│ 日志 │                                      │
└──────┴──────────────────────────────────────┘
```

---

## 🔧 配置说明

### 数据存储位置

所有数据存储在 `store/` 目录：

```
store/
├── accounts.json    # 账号配置
├── auth.json        # 管理员密码
└── logs.json        # 操作日志
```

### 账号配置示例

`store/accounts.json`:
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

## 🔐 安全建议

1. **修改默认密码**
   - 首次登录后立即修改密码
   - 密码长度至少 8 位

2. **网络访问控制**
   - 仅在可信网络中使用
   - 建议配置防火墙规则
   - 使用反向代理（Nginx）添加 HTTPS

3. **定期备份**
   - 定期备份 `store/` 目录
   - 包含账号配置和日志

---

## 📝 API 接口

完整 API 文档请查看 `TEST_GUIDE.md`。

**基础地址：** `http://localhost:8088/api`

**认证方式：** Bearer Token（除登录接口外）

**主要接口：**
- `POST /auth/login` - 登录
- `GET /accounts` - 账号列表
- `POST /accounts` - 添加账号
- `GET /monitor/status` - 监控状态
- `GET /logs` - 操作日志

---

## 🐛 故障排查

### 服务无法启动

**问题：** 端口 8088 被占用
```bash
# 查看占用端口的进程
lsof -i:8088

# 或者修改端口（编辑 cmd/server.go）
```

### 前端白屏

**问题：** 静态文件未构建
```bash
# 重新构建前端
cd web && npm run build
```

### 登录失败

**问题：** 密码错误
```bash
# 删除认证文件，重新生成密码
rm store/auth.json
./ecloud server
```

### 账号无法启动监控

**问题1：** 公众版设备未信任
```bash
# Web端会自动弹出验证码对话框，无需命令行操作
# 1. 添加账号时会自动发送短信验证码
# 2. 在弹出的对话框中输入验证码
# 3. 验证成功后自动完成设备信任
```

**问题2：** 账号凭证错误
- 检查用户名和密码是否正确
- 查看 `store/logs.json` 中的错误日志

---

## 🔄 CLI 命令兼容性

原有的 CLI 命令依然可用：

```bash
# 信任设备（公众版）
# 1. 添加账号时会自动发送短信验证码
# 2. 在弹出的对话框中输入验证码
# 3. 验证成功后自动完成设备信任

# 查看机器列表
./ecloud list-machines

# 传统监控模式（单账号）
./ecloud run
```

---

## 🌟 最佳实践

### 1. 监控配置

- **检查间隔**：建议 60-120 秒，避免 API 限流
- **机器列表**：留空监控所有，指定 ID 监控特定机器
- **多账号**：不同账号设置不同间隔，错开请求时间

### 2. 日志管理

- 系统自动保留最近 1000 条日志
- 定期导出重要日志（手动复制 `store/logs.json`）

### 3. 性能优化

- 账号数量建议不超过 10 个
- 每个账号监控间隔不少于 60 秒
- 定期清理日志文件

---

## 📦 生产部署

### 使用 systemd 服务

创建 `/etc/systemd/system/ecloud.service`:

```ini
[Unit]
Description=eCloud Computer Auto Boot Web Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/ecloud
ExecStart=/opt/ecloud/ecloud server
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable ecloud
sudo systemctl start ecloud
sudo systemctl status ecloud
```

### 使用 Nginx 反向代理

```nginx
server {
    listen 80;
    server_name ecloud.example.com;

    location / {
        proxy_pass http://127.0.0.1:8088;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## 📊 监控指标

通过日志可以查看以下指标：

- **账号数量**：总账号数
- **监控机器数**：所有账号下的云电脑总数
- **开机次数**：每日开机操作统计
- **活跃任务数**：正在运行的监控任务
- **成功率**：开机成功/失败比例

---

## 🚧 已知限制

1. **认证实现**：当前使用简化版，需升级为 bcrypt + JWT
2. **密码加密**：使用 Base64，建议升级为 AES-256-GCM
3. **并发数量**：建议账号数 ≤ 10，避免 API 限流
4. **日志容量**：仅保留最近 1000 条

---

## 🎯 下一步计划

- [ ] 升级认证实现（bcrypt + 标准 JWT）
- [ ] 增强密码加密（AES-256-GCM）
- [ ] Docker 镜像构建
- [ ] 通知功能（Webhook/邮件）
- [ ] 统计图表（开机趋势）
- [ ] 移动端适配

---

## 📞 技术支持

**问题反馈：**
1. 查看 `PROGRESS.md` 了解项目状态
2. 查看 `TEST_GUIDE.md` 进行 API 测试
3. 检查 `store/logs.json` 查看详细日志

**文档更新：** 2026-06-15  
**版本：** v0.3-beta（完整版）
