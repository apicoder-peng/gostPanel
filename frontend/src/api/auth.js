import request from '@/utils/request'

/**
 * 用户登录
 * @param {Object} data - 登录信息 { username, password }
 */
export function login(data) {
    return request({
        url: '/auth/login',
        method: 'post',
        data
    })
}

/**
 * 获取当前用户信息
 */
export function getUserInfo() {
    return request({
        url: '/auth/info',
        method: 'get'
    })
}

/**
 * 修改密码
 * @param {Object} data - 密码信息 { old_password, new_password }
 */
export function changePassword(data) {
    return request({
        url: '/auth/password',
        method: 'put',
        data
    })
}

/**
 * 刷新 Token
 */
export function refreshToken() {
    return request({
        url: '/auth/refresh',
        method: 'post'
    })
}
