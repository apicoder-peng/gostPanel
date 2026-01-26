import request from '@/utils/request'

/**
 * 获取操作日志列表
 */
export function getLogList(params) {
    return request({
        url: '/logs',
        method: 'get',
        params
    })
}
