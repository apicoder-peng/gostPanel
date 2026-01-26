import request from '@/utils/request'

/**
 * 获取转发规则列表
 */
export function getForwardList(params) {
    return request({
        url: '/forwards',
        method: 'get',
        params
    })
}

/**
 * 获取转发规则详情
 */
export function getForward(id) {
    return request({
        url: `/forwards/${id}`,
        method: 'get'
    })
}

/**
 * 创建转发规则
 */
export function createForward(data) {
    return request({
        url: '/forwards',
        method: 'post',
        data
    })
}

/**
 * 更新转发规则
 */
export function updateForward(id, data) {
    return request({
        url: `/forwards/${id}`,
        method: 'put',
        data
    })
}

/**
 * 删除转发规则
 */
export function deleteForward(id) {
    return request({
        url: `/forwards/${id}`,
        method: 'delete'
    })
}

/**
 * 启动转发
 */
export function startForward(id) {
    return request({
        url: `/forwards/${id}/start`,
        method: 'post'
    })
}

/**
 * 停止转发
 */
export function stopForward(id) {
    return request({
        url: `/forwards/${id}/stop`,
        method: 'post'
    })
}

/**
 * 获取转发统计
 */
export function getForwardStats() {
    return request({
        url: '/forwards/stats',
        method: 'get'
    })
}
