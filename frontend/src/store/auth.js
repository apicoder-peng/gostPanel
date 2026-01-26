import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi, getUserInfo as getUserInfoApi } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
    // 状态
    const token = ref(localStorage.getItem('token') || '')
    const userInfo = ref(JSON.parse(localStorage.getItem('userInfo') || 'null'))

    // 计算属性
    const isLoggedIn = computed(() => !!token.value)
    const username = computed(() => userInfo.value?.username || '')

    // 登录
    async function login(loginForm) {
        try {
            const res = await loginApi(loginForm)
            token.value = res.data.token
            userInfo.value = res.data.user
            // 持久化存储
            localStorage.setItem('token', res.data.token)
            localStorage.setItem('userInfo', JSON.stringify(res.data.user))
            return res
        } catch (error) {
            throw error
        }
    }

    // 获取用户信息
    async function fetchUserInfo() {
        try {
            const res = await getUserInfoApi()
            userInfo.value = res.data
            localStorage.setItem('userInfo', JSON.stringify(res.data))
            return res
        } catch (error) {
            throw error
        }
    }

    // 登出
    function logout() {
        token.value = ''
        userInfo.value = null
        localStorage.removeItem('token')
        localStorage.removeItem('userInfo')
    }

    return {
        token,
        userInfo,
        isLoggedIn,
        username,
        login,
        fetchUserInfo,
        logout
    }
})
