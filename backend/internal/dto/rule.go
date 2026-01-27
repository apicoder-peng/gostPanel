package dto

// ==================== 规则管理相关 ====================

// CreateRuleReq 创建规则请求
// 入口选择：NodeID 或 TunnelID 二选一
// - 端口转发 (forward)：NodeID 必填，直接在该节点上创建转发服务
// - 隧道转发 (tunnel)：TunnelID 必填，在隧道的入口节点上创建转发服务
type CreateRuleReq struct {
	NodeID     *uint  `json:"node_id"`                                        // 入口节点 ID（端口转发时必填）
	TunnelID   *uint  `json:"tunnel_id"`                                      // 隧道 ID（隧道转发时必填）
	Name       string `json:"name" binding:"required,min=1,max=100"`          // 规则名称
	Type       string `json:"type" binding:"required,oneof=forward tunnel"`   // 规则类型
	Protocol   string `json:"protocol" binding:"required,oneof=tcp udp"`      // 协议类型
	ListenPort int    `json:"listen_port" binding:"required,min=1,max=65535"` // 监听端口

	Targets   []string `json:"targets"`                                                 // 多目标列表
	Strategy  string   `json:"strategy" binding:"omitempty,oneof=round rand fifo hash"` // 负载均衡策略
	EnableTLS bool     `json:"enable_tls"`                                              // 是否启用 TLS

	Remark string `json:"remark"` // 备注
}

// UpdateRuleReq 更新规则请求
type UpdateRuleReq struct {
	Name       string `json:"name" binding:"required,min=1,max=100"`          // 规则名称
	Protocol   string `json:"protocol" binding:"required,oneof=tcp udp"`      // 协议类型
	ListenPort int    `json:"listen_port" binding:"required,min=1,max=65535"` // 监听端口

	Targets   []string `json:"targets"`                                                 // 多目标列表
	Strategy  string   `json:"strategy" binding:"omitempty,oneof=round rand fifo hash"` // 负载均衡策略
	EnableTLS bool     `json:"enable_tls"`                                              // 是否启用 TLS

	Remark string `json:"remark"` // 备注
}

// RuleListReq 规则列表请求
type RuleListReq struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`             // 页码
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"` // 每页数量
	NodeID   uint   `form:"node_id"`                                    // 节点 ID 筛选
	TunnelID uint   `form:"tunnel_id"`                                  // 隧道 ID 筛选
	Type     string `form:"type"`                                       // 规则类型筛选
	Status   string `form:"status"`                                     // 状态筛选
	Keyword  string `form:"keyword"`                                    // 关键词搜索
}

// SetDefaults 设置默认值
func (r *RuleListReq) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.PageSize == 0 {
		r.PageSize = 10
	}
}
