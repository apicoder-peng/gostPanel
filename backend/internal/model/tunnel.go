package model

import (
	"time"

	"gorm.io/gorm"
)

// TunnelStatus 隧道状态
type TunnelStatus string

const (
	TunnelStatusStopped TunnelStatus = "stopped" // 已停止
	TunnelStatusRunning TunnelStatus = "running" // 运行中
	TunnelStatusError   TunnelStatus = "error"   // 错误
)

// GostTunnel 隧道模型 - 多跳转发
// 流量路径: 客户端 -> 入口节点(EntryNode) -> 出口节点(ExitNode) -> 目标地址(TargetHost:TargetPort)
type GostTunnel struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:100;not null" json:"name"`       // 隧道名称
	EntryNodeID uint   `gorm:"not null" json:"entry_node_id"`       // 入口节点 ID
	ExitNodeID  uint   `gorm:"not null" json:"exit_node_id"`        // 出口节点 ID
	Protocol    string `gorm:"size:20;default:tcp" json:"protocol"` // 协议 tcp/udp
	ListenPort  int    `gorm:"not null" json:"listen_port"`         // 入口节点监听端口

	Targets   []string     `gorm:"type:json;serializer:json" json:"targets"` // 多目标列表
	Strategy  string       `gorm:"size:20;default:round" json:"strategy"`    // 负载均衡策略
	RelayPort int          `gorm:"default:8443" json:"relay_port"`           // 出口节点 relay 服务端口
	Status    TunnelStatus `gorm:"size:20;default:stopped" json:"status"`
	ServiceID string       `gorm:"size:100" json:"service_id"` // Gost 服务 ID
	ChainID   string       `gorm:"size:100" json:"chain_id"`   // Gost 链 ID

	// 流量监控配置
	ObserverID string `gorm:"size:100" json:"observer_id"` // 观察器 ID

	// 流量统计 (由观察器更新)
	InputBytes    int64 `gorm:"default:0" json:"input_bytes"`    // 入站总流量 (bytes)
	OutputBytes   int64 `gorm:"default:0" json:"output_bytes"`   // 出站总流量 (bytes)
	TotalBytes    int64 `gorm:"default:0" json:"total_bytes"`    // 总流量 (Input + Output)
	TotalRequests int64 `gorm:"default:0" json:"total_requests"` // 总请求数

	Remark    string         `gorm:"type:text" json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	EntryNode *GostNode `gorm:"foreignKey:EntryNodeID" json:"entry_node,omitempty"`
	ExitNode  *GostNode `gorm:"foreignKey:ExitNodeID" json:"exit_node,omitempty"`
}

// TableName 指定表名
func (GostTunnel) TableName() string {
	return "gost_tunnels"
}
