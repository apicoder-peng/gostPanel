package dto

// ==================== 隧道管理相关 ====================

// CreateTunnelReq 创建隧道请求
type CreateTunnelReq struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`         // 隧道名称
	EntryNodeID uint   `json:"entry_node_id" binding:"required"`              // 入口节点 ID
	ExitNodeID  uint   `json:"exit_node_id" binding:"required"`               // 出口节点 ID
	Protocol    string `json:"protocol" binding:"required,oneof=tcp udp"`     // 协议类型
	RelayPort   int    `json:"relay_port" binding:"required,min=1,max=65535"` // 出口节点 Relay 端口
	Remark      string `json:"remark"`                                        // 备注
}

// UpdateTunnelReq 更新隧道请求
type UpdateTunnelReq struct {
	Name      string `json:"name" binding:"required,min=1,max=100"`         // 隧道名称
	Protocol  string `json:"protocol" binding:"required,oneof=tcp udp"`     // 协议类型
	RelayPort int    `json:"relay_port" binding:"required,min=1,max=65535"` // 出口节点 Relay 端口
	Remark    string `json:"remark"`                                        // 备注
}

// TunnelListReq 隧道列表请求
type TunnelListReq struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`             // 页码
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"` // 每页数量
	NodeID   uint   `form:"node_id"`                                    // 节点 ID 筛选
	Status   string `form:"status"`                                     // 状态筛选
	Keyword  string `form:"keyword"`                                    // 关键词搜索
}

// SetDefaults 设置默认值
func (r *TunnelListReq) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.PageSize == 0 {
		r.PageSize = 10
	}
}
