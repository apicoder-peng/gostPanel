<template>
  <div class="system-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>系统设置</span>
        </div>
      </template>

      <el-tabs v-model="activeTab" class="demo-tabs">
        <!-- 面板配置 -->
        <el-tab-pane label="面板配置" name="config">
          <el-form ref="configFormRef" :model="configForm" label-width="120px" class="setting-form">
            <el-form-item label="面板地址" prop="panelUrl">
               <el-input v-model="configForm.panelUrl" placeholder="你当前的面板地址" />
            </el-form-item>
             <el-form-item label="站点标题" prop="siteTitle">
              <el-input v-model="configForm.siteTitle" placeholder="Gost Panel" />
            </el-form-item>

            <el-form-item label="Logo地址" prop="logoUrl">
              <el-input v-model="configForm.logoUrl" placeholder="请输入Logo URL" />
            </el-form-item>
            <el-form-item label="版权信息" prop="copyright">
              <el-input v-model="configForm.copyright" placeholder="请输入版权信息" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="loading" @click="handleSave('config')">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 邮箱配置 -->
        <el-tab-pane label="邮箱配置" name="email">
          <el-form ref="emailFormRef" :model="emailForm" label-width="120px" class="setting-form">
            <el-form-item label="SMTP服务器" prop="host">
              <el-input v-model="emailForm.host" placeholder="例如: smtp.gmail.com" />
            </el-form-item>
            <el-form-item label="SMTP端口" prop="port">
              <el-input v-model.number="emailForm.port" placeholder="例如: 465 或 587" />
            </el-form-item>
            <el-form-item label="用户名" prop="username">
              <el-input v-model="emailForm.username" placeholder="请输入SMTP用户名" />
            </el-form-item>
            <el-form-item label="密码" prop="password">
              <el-input v-model="emailForm.password" type="password" show-password placeholder="请输入SMTP密码" />
            </el-form-item>
            <el-form-item label="发件人邮箱" prop="fromEmail">
              <el-input v-model="emailForm.fromEmail" placeholder="例如: noreply@example.com" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="loading" @click="handleSave('email')">保存设置</el-button>
              <el-button @click="handleTestEmail">测试邮件</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>



        <!-- 备份 -->
        <el-tab-pane label="备份" name="backup">
          <el-form ref="backupFormRef" :model="backupForm" label-width="120px" class="setting-form">
            <el-form-item label="自动备份" prop="autoBackup">
              <el-switch v-model="backupForm.autoBackup" />
            </el-form-item>
            <el-form-item label="保留份数" prop="retentionCount">
              <el-input-number v-model="backupForm.retentionCount" :min="1" :max="100" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="loading" @click="handleSave('backup')">保存设置</el-button>
              <el-divider direction="vertical" />
              <el-button type="success" @click="handleBackupNow">立即备份</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 测试邮件弹窗 -->
    <el-dialog v-model="testEmailVisible" title="发送测试邮件" width="400px">
      <el-form :model="testEmailForm" label-width="80px">
        <el-form-item label="收件人">
          <el-input v-model="testEmailForm.toEmail" placeholder="请输入接收测试邮件的邮箱" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="testEmailVisible = false">取消</el-button>
          <el-button type="primary" :loading="testEmailLoading" @click="confirmTestEmail">
            发送
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getSystemConfig, updateSystemConfig, sendTestEmail, backupSystem } from '@/api/system'

const activeTab = ref('config')
const loading = ref(false)
const testEmailVisible = ref(false)
const testEmailLoading = ref(false)
const testEmailForm = reactive({
    toEmail: ''
})


const emailForm = reactive({
  host: '',
  port: 465,
  username: '',
  password: '',
  fromEmail: ''
})

const configForm = reactive({
  panelUrl: '',
  siteTitle: 'Gost Panel',
  logoUrl: '',
  copyright: ''
})



const backupForm = reactive({
  autoBackup: false,
  retentionCount: 7
})

// 获取配置
const fetchConfig = async () => {
    loading.value = true
    try {
        const res = await getSystemConfig()
        if (res.data) {
            // 根据后端返回的数据结构填充表单
            const { panel, email, config, backup } = res.data
            if (panel) {
                configForm.panelUrl = panel.panelUrl
            }
            if (email) Object.assign(emailForm, email)
            if (config) Object.assign(configForm, config)
            if (backup) Object.assign(backupForm, backup)
        }
    } catch (error) {
        console.error('获取系统配置失败:', error)
    } finally {
        loading.value = false
    }
}

// 保存配置
const handleSave = async (type) => {
    loading.value = true
    try {
        let data = {}
        // 根据当前tab类型构建提交数据
        const payload = {
            panel: { panelUrl: configForm.panelUrl },
            email: emailForm,
            config: {
                siteTitle: configForm.siteTitle,
                logoUrl: configForm.logoUrl,
                copyright: configForm.copyright
            },
            backup: backupForm
        }
        
        await updateSystemConfig(payload)
        ElMessage.success('保存成功')
    } catch (error) {
        console.error('保存失败:', error)
    } finally {
        loading.value = false
    }
}

const handleTestEmail = () => {
    // 简单校验
    if (!emailForm.host || !emailForm.port || !emailForm.fromEmail) {
        ElMessage.warning('请先完善邮箱配置信息')
        return
    }
    testEmailForm.toEmail = emailForm.fromEmail
    testEmailVisible.value = true
}

const confirmTestEmail = async () => {
    if (!testEmailForm.toEmail) {
        ElMessage.warning('请输入收件人邮箱')
        return
    }

    testEmailLoading.value = true
    try {
        const payload = {
            ...emailForm,
            toEmail: testEmailForm.toEmail
        }
        await sendTestEmail(payload)
        ElMessage.success('测试邮件发送成功')
        testEmailVisible.value = false
    } catch (error) {
        console.error('测试邮件发送失败:', error)
    } finally {
        testEmailLoading.value = false
    }
}

const handleBackupNow = async () => {
    try {
        await ElMessageBox.confirm('确定要立即执行数据库备份吗？', '提示', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
        })
        
        loading.value = true
        await backupSystem()
        ElMessage.success('备份成功')
    } catch (error) {
        if (error !== 'cancel') {
            console.error('备份失败:', error)
        }
    } finally {
        loading.value = false
    }
}

onMounted(() => {
    fetchConfig()
})
</script>

<style scoped>
.system-container {
  padding: 20px;
}
.setting-form {
    max-width: 600px;
    margin-top: 20px;
}
.ml-2 {
  margin-left: 8px;
}
</style>
