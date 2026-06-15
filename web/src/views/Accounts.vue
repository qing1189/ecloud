<template>
  <div class="accounts">
    <el-card shadow="never">
      <!-- 头部操作栏 -->
      <div class="toolbar">
        <el-button type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          添加账号
        </el-button>
        <el-button @click="loadAccounts">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>

      <!-- 账号列表 -->
      <el-table :data="accounts" v-loading="loading" style="margin-top: 20px">
        <el-table-column prop="name" label="账号名称" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.type === 'public' ? 'success' : 'warning'" size="small">
              {{ row.type === 'public' ? '公众版' : '政企版' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column label="监控状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.monitor_config.enabled ? 'success' : 'info'" size="small">
              {{ row.monitor_config.enabled ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="检查间隔" width="100">
          <template #default="{ row }">
            {{ row.monitor_config.interval }}秒
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" text type="primary" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button
              size="small"
              text
              :type="row.monitor_config.enabled ? 'warning' : 'success'"
              @click="handleToggle(row)"
            >
              {{ row.monitor_config.enabled ? '停用' : '启用' }}监控
            </el-button>
            <el-button size="small" text type="danger" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'add' ? '添加账号' : '编辑账号'"
      width="600px"
      @close="resetForm"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
      >
        <el-form-item label="账号名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入账号名称" />
        </el-form-item>

        <el-form-item label="账号类型" prop="type">
          <el-radio-group v-model="form.type" :disabled="dialogMode === 'edit'">
            <el-radio value="public">公众版</el-radio>
            <el-radio value="business">政企版</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- 公众版字段 -->
        <template v-if="form.type === 'public'">
          <el-form-item label="用户名" prop="username">
            <el-input v-model="form.username" placeholder="手机号" :disabled="dialogMode === 'edit'" />
          </el-form-item>
          <el-form-item label="密码" :prop="dialogMode === 'add' ? 'password' : ''">
            <el-input
              v-model="form.password"
              type="password"
              show-password
              :placeholder="dialogMode === 'edit' ? '留空则不修改' : '请输入密码'"
            />
          </el-form-item>
          <el-alert
            title="提示：公众版账号需要先在本地运行 trust 命令进行设备信任"
            type="warning"
            :closable="false"
            style="margin-bottom: 20px"
          />
        </template>

        <!-- 政企版字段 -->
        <template v-if="form.type === 'business'">
          <el-form-item label="Access Key" prop="access_key">
            <el-input v-model="form.access_key" :disabled="dialogMode === 'edit'" />
          </el-form-item>
          <el-form-item label="Secret Key" :prop="dialogMode === 'add' ? 'secret_key' : ''">
            <el-input
              v-model="form.secret_key"
              type="password"
              show-password
              :placeholder="dialogMode === 'edit' ? '留空则不修改' : '请输入 Secret Key'"
            />
          </el-form-item>
          <el-form-item label="Pool ID" prop="pool_id">
            <el-input v-model="form.pool_id" placeholder="默认：CIDC-CORE-00" />
          </el-form-item>
        </template>

        <el-divider />

        <el-form-item label="启用监控">
          <el-switch v-model="form.monitor_config.enabled" />
        </el-form-item>

        <el-form-item label="检查间隔(秒)" prop="monitor_config.interval">
          <el-input-number
            v-model="form.monitor_config.interval"
            :min="30"
            :max="3600"
            :step="10"
          />
        </el-form-item>

        <el-form-item label="监控机器">
          <el-input
            v-model="machineIdsText"
            type="textarea"
            :rows="3"
            placeholder="留空表示监控所有机器，多个机器ID用逗号分隔"
          />
          <span style="font-size: 12px; color: #999">
            格式：machine_id_1,machine_id_2
          </span>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 设备验证对话框 -->
    <el-dialog
      v-model="verifyDialogVisible"
      title="设备验证"
      width="500px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
    >
      <el-alert type="info" :closable="false" style="margin-bottom: 20px">
        已向 {{ verifyInfo.mobile }} 发送验证码
      </el-alert>

      <el-form label-width="80px">
        <el-form-item label="验证码">
          <el-input
            v-model="verifyCode"
            placeholder="请输入6位验证码"
            maxlength="6"
            @keyup.enter="handleVerifySubmit"
          >
            <template #append>
              <el-button
                :disabled="countdown > 0"
                @click="handleResendCode"
                style="width: 120px"
              >
                {{ countdown > 0 ? `${countdown}秒后重发` : '重新发送' }}
              </el-button>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item>
          <el-text type="info" size="small">
            验证码有效期：{{ verifyInfo.expireTime }} 秒
          </el-text>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="handleCancelVerify">取消</el-button>
        <el-button type="primary" @click="handleVerifySubmit" :loading="verifying">
          提交验证
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { accountAPI } from '@/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'

const loading = ref(false)
const submitting = ref(false)
const accounts = ref([])

const dialogVisible = ref(false)
const dialogMode = ref('add') // 'add' or 'edit'
const formRef = ref(null)

// 验证相关
const verifyDialogVisible = ref(false)
const verifying = ref(false)
const verifyCode = ref('')
const verifyInfo = ref({
  tempId: '',
  mobile: '',
  expireTime: 0
})
const countdown = ref(0)
let countdownTimer = null

const form = reactive({
  id: '',
  name: '',
  type: 'public',
  username: '',
  password: '',
  access_key: '',
  secret_key: '',
  pool_id: 'CIDC-CORE-00',
  monitor_config: {
    enabled: true,
    interval: 60,
    machines: []
  }
})

const machineIdsText = ref('')

const rules = {
  name: [
    { required: true, message: '请输入账号名称', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择账号类型', trigger: 'change' }
  ],
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ],
  access_key: [
    { required: true, message: '请输入 Access Key', trigger: 'blur' }
  ],
  secret_key: [
    { required: true, message: '请输入 Secret Key', trigger: 'blur' }
  ],
  'monitor_config.interval': [
    { required: true, message: '请输入检查间隔', trigger: 'blur' }
  ]
}

// 加载账号列表
const loadAccounts = async () => {
  loading.value = true
  try {
    const res = await accountAPI.list()
    accounts.value = res.data
  } catch (error) {
    console.error('加载账号列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 添加账号
const handleAdd = () => {
  dialogMode.value = 'add'
  dialogVisible.value = true
}

// 编辑账号
const handleEdit = (row) => {
  dialogMode.value = 'edit'
  Object.assign(form, {
    id: row.id,
    name: row.name,
    type: row.type,
    username: row.username || '',
    password: '',
    access_key: row.access_key || '',
    secret_key: '',
    pool_id: row.pool_id || 'CIDC-CORE-00',
    monitor_config: { ...row.monitor_config }
  })
  machineIdsText.value = row.monitor_config.machines.join(',')
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    // 处理机器ID列表
    form.monitor_config.machines = machineIdsText.value
      ? machineIdsText.value.split(',').map(id => id.trim()).filter(id => id)
      : []

    submitting.value = true
    try {
      if (dialogMode.value === 'add') {
        const res = await accountAPI.add(form)

        // 检查是否需要验证
        if (res.data && res.data.needVerify) {
          // 需要设备验证
          verifyInfo.value = {
            tempId: res.data.tempId,
            mobile: res.data.mobile,
            expireTime: res.data.expireTime
          }
          countdown.value = res.data.expireTime
          startCountdown()

          // 关闭添加对话框，打开验证对话框
          dialogVisible.value = false
          verifyDialogVisible.value = true
          ElMessage.info('需要验证设备，请输入验证码')
        } else {
          // 直接成功
          ElMessage.success('添加成功')
          dialogVisible.value = false
          loadAccounts()
        }
      } else {
        const updateData = {
          name: form.name,
          monitor_config: form.monitor_config
        }
        if (form.password) {
          updateData.password = form.password
        }
        if (form.secret_key) {
          updateData.secret_key = form.secret_key
        }
        await accountAPI.update(form.id, updateData)
        ElMessage.success('更新成功')
        dialogVisible.value = false
        loadAccounts()
      }
    } catch (error) {
      console.error('提交失败:', error)
    } finally {
      submitting.value = false
    }
  })
}

// 切换监控状态
const handleToggle = async (row) => {
  try {
    await accountAPI.toggle(row.id)
    ElMessage.success('操作成功')
    loadAccounts()
  } catch (error) {
    console.error('切换监控状态失败:', error)
  }
}

// 删除账号
const handleDelete = (row) => {
  ElMessageBox.confirm(`确定要删除账号 "${row.name}" 吗？此操作不可恢复。`, '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await accountAPI.delete(row.id)
      ElMessage.success('删除成功')
      loadAccounts()
    } catch (error) {
      console.error('删除失败:', error)
    }
  }).catch(() => {})
}

// 重置表单
const resetForm = () => {
  Object.assign(form, {
    id: '',
    name: '',
    type: 'public',
    username: '',
    password: '',
    access_key: '',
    secret_key: '',
    pool_id: 'CIDC-CORE-00',
    monitor_config: {
      enabled: true,
      interval: 60,
      machines: []
    }
  })
  machineIdsText.value = ''
  if (formRef.value) {
    formRef.value.resetFields()
  }
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

// 提交验证码
const handleVerifySubmit = async () => {
  if (!verifyCode.value || verifyCode.value.length !== 6) {
    ElMessage.warning('请输入6位验证码')
    return
  }

  verifying.value = true
  try {
    await accountAPI.verifyDevice({
      tempId: verifyInfo.value.tempId,
      code: verifyCode.value
    })

    ElMessage.success('验证成功，账号添加完成')
    verifyDialogVisible.value = false
    stopCountdown()
    verifyCode.value = ''
    loadAccounts()
  } catch (error) {
    console.error('验证失败:', error)
    ElMessage.error(error.message || '验证码错误或已过期')
  } finally {
    verifying.value = false
  }
}

// 重新发送验证码
const handleResendCode = async () => {
  try {
    const res = await accountAPI.resendCode({
      tempId: verifyInfo.value.tempId
    })

    countdown.value = res.data.expireTime
    verifyInfo.value.expireTime = res.data.expireTime
    startCountdown()
    ElMessage.success('验证码已重新发送')
  } catch (error) {
    console.error('发送失败:', error)
    ElMessage.error(error.message || '发送验证码失败')
  }
}

// 取消验证
const handleCancelVerify = () => {
  ElMessageBox.confirm(
    '取消验证将放弃添加此账号，确定要取消吗？',
    '提示',
    {
      confirmButtonText: '确定',
      cancelButtonText: '继续验证',
      type: 'warning'
    }
  ).then(() => {
    verifyDialogVisible.value = false
    stopCountdown()
    verifyCode.value = ''
  }).catch(() => {})
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

onMounted(() => {
  loadAccounts()
})

onBeforeUnmount(() => {
  stopCountdown()
})
</script>

<style scoped>
.accounts {
  height: 100%;
}

.toolbar {
  display: flex;
  gap: 10px;
}
</style>
