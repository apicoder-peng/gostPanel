import request from '@/utils/request'

/**
 * 获取节点列表
 */
export function getNodeList(params) {
    return request({
        url: '/nodes',
        method: 'get',
        params
    })
}

/**
 * 获取节点详情
 */
export function getNode(id) {
    return request({
        url: `/nodes/${id}`,
        method: 'get'
    })
}

/**
 * 创建节点
 */
export function createNode(data) {
    return request({
        url: '/nodes',
        method: 'post',
        data
    })
}

/**
 * 更新节点
 */
export function updateNode(id, data) {
    return request({
        url: `/nodes/${id}`,
        method: 'put',
        data
    })
}

/**
 * 删除节点
 */
export function deleteNode(id) {
    return request({
        url: `/nodes/${id}`,
        method: 'delete'
    })
}

/**
 * 获取节点统计
 */
export function getNodeStats() {
    return request({
        url: '/nodes/stats',
        method: 'get'
    })
}

/**
 * 获取节点配置
 */
export function getNodeConfig(id) {
    return request({
        url: `/nodes/${id}/config`,
        method: 'get'
    })
}
