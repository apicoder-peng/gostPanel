import request from '@/utils/request'

/**
 * 获取隧道列表
 */
export function getTunnelList(params) {
    return request({
        url: '/tunnels',
        method: 'get',
        params
    })
}

/**
 * 获取隧道详情
 */
export function getTunnel(id) {
    return request({
        url: `/tunnels/${id}`,
        method: 'get'
    })
}

/**
 * 创建隧道
 */
export function createTunnel(data) {
    return request({
        url: '/tunnels',
        method: 'post',
        data
    })
}

/**
 * 更新隧道
 */
export function updateTunnel(id, data) {
    return request({
        url: `/tunnels/${id}`,
        method: 'put',
        data
    })
}

/**
 * 删除隧道
 */
export function deleteTunnel(id) {
    return request({
        url: `/tunnels/${id}`,
        method: 'delete'
    })
}

/**
 * 启动隧道
 */
export function startTunnel(id) {
    return request({
        url: `/tunnels/${id}/start`,
        method: 'post'
    })
}

/**
 * 停止隧道
 */
export function stopTunnel(id) {
    return request({
        url: `/tunnels/${id}/stop`,
        method: 'post'
    })
}
