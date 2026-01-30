import request from '@/utils/request'

/**
 * 获取规则列表
 */
export function getRuleList(params) {
    return request({
        url: '/rules',
        method: 'get',
        params
    })
}

/**
 * 获取规则详情
 */
export function getRule(id) {
    return request({
        url: `/rules/${id}`,
        method: 'get'
    })
}

/**
 * 创建规则
 */
export function createRule(data) {
    return request({
        url: '/rules',
        method: 'post',
        data
    })
}

/**
 * 更新规则
 */
export function updateRule(id, data) {
    return request({
        url: `/rules/${id}`,
        method: 'put',
        data
    })
}

/**
 * 删除规则
 */
export function deleteRule(id) {
    return request({
        url: `/rules/${id}`,
        method: 'delete'
    })
}

/**
 * 启动规则
 */
export function startRule(id) {
    return request({
        url: `/rules/${id}/start`,
        method: 'post'
    })
}

/**
 * 停止规则
 */
export function stopRule(id) {
    return request({
        url: `/rules/${id}/stop`,
        method: 'post'
    })
}