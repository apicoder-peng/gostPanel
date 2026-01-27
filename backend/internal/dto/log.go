package dto

// ==================== 操作日志相关 ====================

// LogListReq 日志列表请求
type LogListReq struct {
	Page         int    `form:"page" binding:"omitempty,min=1"`             // 页码
	PageSize     int    `form:"pageSize" binding:"omitempty,min=1,max=100"` // 每页数量
	Username     string `form:"username"`                                   // 用户名筛选
	Action       string `form:"action"`                                     // 操作类型筛选
	ResourceType string `form:"resource_type"`                              // 资源类型筛选
}

// SetDefaults 设置默认值
func (r *LogListReq) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.PageSize == 0 {
		r.PageSize = 20
	}
}
