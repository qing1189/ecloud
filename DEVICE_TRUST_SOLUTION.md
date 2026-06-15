# 🔐 二次验证（设备信任）解决方案

## 问题描述

移动云公众版账号首次登录需要短信验证码进行设备信任。原 CLI 方式是在命令行输入验证码，但 Web 界面需要更友好的交互方式。

---

## 💡 推荐解决方案

### 方案一：账号添加时的验证流程（推荐）

**用户体验流程：**

```
添加账号
  ↓
填写账号信息（用户名、密码）
  ↓
点击"添加" → 后端尝试登录
  ↓
【如果需要验证】
  ↓
弹出验证码输入框
  - 提示：已向 138****1234 发送验证码
  - 输入框：请输入验证码
  - 倒计时：120 秒
  - 按钮：提交验证 / 重新发送
  ↓
提交验证码 → 后端完成设备信任
  ↓
【验证成功】
  ↓
账号添加成功，开始监控
```

**实现要点：**
1. 添加账号时分两步：先登录检查，再验证（如需要）
2. 前端弹窗显示验证码输入框
3. 后端保持登录状态，等待验证码
4. 验证成功后自动保存账号

---

### 方案二：账号状态标记 + 手动验证

**用户体验流程：**

```
添加账号（未验证）
  ↓
账号列表显示状态："待验证"
  ↓
点击"验证设备"按钮
  ↓
弹出验证窗口
  - 发送验证码
  - 输入验证码
  - 提交验证
  ↓
验证成功 → 状态变为"已验证"
  ↓
可以启用监控
```

**实现要点：**
1. 账号增加 `needVerify` 状态字段
2. 未验证的账号不能启用监控
3. 提供"验证设备"功能入口
4. 验证成功后更新状态

---

### 方案三：CLI 预验证 + Web 管理（临时方案）

**用户体验流程：**

```
【服务器上执行】
./ecloud trust
  ↓ 
在 CLI 中完成设备信任
  ↓
【Web 界面】
添加账号（已信任设备）
  ↓
直接启用监控
```

**实现要点：**
1. Web 界面添加说明文档
2. 告知用户首次需要在服务器上运行 `trust` 命令
3. 验证完成后再在 Web 中添加账号

---

## 🎯 推荐实现方案（方案一）

### 后端 API 设计

#### 1. 添加账号接口增强

```go
// POST /api/accounts
// Request:
{
  "name": "测试账号",
  "type": "public",
  "username": "13800138000",
  "password": "password123",
  "interval": 60
}

// Response (需要验证):
{
  "code": 10001,  // 特殊状态码：需要二次验证
  "msg": "需要设备验证",
  "data": {
    "tempId": "temp_abc123",  // 临时 ID，用于后续验证
    "mobile": "138****1234",   // 脱敏手机号
    "expireTime": 120          // 验证码有效期（秒）
  }
}

// Response (直接成功):
{
  "code": 0,
  "msg": "添加成功",
  "data": {
    "id": "acc_123",
    "name": "测试账号"
  }
}
```

#### 2. 提交验证码接口

```go
// POST /api/accounts/verify
// Request:
{
  "tempId": "temp_abc123",  // 临时 ID
  "code": "123456"          // 验证码
}

// Response:
{
  "code": 0,
  "msg": "验证成功",
  "data": {
    "id": "acc_123",
    "name": "测试账号"
  }
}
```

#### 3. 重新发送验证码接口

```go
// POST /api/accounts/resend-code
// Request:
{
  "tempId": "temp_abc123"
}

// Response:
{
  "code": 0,
  "msg": "验证码已发送",
  "data": {
    "expireTime": 120
  }
}
```

---

### 前端实现要点

#### 1. 添加账号流程

