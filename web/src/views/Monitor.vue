<template>
  <div class="monitor">
    <el-card shadow="never">
      <!-- 头部操作栏 -->
      <div class="toolbar">
        <el-button @click="loadStatus">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="primary" @click="handleReloadAll">
          <el-icon><RefreshRight /></el-icon>
          重新加载所有任务
        </el-button>
        <div class="auto-refresh">
          <el-switch v-model="autoRefresh" />
          <span style="margin-left: 10px">自动刷新（30秒）</span>
        </div>
      </div>

      <!-- 监控状态列表 -->
      <el-table :data="statusList" v-loading="loading" style="margin-top: 20px">
        <el-table-column prop="account_name" label="账号名称" width="200" />
        <el-table-column prop="status" label="任务状态" width="120">
          <template #default="{ row }">
            <el-tag
              :type="getStatusType(row.status)"
              size="small"
            >
              <el-icon style="margin-right: 4px">
                <component :is="getStatusIcon(row.status)" />
              </el-icon>
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="machine_count" label="监控机器数" width="120">
          <template #default="{ row }">
            <el-tag type="info" size="small">
              {{ row.machine_count }} 台
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_check" label="最后检查时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.last_check) }}
          </template>
        </el-table-column>
        <el-table-column prop="last_event" label="最近事件" min-width="200">
          <template #default="{ row }">
            <span v-if="row.last_event" style="color: #666">
              {{ row.last_event }}
            </span>
            <span v-else style="color: #ccc">-</span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 空状态 -->
      <el-empty
        v-if="!loading && statusList.length === 0"
        description="暂无监控任务"
      >
        <el-button type="primary" @click="$router.push('/accounts')">
          去添加账号
        </el-button>
      </el-empty>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { monitorAPI } from '@/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Refresh,
  RefreshRight,
  CircleCheck,
  CircleClose,
  Loading
} from '@element-plus/icons-vue'

const loading = ref(false)
const statusList = ref([])
const autoRefresh = ref(true)
let refreshTimer = null

// 加载监控状态
const loadStatus = async () => {
  loading.value = true
  try {
    const res = await monitorAPI.getAllStatus()
    statusList.value = res.data || []
  } catch (error) {
    console.error('加载监控状态失败:', error)
  } finally {
    loading.value = false
  }
}

// 重新加载所有任务
const handleReloadAll = () => {
  ElMessageBox.confirm(
    '重新加载会重启所有监控任务，确定继续吗？',
    '提示',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      await monitorAPI.reload()
      ElMessage.success('任务已重新加载')
      loadStatus()
    } catch (error) {
      console.error('重新加载失败:', error)
    }
  }).catch(() => {})
}

// 获取状态类型
const getStatusType = (status) => {
  const types = {
    running: 'success',
    stopped: 'info',
    error: 'danger'
  }
  return types[status] || 'info'
}

// 获取状态图标
const getStatusIcon = (status) => {
  const icons = {
    running: CircleCheck,
    stopped: CircleClose,
    error: CircleClose
  }
  return icons[status] || Loading
}

// 获取状态文本
const getStatusText = (status) => {
  const texts = {
    running: '运行中',
    stopped: '已停止',
    error: '错误'
  }
  return texts[status] || status
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return '-'
  const date = new Date(time)
  const now = new Date()
  const diff = Math.floor((now - date) / 1000)

  if (diff < 60) return `${diff}秒前`
  if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`
  if (diff < 86400) return `${Math.floor(diff / 3600)}小时前`
  return date.toLocaleString('zh-CN')
}

// 启动自动刷新
const startAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  refreshTimer = setInterval(() => {
    if (autoRefresh.value) {
      loadStatus()
    }
  }, 30000)
}

// 监听自动刷新开关
watch(autoRefresh, (val) => {
  if (val) {
    startAutoRefresh()
  } else if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})

onMounted(() => {
  loadStatus()
  if (autoRefresh.value) {
    startAutoRefresh()
  }
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.monitor {
  height: 100%;
}

.toolbar {
  display: flex;
  gap: 10px;
  align-items: center;
}

.auto-refresh {
  margin-left: auto;
  display: flex;
  align-items: center;
  font-size: 14px;
  color: #666;
}
</style>
