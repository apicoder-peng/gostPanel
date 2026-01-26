import request from '@/utils/request'

/**
 * 获取系统配置
 */
export function getSystemConfig() {
    return request({
        url: '/system/config',
        method: 'get'
    })
}

/**
 * 获取公开系统配置
 */
export function getPublicSystemConfig() {
    return request({
        url: '/system/public-config',
        method: 'get'
    })
}

/**
 * 更新系统配置
 */
export function updateSystemConfig(data) {
    return request({
        url: '/system/config',
        method: 'put',
        data
    })
}

export function sendTestEmail(data) {
    return request({
        url: '/system/email/test',
        method: 'post',
        data
    })
}

export function backupSystem() {
    return request({
        url: '/system/backup',
        method: 'post'
    })
}