```vue
<template>
  <!-- 添加账号对话框 -->
  <el-dialog v-model="addDialogVisible" title="添加账号">
    <el-form :model="form">
      <!-- 账号信息表单 -->
    </el-form>
    
    <template #footer>
      <el-button @click="addDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="handleAdd" :loading="loading">
        添加
      </el-button>
    </template>
  </el-dialog>
  
  <!-- 验证码对话框 -->
  <el-dialog v-model="verifyDialogVisible" title="设备验证" :close-on-click-modal="false">
    <el-alert type="info" :closable="false" style="margin-bottom: 20px">
      已向 {{ verifyInfo.mobile }} 发送验证码
    </el-alert>
    
    <el-form>
      <el-form-item label="验证码">
        <el-input 
          v-model="verifyCode" 
          placeholder="请输入6位验证码"
          maxlength="6"
          @keyup.enter="handleVerify"
        />
      </el-form-item>
      
      <el-form-item>
        <el-text type="info">
          验证码有效期：{{ countdown }} 秒
        </el-text>
      </el-form-item>
    </el-form>
    
    <template #footer>
      <el-button @click="handleResendCode" :disabled="countdown > 0">
        重新发送 {{ countdown > 0 ? `(${countdown}s)` : '' }}
      </el-button>
      <el-button type="primary" @click="handleVerify" :loading="verifying">
        提交验证
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/api'

const addDialogVisible = ref(false)
const verifyDialogVisible = ref(false)
const loading = ref(false)
const verifying = ref(false)

const form = ref({
  name: '',
  type: 'public',
  username: '',
  password: '',
  interval: 60
})

const verifyInfo = ref({
  tempId: '',
  mobile: '',
  expireTime: 0
})

const verifyCode = ref('')
const countdown = ref(0)
let countdownTimer = null

// 添加账号
const handleAdd = async () => {
  loading.value = true
  try {
    const res = await api.addAccount(form.value)
    
    // 需要二次验证
    if (res.code === 10001) {
      verifyInfo.value = res.data
      countdown.value = res.data.expireTime
      startCountdown()
      
      addDialogVisible.value = false
      verifyDialogVisible.value = true
      return
    }
    
    // 直接成功
    ElMessage.success('添加成功')
    addDialogVisible.value = false
    // 刷新列表...
    
  } catch (error) {
    ElMessage.error(error.message || '添加失败')
  } finally {
    loading.value = false
  }
}

// 提交验证码
const handleVerify = async () => {
  if (!verifyCode.value || verifyCode.value.length !== 6) {
    ElMessage.warning('请输入6位验证码')
    return
  }
  
  verifying.value = true
  try {
    await api.verifyAccount({
      tempId: verifyInfo.value.tempId,
      code: verifyCode.value
    })
    
    ElMessage.success('验证成功')
    verifyDialogVisible.value = false
    stopCountdown()
    // 刷新列表...
    
  } catch (error) {
    ElMessage.error(error.message || '验证失败')
  } finally {
    verifying.value = false
  }
}

// 重新发送验证码
const handleResendCode = async () => {
  try {
    const res = await api.resendCode({
      tempId: verifyInfo.value.tempId
    })
    
    countdown.value = res.data.expireTime
    startCountdown()
    ElMessage.success('验证码已重新发送')
    
  } catch (error) {
    ElMessage.error(error.message || '发送失败')
  }
}

// 倒计时
const startCountdown = () => {
  stopCountdown()
  countdownTimer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      stopCountdown()
    }
  }, 1000)
}

const stopCountdown = () => {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
}
</script>
```

---

### 后端实现要点

#### 1. 临时账号存储

```go
// pkg/service/account/temp_store.go

type TempAccount struct {
    ID        string    `json:"id"`
    Client    *ecloud.Client `json:"-"`
    Account   *Account  `json:"account"`
    CreatedAt time.Time `json:"createdAt"`
}

var tempStore = make(map[string]*TempAccount)
var tempMutex sync.RWMutex

// 保存临时账号
func SaveTempAccount(client *ecloud.Client, account *Account) string {
    tempMutex.Lock()
    defer tempMutex.Unlock()
    
    id := "temp_" + generateID()
    tempStore[id] = &TempAccount{
        ID:        id,
        Client:    client,
        Account:   account,
        CreatedAt: time.Now(),
    }
    
    // 5分钟后自动清理
    time.AfterFunc(5*time.Minute, func() {
        DeleteTempAccount(id)
    })
    
    return id
}

// 获取临时账号
func GetTempAccount(id string) (*TempAccount, error) {
    tempMutex.RLock()
    defer tempMutex.RUnlock()
    
    temp, ok := tempStore[id]
    if !ok {
        return nil, errors.New("临时账号不存在或已过期")
    }
    
    return temp, nil
}

// 删除临时账号
func DeleteTempAccount(id string) {
    tempMutex.Lock()
    defer tempMutex.Unlock()
    
    delete(tempStore, id)
}
```

#### 2. 账号处理器增强

