<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-cards">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">
              <el-icon><Monitor /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.nodes.total }}</div>
              <div class="stat-label">节点总数</div>
              <div class="stat-sub">
                <span class="online">在线 {{ stats.nodes.online }}</span>
                <span class="offline">离线 {{ stats.nodes.offline }}</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);">
              <el-icon><Switch /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.forwards.total }}</div>
              <div class="stat-label">转发规则</div>
              <div class="stat-sub">
                <span class="online">运行 {{ stats.forwards.running }}</span>
                <span class="offline">停止 {{ stats.forwards.stopped }}</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #eb3349 0%, #f45c43 100%);">
              <el-icon><Connection /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.tunnels.total }}</div>
              <div class="stat-label">隧道配置</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);">
              <el-icon><TrendCharts /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.version || 'v1.0.0' }}</div>
              <div class="stat-label">系统版本</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 快速操作 -->
    <el-row :gutter="20" class="quick-actions">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>快速操作</span>
            </div>
          </template>
          <div class="action-buttons">
            <el-button type="primary" :icon="Plus" @click="$router.push('/nodes')">
              添加节点
            </el-button>
            <el-button type="success" :icon="Plus" @click="$router.push('/forwards')">
              添加转发
            </el-button>
            <el-button type="warning" :icon="Plus" @click="$router.push('/tunnels')">
              添加隧道
            </el-button>
            <el-button :icon="Refresh" @click="loadStats">刷新数据</el-button>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>系统信息</span>
            </div>
          </template>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="Gost 版本">v3.2.6</el-descriptions-item>
            <el-descriptions-item label="数据库">SQLite</el-descriptions-item>
            <el-descriptions-item label="登录用户">{{ authStore.username }}</el-descriptions-item>
            <el-descriptions-item label="当前时间">{{ currentTime }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近操作日志 -->
    <el-card shadow="hover" class="recent-logs">
      <template #header>
        <div class="card-header">
          <span>最近操作</span>
          <el-button type="primary" link @click="$router.push('/logs')">查看全部</el-button>
        </div>
      </template>
      <el-table :data="recentLogs" style="width: 100%" v-loading="logsLoading">
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">
            {{ new Date(row.created_at).toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column prop="username" label="用户" width="100" />
        <el-table-column prop="action" label="操作" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="getActionType(row.action)">{{ getActionText(row.action) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="details" label="详情" show-overflow-tooltip />
      </el-table>
      <el-empty v-if="recentLogs.length === 0 && !logsLoading" description="暂无操作记录" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { Monitor, Switch, Connection, TrendCharts, Plus, Refresh } from '@element-plus/icons-vue'
import { useAuthStore } from '@/store/auth'
import { getDashboardStats } from '@/api/stats'
import { getLogList } from '@/api/log'

const authStore = useAuthStore()

// 统计数据
const stats = reactive({
  nodes: { total: 0, online: 0, offline: 0 },
  forwards: { total: 0, running: 0, stopped: 0 },
  tunnels: { total: 0 },
  version: ''
})

// 最近日志
const recentLogs = ref([])
const logsLoading = ref(false)

// 当前时间
const currentTime = ref('')
let timer = null

// 操作类型
const getActionType = (action) => {
  const map = { login: 'success', create: 'primary', update: 'warning', delete: 'danger', start: 'success', stop: 'info' }
  return map[action] || ''
}

const getActionText = (action) => {
  const map = { login: '登录', logout: '登出', create: '创建', update: '更新', delete: '删除', start: '启动', stop: '停止', change_password: '改密' }
  return map[action] || action
}

// 加载统计数据
const loadStats = async () => {
  try {
    const res = await getDashboardStats()
    Object.assign(stats, res.data)
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

// 加载最近日志
const loadRecentLogs = async () => {
  logsLoading.value = true
  try {
    const res = await getLogList({ page: 1, pageSize: 5 })
    recentLogs.value = res.data.list || []
  } catch (error) {
    console.error('获取日志失败:', error)
  } finally {
    logsLoading.value = false
  }
}

// 更新时间
const updateTime = () => {
  currentTime.value = new Date().toLocaleString()
}

onMounted(() => {
  loadStats()
  loadRecentLogs()
  updateTime()
  timer = setInterval(updateTime, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.stat-cards {
  margin-bottom: 0;
}

.stat-card {
  border-radius: 12px;
  height: 100%;
}

.quick-actions .el-card {
  height: 100%;
  border-radius: 12px;
}

.stat-card :deep(.el-card__body) {
  padding: 20px;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 28px;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
  line-height: 1.2;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}

.stat-sub {
  font-size: 12px;
  margin-top: 6px;
  display: flex;
  gap: 12px;
}

.stat-sub .online {
  color: #67c23a;
}

.stat-sub .offline {
  color: #909399;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.action-buttons {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.recent-logs {
  border-radius: 12px;
}
</style>
