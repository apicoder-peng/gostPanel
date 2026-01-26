<template>
  <div class="page-container">
    <div class="page-header">
      <h3>隧道转发</h3>
    </div>
    <el-card shadow="hover">
      <!-- 搜索栏 -->
      <div class="search-bar">
        <div class="filters">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索名称或目标"
            :prefix-icon="Search"
            clearable
            style="width: 250px"
            @clear="handleSearch"
            @keyup.enter="handleSearch"
          />
          <el-select v-model="searchNodeId" placeholder="选择节点" clearable style="width: 180px" @change="handleSearch">
            <el-option v-for="node in nodeList" :key="node.id" :label="node.name" :value="node.id" />
          </el-select>
          <el-select v-model="searchStatus" placeholder="状态" clearable style="width: 120px" @change="handleSearch">
            <el-option label="运行中" value="running" />
            <el-option label="已停止" value="stopped" />
          </el-select>
          <el-button :icon="Search" @click="handleSearch">搜索</el-button>
          <el-button :icon="Refresh" @click="fetchData">刷新</el-button>
        </div>
        <el-button type="primary" :icon="Plus" @click="openDialog()">添加隧道</el-button>
      </div>

      <!-- 表格 -->
      <el-table :data="tunnelList" v-loading="loading" style="width: 100%" border>
        <el-table-column prop="id" label="ID" width="70" align="center" />
        <el-table-column prop="name" label="隧道名称" min-width="120" align="center" show-overflow-tooltip />
        <el-table-column label="入口节点" width="120" align="center">
          <template #default="{ row }">
            <el-tag size="small" type="primary">{{ row.entry_node?.name || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="出口节点" width="120" align="center">
          <template #default="{ row }">
            <el-tag size="small" type="success">{{ row.exit_node?.name || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="protocol" label="协议" width="80" align="center">
          <template #default="{ row }">
            <el-tag size="small">{{ row.protocol?.toUpperCase() }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="listen_port" label="监听端口" width="100" align="center" />
        <el-table-column label="目标地址" min-width="200" align="center" show-overflow-tooltip>
          <template #default="{ row }">
              <span v-if="row.targets && row.targets.length > 0">{{ row.targets[0] }}<span v-if="row.targets.length > 1"> (+{{ row.targets.length - 1 }})</span></span>
              <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="总流量" width="120" align="center">
          <template #default="{ row }">
            {{ formatBytes(row.total_bytes || 0) }}
          </template>
        </el-table-column>
        <el-table-column label="上传流量" width="120" align="center">
          <template #default="{ row }">
            {{ formatBytes(row.output_bytes || 0) }}
          </template>
        </el-table-column>
        <el-table-column label="下载流量" width="120" align="center">
          <template #default="{ row }">
            {{ formatBytes(row.input_bytes || 0) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">{{ getStatusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" align="center" fixed="right">
          <template #default="{ row }">
            <el-button 
              v-if="row.status !== 'running'" 
              type="success" link size="small" 
              @click="handleStart(row)"
            >启动</el-button>
            <el-button 
              v-else 
              type="warning" link size="small" 
              @click="handleStop(row)"
            >停止</el-button>
            <el-button type="primary" link size="small" @click="openDialog(row)">编辑</el-button>
            <el-button type="info" link size="small" @click="handleCopy(row)">复制</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
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

    <!-- 添加/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑隧道' : '添加隧道'"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="隧道名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入隧道名称" :prefix-icon="EditPen" />
        </el-form-item>
        <el-divider content-position="left">路由配置</el-divider>
        <el-form-item label="入口节点" prop="entry_node_id">
          <el-select v-model="form.entry_node_id" placeholder="选择入口节点（客户端连接的节点）" style="width: 100%">
            <el-option 
              v-for="node in nodeList" 
              :label="node.name" 
              :value="node.id" 
              :disabled="node.id === form.exit_node_id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="出口节点" prop="exit_node_id">
          <el-select v-model="form.exit_node_id" placeholder="选择出口节点（访问目标的节点）" style="width: 100%">
            <el-option 
              v-for="node in nodeList" 
              :key="node.id" 
              :label="node.name" 
              :value="node.id" 
              :disabled="node.id === form.entry_node_id"
            />
          </el-select>
        </el-form-item>
        <el-divider content-position="left">端口配置</el-divider>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="协议" prop="protocol">
              <el-select v-model="form.protocol" style="width: 100%">
                <el-option label="TCP" value="tcp" />
                <el-option label="UDP" value="udp" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="监听端口" prop="listen_port">
              <el-input-number v-model="form.listen_port" :min="1" :max="65535" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-divider content-position="left">目标地址（最终访问的地址）</el-divider>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="负载均衡" prop="strategy">
               <el-select v-model="form.strategy" placeholder="默认为轮询">
                  <el-option label="轮询 (Round Robin)" value="round"/>
                  <el-option label="随机 (Random)" value="rand"/>
                  <el-option label="先进先出 (FIFO)" value="fifo"/>
                  <el-option label="哈希 (Hash)" value="hash"/>
               </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="目标列表" style="margin-bottom: 0;">
           <el-table :data="form.targetList" border style="width: 100%" size="small" :show-header="true">
              <el-table-column label="目标地址 (IP:Port)" min-width="250">
                  <template #default="{ row }">
                      <el-input v-model="row.address" placeholder="例如: 192.168.1.100:8080" />
                  </template>
              </el-table-column>
              <el-table-column label="操作" width="60" align="center">
                  <template #default="{ $index }">
                      <el-button type="danger" link :icon="UseRemove" @click="removeTarget($index)" />
                  </template>
              </el-table-column>
           </el-table>
           <div style="margin-top: 10px; text-align: center; width: 100%;">
               <el-button type="primary" link :icon="Plus" @click="addTarget" style="width: 100%; border: 1px dashed #dcdfe6;">添加目标地址</el-button>
           </div>
        </el-form-item>


        <el-form-item label="备注说明" prop="remark">
          <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="备注信息" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search, EditPen, Connection, Link, Location, DataLine, Top, Bottom, Remove as UseRemove } from '@element-plus/icons-vue'
import { getTunnelList, createTunnel, updateTunnel, deleteTunnel, startTunnel, stopTunnel } from '@/api/tunnel'
import { getNodeList } from '@/api/node'

// 节点列表
const nodeList = ref([])

// 列表数据
const tunnelList = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 搜索
const searchKeyword = ref('')
const searchNodeId = ref('')
const searchStatus = ref('')

// 对话框
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const submitLoading = ref(false)
const formRef = ref(null)

const form = reactive({
  name: '',
  entry_node_id: '',
  exit_node_id: '',
  protocol: 'tcp',
  listen_port: 0,

  targetList: [{ address: '' }],
  strategy: 'round',
  relay_port: 0,
  remark: ''
})

const rules = {
  name: [{ required: true, message: '请输入隧道名称', trigger: 'blur' }],
  entry_node_id: [{ required: true, message: '请选择入口节点', trigger: 'change' }],
  exit_node_id: [{ required: true, message: '请选择出口节点', trigger: 'change' }],
  protocol: [{ required: true, message: '请选择协议', trigger: 'change' }],
  listen_port: [{ required: true, message: '请输入监听端口', trigger: 'blur' }],
  listen_port: [{ required: true, message: '请输入监听端口', trigger: 'blur' }]
}

// 状态处理
const getStatusType = (status) => {
  const map = { running: 'success', stopped: 'info', error: 'danger' }
  return map[status] || 'info'
}

const getStatusText = (status) => {
  const map = { running: '运行中', stopped: '已停止', error: '错误' }
  return map[status] || status
}

// 格式化字节数
const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
}

// 获取节点列表
const fetchNodes = async () => {
  try {
    const res = await getNodeList({ pageSize: 100 })
    nodeList.value = res.data.list || []
  } catch (error) {
    console.error('获取节点列表失败:', error)
  }
}

// 搜索
const handleSearch = () => {
  page.value = 1
  fetchData()
}

// 获取数据
const fetchData = async (isSilent = false) => {
  if (!isSilent) loading.value = true
  try {
    const res = await getTunnelList({
      page: page.value,
      pageSize: pageSize.value,
      node_id: searchNodeId.value,
      status: searchStatus.value,
      keyword: searchKeyword.value
    })
    tunnelList.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (error) {
    console.error('获取隧道列表失败:', error)
  } finally {
    if (!isSilent) loading.value = false
  }
}

// 打开对话框
const openDialog = (row = null) => {
  isEdit.value = !!row
  editId.value = row?.id || null
  
  if (row) {
    // 解析 targets
    let tList = []
    if (row.targets && row.targets.length > 0) {
        tList = row.targets.map(t => ({ address: t }))
    }

    Object.assign(form, {
      name: row.name,
      entry_node_id: row.entry_node_id,
      exit_node_id: row.exit_node_id,
      protocol: row.protocol || 'tcp',
      listen_port: row.listen_port,
      targetList: tList,
      strategy: row.strategy || 'round',
      remark: row.remark || ''
    })
  } else {
    Object.assign(form, {
      name: '',
      entry_node_id: '',
      exit_node_id: '',
      protocol: 'tcp',
      listen_port: 8080,
      targetList: [{ address: '' }],
      strategy: 'round',
      remark: ''
    })
  }
  
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitLoading.value = true
    try {
      // 准备提交数据
      // 准备提交数据
      const targets = form.targetList.map(item => item.address).filter(t => t.trim() !== '')

      const { targetList, ...rest } = form
      const submitData = {
        ...rest,
        targets: targets
      }

      if (isEdit.value) {
        await updateTunnel(editId.value, submitData)
        ElMessage.success('更新成功')
      } else {
        await createTunnel(submitData)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      fetchData()
    } catch (error) {
      console.error('操作失败:', error)
    } finally {
      submitLoading.value = false
    }
  })
}

// 复制隧道
const handleCopy = (row) => {
  isEdit.value = false
  editId.value = null
  
  Object.assign(form, {
    name: row.name,
    entry_node_id: row.entry_node_id,
    exit_node_id: row.exit_node_id,
    protocol: row.protocol || 'tcp',
    listen_port: row.listen_port + 1,
    targetList: (row.targets && row.targets.length > 0) ? row.targets.map(t => ({ address: t })) : [{ address: '' }],
    strategy: row.strategy || 'round',
    remark: row.remark || ''
  })
  
  if (row.targets && row.targets.length > 0) {
      form.targetList = row.targets.map(t => ({ address: t }))
  }
  
  dialogVisible.value = true
}

// 删除隧道
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要删除隧道 "${row.name}" 吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await deleteTunnel(row.id)
    ElMessage.success('删除成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
    }
  }
}

// 启动隧道
const handleStart = async (row) => {
  try {
    await startTunnel(row.id)
    ElMessage.success('启动成功')
    fetchData()
  } catch (error) {
    console.error('启动失败:', error)
  }
}

// 停止隧道
const handleStop = async (row) => {
  try {
    await stopTunnel(row.id)
    ElMessage.success('停止成功')
    fetchData()
  } catch (error) {
    console.error('停止失败:', error)
  }
}

// Ping 类型判断
const getPingType = (ping) => {
  if (ping < 100) return 'success'
  if (ping < 300) return 'warning'
  return 'danger'
}

// 定时刷新
let refreshTimer = null

onMounted(() => {
  fetchNodes()
  fetchData()
  
  // 每 5 秒刷新一次 (静默刷新)
  refreshTimer = setInterval(() => {
    fetchData(true)
  }, 5000)
})

onBeforeUnmount(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})

// 添加目标
const addTarget = () => {
    form.targetList.push({ address: '' })
}

// 移除目标
const removeTarget = (index) => {
    form.targetList.splice(index, 1)
}
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

.page-header h3 {
  margin: 0 0 16px 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.search-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.filters {
  display: flex;
  gap: 12px;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

/* 表格行高优化 */
:deep(.el-table .el-table__cell) {
  padding: 12px 0;
}

/* 限速限流标签 */
.limit-tags {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

/* 流量统计 */
.stats-info {
  display: flex;
  flex-direction: row;
  gap: 8px;
  font-size: 12px;
  align-items: center;
  justify-content: center;
}

.stat-divider {
  color: #dcdfe6;
  margin: 0 4px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.text-muted {
  color: #909399;
  font-size: 12px;
}

/* 表单提示 */
.form-tip {
  margin-left: 8px;
  color: #909399;
  font-size: 12px;
}

.upload-icon {
  color: #67c23a;
}

.download-icon {
  color: #409eff;
}

.ml-2 {
  margin-left: 8px;
}
</style>
