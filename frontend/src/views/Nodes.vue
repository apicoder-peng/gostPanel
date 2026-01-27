<template>
  <div class="page-container">
    <div class="page-header">
      <h3>节点管理</h3>
    </div>
    <el-card shadow="hover">
      <!-- 搜索栏 -->
      <div class="search-bar">
        <div class="filters">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索节点名称或地址"
            :prefix-icon="Search"
            clearable
            style="width: 250px"
            @clear="handleSearch"
            @keyup.enter="handleSearch"
          />
          <el-select v-model="searchStatus" placeholder="状态" clearable style="width: 120px" @change="handleSearch">
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
          </el-select>
          <el-button :icon="Search" @click="handleSearch">搜索</el-button>
          <el-button :icon="Refresh" @click="fetchData">刷新</el-button>
        </div>
        <el-button type="primary" :icon="Plus" @click="openDialog()">添加节点</el-button>
      </div>

      <!-- 表格 -->
      <el-table :data="nodeList" v-loading="loading" style="width: 100%" border>
        <el-table-column prop="id" label="ID" width="70" align="center" />
        <el-table-column prop="name" label="节点名称" min-width="120" align="center" show-overflow-tooltip />
        <el-table-column prop="address" label="IP/域名" min-width="150" align="center" show-overflow-tooltip />
        <el-table-column prop="port" label="端口" width="100" align="center" />

        <el-table-column prop="total_bytes" label="总流量" width="120" align="center">
          <template #default="{ row }">
            {{ formatBytes(row.total_bytes || 0) }}
          </template>
        </el-table-column>
        <el-table-column prop="input_bytes" label="上传流量" width="120" align="center">
          <template #default="{ row }">
            <span style="color: #67c23a">{{ formatBytes(row.input_bytes || 0) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="output_bytes" label="下载流量" width="120" align="center">
          <template #default="{ row }">
            <span style="color: #409eff">{{ formatBytes(row.output_bytes || 0) }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'online' ? 'success' : 'danger'" size="small">
              {{ row.status === 'online' ? '在线' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="last_check_at" label="最后检查" width="170" align="center">
          <template #default="{ row }">
            {{ row.last_check_at ? new Date(row.last_check_at).toLocaleString() : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" align="center" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status !== 'online'" type="warning" link size="small" @click="showInstallCommand(row)">安装</el-button>
            <el-button type="success" link size="small" @click="handleViewConfig(row)">配置</el-button>
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
      :title="isEdit ? '编辑节点' : '添加节点'"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="节点名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入节点名称" :prefix-icon="Management" />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="IP/域名" prop="address">
              <el-input v-model="form.address" placeholder="例如: 1.2.3.4" :prefix-icon="Link" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="API 端口" prop="port">
              <el-input-number v-model="form.port" :min="1" :max="65535" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="认证用户" prop="username">
              <el-input v-model="form.username" placeholder="用户名" :prefix-icon="User" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="认证密码" prop="password">
              <el-input v-model="form.password" placeholder="密码" :prefix-icon="Lock" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="备注说明" prop="remark">
          <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="备注信息" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 配置对话框 -->
    <el-dialog v-model="configDialogVisible" title="节点配置" width="800px">
      <el-tabs v-model="configActiveTab" type="border-card">
        <el-tab-pane label="服务(Services)" name="services">
          <el-table :data="nodeConfig?.services || []" style="width: 100%" border size="small">
             <el-table-column prop="name" label="名称" width="150" />
             <el-table-column prop="addr" label="监听地址" width="180" />
             <el-table-column prop="handler.type" label="协议" width="100">
               <template #default="{ row }">
                 <el-tag size="small">{{ row.handler?.type || '-' }}</el-tag>
               </template>
             </el-table-column>
             <el-table-column prop="handler.chain" label="关联链" />
             <el-table-column prop="forwarder.nodes" label="转发目标">
                <template #default="{ row }">
                   <span v-if="row.forwarder?.nodes?.length">{{ row.forwarder.nodes.length }} 个目标</span>
                   <span v-else>-</span>
                </template>
             </el-table-column>
             <el-table-column type="expand" label="详情" width="60">
                <template #default="{ row }">
                   <pre style="padding: 10px; background: #f5f7fa; border-radius: 4px; font-size: 12px;">{{ JSON.stringify(row, null, 2) }}</pre>
                </template>
             </el-table-column>
          </el-table>
        </el-tab-pane>
        
        <el-tab-pane label="转发链(Chains)" name="chains">
          <el-table :data="nodeConfig?.chains || []" style="width: 100%" border size="small">
             <el-table-column prop="name" label="名称" width="150" />
             <el-table-column label="跳数 (Hops)">
                <template #default="{ row }">
                   {{ row.hops?.length || 0 }}
                </template>
             </el-table-column>
             <el-table-column type="expand" label="详情" width="60">
                <template #default="{ row }">
                   <pre style="padding: 10px; background: #f5f7fa; border-radius: 4px; font-size: 12px;">{{ JSON.stringify(row, null, 2) }}</pre>
                </template>
             </el-table-column>
          </el-table>
        </el-tab-pane>



        <el-tab-pane label="观察器(Observers)" name="observers">
           <el-table :data="nodeConfig?.observers || []" style="width: 100%" border size="small">
             <el-table-column prop="name" label="名称" width="150" />
             <el-table-column prop="plugin.type" label="类型" width="100" />
             <el-table-column prop="plugin.addr" label="地址" />
             <el-table-column type="expand" label="详情" width="60">
                <template #default="{ row }">
                   <pre style="padding: 10px; background: #f5f7fa; border-radius: 4px; font-size: 12px;">{{ JSON.stringify(row, null, 2) }}</pre>
                </template>
             </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>

    <!-- 安装脚本对话框 -->
    <el-dialog v-model="installDialogVisible" :title="'安装节点: ' + currentInstallNode?.name" width="650px" :close-on-click-modal="false">
      <el-alert type="info" :closable="false" style="margin-bottom: 20px;">
        <template #title>
          在目标服务器上执行以下命令，将自动安装 Gost 并配置为当前节点设置的参数。
        </template>
      </el-alert>

      <!-- 节点配置信息 -->
      <div class="node-info-section" v-if="currentInstallNode">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="API 端口">{{ extractPort(currentInstallNode.api_url) }}</el-descriptions-item>
          <el-descriptions-item label="用户名">{{ currentInstallNode.username || 'admin' }}</el-descriptions-item>
          <el-descriptions-item label="密码">{{ currentInstallNode.password || '(自动生成)' }}</el-descriptions-item>
        </el-descriptions>
      </div>

      <div class="install-command-section">
        <div class="command-header">
          <span class="command-title">一键安装命令</span>
          <el-button type="primary" size="small" :icon="CopyDocument" @click="copyInstallCommand">复制命令</el-button>
        </div>
        <div class="command-box">
          <code>{{ installCommand }}</code>
        </div>
      </div>

      <div class="install-tips">
        <p><strong>说明：</strong></p>
        <ol>
          <li>复制上方命令，在目标服务器上以 <strong>root</strong> 用户执行</li>
          <li>脚本将自动下载 Gost、生成配置文件并启动服务</li>
          <li>安装完成后节点将自动上线，无需手动操作</li>
        </ol>
      </div>

      <template #footer>
        <el-button @click="installDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh, CopyDocument, Management, Link, User, Lock } from '@element-plus/icons-vue'
import { getNodeList, createNode, updateNode, deleteNode, getNodeConfig } from '@/api/node'

// 安装脚本 URL（GitHub Raw）
const INSTALL_SCRIPT_URL = 'https://raw.githubusercontent.com/qiuapeng921/gostPanel/master/scripts/install_node.sh'

// 列表数据
const nodeList = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 搜索
const searchKeyword = ref('')
const searchStatus = ref('')

// 对话框
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const submitLoading = ref(false)
const formRef = ref(null)

// 配置对话框
const configDialogVisible = ref(false)
const configActiveTab = ref('services')
const nodeConfig = ref(null)

// 安装脚本对话框
const installDialogVisible = ref(false)
const currentInstallNode = ref(null)

// 格式化流量
const formatBytes = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 从 API URL 中提取端口号
const extractPort = (apiUrl) => {
  if (!apiUrl) return '39000'
  try {
    const url = new URL(apiUrl)
    return url.port || '39000'
  } catch {
    // 尝试用正则匹配端口
    const match = apiUrl.match(/:(\d+)/)
    return match ? match[1] : '39000'
  }
}

// 生成安装命令（根据当前节点配置）
const installCommand = computed(() => {
  if (!currentInstallNode.value) return ''
  
  const node = currentInstallNode.value
  const port = extractPort(node.api_url)
  const username = node.username
  const password = node.password
  
  // 构建带参数的安装命令
  // 参数顺序: 端口 用户名 密码
  let cmd = `bash <(curl -sL ${INSTALL_SCRIPT_URL}) ${port} ${username} ${password}`
  return cmd
})

// 显示安装命令对话框
const showInstallCommand = (row) => {
  currentInstallNode.value = row
  installDialogVisible.value = true
}

// 复制安装命令
const copyInstallCommand = async () => {
  try {
    await navigator.clipboard.writeText(installCommand.value)
    ElMessage.success('安装命令已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败，请手动复制')
  }
}

const form = reactive({
  name: '',
  address: '',
  port: 39000,
  username: '',
  password: '',
  remark: ''
})

const rules = {
  name: [{ required: true, message: '请输入节点名称', trigger: 'blur' }],
  address: [{ required: true, message: '请输入 IP 或域名', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口', trigger: 'blur' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

// 获取数据
const fetchData = async (isSilent = false) => {
  if (!isSilent) loading.value = true
  try {
    const res = await getNodeList({
      page: page.value,
      pageSize: pageSize.value,
      keyword: searchKeyword.value,
      status: searchStatus.value
    })
    nodeList.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (error) {
    console.error('获取节点列表失败:', error)
  } finally {
    if (!isSilent) loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  page.value = 1
  fetchData()
}

// 打开对话框
const openDialog = (row = null) => {
  isEdit.value = !!row
  editId.value = row?.id || null
  
  if (row) {
    Object.assign(form, {
      name: row.name,
      address: row.address,
      port: row.port,
      username: row.username,
      password: row.password,
      remark: row.remark
    })
  } else {
    Object.assign(form, {
      name: '',
      address: '127.0.0.1',
      port: 39000,
      username: 'admin',
      password: '123456',
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
      if (isEdit.value) {
        await updateNode(editId.value, form)
        ElMessage.success('更新成功')
      } else {
        await createNode(form)
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

// 删除节点
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要删除节点 "${row.name}" 吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await deleteNode(row.id)
    ElMessage.success('删除成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
    }
  }
}

// 复制节点
const handleCopy = (row) => {
  isEdit.value = false
  editId.value = null
  
  Object.assign(form, {
    name: row.name,
    api_url: row.api_url || '',
    username: row.username || '',
    password: row.password || '',
    remark: row.remark || ''
  })
  
  dialogVisible.value = true
}

// 查看配置
const handleViewConfig = async (row) => {
  try {
    const res = await getNodeConfig(row.id)
    nodeConfig.value = res.data
    configActiveTab.value = 'services'
    configDialogVisible.value = true
  } catch (error) {
    console.error('获取配置失败:', error)
    ElMessage.error('获取节点配置失败')
  }
}




// 定时刷新
let refreshTimer = null

onMounted(() => {
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

/* 安装命令相关样式 */
.node-info-section {
  margin-bottom: 20px;
}

.install-command-section {
  margin-top: 24px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.command-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.command-title {
  font-weight: 600;
  color: #303133;
}

.command-box {
  background: #1e1e1e;
  border-radius: 6px;
  padding: 16px;
  overflow-x: auto;
}

.command-box code {
  color: #4fc08d;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  word-break: break-all;
  white-space: pre-wrap;
}

.install-tips {
  margin-top: 20px;
  padding: 12px;
  background: #fdf6ec;
  border-radius: 6px;
  border-left: 4px solid #e6a23c;
}

.install-tips p {
  margin: 0 0 8px 0;
  color: #606266;
}

.install-tips ol {
  margin: 0;
  padding-left: 20px;
  color: #606266;
}

.install-tips li {
  margin-bottom: 4px;
}
</style>
