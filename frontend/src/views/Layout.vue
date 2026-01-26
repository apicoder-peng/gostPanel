<template>
  <div class="layout-container">
    <!-- 头部 (Top Header) -->
    <el-header class="header">
      <div class="header-left">
        <div class="logo">
          <img v-if="logoUrl" :src="logoUrl" alt="logo" class="logo-img" />
          <span>{{ siteTitle }}</span>
        </div>
      </div>
      <div class="header-right">
        <div class="user-group">
          <div class="username-box">
            {{ authStore.username || 'admin' }}
          </div>
          <el-dropdown trigger="click">
            <div class="user-icon-box">
              <el-icon><User /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="showPasswordDialog = true">
                  <el-icon><Key /></el-icon>改密
                </el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">
                  <el-icon><SwitchButton /></el-icon>登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </el-header>

    <!-- 下方主体 (Body: Sidebar + Content) -->
    <el-container class="body-container">
      <!-- 侧边栏 (Floating Sidebar) -->
      <el-aside :width="isCollapse ? '64px' : '220px'" class="aside">
        <el-menu
          :default-active="currentRoute"
          :collapse="isCollapse"
          :collapse-transition="false"
          router
          class="sidebar-menu"
          text-color="#606266"
          active-text-color="#409eff"
        >
          <el-menu-item 
            v-for="item in menuItems" 
            :key="item.path" 
            :index="item.path"
          >
            <el-icon><component :is="item.icon" /></el-icon>
            <template #title>{{ item.title }}</template>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 主内容 (Content) -->
      <el-main class="main">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>

    <!-- 修改密码对话框 -->
    <el-dialog
      v-model="showPasswordDialog"
      title="修改密码"
      width="400px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        label-width="80px"
      >
        <el-form-item label="原密码" prop="old_password">
          <el-input
            v-model="passwordForm.old_password"
            type="password"
            placeholder="请输入原密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="新密码" prop="new_password">
          <el-input
            v-model="passwordForm.new_password"
            type="password"
            placeholder="请输入新密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirm_password">
          <el-input
            v-model="passwordForm.confirm_password"
            type="password"
            placeholder="请确认新密码"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPasswordDialog = false">取消</el-button>
        <el-button type="primary" :loading="passwordLoading" @click="handleChangePassword">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Key, SwitchButton,
  Odometer, Monitor, Switch, Connection, Document, User, Setting, InfoFilled
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/store/auth'
import { useSystemStore } from '@/store/system'
import { changePassword } from '@/api/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const systemStore = useSystemStore()

const siteTitle = computed(() => systemStore.siteTitle)
const logoUrl = computed(() => systemStore.logoUrl)

onMounted(() => {
  systemStore.fetchSystemConfig()
})

// 菜单配置
const menuItems = [
  { path: '/dashboard', title: '仪表盘', icon: Odometer },
  { path: '/nodes', title: '节点管理', icon: Monitor },
  { path: '/forwards', title: '端口转发', icon: Switch },
  { path: '/tunnels', title: '隧道转发', icon: Connection },
  { path: '/logs', title: '操作日志', icon: Document },
  { path: '/system', title: '系统设置', icon: Setting },
  { path: '/about', title: '关于系统', icon: InfoFilled }
]

// 侧边栏折叠
const isCollapse = ref(false)
const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

// 当前路由
const currentRoute = computed(() => route.path)
const currentTitle = computed(() => {
  const item = menuItems.find(m => m.path === route.path)
  return item?.title || ''
})

// 修改密码
const showPasswordDialog = ref(false)
const passwordLoading = ref(false)
const passwordFormRef = ref(null)
const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const validateConfirmPassword = (rule, value, callback) => {
  if (value !== passwordForm.new_password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const passwordRules = {
  old_password: [
    { required: true, message: '请输入原密码', trigger: 'blur' }
  ],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 50, message: '密码长度为 6-50 个字符', trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const handleChangePassword = async () => {
  if (!passwordFormRef.value) return
  
  await passwordFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    passwordLoading.value = true
    try {
      await changePassword({
        old_password: passwordForm.old_password,
        new_password: passwordForm.new_password
      })
      ElMessage.success('密码修改成功，请重新登录')
      showPasswordDialog.value = false
      authStore.logout()
      router.push('/login')
    } catch (error) {
      console.error('修改密码失败:', error)
    } finally {
      passwordLoading.value = false
    }
  })
}

// 退出登录
const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    authStore.logout()
    router.push('/login')
    ElMessage.success('已退出登录')
  } catch {
    // 取消操作
  }
}
</script>

<style scoped>
.layout-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  height: 64px;
  position: relative;
  z-index: 10;
}

.body-container {
  flex: 1;
  display: flex;
  background-color: #f0f2f5;
  overflow: hidden;
}

.aside {
  background-color: #fff !important;
  margin: 16px 0 16px 16px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  overflow: hidden;
  border-right: none;
  transition: width 0.3s;
}

.sidebar-menu {
  border-right: none;
  height: 100%;
}

.main {
  flex: 1;
  background: transparent;
  padding: 0;
  margin: 16px;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.logo {
  font-size: 20px;
  font-weight: 700;
  color: #333;
  margin-right: 32px;
  display: flex;
  align-items: center;
}

.logo-img {
  width: 32px;
  height: 32px;
  margin-right: 10px;
  border-radius: 4px;
}

.header-left {
  display: flex;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-group {
  display: flex;
  align-items: center;
}

.username-box {
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 16px;
  min-width: 100px;
  border: 1px solid #d9d9d9;
  border-right: none;
  border-top-left-radius: 4px;
  border-bottom-left-radius: 4px;
  background: #fff;
  font-size: 13px;
  font-weight: 500;
  cursor: default;
  color: #606266;
}

.user-icon-box {
  height: 30px;
  width: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #d9d9d9;
  border-top-right-radius: 4px;
  border-bottom-right-radius: 4px;
  background: #fff;
  color: #606266;
  cursor: pointer;
  transition: all 0.3s;
  font-size: 14px;
  margin-left: -1px;
}

.user-icon-box:hover {
  border-color: #409eff;
  color: #409eff;
  position: relative;
  z-index: 1;
}

/* Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