```go
// pkg/api/handlers/account.go

// 添加账号（增强版）
func (h *AccountHandler) Add(c *gin.Context) {
    var req AddAccountRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, "参数错误")
        return
    }
    
    // 创建账号
    account := &account.Account{
        ID:       generateID(),
        Name:     req.Name,
        Type:     req.Type,
        Username: req.Username,
        Password: req.Password,
        Interval: req.Interval,
        Enabled:  false,
    }
    
    // 仅公众版需要处理设备信任
    if req.Type == "public" {
        // 创建客户端
        client, err := ecloud.NewClient(req.Username, req.Password)
        if err != nil {
            response.Error(c, "创建客户端失败")
            return
        }
        
        // 登录
        if _, err := client.Login(); err != nil {
            response.Error(c, "登录失败："+err.Error())
            return
        }
        
        // 检查是否需要设备信任
        if !client.HasTrustDeviceRecord() {
            // 发送验证码
            resp, err := client.SendTrustDeviceVerifySms()
            if err != nil {
                response.Error(c, "发送验证码失败")
                return
            }
            
            // 保存临时账号
            tempID := account.SaveTempAccount(client, account)
            
            // 返回需要验证的响应
            c.JSON(http.StatusOK, gin.H{
                "code": 10001,
                "msg":  "需要设备验证",
                "data": gin.H{
                    "tempId":     tempID,
                    "mobile":     maskMobile(client.GetSession().Mobile),
                    "expireTime": int(resp.Body.(map[string]interface{})["expireTime"].(float64)),
                },
            })
            return
        }
    }
    
    // 不需要验证或政企版，直接保存
    if err := h.accountManager.Add(account); err != nil {
        response.Error(c, err.Error())
        return
    }
    
    response.Success(c, account)
}

// 提交验证码
func (h *AccountHandler) Verify(c *gin.Context) {
    var req struct {
        TempID string `json:"tempId" binding:"required"`
        Code   string `json:"code" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, "参数错误")
        return
    }
    
    // 获取临时账号
    temp, err := account.GetTempAccount(req.TempID)
    if err != nil {
        response.Error(c, err.Error())
        return
    }
    
    // 提交验证码
    resp, err := temp.Client.TrustDevice(req.Code)
    if err != nil {
        response.Error(c, "验证失败："+err.Error())
        return
    }
    
    if !resp.Success() {
        response.Error(c, "验证失败："+resp.ErrorMessage)
        return
    }
    
    // 验证成功，保存账号
    if err := h.accountManager.Add(temp.Account); err != nil {
        response.Error(c, err.Error())
        return
    }
    
    // 清理临时账号
    account.DeleteTempAccount(req.TempID)
    
    response.Success(c, temp.Account)
}

// 重新发送验证码
func (h *AccountHandler) ResendCode(c *gin.Context) {
    var req struct {
        TempID string `json:"tempId" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, "参数错误")
        return
    }
    
    // 获取临时账号
    temp, err := account.GetTempAccount(req.TempID)
    if err != nil {
        response.Error(c, err.Error())
        return
    }
    
    // 重新发送验证码
    resp, err := temp.Client.SendTrustDeviceVerifySms()
    if err != nil {
        response.Error(c, "发送失败："+err.Error())
        return
    }
    
    response.Success(c, gin.H{
        "expireTime": int(resp.Body.(map[string]interface{})["expireTime"].(float64)),
    })
}

// 手机号脱敏
func maskMobile(mobile string) string {
    if len(mobile) != 11 {
        return mobile
    }
    return mobile[:3] + "****" + mobile[7:]
}
```

#### 3. 路由注册

```go
// pkg/api/router.go

// 账号管理
accounts := api.Group("/accounts")
accounts.Use(middleware.Auth())
{
    accounts.POST("", accountHandler.Add)
    accounts.POST("/verify", accountHandler.Verify)      // 新增
    accounts.POST("/resend-code", accountHandler.ResendCode) // 新增
    accounts.GET("", accountHandler.List)
    accounts.PUT("/:id", accountHandler.Update)
    accounts.DELETE("/:id", accountHandler.Delete)
}
```

---

## 📝 使用说明文档更新

在 `WEB_USAGE.md` 中添加：

```markdown
### 设备信任验证

**公众版账号首次添加时需要验证：**

1. 填写账号信息后点击"添加"
2. 系统会向绑定手机号发送验证码
3. 在弹出的对话框中输入验证码
4. 点击"提交验证"完成设备信任
5. 验证成功后账号自动保存并可启用监控

**注意事项：**
- 验证码有效期为 120 秒
- 可点击"重新发送"获取新验证码
- 政企版账号不需要此验证步骤
- 同一账号仅需验证一次
```

---

## 🎯 优先级建议

1. **立即实现**（如果用户有公众版账号）
   - 方案一：完整的验证流程
   - 提供最佳用户体验

2. **临时方案**（快速上线）
   - 方案三：文档说明 + CLI 预验证
   - 快速但需要手动操作

3. **后续优化**（可选）
   - 方案二：账号状态管理
   - 更灵活但复杂度更高

---

## ✅ 建议

**我推荐实现方案一**，因为：
1. 用户体验最好，流程最顺畅
2. 完全在 Web 界面完成，无需切换到 CLI
3. 代码结构清晰，易于维护
4. 后端临时存储机制简单可靠

你觉得哪个方案更合适？我可以立即实现！
