// Package dto 定义数据传输对象
package dto

// ==================== 转发相关 ====================

// CreateForwardReq 创建转发请求
type CreateForwardReq struct {
	NodeID     uint   `json:"node_id" binding:"required"`                     // 节点 ID
	Name       string `json:"name" binding:"required,min=1,max=100"`          // 规则名称
	Protocol   string `json:"protocol" binding:"required,oneof=tcp udp"`      // 协议类型
	ListenPort int    `json:"listen_port" binding:"required,min=1,max=65535"` // 监听端口

	Targets   []string `json:"targets"`                                                 // 多目标列表
	Strategy  string   `json:"strategy" binding:"omitempty,oneof=round rand fifo hash"` // 负载均衡策略
	EnableTLS bool     `json:"enable_tls"`                                              // 是否启用 TLS

	Remark string `json:"remark"` // 备注
}

// UpdateForwardReq 更新转发请求
type UpdateForwardReq struct {
	Name       string `json:"name" binding:"required,min=1,max=100"`          // 规则名称
	Protocol   string `json:"protocol" binding:"required,oneof=tcp udp"`      // 协议类型
	ListenPort int    `json:"listen_port" binding:"required,min=1,max=65535"` // 监听端口

	Targets   []string `json:"targets"`                                                 // 多目标列表
	Strategy  string   `json:"strategy" binding:"omitempty,oneof=round rand fifo hash"` // 负载均衡策略
	EnableTLS bool     `json:"enable_tls"`                                              // 是否启用 TLS

	Remark string `json:"remark"` // 备注
}

// ForwardListReq 转发列表请求
type ForwardListReq struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`             // 页码
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"` // 每页数量
	NodeID   uint   `form:"node_id"`                                    // 节点 ID 筛选
	Status   string `form:"status"`                                     // 状态筛选
	Keyword  string `form:"keyword"`                                    // 关键词搜索
}

// SetDefaults 设置默认值
func (r *ForwardListReq) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.PageSize == 0 {
		r.PageSize = 10
	}
}
