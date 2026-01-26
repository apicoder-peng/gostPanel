import request from '@/utils/request'

/**
 * 获取仪表盘统计
 */
export function getDashboardStats() {
    return request({
        url: '/dashboard/stats',
        method: 'get'
    })
}
