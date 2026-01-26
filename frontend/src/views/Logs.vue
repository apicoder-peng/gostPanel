<template>
  <div class="page-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>操作日志</span>
          <el-button :icon="Refresh" @click="fetchData">刷新</el-button>
        </div>
      </template>

      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-input
          v-model="searchUsername"
          placeholder="搜索用户名"
          :prefix-icon="Search"
          clearable
          style="width: 150px"
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        />
        <el-select v-model="searchAction" placeholder="操作类型" clearable style="width: 120px" @change="handleSearch">
          <el-option label="登录" value="login" />
          <el-option label="创建" value="create" />
          <el-option label="更新" value="update" />
          <el-option label="删除" value="delete" />
          <el-option label="启动" value="start" />
          <el-option label="停止" value="stop" />
        </el-select>
        <el-select v-model="searchResourceType" placeholder="资源类型" clearable style="width: 120px" @change="handleSearch">
          <el-option label="节点" value="node" />
          <el-option label="转发" value="forward" />
          <el-option label="隧道" value="tunnel" />
        </el-select>
        <el-button :icon="Search" @click="handleSearch">搜索</el-button>
      </div>

      <!-- 表格 -->
      <el-table :data="logList" v-loading="loading" style="width: 100%" border>
        <el-table-column prop="id" label="ID" width="70" align="center" />
        <el-table-column prop="username" label="用户" width="120" align="center">
          <template #default="{ row }">
            <el-tag type="info" size="small" effect="plain">{{ row.username }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="action" label="操作" width="100" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="getActionType(row.action)">{{ getActionText(row.action) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="resource_type" label="资源类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.resource_type" :type="getResourceTagType(row.resource_type)" effect="light" size="small">
              {{ getResourceText(row.resource_type) }}
            </el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="170" align="center">
          <template #default="{ row }">
            {{ new Date(row.created_at).toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column prop="ip_address" label="IP 地址" width="150" align="center">
          <template #default="{ row }">
            <el-tooltip v-if="row.user_agent" :content="row.user_agent" placement="top">
              <span style="font-family: monospace; cursor: help">{{ row.ip_address || '-' }}</span>
            </el-tooltip>
            <span v-else style="font-family: monospace">{{ row.ip_address || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="details" label="详情" min-width="200" header-align="center" show-overflow-tooltip />
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchData"
          @current-change="fetchData"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Refresh, Search } from '@element-plus/icons-vue'
import { getLogList } from '@/api/log'

// 列表数据
const logList = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 搜索
const searchUsername = ref('')
const searchAction = ref('')
const searchResourceType = ref('')

// 操作类型
const getActionType = (action) => {
  const map = { login: 'warning', create: 'primary', update: 'warning', delete: 'danger', start: 'success', stop: 'info' }
  return map[action] || ''
}

const getActionText = (action) => {
  const map = { login: '登录', logout: '登出', create: '创建', update: '更新', delete: '删除', start: '启动', stop: '停止', change_password: '改密' }
  return map[action] || action
}

const getResourceText = (type) => {
  const map = { node: '节点', forward: '转发', tunnel: '隧道' }
  return map[type] || type || '-'
}

const getResourceTagType = (type) => {
  const map = { node: '', forward: 'success', tunnel: 'warning' }
  return map[type] || 'info'
}

// 获取数据
const fetchData = async () => {
  loading.value = true
  try {
    const res = await getLogList({
      page: page.value,
      pageSize: pageSize.value,
      username: searchUsername.value,
      action: searchAction.value,
      resource_type: searchResourceType.value
    })
    logList.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (error) {
    console.error('获取日志列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  page.value = 1
  fetchData()
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.page-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
