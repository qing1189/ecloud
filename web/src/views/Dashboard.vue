<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: #409eff">
              <el-icon :size="30"><User /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.accountCount }}</div>
              <div class="stat-label">总账号数</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: #67c23a">
              <el-icon :size="30"><Monitor /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.machineCount }}</div>
              <div class="stat-label">监控中的云电脑</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: #e6a23c">
              <el-icon :size="30"><VideoPlay /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.bootCount }}</div>
              <div class="stat-label">今日开机次数</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: #f56c6c">
              <el-icon :size="30"><Histogram /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.activeTaskCount }}</div>
              <div class="stat-label">活跃任务数</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 近期操作日志 -->
    <el-card class="recent-logs" shadow="never">
      <template #header>
        <div class="card-header">
          <span>近期操作</span>
          <el-button text type="primary" @click="$router.push('/logs')">
            查看全部 <el-icon><ArrowRight /></el-icon>
          </el-button>
        </div>
      </template>

      <el-table :data="recentLogs" v-loading="loading">
        <el-table-column prop="timestamp" label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.timestamp) }}
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getEventTypeTag(row.type)" size="small">
              {{ getEventTypeName(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="内容" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusTag(row.status)" size="small">
              {{ getStatusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 账号监控状态 -->
    <el-card class="monitor-status" shadow="never">
      <template #header>
        <div class="card-header">
          <span>监控状态</span>
          <el-button text type="primary" @click="$router.push('/monitor')">
            查看详情 <el-icon><ArrowRight /></el-icon>
          </el-button>
        </div>
      </template>

      <el-table :data="monitorStatus" v-loading="statusLoading">
        <el-table-column prop="account_name" label="账号名称" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="row.status === 'running' ? 'success' : 'info'" size="small">
              {{ row.status === 'running' ? '运行中' : '已停止' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_check" label="最后检查" width="180">
          <template #default="{ row }">
            {{ formatTime(row.last_check) }}
          </template>
        </el-table-column>
        <el-table-column prop="last_event" label="最近事件" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { accountAPI, monitorAPI, logAPI } from '@/api'
import { User, Monitor, VideoPlay, Histogram, ArrowRight } from '@element-plus/icons-vue'

const loading = ref(false)
const statusLoading = ref(false)

const stats = ref({
  accountCount: 0,
  machineCount: 0,
  bootCount: 0,
  activeTaskCount: 0
})

const recentLogs = ref([])
const monitorStatus = ref([])

// 加载统计数据
const loadStats = async () => {
  try {
    const [accounts, status, logs] = await Promise.all([
      accountAPI.list(),
      monitorAPI.getAllStatus(),
      logAPI.list({ page: 1, limit: 10 })
    ])

    stats.value.accountCount = accounts.data.length
    stats.value.activeTaskCount = status.data.filter(s => s.status === 'running').length

    // 计算今日开机次数
    const today = new Date().toISOString().split('T')[0]
    stats.value.bootCount = logs.data.events.filter(log =>
      log.type === 'boot_success' && log.timestamp.startsWith(today)
    ).length

    recentLogs.value = logs.data.events.slice(0, 10)
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

// 加载监控状态
const loadMonitorStatus = async () => {
  statusLoading.value = true
  try {
    const res = await monitorAPI.getAllStatus()
    monitorStatus.value = res.data
  } catch (error) {
    console.error('加载监控状态失败:', error)
  } finally {
    statusLoading.value = false
  }
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

// 获取事件类型标签
const getEventTypeTag = (type) => {
  const tags = {
    boot: 'info',
    boot_success: 'success',
    boot_failed: 'danger',
    account_add: 'success',
    account_update: 'warning',
    account_delete: 'danger'
  }
  return tags[type] || 'info'
}

// 获取事件类型名称
const getEventTypeName = (type) => {
  const names = {
    boot: '开机',
    boot_success: '开机成功',
    boot_failed: '开机失败',
    account_add: '添加账号',
    account_update: '更新账号',
    account_delete: '删除账号',
    task_start: '任务启动',
    task_stop: '任务停止'
  }
  return names[type] || type
}

// 获取状态标签
const getStatusTag = (status) => {
  const tags = {
    success: 'success',
    failed: 'danger',
    info: 'info'
  }
  return tags[status] || 'info'
}

// 获取状态名称
const getStatusName = (status) => {
  const names = {
    success: '成功',
    failed: '失败',
    info: '信息'
  }
  return names[status] || status
}

onMounted(() => {
  loadStats()
  loadMonitorStatus()

  // 每30秒刷新一次
  const interval = setInterval(() => {
    loadStats()
    loadMonitorStatus()
  }, 30000)

  // 组件卸载时清除定时器
  return () => clearInterval(interval)
})
</script>

<style scoped>
.dashboard {
  height: 100%;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  border-radius: 8px;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 20px;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #333;
  margin-bottom: 5px;
}

.stat-label {
  font-size: 14px;
  color: #999;
}

.recent-logs,
.monitor-status {
  margin-bottom: 20px;
  border-radius: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
