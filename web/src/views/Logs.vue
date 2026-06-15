<template>
  <div class="logs">
    <el-card shadow="never">
      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-form :inline="true" :model="filters">
          <el-form-item label="事件类型">
            <el-select
              v-model="filters.type"
              placeholder="全部类型"
              clearable
              style="width: 150px"
            >
              <el-option label="开机" value="boot" />
              <el-option label="开机成功" value="boot_success" />
              <el-option label="开机失败" value="boot_failed" />
              <el-option label="添加账号" value="account_add" />
              <el-option label="更新账号" value="account_update" />
              <el-option label="删除账号" value="account_delete" />
              <el-option label="任务启动" value="task_start" />
              <el-option label="任务停止" value="task_stop" />
            </el-select>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="handleSearch">
              <el-icon><Search /></el-icon>
              查询
            </el-button>
            <el-button @click="handleReset">
              <el-icon><Refresh /></el-icon>
              重置
            </el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 日志列表 -->
      <el-table :data="logs" v-loading="loading" style="margin-top: 20px">
        <el-table-column prop="timestamp" label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.timestamp) }}
          </template>
        </el-table-column>

        <el-table-column prop="type" label="事件类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getEventTypeTag(row.type)" size="small">
              {{ getEventTypeName(row.type) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="message" label="内容" min-width="300" />

        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusTag(row.status)" size="small">
              {{ getStatusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="详情" width="100">
          <template #default="{ row }">
            <el-button
              v-if="row.details"
              text
              type="primary"
              size="small"
              @click="showDetails(row)"
            >
              查看
            </el-button>
            <span v-else style="color: #ccc">-</span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.limit"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadLogs"
          @current-change="loadLogs"
        />
      </div>
    </el-card>

    <!-- 详情对话框 -->
    <el-dialog v-model="detailsVisible" title="事件详情" width="600px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="事件ID">
          {{ currentLog?.id }}
        </el-descriptions-item>
        <el-descriptions-item label="时间">
          {{ formatTime(currentLog?.timestamp) }}
        </el-descriptions-item>
        <el-descriptions-item label="类型">
          <el-tag :type="getEventTypeTag(currentLog?.type)" size="small">
            {{ getEventTypeName(currentLog?.type) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusTag(currentLog?.status)" size="small">
            {{ getStatusName(currentLog?.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="消息">
          {{ currentLog?.message }}
        </el-descriptions-item>
      </el-descriptions>

      <div v-if="currentLog?.details" style="margin-top: 20px">
        <h4>详细信息</h4>
        <pre style="background: #f5f5f5; padding: 10px; border-radius: 4px; overflow-x: auto">{{
          JSON.stringify(currentLog.details, null, 2)
        }}</pre>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { logAPI } from '@/api'
import { Search, Refresh } from '@element-plus/icons-vue'

const loading = ref(false)
const logs = ref([])
const detailsVisible = ref(false)
const currentLog = ref(null)

const filters = reactive({
  type: '',
  account_id: ''
})

const pagination = reactive({
  page: 1,
  limit: 20,
  total: 0
})

// 加载日志列表
const loadLogs = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      limit: pagination.limit
    }

    if (filters.type) {
      params.type = filters.type
    }
    if (filters.account_id) {
      params.account_id = filters.account_id
    }

    const res = await logAPI.list(params)
    logs.value = res.data.events || []
    pagination.total = res.data.total || 0
  } catch (error) {
    console.error('加载日志失败:', error)
  } finally {
    loading.value = false
  }
}

// 查询
const handleSearch = () => {
  pagination.page = 1
  loadLogs()
}

// 重置
const handleReset = () => {
  filters.type = ''
  filters.account_id = ''
  pagination.page = 1
  loadLogs()
}

// 显示详情
const showDetails = (row) => {
  currentLog.value = row
  detailsVisible.value = true
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
    account_delete: 'danger',
    task_start: 'success',
    task_stop: 'info',
    login: 'success',
    login_failed: 'danger',
    config_change: 'warning'
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
    task_stop: '任务停止',
    login: '登录',
    login_failed: '登录失败',
    config_change: '配置变更'
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
  loadLogs()
})
</script>

<style scoped>
.logs {
  height: 100%;
}

.filter-bar {
  padding-bottom: 10px;
  border-bottom: 1px solid #eee;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}
</style>
