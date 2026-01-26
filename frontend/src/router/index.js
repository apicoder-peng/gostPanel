import { createRouter, createWebHashHistory } from 'vue-router'
import { useAuthStore } from '@/store/auth'

// 路由配置
const routes = [
    {
        path: '/login',
        name: 'Login',
        component: () => import('@/views/Login.vue'),
        meta: { requiresAuth: false, title: '登录' }
    },
    {
        path: '/',
        name: 'Layout',
        component: () => import('@/views/Layout.vue'),
        redirect: '/dashboard',
        meta: { requiresAuth: true },
        children: [
            {
                path: 'dashboard',
                name: 'Dashboard',
                component: () => import('@/views/Dashboard.vue'),
                meta: { title: '仪表盘', icon: 'Odometer' }
            },
            {
                path: 'nodes',
                name: 'Nodes',
                component: () => import('@/views/Nodes.vue'),
                meta: { title: '节点管理', icon: 'Monitor' }
            },
            {
                path: 'forwards',
                name: 'Forwards',
                component: () => import('@/views/Forwards.vue'),
                meta: { title: '端口转发', icon: 'Switch' }
            },
            {
                path: 'tunnels',
                name: 'Tunnels',
                component: () => import('@/views/Tunnels.vue'),
                meta: { title: '隧道转发', icon: 'Connection' }
            },
            {
                path: 'logs',
                name: 'Logs',
                component: () => import('@/views/Logs.vue'),
                meta: { title: '操作日志', icon: 'Document' }
            },
            {
                path: 'system',
                name: 'System',
                component: () => import('@/views/System.vue'),
                meta: { title: '系统设置', icon: 'Setting' }
            },
            {
                path: 'about',
                name: 'About',
                component: () => import('@/views/About.vue'),
                meta: { title: '关于系统', icon: 'InfoFilled' }
            }
        ]
    },
    {
        path: '/:pathMatch(.*)*',
        redirect: '/'
    }
]

const router = createRouter({
    history: createWebHashHistory(),
    routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
    const authStore = useAuthStore()
    const requiresAuth = to.matched.some(record => record.meta.requiresAuth !== false)

    if (requiresAuth && !authStore.isLoggedIn) {
        // 需要认证但未登录
        next({ name: 'Login', query: { redirect: to.fullPath } })
    } else if (to.name === 'Login' && authStore.isLoggedIn) {
        // 已登录但访问登录页
        next({ name: 'Dashboard' })
    } else {
        next()
    }
})

export default router
